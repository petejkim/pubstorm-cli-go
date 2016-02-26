package auth

import (
	"fmt"
	"log"
	"net/url"

	"github.com/codegangsta/cli"
	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Signup(c *cli.Context) {
	email, err := readline.Read("Enter Email: ")
	util.ExitIfError(err)

	var password, passwordConf string

	readPw := func() {
		var err error
		password, err = readline.ReadSecurely("Enter Password: ")
		util.ExitIfError(err)

		passwordConf, err = readline.ReadSecurely("Confirm Password: ")
		util.ExitIfError(err)
	}

	readPw()
	for password != passwordConf {
		fmt.Println("Passwords do not match. Please re-enter password.")
		readPw()
	}

	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/users",
		ContentType: "application/x-www-form-urlencoded",

		Body: url.Values{
			"email":    {email},
			"password": {password},
		}.Encode(),
	}.Do()

	if err != nil {
		log.Fatal(err)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		log.Fatal(err)
	}

	if res.StatusCode == 422 && j["error"] == "invalid_params" {
		fmt.Println("There were errors in your input. Please try again")
		if errs, ok := j["errors"].(map[string]interface{}); ok {
			for k, v := range errs {
				fmt.Println("*", k, v)
			}
		}
	}
}
