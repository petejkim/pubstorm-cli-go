package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func ExitIfError(err error) {
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}

func ValidationErrorsToString(j map[string]interface{}) string {
	msgs := []string{}
	if errs, ok := j["errors"].(map[string]interface{}); ok {
		for k, v := range errs {
			msgs = append(msgs, fmt.Sprintf("* %s %s", k, v))
		}
	}
	return strings.Join(msgs, "\n")
}
