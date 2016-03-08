package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Prints and exits if err is not nil
func ExitIfError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Exits if err is not nil. If err is not an EOF error, prints error message
func ExitIfErrorOrEOF(err error) {
	if err != nil {
		if err != io.EOF {
			log.Fatalln(err)
		}
		os.Exit(1)
	}
}

func ExitSomethingWentWrong() {
	log.Fatalln("Error: Something went wrong. Please try again.")
}

func ValidationErrorsToString(j map[string]interface{}) string {
	if j == nil {
		return ""
	}
	msgs := []string{}
	if errs, ok := j["errors"].(map[string]interface{}); ok {
		for k, v := range errs {
			msgs = append(msgs, fmt.Sprintf("%s %s", Capitalize(k), v))
		}
	}
	return strings.Join(msgs, ", ")
}

func Capitalize(s string) string {
	if len(s) <= 1 {
		return strings.ToUpper(s)
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}
