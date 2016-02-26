package util

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/franela/goreq"
)

func ExitIfError(err error) {
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}

func HandleErrorResponse(res *goreq.Response) {
	sc := res.StatusCode
	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err == nil {
		e := j["error"]
		ed := j["error_description"]

		if e != nil {
			if ed != nil {
				log.Fatal(fmt.Sprintf("%d: %s - %s", sc, e, ed))
			}
			log.Fatal(fmt.Sprintf("%d: %s", sc, e))
		}
	} else if str, err := res.Body.ToString(); err == nil {
		log.Fatal(fmt.Sprintf("%d: %s", sc, str))
	}

	log.Fatal(fmt.Sprintf("%d: Something went wrong. Please try again.", sc))
}

func SomethingWentWrong(err error) {
	log.Fatal("Something went wrong. Please try again.\n", err)
}

func CouldNotMakeRequest(err error) {
	log.Fatal("Failed to make request to Rise server. Please check your Internet connection and try again.", err)
}
