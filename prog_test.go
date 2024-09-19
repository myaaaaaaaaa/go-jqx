package jqx

import (
	"bytes"
	"strings"
	"testing"
)

func testRun(t *testing.T, stdin, want string, p Program) {
	t.Helper()

	var got bytes.Buffer

	p.Stdin = bytes.NewBufferString(stdin)
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
	testRun(t, "[]  []", "[] []", Program{})
	testRun(t, "[]  [}", "error", Program{})
	testRun(t, "[]  []", "{}", Program{StdinIsTerminal: true})

	testRun(t, "[10]", "[10]", Program{})
	testRun(t, "[10]", "[\n\t10\n]", Program{StdoutIsTerminal: true})

	testRun(t, `"a"`, `"a"`, Program{})
	testRun(t, `"a"`, `a`, Program{StdoutIsTerminal: true})

	testRun(t, "[10]", "[10]", Program{})
	testRun(t, "[10]", "[10]", Program{Args: []string{"."}})
	testRun(t, "[10]", "10", Program{Args: []string{".[]"}})
	testRun(t, "[10]", "error", Program{Args: []string{"|"}})
}
func TestFS(t *testing.T) {
	testFS := State{Files: map[string]any{
		"a.json":    "[1][2][3]",
		"b.json":    "[1,2,3]",
		"c.notjson": "[}",
	}}.FS()

	p := Program{StdinIsTerminal: true}
	testRun(t, "", "{}", p)

	p.Open = testFS.Open
	testRun(t, "", "{}", p)

	p.Args = []string{`.[][][]`, "a.json"}
	testRun(t, "", "1 2 3", p)
	p.Args = []string{`.[][]`, "b.json"}
	testRun(t, "", "1 2 3", p)

	p.Args = []string{`.`, "c.json"}
	testRun(t, "", "error", p)
	p.Args = []string{`.`, "c.notjson"}
	testRun(t, "", "error", p)
}
