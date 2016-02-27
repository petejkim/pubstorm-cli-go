package apperror

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Error struct {
	Code        string
	Err         error
	Description string
	IsFatal     bool
}

func New(code string, err error, description string, isFatal bool) *Error {
	return &Error{code, err, description, isFatal}
}

func (e *Error) Error() string {
	var m []string

	if e.IsFatal {
		m = []string{"Fatal Error:"}
	} else {
		m = []string{"Error:"}
	}

	if e.Description != "" {
		m = append(m, e.Description)
	} else {
		if e.Code != "" && e.Err != nil {
			m = append(m, fmt.Sprintf("%s (%s)", e.Err.Error(), e.Code))
		} else if e.Code != "" {
			m = append(m, e.Code)
		} else if e.Err != nil {
			m = append(m, e.Err.Error())
		} else {
			m = append(m, "Something went wrong!")
		}
	}

	return strings.Join(m, " ")
}

func (e *Error) Print() {
	fmt.Fprintln(os.Stderr, e)
}

func (e *Error) Handle() {
	if e.IsFatal {
		log.Fatalln(e.Error())
	} else {
		e.Print()
	}
}
