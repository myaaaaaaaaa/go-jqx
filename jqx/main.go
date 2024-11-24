package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"

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

func main() {
	prog := jqx.Program{
		Args: os.Args[1:],

		Open: func(f string) (fs.File, error) { return os.Open(f) },
		Find: os.DirFS,

		Stdin:  os.Stdin,
		Stdout: os.Stdout,

		StdinIsTerminal:  isTerminal(os.Stdin),
		StdoutIsTerminal: isTerminal(os.Stdout),
	}

	if prog.StdoutIsTerminal {
		cmd := exec.Command("less", "-S")
		pipe := must(cmd.StdinPipe())
		cmd.Stdout = os.Stdout
		prog.Stdout = pipe

		try(cmd.Start())

		defer cmd.Wait()
		defer pipe.Close()
	}

	try(prog.Main())
	try(os.CopyFS(".", prog.OutFS))
}
