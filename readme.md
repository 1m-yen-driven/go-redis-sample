# Go Redis Sample

Redis + Go + MSGPack を使う際のサンプル集。

- systemctl で enable して再起動後の動作保証を忘れないこと。

## 基本

`go run basis.go util.go`

基本操作。MGet/MSetしないと遅い。パイプラインで改善可能。

## パイプライン(トランザクションなし)

`go run pipe.go util.go`

Execute() をするまでは値が定まらないが、一括で実行できるので、爆速。

## トランザクション+パイプライン

`go run tx.go util.go`

楽観ロック

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
