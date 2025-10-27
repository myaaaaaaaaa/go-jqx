package proptest

import (
	"math/rand/v2"
	"strings"
)

type random struct {
	rand.Rand
}

func Rand(seed int) random {
	s := uint64(seed)
	return random{*rand.New(rand.NewPCG(s, ^s))}
}
func (r random) Strings(stringSet ...string) string {
	var rt strings.Builder
	for range r.IntN(30) {
		rt.WriteString(stringSet[r.IntN(len(stringSet))])
	}
	return rt.String()
}
func (r random) Chars(charSet string) string {
	return r.Strings(strings.Split(charSet, "")...)
}
