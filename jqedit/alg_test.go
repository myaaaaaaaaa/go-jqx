package main

import (
	"runtime"
	"testing"
)

func TestDebounce(t *testing.T) {
	push, wait := debounceQueue[int]()

	push(3)
	push(5)
	push(7)

	i := 0
	wait(&i)
	if i != 7 {
		t.Error(i)
	}

	go func() {
		for range 100 {
			push(7)
			runtime.Gosched()
		}
		push(1)
	}()

	wait(&i)
	if i != 1 {
		t.Error(i)
	}
}
