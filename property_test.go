package jqx

import (
	"math/rand"
	"strings"
	"testing"
)

const charSet = "ab\n"

func generate(r *rand.Rand) string {
	n := r.Intn(100)
	var builder strings.Builder
	for i := 0; i < n; i++ {
		builder.WriteByte(charSet[r.Intn(len(charSet))])
	}
	return builder.String()
}

func TestProperty(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	for i := 0; i < 100; i++ {
		s := generate(r)
		for _, c := range s {
			if !strings.ContainsRune(charSet, c) {
				t.Fatalf("generated string %q contains invalid character %q", s, c)
			}
		}
		checkPagetrim(t, s)
	}
}
