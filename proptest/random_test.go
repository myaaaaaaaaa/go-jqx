package proptest

import (
	"slices"
	"strings"
	"testing"
)

func TestChars(t *testing.T) {
	r := Rand(10)
	for range 200 {
		for i := range 5 {
			got := r.Chars(strings.Repeat("a", i+1))
			if strings.Trim(got, "a") != "" {
				t.Error(got, "contains non-a")
			}
		}

		got := r.Chars("abcde")
		if strings.Trim(got, "abcde") != "" {
			t.Error(got, "contains non-abcde")
		}
	}
}
func TestCharsNotIn(t *testing.T) {
	const charSet = "abcdefg"

	r := Rand(10)
	got := ""
	for range 20 {
		got += r.Chars(charSet)
	}

	s := strings.Split(got, "")
	slices.Sort(s)
	s = slices.Compact(s)
	if strings.Join(s, "") != charSet {
		t.Error(got)
	}
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

func TestStrings(t *testing.T) {
	r := Rand(15)
	for range 200 {
		got := r.Strings("   ", " hello ")
		for s := range strings.FieldsSeq(got) {
			if s != "hello" {
				t.Error(s)
			}
		}

		for i := range 5 {
			got := r.Strings(make([]string, i+1)...)
			if got != "" {
				t.Error(got)
			}
		}
	}
}
func TestStringsNotIn(t *testing.T) {
	r := Rand(22)
	got := ""
	for range 10 {
		got += r.Strings(" hello ", "    world!    ")
	}

	s := strings.Split(got, " ")
	slices.Sort(s)
	s = slices.Compact(s)
	if strings.Join(s, "") != "helloworld!" {
		t.Error(got)
	}
}
