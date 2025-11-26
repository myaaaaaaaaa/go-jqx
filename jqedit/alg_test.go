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

	got := 0
	wait(&got)
	if got != 7 {
		t.Error(got)
	}

	go func() {
		for range 100 {
			push(7)
			runtime.Gosched()
		}
		push(1)
	}()

	wait(&got)
	if got != 1 {
		t.Error(got)
	}

	go func() {
		for i := range 10 {
			for range 100 {
				push(i)
				runtime.Gosched()
			}
		}
		push(10)
	}()

	for want := range 11 {
		wait(&got)
		if got != want {
			t.Error(got)
		}
	}
}
