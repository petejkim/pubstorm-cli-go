package util

import (
	"io"
	"log"
	"os"
)

func ExitIfError(err error) {
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}
