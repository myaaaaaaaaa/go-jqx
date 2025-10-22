package jqx

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"iter"
	"math/rand/v2"
	"slices"

	"github.com/itchyny/gojq"
)

var builtins = must(gojq.Parse(`
	def shuffle: shuffle("A seed");
`)).FuncDefs

type constString string
type FanOut func(any) iter.Seq[any]

type nexter func() (any, bool)

func (f nexter) Next() (any, bool) { return f() }

func iterTest(input any, _ []any) gojq.Iter {
	switch input := input.(type) {
	case int:
		return nexter(func() (any, bool) {
			input--
			return input, input >= 0
		})
	case string:
		return nexter(func() (any, bool) {
			rt := input
			if len(input) > 0 {
				input = input[1:]
			}
			return rt, len(rt) > 0
		})
	}

	return gojq.NewIter(errors.New("oops"))
}

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

func pagetrim(input any, _ []any) any {
	s := []byte(input.(string))

	s = bytes.TrimSpace(s)
	lines := bytes.Split(s, []byte("\n"))
	i := 0
	for _, line := range lines {
		lines[i] = bytes.TrimSpace(line)
		if len(lines[i]) == 0 && i > 0 && len(lines[i-1]) == 0 {
		} else {
			i++
		}
	}
	s = bytes.Join(lines[:i], []byte("\n"))

	return string(s)
}

func htmlq(input any, args []any) any {
	selector := args[0]
	rt, err := htmlQuerySelector(input.(string), selector.(string))
	if err != nil {
		return err
	}
	return rt
}
func htmlt(input any, _ []any) any {
	rt := htmlExtractText(input.(string))
	return rt
}

func (s *State) Compile(code constString) FanOut {
	parsed, err := gojq.Parse(string(code))
	failif(err, "parsing query")

	builtins := slices.Clone(builtins)
	parsed.FuncDefs = append(builtins, parsed.FuncDefs...)

	globalKeys := slices.Sorted(func(yield func(string) bool) {
		for key := range s.Globals {
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
		gojq.WithIterFunction("_itertest", 0, 0, iterTest),
		gojq.WithFunction("snapshot", 2, 2, s.snapshot),
		gojq.WithFunction("shuffle", 1, 1, shuffle),
		gojq.WithFunction("md5", 0, 0, hasher(md5.New)),
		gojq.WithFunction("sha1", 0, 0, hasher(sha1.New)),
		gojq.WithFunction("sha256", 0, 0, hasher(sha256.New)),
		gojq.WithFunction("sha512", 0, 0, hasher(sha512.New)),
		gojq.WithFunction("pagetrim", 0, 0, pagetrim),
		gojq.WithFunction("_htmlq", 1, 1, htmlq),
		gojq.WithFunction("_htmlt", 0, 0, htmlt),
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
