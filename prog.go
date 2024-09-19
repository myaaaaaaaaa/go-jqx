package jqx

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"iter"
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

func (p *Program) Main() (rtErr error) {
	defer catch[failError](&rtErr)

	script := "."
	filenames := p.Args
	if len(filenames) > 0 {
		script, filenames = filenames[0], filenames[1:]
	}

	files := map[string]any{}
	for _, filename := range filenames {
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

	query := p.State.Compile(constString(script))
	for v := range input {
		for v := range query(v) {
			v := must(marshal(v))
			fmt.Fprintln(p.Stdout, string(v))
		}
	}

	return nil
}
