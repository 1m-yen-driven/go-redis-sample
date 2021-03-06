package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

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

var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
	DB:   0, // 0 - 15
})

// トランザクションのサンプル
func main() {
	fmt.Println("RUN: TRANSACTION SAMPLE")
	ctx := context.Background()
	count := 2000
	key := "hoge"

	// 通常のIncrByはTransaction不要であることを確認
	Measure("IncrBy", func() {
		rdb.FlushDB(ctx)
		parallel(count, func() {
			rdb.IncrBy(ctx, key, 1)
		})
		AssertEq(rdb.Get(ctx, key).Val(), strconv.Itoa(count))
	})

	// User.Count を Transactionを使わずIncrByする
	Measure("IncrBy User.Count", func() {
		parallelUser(ctx, count, key, func() {
			user := User{}
			DecodePtrStringCmd(rdb.Get(ctx, key), &user)
			user.Count += 1
			rdb.Set(ctx, key, EncodePtr(&user), 0)
		})
	})

	// User.Count を Transactionを使ってIncrByする
	Measure("IncrBy User.Count (Transaction)", func() {
		parallelUser(ctx, count, key, func() {
			OptmisticLockPipe(ctx, func(tx *redis.Tx, pipe redis.Pipeliner) {
				// tx を使ってRead操作を全て行ってから
				user := User{}
				DecodePtrStringCmd(tx.Get(ctx, key), &user)
				user.Count += 1
				// pipe を使ってWrite操作を全て行う
				pipe.Set(ctx, key, EncodePtr(&user), 0)
			}, key)
		})
	})

	// SetNXを使えば悲観ロックできるのでそのサンプル
	// http://redis.shibu.jp/commandreference/strings.html

	// 悲観ロック用のDBを使う方式
	Measure("IncrBy User.Count (SetNX)", func() {
		rdbForLock.FlushDB(ctx)
		parallelUser(ctx, count, key, func() {
			PessimiticLock(rdb, ctx, key, func() {
				user := User{}
				DecodePtrStringCmd(rdb.Get(ctx, key), &user)
				user.Count += 1
				rdb.Set(ctx, key, EncodePtr(&user), 0)
			})
		})
	})

	// 現在参照しているDBにロック用のキーを追加する方式
	// キーが増えてしまうが、Pipelineが効くので少し速い
	Measure("IncrBy User.Count (SetNX Pipe)", func() {
		parallelUser(ctx, count, key, func() {
			PessimiticLockPipe(ctx, key, func(pipe redis.Pipeliner) {
				user := User{}
				DecodePtrStringCmd(rdb.Get(ctx, key), &user)
				user.Count += 1
				pipe.Set(ctx, key, EncodePtr(&user), 0)
			})
		})
	})
}

// -------------- ロック用 ------------------------------------

// 楽観ロック
func OptmisticLockPipe(ctx context.Context, f func(tx *redis.Tx, pipe redis.Pipeliner), keys ...string) {
	// errが出ず成功するまで試行する。key は複数指定可能
	for nil != rdb.Watch(ctx, func(tx *redis.Tx) error {
		// ここの err が大事
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			f(tx, pipe)
			return nil
		})
		return err
	}, keys...) {
	}
}

// 悲観ロック用のDBを使わずPipelineを使用するもの
func PessimiticLockPipe(ctx context.Context, key string, f func(redis.Pipeliner)) {
	// 他のキーと競合しないようにmd5しておく(キー数が増えてしまうので注意)
	lockHash := md5.Sum([]byte(key))
	lockKey := hex.EncodeToString(lockHash[:])
	for {
		// デッドロック回避のために、時間経過でもロックが解除されるようにしている
		if !rdb.SetNX(ctx, lockKey, "1", 5*time.Second).Val() {
			// どのくらいSleepするかは要チューニング
			time.Sleep(time.Microsecond * time.Duration(1000))
			continue
		}
		// ２つ以上のコマンドを打つので、効率重視でPipeline使用
		pipe := rdb.Pipeline()
		defer pipe.Close()
		f(pipe)
		pipe.Del(ctx, lockKey)
		pipe.Exec(ctx)
		return
	}
}

// 悲観ロック用のDBを作成
var rdbForLock = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
	DB:   15, // とりあえず最後の
})

// 悲観ロック用のDBを使うもの
func PessimiticLock(rdb *redis.Client, ctx context.Context, key string, f func()) {
	lockHash := md5.Sum([]byte(key))
	lockKey := hex.EncodeToString(lockHash[:]) + strconv.Itoa(rdb.Options().DB)
	for {
		// デッドロック回避のために、時間経過でもロックが解除されるようにしている
		if !rdbForLock.SetNX(ctx, lockKey, "1", 5*time.Second).Val() {
			// どのくらいSleepするかは要チューニング
			time.Sleep(time.Microsecond * time.Duration(1000))
			continue
		}
		f()
		rdbForLock.Del(ctx, lockKey)
		return
	}
}

// -------------- テスト用 ------------------------------------

func parallel(count int, f func()) {
	maxGoroutineNum := 20 // go routine 生成コストを鑑みて制限つきにしている
	var wg sync.WaitGroup
	for i := 0; i < maxGoroutineNum; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < count/maxGoroutineNum; j++ {
				f()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func parallelUser(ctx context.Context, count int, key string, f func()) {
	rdb.FlushDB(ctx)
	user := RandUser()
	user.Count = 0
	rdb.Set(ctx, key, EncodePtr(&user), 0)
	parallel(count, f)
	DecodePtrStringCmd(rdb.Get(ctx, key), &user)
	fmt.Println("Sum:", user.Count, ", Expect:", count)
}
