package main

import (
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

type NoLogger struct{}

func NewNoLogger() *NoLogger {
	return &NoLogger{}
}

func (n *NoLogger) Printf(string, ...interface{}) {}

func main() {
	wg := new(sync.WaitGroup)
	conn, _, err := zk.Connect([]string{"localhost:2181"}, time.Second*10, zk.WithLogger(NewNoLogger()))
	if err != nil {
		panic(err)
	}

	returns := make([]chan string, 100000)

	for i := 0; i < 100000; i++ {
		returns[i] = make(chan string, 1)
		wg.Add(1)
		go func(ch chan string) {
			defer wg.Done()
			d, _, _ := conn.Get("/kakao/commerce/kcdl/v2/jobs/meta/dev-hadoop/local/test.app/config/key1")
			ch <- string(d)
		}(returns[i])
		//if err != nil {
		//	panic(err)
		//}
	}
	wg.Wait()

	//println(string(v))
}
