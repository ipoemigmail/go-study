package util

import (
	"fmt"
	"math/rand"
	"time"
)

func GetRateLimiter(duration time.Duration, limit int) <-chan time.Time {
	ticker := make(chan time.Time)
	rand.Seed(time.Now().Unix())
	go func() {
		for {
			time.Sleep(1*time.Second + time.Duration(rand.Intn(1000)))
			ticker <- time.Now()
		}
	}()
	return ticker
}

func PanicError(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s(%s)", msg, err))
	}
}
