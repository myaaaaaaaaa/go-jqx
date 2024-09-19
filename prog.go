package jqx

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"iter"
	"maps"
	"runtime/debug"
	"slices"
)

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
			yield(v)
		}
	}
}

type Program struct {
	Args []string

	Open  func(string) (fs.File, error)
	State State

	Stdin  io.Reader
	Stdout io.Writer

	StdinIsTerminal  bool
	StdoutIsTerminal bool
}
type flags struct {
	script    string
	filenames []string

	raw bool
	dry bool
}

func (f *flags) populate(args []string) {
	fset := flag.NewFlagSet("", flag.ExitOnError)
	fset.BoolVar(&f.raw, "r", false, `stdin, stdout, and files are newline-separated strings (unimplemented)`)
	fset.BoolVar(&f.dry, "dry-run", false, `don't persist snapshots`)

	usage := fset.Usage
	fset.Usage = func() {
		usage()
		out := fset.Output()

		info, _ := debug.ReadBuildInfo()
		if info != nil {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "jqx", info.Main.Version, "built with", info.GoVersion)
			fmt.Fprintf(out, "Update: go install %s@latest\n\n", info.Path)
		}
	}

	fset.Parse(args)

	args = fset.Args()
	f.script = "."
	if len(args) > 0 {
		f.script, f.filenames = args[0], args[1:]
	}
}

func (p *Program) Main() (rtErr error) {
	defer catch[failError](&rtErr)

	var f flags
	f.populate(p.Args)

	files := map[string]any{}
	for _, filename := range f.filenames {
		file, err := p.Open(filename)
		failif(err, "loading")
		defer file.Close()

		v := slices.Collect(decoder(file, filename))
		if len(v) == 1 {
			files[filename] = v[0]
		} else {
			files[filename] = v
		}
	}

	input := decoder(p.Stdin, "stdin")
	if p.StdinIsTerminal {
		input = func(yield func(any) bool) { yield(files) }
	}

	marshal := json.Marshal
	if p.StdoutIsTerminal {
		marshal = func(v any) ([]byte, error) {
			if v, ok := v.(string); ok {
				return []byte(v), nil
			}
			return json.MarshalIndent(v, "", "\t")
		}
	}

	query := p.State.Compile(constString(f.script))
	for v := range input {
		for v := range query(v) {
			v := must(marshal(v))
			fmt.Fprintln(p.Stdout, string(v))
		}
	}

	if f.dry {
		for _, file := range slices.Collect(maps.Keys(p.State.Files)) {
			fmt.Fprintln(p.Stdout, file)
		}
		p.State.Files = nil
	}

	return nil
}
