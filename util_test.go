package jqx

import (
	"fmt"
	"testing"
)

func assertEqual[T comparable](t *testing.T, got T, want T) {
	t.Helper()
	if got != want {
		t.Error("got", got)
		t.Error("want", want)
	}
}
func assertString(t *testing.T, got any, want string) {
	t.Helper()
	assertEqual(t, fmt.Sprint(got), want)
}
