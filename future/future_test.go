package future

import (
	"reflect"
	"testing"
	"time"
)

func TestNewFuture(t *testing.T) {
	result := NewFuture(func() interface{} { return 1 }).Get()
	expected := 1
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestJoin(t *testing.T) {
	result1 := NewFuture(func() interface{} { return 1 })
	result2 := NewFuture(func() interface{} { return 2 })
	result := Join(result1, result2).Get()
	expected := [2]interface{}{1, 2}
	if reflect.DeepEqual(result, expected) {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFuture(t *testing.T) {
	result1 := NewFuture(func() interface{} { time.Sleep(2 * time.Second); return 1 })
	result2 := NewFuture(func() interface{} { return 2 })
	result := First(result1, result2).Get()
	expected := 2

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}

	result1 = NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 = NewFuture(func() interface{} { return 2 })
	result = First(result2, result1).Get()
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFutures(t *testing.T) {
	result1 := NewFuture(func() interface{} { time.Sleep(2 * time.Second); return 1 })
	result2 := NewFuture(func() interface{} { time.Sleep(2 * time.Second); return 2 })
	result3 := NewFuture(func() interface{} { return 3 })
	result := First(result1, result2, result3).Get()
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestAndThen(t *testing.T) {
	result1 := NewFuture(func() interface{} { time.Sleep(3 * time.Second); return 1 })
	result2 := NewFuture(func() interface{} { return 3 })
	result3 := AndThen(result2, func(a interface{}) Future {
		return NewFutureValue(a)
	})
	result := First(result1, result3).Get()
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestMap(t *testing.T) {
	result1 := NewFuture(func() interface{} { time.Sleep(3 * time.Second); return 1 })
	result2 := NewFuture(func() interface{} { return 3 })
	result3 := Map(result2, func(a interface{}) interface{} {
		return a
	})
	result := First(result1, result3).Get()
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func BenchmarkNewFuture(b *testing.B) {
	v := make([]Future, 1000000)
	for i := 0; i < b.N; i++ {
		for i := range v {
			v[i] = NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 1 })
		}
		for _, f := range v {
			f.Get()
		}
	}
}
