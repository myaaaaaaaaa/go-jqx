package main

import (
	"runtime"
	"testing"
)

func TestWatcherDebounce(t *testing.T) {
	set, wait := watcher[int]()
	set(3)
	set(5)
	set(7)

	got := 0
	wait(&got)
	if got != 7 {
		t.Error(got)
	}
}
func TestWatcherWait(t *testing.T) {
	set, wait := watcher[int]()
	go func() {
		for range 100 {
			set(0)
			runtime.Gosched()
		}
		set(4)
	}()

	got := 0
	wait(&got)
	if got != 4 {
		t.Error(got)
	}
}
func TestWatcherFull(t *testing.T) {
	set, wait := watcher[int]()
	go func() {
		for i := range 10 {
			for range 100 {
				set(i)
				runtime.Gosched()
			}
		}
		set(10)
	}()

	got := 100
	for want := range 11 {
		wait(&got)
		if got != want {
			t.Error(got)
		}
	}
}
