package future

import (
	"reflect"
	"testing"
	"time"
)

func TestNewFuture(t *testing.T) {
	f := NewFuture(func() interface{} { return 1 })
	result := f.Get()
	expected := 1
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestJoin(t *testing.T) {
	f1 := NewFuture(func() interface{} { return 1 })
	f2 := NewFuture(func() interface{} { return 2 })
	f3 := Join(&f1, &f2)
	result := f3.Get()
	expected := [2]interface{}{1, 2}
	if reflect.DeepEqual(result, expected) {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFuture(t *testing.T) {
	f1 := NewFuture(func() interface{} { time.Sleep(2 * time.Second); return 1 })
	f2 := NewFuture(func() interface{} { return 2 })
	f3 := First(&f1, &f2)
	result := f3.Get()
	expected := 2

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}

	f1 = NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	f2 = NewFuture(func() interface{} { return 2 })
	f3 = First(&f1, &f2)
	result = f3.Get()
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFutures(t *testing.T) {
	f1 := NewFuture(func() interface{} { time.Sleep(2 * time.Second); return 1 })
	f2 := NewFuture(func() interface{} { time.Sleep(2 * time.Second); return 2 })
	f3 := NewFuture(func() interface{} { return 3 })
	f4 := First(&f1, &f2, &f3)
	result := f4.Get()
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestAndThen(t *testing.T) {
	f1 := NewFuture(func() interface{} { time.Sleep(3 * time.Second); return 1 })
	f2 := NewFuture(func() interface{} { return 3 })
	f3 := AndThen(&f2, func(a interface{}) Future {
		return NewFutureValue(a)
	})
	f4 := First(&f1, &f3)
	result := f4.Get()
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestMap(t *testing.T) {
	f1 := NewFuture(func() interface{} { time.Sleep(3 * time.Second); return 1 })
	f2 := NewFuture(func() interface{} { return 3 })
	f3 := Map(&f2, func(a interface{}) interface{} {
		return a
	})
	f4 := First(&f1, &f3)
	result := f4.Get()
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
		for i := range v {
			v[i].Get()
		}
	}
}
