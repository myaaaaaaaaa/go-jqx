package jqx

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func testRun(t *testing.T, stdin, want string, p *Program) fs.FS {
	t.Helper()

	var got bytes.Buffer

	p.Stdin = bytes.NewBufferString(strings.ReplaceAll(stdin, " ", "\n"))
	p.Println = func(s string) { fmt.Fprintln(&got, s) }

	rt, err := p.Main()
	if err != nil {
		t.Log(err)
		if want != "error" {
			t.Fail()
		}
		return nil
	}

	want = strings.ReplaceAll(want, " ", "\n") + "\n"
	assertEqual(t, got.String(), want)

	return rt
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

	testRun(t, `null`, "env files", &Program{Args: []string{"-e", "$vars | keys[]"}})
	testRun(t, `null`, "files", &Program{Args: []string{"$vars | keys[]"}})

	t.Setenv("XYZ", "_____")
	testRun(t, `"XYZ"`, "_____", &Program{Args: []string{"-e", "$env[.]"}})
	testRun(t, `"XYZ"`, "_____", &Program{Args: []string{"-e", "$vars.env[.]"}})
	testRun(t, `"XYZ"`, "error", &Program{Args: []string{"$env[.]"}})
}
func TestOpen(t *testing.T) {
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

	p.Args = []string{`"a.json" | $files[.][][]`, "a.json"}
	testRun(t, "", "1 2 3", &p)
	p.Args = []string{`.[][][]`, "a.json"}
	testRun(t, "", "1 2 3", &p)
	p.Args = []string{`.[][]`, "b.json"}
	testRun(t, "", "1 2 3", &p)

	for _, file := range []string{"c.json", "c.notjson", "d.txt", "e.txt"} {
		p.Args = []string{".", file}
		testRun(t, "", "error", &p)
	}

	p.Args = []string{"-r", ".[]", "d.txt", "e.txt"}
	testRun(t, "", "foo q w e r t y", &p)
}

func TestFS(t *testing.T) {
	testfs := fstest.MapFS{
		"x/d":      &fstest.MapFile{Mode: fs.ModeDir},
		"x/dd/a/d": &fstest.MapFile{Mode: fs.ModeDir},
		"x/f":      &fstest.MapFile{Data: []byte("file")},
		"x/ff/a/f": &fstest.MapFile{Data: []byte("file")},
	}

	p := Program{StdinIsTerminal: true}
	p.Find = func(s string) fs.FS {
		return must(fs.Sub(testfs, s))
	}

	p.Args = []string{`$find | keys[]`}
	testRun(t, "", "error", &p)

	p.Args = []string{"-find", "x", `$find | keys[]`}
	testRun(t, "", "d/ dd/ dd/a/ dd/a/d/ f ff/ ff/a/ ff/a/f", &p)
}

func TestDry(t *testing.T) {
	const q = `snapshot("\(.).json"; [.])`
	var fsys fs.FS
	p := Program{}

	ls := func() string {
		rt := must(fs.Glob(fsys, "*"))
		return strings.Join(rt, " ")
	}
	cat := func(filename string) string {
		data, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return "error"
		}
		return string(data)
	}

	for range 3 {
		p.Args = []string{"--dry-run", q}
		fsys = testRun(t, `false`, `false false.json`, &p)
		assertEqual(t, ls(), "")
		assertEqual(t, cat("false.json"), "error")

		p.Args = []string{q}
		fsys = testRun(t, `false`, `false`, &p)
		assertEqual(t, ls(), "false.json")
		assertEqual(t, cat("false.json"), "[false]")

		p.Args = []string{"-t", q}
		fsys = testRun(t, `false`, `false`, &p)
		assertEqual(t, ls(), "false.json")
		assertEqual(t, cat("false.json"), "[\n\tfalse\n]")
	}
}
