package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/shamaton/msgpack"
	"math/rand"
	"strconv"
	"time"
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
	// 取得：Z(REV)RANK  / ZCARD / ZSCORE
	// 追加削除： ZADD / ZREM
	// スコア更新： ZINCRBY
	// 順序クエリ：Z{,REV,REM}RANGE{,BYSCORE} / ZCOUNT
	// 集合操作(キー複合時はSum/Min/Maxで選べる)： ZUIONSTORE / ZINTERSTORE
	fmt.Println("RUN: BASIC SAMPLE")
	ctx := context.Background()
	key := "hoge"
	rdb.FlushDB(ctx)
	// create data (score == UnixTimeMicroSec)
	zs := []*redis.Z{}
	for i := 0; i < 10; i++ {
		user := RandUser()
		user.Count = i
		user.Name = strconv.Itoa(i)
		now := time.Now()
		user.CreatedAt = now.Truncate(time.Microsecond)
		z := redis.Z{float64(now.UnixNano() / 1000), EncodePtr(&user)}
		zs = append(zs, &z)
		fmt.Println(user)
	}
	rand.Shuffle(len(zs), func(i, j int) { zs[i], zs[j] = zs[j], zs[i] })
	// ZAdd
	rdb.ZAdd(ctx, key, zs...)
	rdb.ZAdd(ctx, key, zs...)
	// ZCard
	Assert(rdb.ZCard(ctx, key).Val() == 10)
	// ZRank
	fmt.Println(rdb.ZRank(ctx, key, zs[1].Member.(string)).Val())
	fmt.Println(rdb.ZRevRank(ctx, key, zs[1].Member.(string)).Val())
	// ZRevRange (get latest)
	for i, v := range rdb.ZRevRange(ctx, key, 0, 1).Val() {
		// NOTE: ↑ stop including index
		user := User{}
		DecodePtrSliceCmdElem(v, &user)
		fmt.Println(user)
		// 0-indexed
		Assert(int64(i) == rdb.ZRevRank(ctx, key, EncodePtr(&user)).Val())
	}
	// ZRevRangeByScore (get latest)
	for i, v := range rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{"-inf", "+inf", 0, 2}).Val() {
		user := User{}
		DecodePtrSliceCmdElem(v, &user)
		fmt.Println(user)
		// 0-indexed
		Assert(int64(i) == rdb.ZRevRank(ctx, key, EncodePtr(&user)).Val())
	}
	// ZRangeByScore (get oldest)
	for _, v := range rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{"-inf", "+inf", 0, 2}).Val() {
		user := User{}
		DecodePtrSliceCmdElem(v, &user)
		fmt.Println(user)
	}
}
