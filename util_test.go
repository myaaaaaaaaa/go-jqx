package jqx

import (
	"io/fs"
	"testing"
)

func assertEqual[T comparable](t *testing.T, got T, want T) {
	t.Helper()
	if got != want {
		t.Error("got", got)
		t.Error("want", want)
	}
}

func fsGetText(fsys fs.FS, filename string) string {
	data := must(fs.ReadFile(fsys, filename))
	return string(data)
}
