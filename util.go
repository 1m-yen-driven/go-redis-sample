package main

// only for Random / User Type / Assert / Measure

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

//
// User Type & Random -----------------------------
//
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

//
// Assert --------------------------------
//
func Assert(cond bool) {
	if !cond {
		panic("assertion failed")
	}
}
func AssertEq(a, b interface{}) {
	if a != b {
		fmt.Println(a)
		fmt.Println("IS NOT")
		fmt.Println(b)
		panic("assertion failed")
	}
}
func AssertNotEq(a, b interface{}) {
	if a == b {
		fmt.Println(a)
		fmt.Println("IS")
		fmt.Println(b)
		panic("assertion failed")
	}
}

//
// Measure ----------------------------------
//
func Measure(title string, f func()) {
	now := time.Now()
	f()
	fmt.Println(title)
	fmt.Println(time.Since(now))
}
