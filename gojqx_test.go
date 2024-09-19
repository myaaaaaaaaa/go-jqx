package jqx

import (
	"io/fs"
	"slices"
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	query := new(State).Compile(`range(.)*2+1 | tostring`)
	got := slices.Collect(query(3))

	assertEqual(t, len(got), 3)
	assertEqual(t, [3]any(got), [...]any{"1", "3", "5"})
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
func TestSnapshot(t *testing.T) {
	state := State{}
	query := state.Compile(`to_entries[] | snapshot(.key;.value) | .key`)

	entries := map[string]any{
		"a.txt":    "aaa",
		"b/b.json": "ugh",
		"c/c.json": []any{2, 8},
		"d.json":   map[string]any{"e": 10},
		"f.json":   8,
	}

	filenames := slices.Collect(query(entries))
	cat := []string{}

	fsys := state.FS()

	for _, filename := range filenames {
		filename := filename.(string)
		data := must(fs.ReadFile(fsys, filename))
		cat = append(cat, string(data))
	}

	assertEqual(t, len(filenames), 5)

	assertEqual(t,
		[5]any(filenames),
		[...]any{"a.txt", "b/b.json", "c/c.json", "d.json", "f.json"},
	)

	assertEqual(t,
		strings.Join(cat, " "),
		`aaa ugh [2,8] {"e":10} 8`,
	)
}
