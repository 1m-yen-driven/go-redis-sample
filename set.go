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
	fmt.Println("RUN: SET SAMPLE")
	ctx := context.Background()
	key := "hoge"
	// 取得：SISMEMBER / SCARD / SMEMBERS
	// 追加削除： SADD / SREM
	// セット操作： S{INTER, UNION, DIFF}(STORE)  / SMOVE
	rdb.FlushDB(ctx)
	// SAdd
	user := RandUser()
	user2 := RandUser()
	rdb.SAdd(ctx, key, EncodePtr(&user), EncodePtr(&user), EncodePtr(&user2))
	// SCard
	Assert(rdb.SCard(ctx, key).Val() == 2)
	// SIsMember
	Assert(rdb.SIsMember(ctx, key, EncodePtr(&user)).Val())
	Assert(!rdb.SIsMember(ctx, key, "aaa").Val())
	// SMembers
	for _, v := range rdb.SMembers(ctx, key).Val() {
		user3 := User{}
		DecodePtrSliceCmdElem(v, &user3)
		fmt.Println(user3)
	}
	// SMove
	key2 := "hoge2"
	rdb.SMove(ctx, key, key2, EncodePtr(&user2))
	Assert(rdb.SCard(ctx, key).Val() == 1)
	Assert(rdb.SCard(ctx, key2).Val() == 1)
	Assert(rdb.SIsMember(ctx, key, EncodePtr(&user)).Val())
	Assert(rdb.SIsMember(ctx, key2, EncodePtr(&user2)).Val())
	// SUnionStore
	rdb.SUnionStore(ctx, key2, key, key2)
	Assert(rdb.SCard(ctx, key2).Val() == 2)
	Assert(rdb.SIsMember(ctx, key2, EncodePtr(&user)).Val())
	Assert(rdb.SIsMember(ctx, key2, EncodePtr(&user2)).Val())
	// SDiffStore
	rdb.SDiffStore(ctx, key2, key2, key)
	Assert(rdb.SCard(ctx, key2).Val() == 1)
	Assert(!rdb.SIsMember(ctx, key2, EncodePtr(&user)).Val())
	Assert(rdb.SIsMember(ctx, key2, EncodePtr(&user2)).Val())
	// SInter
	Assert(len(rdb.SInter(ctx, key, key2).Val()) == 0)
	// SUnion
	for _, v := range rdb.SUnion(ctx, key, key2).Val() {
		user3 := User{}
		DecodePtrSliceCmdElem(v, &user3)
		fmt.Println(user3)
	}
}
