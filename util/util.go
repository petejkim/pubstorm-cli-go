package util

import (
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Prints and exits if err is not nil
func ExitIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Exits if err is not nil. If err is not an EOF error, prints error message
func ExitIfErrorOrEOF(err error) {
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}

func ExitSomethingWentWrong() {
	log.Fatal("Something went wrong. Please try again.")
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

func FindInt(list []int, what int) int {
	for i, v := range list {
		if v == what {
			return i
		}
	}
	return -1
}

func ContainsInt(list []int, what int) bool {
	return FindInt(list, what) != -1
}

func SanitizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	labels := strings.Split(domain, ".")
	if len(labels) == 2 {
		domain = "www." + domain
	}
	return domain
}
