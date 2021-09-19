package main

// only for Random / User Type / Assert / Measure

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

//
// User Type & Random -----------------------------
//
type User struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Count     int       `db:"count"`
	CreatedAt time.Time `db:"created_at"`
	// time.Time は truncate必須。pointer型は不可
}

func Random() int {
	return rand.Intn(100)
}
func RandStr() string {
	uuid, _ := uuid.NewUUID()
	return uuid.String()
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
	fmt.Println("###", title, "###")
	now := time.Now()
	f()
	fmt.Println(time.Since(now))
}
