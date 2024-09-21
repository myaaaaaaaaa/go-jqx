package jqx

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
)

func testRun(t *testing.T, stdin, want string, p *Program) {
	t.Helper()

	var got bytes.Buffer

	p.Stdin = bytes.NewBufferString(strings.ReplaceAll(stdin, " ", "\n"))
	p.Stdout = &got

	err := p.Main()
	if err != nil {
		t.Log(err)
		if want != "error" {
			t.Fail()
		}
		return
	}

	want = strings.ReplaceAll(want, " ", "\n") + "\n"
	assertEqual(t, got.String(), want)
}

func TestProgram(t *testing.T) {
	testRun(t, "[]  []", "[] []", &Program{})
	testRun(t, "[]  [}", "error", &Program{})
	testRun(t, "[]  []", "{}", &Program{StdinIsTerminal: true})

	testRun(t, "[10]", "[10]", &Program{})
	testRun(t, "[10]", "[10]", &Program{Args: []string{"-j"}})
	testRun(t, "[10]", "[10]", &Program{Args: []string{"-j"}, StdoutIsTerminal: true})
	testRun(t, "[10]", "[\n\t10\n]", &Program{StdoutIsTerminal: true})
	testRun(t, "[10]", "[\n\t10\n]", &Program{Args: []string{"-t"}})
	testRun(t, "[10]", "[\n\t10\n]", &Program{Args: []string{"-t", "-j"}})

	testRun(t, `"a"`, `a`, &Program{})
	testRun(t, `"a"`, `a`, &Program{StdoutIsTerminal: true})
	testRun(t, `"a"`, `"a"`, &Program{Args: []string{"-j"}})

	testRun(t, "[10]", "[10]", &Program{})
	testRun(t, "[10]", "[10]", &Program{Args: []string{"."}})
	testRun(t, "[10]", "10", &Program{Args: []string{".[]"}})
	testRun(t, "[10]", "error", &Program{Args: []string{"|"}})

	testRun(t, `] })`, `] })`, &Program{Args: []string{"-r"}})
	testRun(t, `] })`, `] [1] }) [2]`, &Program{Args: []string{"-r", "., [length]"}})
}
func TestFS(t *testing.T) {
	testFiles := map[string]any{
		"a.json":    "[1][2][3]",
		"b.json":    "[1,2,3]",
		"c.notjson": "[}",
		"d.txt":     "foo",
		"e.txt":     "q\nw\ne\nr\nt\ny",
	}

	p := Program{StdinIsTerminal: true}
	testRun(t, "", "{}", &p)

	p.Open = toFS(testFiles, nil).Open
	testRun(t, "", "{}", &p)

	p.Args = []string{`.[][][]`, "a.json"}
	testRun(t, "", "1 2 3", &p)
	p.Args = []string{`.[][]`, "b.json"}
	testRun(t, "", "1 2 3", &p)

	for _, file := range []string{"c.json", "c.notjson", "d.txt", "e.txt"} {
		p.Args = []string{".", file}
		testRun(t, "", "error", &p)
	}

	p.Args = []string{"-r", ".[][] | .+.", "d.txt"}
	testRun(t, "", "foofoo", &p)
	p.Args = []string{"-r", ".[][] | .+.", "e.txt"}
	testRun(t, "", "qq ww ee rr tt yy", &p)
}
func TestDry(t *testing.T) {
	const q = `snapshot("\(.).json"; [.])`
	p := Program{}
	files := func() string {
		rt := must(fs.Glob(p.FS, "*"))
		return strings.Join(rt, " ")
	}

	for range 3 {
		p.Args = []string{"--dry-run", q}
		testRun(t, `false`, `false false.json`, &p)
		assertEqual(t, files(), "")
		assertEqual(t, fileText(p.FS, "false.json"), "error")

		p.Args = []string{q}
		testRun(t, `false`, `false`, &p)
		assertEqual(t, files(), "false.json")
		assertEqual(t, fileText(p.FS, "false.json"), "[false]")

		p.Args = []string{"-t", q}
		testRun(t, `false`, `false`, &p)
		assertEqual(t, files(), "false.json")
		assertEqual(t, fileText(p.FS, "false.json"), "[\n\tfalse\n]")
	}
}
