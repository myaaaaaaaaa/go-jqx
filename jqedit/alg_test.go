package main

import (
	"runtime"
	"testing"
)

func TestWatcher(t *testing.T) {
	set, wait := watcher[int]()

	set(3)
	set(5)
	set(7)

	got := 0
	wait(&got)
	if got != 7 {
		t.Error(got)
	}

	go func() {
		for range 100 {
			set(7)
			runtime.Gosched()
		}
		set(1)
	}()

	wait(&got)
	if got != 1 {
		t.Error(got)
	}

	go func() {
		for i := range 10 {
			for range 100 {
				set(i)
				runtime.Gosched()
			}
		}
		set(10)
	}()

	for want := range 11 {
		wait(&got)
		if got != want {
			t.Error(got)
		}
	}
}
