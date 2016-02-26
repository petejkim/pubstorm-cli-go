package readline

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nitrous-io/rise-cli-go/pkg/term"
)

var (
	Input  = os.Stdin
	Output = os.Stdout
)

func Read(prompt string) (string, error) {
	fmt.Fprint(Output, prompt)

	in := bufio.NewReader(Input)
	s, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(s, "\r\n"), nil
}

func ReadSecurely(prompt string) (string, error) {
	s, err := doReadSecurely(prompt)
	fmt.Fprintln(Output)
	if err != nil {
		return "", err
	}
	if s == "" {
		return ReadSecurely(prompt)
	}
	return s, err
}

func doReadSecurely(prompt string) (string, error) {
	useAnsi := false // if true, use ansi to conceal password

	if Input == os.Stdin {
		stdin, _, _ := term.StdStreams()
		stdinFd, stdinIsTerminal := term.GetFdInfo(stdin)

		if stdinIsTerminal {
			oldState, err := term.SetRawTerminal(stdinFd)
			if err == nil {
				useAnsi = false
				defer func() {
					term.RestoreTerminal(stdinFd, oldState)
				}()
			}
		} else {
			// can't make terminal raw, use ansi
			useAnsi = true
		}
	}

	fmt.Fprint(Output, prompt)

	if useAnsi {
		fmt.Fprint(Output, "\033[8m")
	}

	in := bufio.NewReader(Input)

	s := []byte{}
	for {
		b, err := in.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 10 /* LF */ || b == 13 /* CR */ {
			break
		}
		if b == 3 /* ^C */ || b == 4 /* ^D */ {
			return "", io.EOF
		}
		if b == 8 /* BS */ || b == 127 /* DEL */ {
			if len(s) > 0 {
				s = s[:len(s)-1]
			}
		}

		if b >= 32 && b <= 126 { // only allow printable ascii chars
			s = append(s, b)
		}
	}

	if useAnsi {
		fmt.Fprint(Output, "\033[28m\033[F"+prompt+"\033[K")
	}

	return string(s), nil
}
