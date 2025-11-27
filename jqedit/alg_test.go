package main

import (
	"testing"
	"testing/synctest"
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
	assertEqual(t, got, 7)
}
func TestWatcherWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		set, wait := watcher[int]()

		got := 7
		go wait(&got)

		synctest.Wait()

		for range 5 {
			for range 100 {
				set(7)
			}
			synctest.Wait()
			assertEqual(t, got, 7)
		}

		set(10)
		synctest.Wait()
		assertEqual(t, got, 10)
	})
}

func TestTruncLines(t *testing.T) {
	got := truncLines("abcdefg", 20)
	assertEqual(t, got, "abcdefg")

	got = truncLines("abcdefg", 6)
	assertEqual(t, got, "abc...")

	t.Run("over limit", func(t *testing.T) {
		for n := range 8 {
			got := truncLines("1234567", n)
			assertEqual(t, len(got), n)
		}
	})
	t.Run("under limit", func(t *testing.T) {
		for n := range 10 {
			want := "abcdefg"
			got := truncLines(want, n+7)
			assertEqual(t, got, want)
		}
	})
	t.Run("no overflow", func(t *testing.T) {
		for n := range 10 {
			got := truncLines("abcdefg", n)
			assertEqual(t, len(got) <= n, true)
		}
	})
}
