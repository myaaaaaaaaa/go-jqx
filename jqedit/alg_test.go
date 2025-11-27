package main

import (
	"runtime"
	"testing"
)

func assertEqual[T comparable](tb testing.TB, got T, want T) {
	tb.Helper()
	if got != want {
		tb.Error("got", got)
		tb.Error("want", want)
	}
}

func TestWatcherDebounce(t *testing.T) {
	set, wait := watcher[int]()
	set(3)
	set(5)
	set(7)

	got := 0
	wait(&got)
	if got != 7 {
		t.Error("want 7  got", got)
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
		t.Error("want 4  got", got)
	}
}
func TestWatcherFull(t *testing.T) {
	set, wait := watcher[int]()
	go func() {
		for i := range 30 {
			for range 100 {
				set(i)
				runtime.Gosched()
			}
		}
		set(30)
	}()

	got := 100
	for want := range 31 {
		wait(&got)
		if got != want {
			t.Error("got", got)
			t.Error("want", want)
		}
	}
}

func TestTruncLines(t *testing.T) {
	got := truncLines("abcdefg", 10)
	assertEqual(t, got, "abcdefg")

	got = truncLines("abcdefg", 6)
	assertEqual(t, got, "abc...")
}
