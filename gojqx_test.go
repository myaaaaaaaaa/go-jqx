package jqx

import (
	"slices"
	"testing"
)

func TestCompile(t *testing.T) {
	query := new(State).Compile(`range(.)*2+1 | tostring`)
	got := slices.Collect(query(3))

	assertString(t, got, `[1 3 5]`)

	i := 0
	for v := range query(100) {
		if len(v.(string)) != 1 {
			break
		}
		i++
	}
	assertEqual(t, i, 5)
}
func TestError(t *testing.T) {
	err := func(code string) (rt error) {
		defer catch[error](&rt)
		new(State).Compile(constString(code))
		return nil
	}

	assertEqual(t, err("."), nil)
	assertEqual(t, err("true"), nil)
	assertEqual(t, err("false"), nil)

	assertEqual(t, err("|") != nil, true)
	assertEqual(t, err("maybe") != nil, true)
}
func TestState(t *testing.T) {
	state := State{}
	query := state.Compile(`
		fromjson | to_entries[] |
		snapshot(.key;.value+.value) | .key+.key
	`)

	keys := slices.Collect(query(`{"a":"aa", "c":[3], "q":{"e":10}, "r":5}`))

	assertEqual(t, len(keys), 4)
	assertEqual(t, len(state.Files), 4)

	state = State{Globals: map[string]any{
		"k": keys,
		"v": state.Files,
	}}
	query = state.Compile(`$k[],$v[] | tostring`)

	got := slices.Collect(query(nil))
	assertString(t, got, `[aa cc qq rr aaaa [3,3] {"e":10} 10]`)
}

func TestHash(t *testing.T) {
	query := new(State).Compile(`md5`)

	assertString(t, slices.Collect(query("")), `[d41d8cd98f00b204e9800998ecf8427e]`)
	assertString(t, slices.Collect(query("\n")), `[68b329da9893e34099c7d8ad5cb9c940]`)
	assertString(t, slices.Collect(query("hi")), `[49f68a5c8493ec2c0bf489821c21fc3b]`)
}
func TestShuffle(t *testing.T) {
	query := new(State).Compile(`range(4) | [range(.)] | [shuffle(range(30))|join("")] | unique | join("-")`)
	assertString(t, slices.Collect(query(nil)), `[ 0 01-10 012-021-102-120-201-210]`)

	query = new(State).Compile(`def seq(f): range(4) | [range(.)] | [shuffle(range(30)|f)]; seq(.),seq(5) | unique | length`)
	assertString(t, slices.Collect(query(nil)), `[1 1 2 6 1 1 1 1]`)
}
