package main

import (
	"fmt"
	"time"
)

// The playground now uses square brackets for type parameters. Otherwise,
// the syntax of type parameter lists matches the one of regular parameter
// lists except that all type parameters must have a name, and the type
// parameter list cannot be empty. The predeclared identifier "any" may be
// used in the position of a type parameter constraint (and only there);
// it indicates that there are no constraints.

type Future[T any] struct {
	calc <-chan T
	result *T
}

type Tuple[T1 any, T2 any] struct {
	Left T1
	Right T2
}

func (f1 *Future[T]) Get() T {
	if f1.result == nil {
		v := <-f1.calc
		f1.result = &v
	}
	return *f1.result
}

func NewFutureValue[T any](v T) Future[T] {
	r := make(chan T, 1)
	return Future[T]{calc:r, result:&v}
}

func NewFuture[T any](body func() T) Future[T] {
	r := make(chan T, 1)
	go func() {
		r <- body()
	}()
	return newFutureFromChan(r)
}

func newFutureFromChan[T any](c <-chan T) Future[T] {
	return Future[T]{calc:c, result:nil}
}

func Join[T any](fs ...Future[T]) Future[[]T] {
	r := make(chan []T, 1)
	go func() {
		ar := make([]T, len(fs))
		for i, f := range fs {
			ar[i] = f.Get()
		}
		r <- ar
	}()
	return newFutureFromChan(r)
}

func firstOf[T any](f1 *Future[T], f2 *Future[T]) Future[T] {
	if f1.result != nil {
		return NewFutureValue[T](*f1.result)
	} else if f2.result != nil {
		return NewFutureValue[T](*f2.result)
	} else {
		fmt.Println("inin")
		r := make(chan T, 1)
		go func() {
			select {
			case v := <-f1.calc:
				f1.result = &v
				r <- v
			case v := <-f2.calc:
				f2.result = &v
				r <- v
			}
		}()
		return newFutureFromChan(r)
	}
}

func First[T any](f1 *Future[T], fs ...*Future[T]) Future[T] {
	r := *f1
	for _, f := range fs {
		r = firstOf(&r, f)
	}
	return r
}

func main() {
	f1 := NewFuture(func() int { time.Sleep(5 * time.Second); return 1 })
	f2 := NewFutureValue(2)
	f3 := First(&f1, &f2)
	fmt.Println(f3.Get())
	fmt.Println(f3.Get())
	fmt.Println(f1.Get())
	fmt.Println(f1.Get())
}

