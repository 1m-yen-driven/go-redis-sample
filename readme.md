# Go Redis Sample

Redis + Go + MSGPack を使う際のサンプル集。

- systemctl で enable して再起動後の動作保証を忘れないこと。
- Redisはコマンドを送るまでは落ちないので、起動順序問題は通常は大丈夫
- https://pkg.go.dev/github.com/go-redis/redis/v8

## 基本

`go run basic.go util.go`

- 基本の文字列型＋MsgPackの操作。速度面でMGet/MSet推奨。パイプラインで高速化可能。
- Exists / Del / Keys / Rename(NX) / DBSize / FlushDB
- Get / Set / GetSet / MGet / MSet / IncrBy / Append

## パイプライン(トランザクションなし)

`go run pipe.go util.go`

- Execute() をするまでは値が定まらないが、一括で実行できるので、高速。
- ただし、MSet/MGetを使ったほうが速い
- Get / Set / MGet / MSet

## トランザクション

`go run tx.go util.go`

- 通常の書き方では楽観ロックなので成功するまでやる必要がある
  - 同じキーに対しての書き込みが多いと効率が悪い
- [SetNXを使えば悲観ロックも可能](http://redis.shibu.jp/commandreference/strings.html)なので、そのサンプルもあり
  - トランザクションなし:  40ms
  - 楽観ロック: 600ms
  - 悲観ロック: 134ms
  - ISUCONでは適宜悲観ロックした方がいい
- Watch / TxPipelined / SetNX

## Echo 上での動作サンプル

`go run echo.go util.go`

- `ctx := c.Request().Context()`

## Redis-Cli 連携

`go run rediscli.go util.go`

- MSGPack されたものをパースするサンプル
  - `echo 'keys *' | redis-cli | sed 's/^/get /' | redis-cli | go run rediscli.go util.go parse`
- [SlowLog](https://redis.io/commands/slowlog)
  - `> slowlog get 10`
  - time(micro sec) / command は把握可能

## コレクションの型別

多くの種類のコレクションが使えるので、そのサンプル

- http://redis.shibu.jp/commandreference/index.html
- http://mogile.web.fc2.com/redis/commands/geoadd.html

###　双方向リスト

`go run list.go util.go`

- 取得更新：LINDEX / LSET / LRANGE / LLEN
- 追加削除：{L,R}PUSH / {L,R}POP / RPOPLPUSH
- リスト操作：LTRIM / LREM

### セット

`go run set.go util.go`

### ソート済みセット

`go run sortedset.go util.go`

### ハッシュテーブル

`go run hashtable.go util.go`

### ジオメトリ操作

`go run geo.go util.go`
