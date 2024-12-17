package jqx

import (
	"fmt"
	"testing"
)

func assertEqual[T comparable](tb testing.TB, got T, want T) {
	tb.Helper()
	if got != want {
		tb.Error("got", got)
		tb.Error("want", want)
	}
}
func assertString(t *testing.T, got any, want string) {
	t.Helper()
	assertEqual(t, fmt.Sprint(got), want)
}
