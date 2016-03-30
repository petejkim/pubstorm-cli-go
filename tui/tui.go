package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

var Out io.Writer = os.Stdout

func init() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		Out = colorable.NewColorableStdout()
	}
}

func Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(Out, a...)
}

func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(Out, format, a...)
}

func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(Out, a...)
}
