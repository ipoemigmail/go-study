package future

import (
	"reflect"
	"testing"
	"time"
)

func TestSpawn(t *testing.T) {
	result := <-Spawn(func() interface{} { return 1 })
	expected := 1
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestJoin(t *testing.T) {
	result1 := Spawn(func() interface{} { return 1 })
	result2 := Spawn(func() interface{} { return 2 })
	result := <-Join(result1, result2)
	expected := [2]interface{}{1, 2}
	if reflect.DeepEqual(result, expected) {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFuture(t *testing.T) {
	result1 := Spawn(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 := Spawn(func() interface{} { return 2 })
	result := <-First(result1, result2)
	expected := 2

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}

	result1 = Spawn(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 = Spawn(func() interface{} { return 2 })
	result = <-First(result2, result1)
	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}

func TestFirstFutures(t *testing.T) {
	result1 := Spawn(func() interface{} { time.Sleep(1 * time.Second); return 1 })
	result2 := Spawn(func() interface{} { time.Sleep(1 * time.Second); return 2 })
	result3 := Spawn(func() interface{} { return 3 })
	result := <-First(result1, result2, result3)
	expected := 3

	if result != expected {
		t.Errorf("expected:%d actual:%d", expected, result)
	}
}
