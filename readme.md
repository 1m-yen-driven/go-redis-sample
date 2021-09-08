# Go Redis Sample

Redis + Go + MSGPack を使う際のサンプル集。

- systemctl で enable して再起動後の動作保証を忘れないこと。
- Redisはコマンドを送るまでは落ちないので、起動順序問題は通常は大丈夫

## 基本

`go run basic.go util.go`

- 基本の文字列型＋MsgPackの操作。速度面でMGet/MSet推奨。パイプラインで改善可能。
- Exists / Del / Keys / Rename(NX) / DBSize / FlushDB
- Get / Set / GetSet / MGet / MSet / IncrBy / Append

## パイプライン(トランザクションなし)

`go run pipe.go util.go`

- Execute() をするまでは値が定まらないが、一括で実行できるので、高速。
- ただし、MSet/MGetを使ったほうが速い
- Get / Set / MGet / MSet

## トランザクション+パイプライン

`go run tx.go util.go`

- 楽観ロックなので成功するまでやる必要がある

## Echo 上での動作サンプル

`go run echo.go util.go`

## コンテナの型別

###　双方向リスト

`go run list.go util.go`

### セット

`go run set.go util.go`

### ソート済みセット

`go run sortedset.go util.go`

### ハッシュテーブル

`go run hashtable.go util.go`

### ジオメトリ操作

`go run geo.go util.go`
