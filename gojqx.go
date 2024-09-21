package jqx

import (
	"fmt"
	"iter"

	"github.com/itchyny/gojq"
)

type constString string
type FanOut func(any) iter.Seq[any]

type State struct {
	Files map[string]any
}

func (s *State) snapshot(input any, kv []any) (rt any) {
	rt = input
	filename, jsonData := kv[0], kv[1]

	if s.Files == nil {
		s.Files = map[string]any{}
	}
	s.Files[filename.(string)] = jsonData

	return
}

func (s *State) Compile(code constString) FanOut {
	parsed, err := gojq.Parse(string(code))
	failif(err, "parsing query")

	compiled, err := gojq.Compile(
		parsed,
		gojq.WithFunction("snapshot", 2, 2, s.snapshot),
	)
	failif(err, "compiling query")

	return func(v any) iter.Seq[any] {
		return func(yield func(any) bool) {
			iter := compiled.Run(v)

			for {
				v, ok := iter.Next()
				if !ok {
					break
				}

				if err, ok := v.(error); ok {
					fmt.Println(err)
					continue
				}

				if !yield(v) {
					break
				}
			}
		}
	}
}
