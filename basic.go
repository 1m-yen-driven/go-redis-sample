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
func DecodePtrSliceCmdElem(partsOfSliceCmd interface{}, valuePtr interface{}) {
	msgpack.Decode([]byte(partsOfSliceCmd.(string)), valuePtr)
}

var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
	DB:   0, // 0 - 15
})

func main() {
	fmt.Println("RUN: BASIC SAMPLE")
	ctx := context.Background()
	// Set / Get / GetSet
	{
		user := RandUser()
		rdb.Set(ctx, "aaa", EncodePtr(&user), 0)
		user2 := User{}
		DecodePtrStringCmd(rdb.Get(ctx, "aaa"), &user2)
		AssertEq(user, user2)
		user3 := RandUser()
		DecodePtrStringCmd(rdb.GetSet(ctx, "aaa", EncodePtr(&user3)), &user3)
		AssertEq(user3, user2)
		DecodePtrStringCmd(rdb.Get(ctx, "aaa"), &user3)
		AssertNotEq(user3, user2)
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
		Assert(len(keys) == 1)
		AssertEq(keys[0], "aaa")
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
			DecodePtrSliceCmdElem(got, &user)
			AssertEq(user, users[i])
		}
	}

	// IncrBy / Append
	{
		rdb.FlushDB(ctx)
		rdb.Set(ctx, "aaa", strconv.Itoa(100), 0)
		rdb.IncrBy(ctx, "aaa", 234)
		x, _ := strconv.Atoi(rdb.Get(ctx, "aaa").Val())
		AssertEq(x, 334)
		rdb.Append(ctx, "aaa", "0")
		x, _ = strconv.Atoi(rdb.Get(ctx, "aaa").Val())
		AssertEq(x, 3340)
	}
}
