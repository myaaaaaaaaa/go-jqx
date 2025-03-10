package jqx

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"iter"
	"maps"
	"math/rand/v2"
	"slices"

	"github.com/itchyny/gojq"
)

var builtins = must(gojq.Parse(`
	def shuffle: shuffle("A seed");
`)).FuncDefs

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
func shuffle(input any, seed []any) any {
	s := fmt.Sprint(seed[0])
	r := rand.PCG{}
	r.Seed(0x701877fa59de0c16, 0x99f94bdb8143b770)
	for _, c := range s {
		hi := r.Uint64() ^ uint64(len(s))
		lo := r.Uint64() ^ uint64(c)
		r.Seed(hi, lo)
	}

	rt := slices.Clone(input.([]any))
	rand.New(&r).Shuffle(len(rt), func(i, j int) {
		rt[i], rt[j] = rt[j], rt[i]
	})
	return rt
}
func hasher(f func() hash.Hash) func(any, []any) any {
	return func(input any, _ []any) any {
		h := f()
		h.Write([]byte(input.(string)))
		return hex.EncodeToString(h.Sum(nil))
	}
}

func (s *State) Compile(code constString) FanOut {
	parsed, err := gojq.Parse(string(code))
	failif(err, "parsing query")

	builtins := slices.Clone(builtins)
	parsed.FuncDefs = append(builtins, parsed.FuncDefs...)

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
		gojq.WithFunction("shuffle", 1, 1, shuffle),
		gojq.WithFunction("md5", 0, 0, hasher(md5.New)),
		gojq.WithFunction("sha1", 0, 0, hasher(sha1.New)),
		gojq.WithFunction("sha256", 0, 0, hasher(sha256.New)),
		gojq.WithFunction("sha512", 0, 0, hasher(sha512.New)),
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
