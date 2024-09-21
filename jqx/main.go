package main

import (
	"fmt"
	"io/fs"
	"os"

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

func main() {
	prog := jqx.Program{
		Args: os.Args[1:],

		Open: func(f string) (fs.File, error) { return os.Open(f) },

		Stdin:  os.Stdin,
		Stdout: os.Stdout,

		StdinIsTerminal:  isTerminal(os.Stdin),
		StdoutIsTerminal: isTerminal(os.Stdout),
	}

	try(prog.Main())
	try(os.CopyFS(".", prog.FS))
}
