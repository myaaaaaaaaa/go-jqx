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

func getMarshaler(tab bool, str bool) func(v any) []byte {
	return func(v any) []byte {
		if str {
			if v, ok := v.(string); ok {
				return []byte(v)
			}
		}

		if tab {
			return must(json.MarshalIndent(v, "", "\t"))
		}
		return must(json.Marshal(v))
	}
}
func toFS(m map[string]any, marshaler func(any) []byte) fs.FS {
	if marshaler == nil {
		marshaler = getMarshaler(false, true)
	}

	rt := fstest.MapFS{}

	for k, v := range m {
		rt[k] = &fstest.MapFile{Data: marshaler(v)}
	}

	return rt
}
