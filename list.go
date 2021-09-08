package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/shamaton/msgpack"
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
	fmt.Println("RUN: LIST SAMPLE")
	ctx := context.Background()
	key := "hoge"
	// 取得更新：LINDEX / LSET / LRANGE / LLEN
	// 追加削除：{L,R}PUSH / {L,R}POP / RPOPLPUSH
	// リスト操作：LTRIM / LREM
	rdb.FlushDB(ctx)
	// LPush
	user := RandUser()
	rdb.LPush(ctx, key, EncodePtr(&user), EncodePtr(&user))
	// RPop
	user2 := User{}
	DecodePtrStringCmd(rdb.RPop(ctx, key), &user2)
	AssertEq(user, user2)
	// RPopLPush
	user3 := User{}
	DecodePtrStringCmd(rdb.RPopLPush(ctx, key, key), &user3)
	AssertEq(user, user3)
	// LLen
	rdb.LPush(ctx, key, EncodePtr(&user), EncodePtr(&user))
	Assert(rdb.LLen(ctx, key).Val() == 3)
	// index 0
	user4 := User{}
	DecodePtrStringCmd(rdb.LIndex(ctx, key, 0), &user4)
	AssertEq(user, user4)
	// last index
	user5 := User{}
	DecodePtrStringCmd(rdb.LIndex(ctx, key, -1), &user5)
	AssertEq(user, user5)
	// LRange
	// stop: including index
	for _, val := range rdb.LRange(ctx, key, 0, -1).Val() {
		user6 := User{}
		DecodePtrSliceCmdElem(val, &user6)
		fmt.Println(user6)
	}
}
