package jqx

import (
	"math/rand/v2"
	"testing"
)

func randString(r *rand.Rand, charSet string) string {
	rt := make([]byte, r.IntN(30))
	for i := range rt {
		rt[i] = charSet[r.IntN(len(charSet))]
	}
	return string(rt)
}

func TestProperty(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for range 200 {
		s := randString(r, "ab  \t\r\n\n")
		checkPagetrim(t, s)
	}
}
