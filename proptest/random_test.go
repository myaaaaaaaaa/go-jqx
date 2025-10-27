package proptest

import (
	"slices"
	"strings"
	"testing"
)

func TestChars(t *testing.T) {
	r := Rand(10)
	for range 200 {
		got := r.Chars("aaaaaa")
		if strings.Trim(got, "a") != "" {
			t.Error(got, "contains non-a")
		}

		got = r.Chars("abcde")
		if strings.Trim(got, "abcde") != "" {
			t.Error(got, "contains non-abcde")
		}
	}
}
func TestCharsNotIn(t *testing.T) {
	r := Rand(10)
	for range 5 {
		got := r.Chars("abc")

		s := strings.Split(got, "")
		slices.Sort(s)
		s = slices.Compact(s)
		if strings.Join(s, "") == "abc" {
			return
		}
	}
	t.Error("statistical anomaly")
}

func TestCharsActuallyRandom(t *testing.T) {
	const charSet = "abcdefg"

	r := Rand(50)
	doNotWant := r.Chars(charSet)
	for range 100 {
		got := r.Chars(charSet)
		if doNotWant == got {
			t.Error("not random enough")
		}
	}
}
