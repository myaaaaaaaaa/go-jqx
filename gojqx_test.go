package jqx

import (
	"slices"
	"testing"
)

func TestCompile(t *testing.T) {
	query := new(State).Compile(`range(.)*2+1 | tostring`)
	got := slices.Collect(query(3))

	assertEqual(t, len(got), 3)
	assertEqual(t, [3]any(got), [...]any{"1", "3", "5"})

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
		"$k": keys,
		"$v": state.Files,
	}}
	query = state.Compile(`[$k[],$v[]] | tostring`)

	got := slices.Collect(query(nil))
	assertEqual(t, len(got), 1)
	assertEqual(t, got[0], `["aa","cc","qq","rr","aaaa",[3,3],{"e":10},10]`)
}
