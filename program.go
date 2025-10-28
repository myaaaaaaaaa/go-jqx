package jqx

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"maps"
	"os"
	"runtime/debug"
	"slices"
	"strings"
)

func lines(r io.Reader, name string) iter.Seq[any] {
	return func(yield func(any) bool) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				break
			}
		}
		err := scanner.Err()
		failif(err, "scanning %s", name)
	}
}

func decoder(r io.Reader, name string) iter.Seq[any] {
	return func(yield func(any) bool) {
		decoder := json.NewDecoder(r)
		for {
			var v any
			err := decoder.Decode(&v)
			if err == io.EOF {
				return
			}
			failif(err, "decoding %s", name)
			if !yield(v) {
				break
			}
		}
	}
}

type Program struct {
	Args []string

	Open func(string) (fs.File, error)
	Find func(string) fs.FS

	Stdin   io.Reader
	Println func(string)

	StdinIsTerminal  bool
	StdoutIsTerminal bool
}
type flags struct {
	args []string

	dry     bool
	tab     bool
	rawIn   bool
	jsonOut bool
	env     bool

	find string
}

func (f *flags) populate(args []string) {
	fset := flag.NewFlagSet("", flag.ExitOnError)
	fset.BoolVar(&f.dry, "dry-run", false, `don't persist snapshots`)
	fset.BoolVar(&f.tab, "t", false, `(tab) always indent output`)
	fset.BoolVar(&f.rawIn, "r", false, `(raw) inputs are newline-separated strings`)
	fset.BoolVar(&f.jsonOut, "j", false, `(json) always output json (strings are unwrapped by default)`)
	fset.BoolVar(&f.env, "e", false, `(env) enable $env`)

	fset.StringVar(&f.find, "find", "", `enable $find`)

	usage := fset.Usage
	fset.Usage = func() {
		usage()
		out := fset.Output()

		info, _ := debug.ReadBuildInfo()
		if info != nil {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "jqx", info.Main.Version, "built with", info.GoVersion)
			fmt.Fprintf(out, "\n\tgo install %s@latest\n\n", info.Path)
		}
	}

	fset.Parse(args)
	f.args = fset.Args()
}

func (p Program) Main() (fsys fs.FS, rtErr error) {
	defer catch[failError](&rtErr)

	var f flags
	f.populate(p.Args)

	script, filenames := ".", f.args
	if len(filenames) > 0 {
		script, filenames = filenames[0], filenames[1:]
	}

	files := map[string]any{}
	slices.Values(filenames)(func(filename string) bool {
		file, err := p.Open(filename)
		failif(err, "loading")
		defer file.Close()

		var data any

		if f.rawIn {
			b, err := io.ReadAll(file)
			failif(err, "reading")
			data = string(b)
		} else {
			v := slices.Collect(decoder(file, filename))
			if len(v) == 1 {
				data = v[0]
			} else {
				data = v
			}
		}

		files[filename] = data

		return true
	})

	input := decoder(p.Stdin, "stdin")
	if f.rawIn {
		input = lines(p.Stdin, "stdin")
	}
	if p.StdinIsTerminal {
		input = func(yield func(any) bool) { yield(files) }
	}

	marshal := getMarshaler(
		f.tab || (p.StdoutIsTerminal && !f.jsonOut),
		!f.jsonOut,
	)

	state := State{
		Globals: map[string]any{"files": files},
	}
	if f.env {
		envVars := map[string]any{}
		for _, v := range os.Environ() {
			k, v, _ := strings.Cut(v, "=")
			envVars[k] = v
		}
		state.Globals["env"] = envVars
	}
	if f.find != "" {
		find := map[string]any{}
		err := fs.WalkDir(p.Find(f.find), ".", func(path string, d fs.DirEntry, err error) error {
			if path == "." {
				return err
			}

			if d.IsDir() {
				path = path + "/"
			}
			find[path] = true
			return err
		})
		failif(err, "finding subdirs")
		state.Globals["find"] = find
	}
	query := state.Compile(constString(script))
	for v := range input {
		for v := range query(v) {
			v := marshal(v)
			p.Println(string(v))
		}
	}

	if f.dry {
		for _, file := range slices.Sorted(maps.Keys(state.Files)) {
			p.Println(file)
		}
		state.Files = nil
	}

	fsys = toFS(state.Files, getMarshaler(f.tab, !f.jsonOut))
	rtErr = nil

	return
}
