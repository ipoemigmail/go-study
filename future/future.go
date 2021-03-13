package future

import (
	"sync"
)

type Future struct {
	calc     <-chan interface{}
	result   interface{}
	isFinish bool
	guard    sync.Mutex
}

func (f *Future) Get() interface{} {
	if f.result == nil {
		f.guard.Lock()
		f.result = <-f.calc
		f.isFinish = true
		f.guard.Unlock()
	}
	return f.result
}

// NewFutureValue is
func NewFutureValue(a interface{}) Future {
	r := make(chan interface{}, 1)
	r <- a
	return newFutureWithChan(r)
}

// NewFuture is
func NewFuture(body func() interface{}) Future {
	r := make(chan interface{}, 1)
	go func() {
		r <- body()
	}()
	return newFutureWithChan(r)
}

// NewFuture is
func newFutureWithChan(c chan interface{}) Future {
	return Future{calc: c}
}

// Join is
func Join(fs ...*Future) Future {
	r := make(chan interface{}, 1)
	ra := make([]interface{}, len(fs))
	go func() {
		for i, f := range fs {
			ra[i] = f.Get()
		}
		r <- ra
	}()
	return newFutureWithChan(r)
}

func firstOfTwo(a *Future, b *Future) Future {
	r := make(chan interface{}, 1)
	go func() {
		if a.isFinish {
			r <- a.result
		} else if b.isFinish {
			r <- b.result
		} else {
			select {
			case a1 := <-a.calc:
				(*a).guard.Lock()
				(*a).result = a1
				(*a).isFinish = true
				(*a).guard.Unlock()
				r <- a1
			case b1 := <-b.calc:
				(*b).guard.Lock()
				(*b).result = b1
				(*b).isFinish = true
				(*b).guard.Unlock()
				r <- b1
			}
		}
	}()
	return newFutureWithChan(r)
}

// First is
func First(chans ...*Future) Future {
	var r Future
	for _, c := range chans {
		r = firstOfTwo(chans[0], c)
	}
	return r
}

// AndThen is
func AndThen(f *Future, body func(a interface{}) Future) Future {
	r := make(chan interface{}, 1)
	go func() {
		vv := body(f.Get())
		v2 := vv.Get()
		r <- v2
	}()
	return newFutureWithChan(r)
}

// Map is
func Map(f *Future, body func(a interface{}) interface{}) Future {
	r := make(chan interface{}, 1)
	go func() {
		r <- body(f.Get())
	}()
	return newFutureWithChan(r)
}
