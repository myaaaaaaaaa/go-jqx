package jqx

import "testing"

func assertEqual[T comparable](t *testing.T, got T, want T) {
	t.Helper()
	if got != want {
		t.Error("got", got)
		t.Error("want", want)
	}
}
