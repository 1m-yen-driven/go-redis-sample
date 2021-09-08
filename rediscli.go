package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/shamaton/msgpack"
	"os"
)

func Show(valuePtr interface{}, codes ...string) {
	for _, code := range codes {
		msgpack.Decode([]byte(code), valuePtr)
		m, _ := json.MarshalIndent(valuePtr, "", "  ")
		fmt.Println(string(m))
	}
}

func main() {
	for _, v := range os.Args {
		if v == "parse" {
			// パーサとして使用
			// echo 'keys *' | redis-cli | sed 's/^/get /' | redis-cli | go run rediscli.go util.go parse
			// 改行で死ぬので参考程度で
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				Show(&User{}, scanner.Text())
			}
			return
		}
	}
	// MsgPack のデコードサンプル
	// $ redis-cli
	// > dbsize
	// > keys *  // all
	// > scan 0  // partial
	// > get hoge
	got := "\x84\xa2ID9\xa4Name\xd9*bb6b99c0-10c1-11ec-a036-1e0023143188(READ)\xa5Count/\xa9CreatedAt\xd6\xffa8\xe4\x90"
	// 型は分かっている必要あり
	Show(&User{}, got)
}
