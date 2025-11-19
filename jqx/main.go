package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/myaaaaaaaaa/go-jqx"
)

func isTerminal(f fs.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&fs.ModeCharDevice != 0
}

func try(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
	}
}
func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

type SigPipeWriter struct {
	io.WriteCloser
}

func (w SigPipeWriter) Write(b []byte) (int, error) {
	rt, err := w.WriteCloser.Write(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(13)
	}
	return rt, nil
}

func main() {
	prog := jqx.Program{
		Args: os.Args[1:],

		Open: func(f string) (fs.File, error) { return os.Open(f) },
		Find: os.DirFS,

		Stdin:   os.Stdin,
		Println: func(s string) { fmt.Println(s) },

		StdinIsTerminal:  isTerminal(os.Stdin),
		StdoutIsTerminal: isTerminal(os.Stdout),
	}

	if prog.StdinIsTerminal && prog.StdoutIsTerminal && len(prog.Args) == 0 {
		prog.Args = []string{"-h"}
	} else if prog.StdoutIsTerminal {
		os.Setenv("LESSCHARSET", "utf-8")

		cmd := exec.Command("less", "-RS")
		pipe := must(cmd.StdinPipe())
		pipe = SigPipeWriter{pipe}
		cmd.Stdout = os.Stdout

		if err := cmd.Start(); err != nil {
			defer fmt.Fprintf(os.Stderr, "warning: failed to pipe output to less: %v\n", err)
		} else {
			prog.Println = func(s string) {
				if s != "" {
					switch s[:1] + s[len(s)-1:] {
					case `{}`, `[]`, `""`:
						must(0, quick.Highlight(pipe, s+"\n", "json", "terminal256", "github"))
						return
					}
				}
				fmt.Fprintln(pipe, s)
			}
			defer cmd.Wait()
			defer pipe.Close()
		}
	}

	fsys, err := prog.Main()
	try(err)
	try(os.CopyFS(".", fsys))
}
