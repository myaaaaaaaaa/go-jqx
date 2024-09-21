package jqx

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"testing/fstest"
)

type failError string

func (f failError) Error() string { return string(f) }

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
func failif(err error, format string, args ...any) {
	if err == nil {
		return
	}

	s := fmt.Sprintf(format, args...)
	s = fmt.Sprintf("error while %s: %v", s, err)
	panic(failError(s))
}
func catch[T error](rt *error) {
	err := recover()
	switch err := err.(type) {
	case nil:
	case T:
		*rt = err
	default:
		panic(err)
	}
}

func toFS(m map[string]any, tab bool) fs.FS {
	rt := fstest.MapFS{}

	for k, v := range m {
		var data []byte

		switch v := v.(type) {
		case string:
			data = []byte(v)
		default:
			if tab {
				data = must(json.MarshalIndent(v, "", "\t"))
			} else {
				data = must(json.Marshal(v))
			}
		}

		rt[k] = &fstest.MapFile{Data: data}
	}

	return rt
}
