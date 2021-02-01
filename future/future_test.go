package future

import (
	"reflect"
	"testing"
	"time"
)

func TestNewFuture(t *testing.T) {
	result := <-NewFuture(func() interface{} { return 1 })
	expected := 1
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestJoinFutures(t *testing.T) {
	result1 := NewFuture(func() interface{} { return 1 })
	result2 := NewFuture(func() interface{} { return 2 })
	result := <-JoinFutures(result1, result2)
	expected := [2]interface{}{1, 2}
	if reflect.DeepEqual(result, expected) {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFuture(t *testing.T) {
	result1 := NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 := NewFuture(func() interface{} { return 2 })
	result := <-FirstFuture(result1, result2)
	expected := 2

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}

	result1 = NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 = NewFuture(func() interface{} { return 2 })
	result = <-FirstFuture(result2, result1)
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFutures(t *testing.T) {
	result1 := NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 := NewFuture(func() interface{} { time.Sleep(1 * time.Second); return 2 })
	result3 := NewFuture(func() interface{} { return 3 })
	result := <-FirstFutures(result1, result2, result3)
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}
