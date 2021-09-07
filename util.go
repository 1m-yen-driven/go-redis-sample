package main

import (
	"math/rand"
	"strconv"
	"time"
)

type User struct {
	ID        int64
	Name      string
	Count     int
	CreatedAt time.Time // time.Time は truncate必須。pointer型は不可
}

func Random() int {
	return rand.Intn(100)
}
func RandStr() string {
	result := "あいうえおかきくけこ"
	for i := 0; i < 100; i++ {
		result += strconv.Itoa(Random())
	}
	return result
}
func RandUser() User {
	return User{
		ID:        int64(Random()),
		Name:      RandStr(),
		Count:     Random(),
		CreatedAt: time.Now().Truncate(time.Second),
	}
}

func Assert(cond bool) {
	if !cond {
		panic("assertion failed")
	}
}