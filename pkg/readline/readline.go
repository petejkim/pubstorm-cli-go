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
	Input  io.Reader = os.Stdin
	Output io.Writer = os.Stdout
)

func Read(prompt string, retry bool) (string, error) {
	var (
		s   string
		err error
	)

	in := bufio.NewReader(Input)

	for {
		fmt.Fprint(Output, prompt)

		s, err = in.ReadString('\n')
		if err != nil {
			return "", err
		}

		s = strings.TrimRight(s, "\r\n")

		if s != "" || !retry {
			break
		}
	}

	return s, nil
}

func ReadSecurely(prompt string, retry bool) (string, error) {
	var (
		s   string
		err error
	)

	in := bufio.NewReader(Input)

	for {
		fmt.Fprint(Output, prompt)

		s, err = doReadSecurely(prompt, in)
		if err != nil {
			return "", err
		}

		if s != "" || !retry {
			break
		}
	}

	return s, nil
}

func doReadSecurely(prompt string, in *bufio.Reader) (string, error) {
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

	if useAnsi {
		fmt.Fprint(Output, "\033[8m")
	}

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
