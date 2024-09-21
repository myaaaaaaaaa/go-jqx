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

func fileText(fsys fs.FS, filename string) string {
	data, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return "error"
	}
	return string(data)
}
