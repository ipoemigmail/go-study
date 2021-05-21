package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg = new(sync.WaitGroup)
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(1 * time.Second)
			wg.Done()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println(elapsed)
}
