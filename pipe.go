package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/shamaton/msgpack"
	// "strconv"
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

var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
	DB:   0, // 0 - 15
})

// 通常の Get / Set をパイプライン化して高速化されることを確認するサンプル
func main() {
	fmt.Println("RUN: PIPELINE SAMPLE")
	ctx := context.Background()

	// 通常: Get / Set
	keys, users := initialize(ctx)
	Measure("WITHOUT PIPELINE", func() {
		for i, _ := range keys {
			user := User{}
			DecodePtrStringCmd(rdb.Get(ctx, keys[i]), &user)
			mark(&user)
			rdb.Set(ctx, keys[i], EncodePtr(&user), 0)
		}
	})
	check(ctx, keys, users)

	// パイプライン: Get / Set
	keys, users = initialize(ctx)
	Measure("WITH PIPELINE", func() {
		// Read All -> Write All の順にするとトランザクションにそのまま乗せられる
		pipe := rdb.Pipeline()
		defer pipe.Close()
		// Read All
		gots := []*redis.StringCmd{}
		for i, _ := range keys {
			gots = append(gots, pipe.Get(ctx, keys[i]))
		}
		pipe.Exec(ctx) // gots の値がここで確定する
		// Write All
		for i, _ := range keys {
			user := User{}
			DecodePtrStringCmd(gots[i], &user)
			mark(&user)
			pipe.Set(ctx, keys[i], EncodePtr(&user), 0)
		}
		pipe.Exec(ctx)
	})
	check(ctx, keys, users)

	// 通常: MGet / MSet
	keys, users = initialize(ctx)
	Measure("MGet / MSet", func() {
		mset := map[string]interface{}{}
		for i, got := range rdb.MGet(ctx, keys...).Val() {
			user := User{}
			DecodePtrSliceCmdElem(got, &user)
			mark(&user)
			mset[keys[i]] = EncodePtr(&user)
		}
		rdb.MSet(ctx, mset)
	})
	check(ctx, keys, users)
}

// -------------------- テスト用 ------------------------------

func initialize(ctx context.Context) (keys []string, users []User) {
	keys = []string{}
	users = []User{}
	for i := 0; i < 10000; i++ {
		keys = append(keys, RandStr())
		users = append(users, RandUser())
	}
	// initialize
	rdb.FlushDB(ctx)
	for i, _ := range keys {
		rdb.Set(ctx, keys[i], EncodePtr(&users[i]), 0)
	}
	return keys, users
}
func mark(user *User) {
	user.Name += "(READ)"
}
func check(ctx context.Context, keys []string, users []User) {
	for i, _ := range keys {
		user := User{}
		DecodePtrStringCmd(rdb.Get(ctx, keys[i]), &user)
		user2 := users[i]
		mark(&user2)
		AssertEq(user, user2)
	}
}
