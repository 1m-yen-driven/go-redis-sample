package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/shamaton/msgpack"
	"strings"
)

// Encode / Decode
func EncodePtr(valuePtr interface{}) string {
	d, _ := msgpack.Encode(valuePtr)
	return string(d)
}
func DecodePtrStringCmd(input *redis.StringCmd, valuePtr interface{}) {
	msgpack.Decode([]byte(input.Val()), valuePtr)
}
func DecodePtrSliceCmdElem(partsOfSliceCmd interface{}, valuePtr interface{}) {
	msgpack.Decode([]byte(partsOfSliceCmd.(string)), valuePtr)
}

// MultiStatement の速度比較
var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
	DB:   0, // 0 - 15
})
var mdb = OpenDB(true)

func main() {
	ctx := context.Background()
	keys, users := initialize()
	Measure("Redis: Pipeline Set", func() {
		pipe := rdb.Pipeline()
		for i, _ := range keys {
			pipe.Set(ctx, keys[i], EncodePtr(&users[i]), 0)
		}
		pipe.Exec(ctx)
	})
	Measure("MySQL: Bulk Insert 1", func() {
		if _, err := mdb.NamedExec(
			"INSERT INTO user(id, name, count, created_at)"+
				"VALUES(:id, :name, :count, :created_at)", users); err != nil {
			panic(err)
		}
	})
	mdb.Exec("TRUNCATE TABLE user")
	Measure("MySQL: Bulk Insert 2", func() {
		args := []interface{}{}
		for _, u := range users {
			args = append(args, u.ID, u.Name, u.Count, u.CreatedAt)
		}
		query := "INSERT INTO user(id, name, count, created_at) VALUES "
		query += strings.Repeat("(?,?,?,?),", len(users))
		query = query[:len(query)-1]
		if _, err := mdb.Exec(query, args...); err != nil {
			panic(err)
		}
	})
	// MySQL multi statement is too slow...
	// Measure("MySQL: Bulk Insert 3", func() {
	// 	args := []interface{}{}
	// 	for _, u := range users {
	// 		args = append(args, u.ID, u.Name, u.Count, u.CreatedAt)
	// 	}
	// 	query := "INSERT INTO user(id, name, count, created_at) VALUES (?,?,?,?);"
	// 	query = strings.Repeat(query, len(users))
	// 	if _, err := mdb.Exec(query, args...); err != nil {
	// 		panic(err)
	// 	}
	// })

	// select time
	user := []User{}
	mdb.Select(&user, "SELECT * FROM user")
	fmt.Println(users[:10])
	count := 0
	mdb.Get(&count, "SELECT COUNT(*) FROM user")
	fmt.Println(count)
}

// -------------------- テスト用 ------------------------------
func initialize() (keys []string, users []User) {
	ctx := context.Background()
	keys = []string{}
	users = []User{}
	for i := 0; i < 20000; i++ {
		keys = append(keys, RandStr())
		users = append(users, RandUser())
	}
	// initialize Redis
	rdb.FlushDB(ctx)
	// initialize MySQL
	mdb.Exec("DROP TABLE IF EXISTS user")
	mdb.Exec("CREATE TABLE user(id int, name varchar(64), count int, created_at DATETIME)")
	return keys, users
}

// -------------------------------- テスト用 ---------------

func OpenDB(batch bool) *sqlx.DB {
	// before this, exec `CREATE DATABASE IF NOT EXISTS sandbox`
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = "127.0.0.1"
	mysqlConfig.User = "root"
	mysqlConfig.Passwd = ""
	mysqlConfig.DBName = "sandbox"
	mysqlConfig.Params = map[string]string{
		"time_zone":         "'+00:00'",
		"interpolateParams": "true",
	}
	mysqlConfig.ParseTime = true
	mysqlConfig.MultiStatements = batch
	result, _ := sqlx.Open("mysql", mysqlConfig.FormatDSN())
	result.SetMaxOpenConns(100)
	result.SetMaxIdleConns(100)
	return result
}
