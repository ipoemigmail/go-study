package future

import "fmt"

type Future struct {
	calc <-chan interface{}
}

func (f Future) Get() interface{} {
	return <-f.calc
}

// NewFutureValue is
func NewFutureValue(a interface{}) Future {
	r := make(chan interface{}, 1)
	r <- a
	return Future{calc: r}
}

// NewFuture is
func NewFuture(body func() interface{}) Future {
	r := make(chan interface{}, 1)
	go func() {
		r <- body()
	}()
	return Future{calc: r}
}

// Join is
func Join(fs ...Future) Future {
	r := make(chan interface{}, 1)
	ra := make([]interface{}, len(fs))
	go func() {
		for i, f := range fs {
			ra[i] = <-f.calc
		}
		r <- ra
	}()
	return Future{calc: r}
}

func firstOfTwo(a Future, b Future) Future {
	r := make(chan interface{}, 1)
	go func() {
		select {
		case a1 := <-a.calc:
			r <- a1
		case b1 := <-b.calc:
			r <- b1
		}
	}()
	return Future{calc: r}
}

// First is
func First(chans ...Future) Future {
	r := chans[0]
	for _, c := range chans {
		r = firstOfTwo(r, c)
	}
	return r
}

// AndThen is
func AndThen(f Future, body func(a interface{}) Future) Future {
	r := make(chan interface{}, 1)
	go func() {
		v := <-f.calc
		fmt.Println("1")
		v2 := body(v).Get()
		r <- v2
	}()
	return Future{calc: r}
}

// Map is
func Map(f Future, body func(a interface{}) interface{}) Future {
	r := make(chan interface{}, 1)
	go func() {
		r <- body(<-f.calc)
	}()
	return Future{calc: r}
}
