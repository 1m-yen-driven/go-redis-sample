package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/shamaton/msgpack"
	"strconv"
)

// Encode / Decode
func EncodePtr(valuePtr interface{}) string {
	d, _ := msgpack.Encode(valuePtr)
	return string(d)
}
func DecodePtrStringCmd(input *redis.StringCmd, valuePtr interface{}) {
	msgpack.Decode([]byte(input.Val()), valuePtr)
}
func DecodePtrSliceCmd(partsOfSliceCmd interface{}, valuePtr interface{}) {
	msgpack.Decode([]byte(partsOfSliceCmd.(string)), valuePtr)
}

// コマンドを送るまでは落ちないので、起動順序問題は通常大丈夫
var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
	DB:   0, // 0 - 15
})

func main() {
	fmt.Println("START")
	ctx := context.Background()
	// Set / Get / GetSet
	{
		user := RandUser()
		// ex1: no pipe, no transaction
		rdb.Set(ctx, "aaa", EncodePtr(&user), 0)
		user2 := User{}
		DecodePtrStringCmd(rdb.Get(ctx, "aaa"), &user2)
		fmt.Println(user, "\n", user2)
		Assert(user == user2)
		user3 := RandUser()
		DecodePtrStringCmd(rdb.GetSet(ctx, "aaa", EncodePtr(&user3)), &user3)
		Assert(user3 == user2)
		DecodePtrStringCmd(rdb.Get(ctx, "aaa"), &user3)
		Assert(user3 != user2)
	}

	// Exists / DBSize / Keys / Rename(NX) / Del
	{
		rdb.FlushDB(ctx)
		rdb.Set(ctx, "aaa", "v0", 0)
		ex := rdb.Exists(ctx, "aaa", "bbb").Val()
		Assert(ex == 1)
		cnt := rdb.DBSize(ctx).Val()
		Assert(cnt == 1)
		keys := rdb.Keys(ctx, "*").Val()
		Assert(len(keys) == 1 && keys[0] == "aaa")
		rdb.Rename(ctx, "aaa", "bbb")
		del := rdb.Del(ctx, "aaa", "bbb").Val()
		Assert(del == 1)
	}

	// MGet / MSet
	{
		rdb.FlushDB(ctx)
		mset := map[string]interface{}{}
		keys := []string{}
		users := []User{}
		for i := 0; i < 4; i++ {
			keys = append(keys, RandStr())
			users = append(users, RandUser())
			mset[keys[i]] = EncodePtr(&users[i])
		}
		rdb.MSet(ctx, mset)
		// parse in for loop
		for i, got := range rdb.MGet(ctx, keys...).Val() {
			user := User{}
			DecodePtrSliceCmd(got, &user)
			Assert(user == users[i])
		}
	}

	// IncrBy / Append
	{
		rdb.FlushDB(ctx)
		rdb.Set(ctx, "aaa", strconv.Itoa(100), 0)
		rdb.IncrBy(ctx, "aaa", 234)
		x, _ := strconv.Atoi(rdb.Get(ctx, "aaa").Val())
		Assert(x == 334)
		rdb.Append(ctx, "aaa", "0")
		x, _ = strconv.Atoi(rdb.Get(ctx, "aaa").Val())
		Assert(x == 3340)
	}
}
