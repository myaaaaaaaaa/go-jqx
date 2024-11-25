package jqx

import (
	"iter"
	"maps"
	"slices"

	"github.com/itchyny/gojq"
)

type constString string
type FanOut func(any) iter.Seq[any]

type State struct {
	Files   map[string]any
	Globals map[string]any
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

	globalKeys := slices.Sorted(func(yield func(string) bool) {
		for key := range maps.Keys(s.Globals) {
			yield("$" + key)
		}
	})
	globalValues := slices.Collect(func(yield func(any) bool) {
		for _, globalKey := range globalKeys {
			yield(s.Globals[globalKey[1:]])
		}
	})

	globalKeys = append(globalKeys, "$vars")
	globalValues = append(globalValues, s.Globals)

	compiled, err := gojq.Compile(
		parsed,
		gojq.WithFunction("snapshot", 2, 2, s.snapshot),
		gojq.WithVariables(globalKeys),
	)
	failif(err, "compiling query")

	return func(v any) iter.Seq[any] {
		return func(yield func(any) bool) {
			iter := compiled.Run(v, globalValues...)

			for {
				v, ok := iter.Next()
				if !ok {
					break
				}

				err, _ := v.(error)
				failif(err, "running query")

				if !yield(v) {
					break
				}
			}
		}
	}
}
