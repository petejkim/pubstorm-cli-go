package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
		util.CouldNotMakeRequest(err)
	}

	if res.StatusCode != http.StatusCreated && res.StatusCode != 422 {
		util.HandleErrorResponse(res)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		util.SomethingWentWrong(err)
	}

	if res.StatusCode == 422 && j["error"] == "invalid_params" {
		fmt.Println("There were errors in your input. Please try again")
		if errs, ok := j["errors"].(map[string]interface{}); ok {
			for k, v := range errs {
				fmt.Println("*", k, v)
			}
		}
	} else if res.StatusCode == http.StatusCreated {
		fmt.Println("Your account has been created. You will receive your confirmation code shortly.")

		for {
			confirmationCode, err := readline.Read("Enter Confirmation Code (Check your inbox!): ")
			util.ExitIfError(err)

			res, err := goreq.Request{
				Method:      "POST",
				Uri:         config.Host + "/user/confirm",
				ContentType: "application/x-www-form-urlencoded",

				Body: url.Values{
					"email":             {email},
					"confirmation_code": {confirmationCode},
				}.Encode(),
			}.Do()

			if err != nil {
				util.CouldNotMakeRequest(err)
			}

			if res.StatusCode != http.StatusOK && res.StatusCode != 422 {
				util.HandleErrorResponse(res)
			}

			if res.StatusCode == 422 {
				resText, err := res.Body.ToString()
				if err != nil {
					util.SomethingWentWrong(err)
				}
				if strings.Contains(resText, "invalid email or confirmation_code") {
					fmt.Println("You've entered an incorrect confirmation code. Please try again.")
				}
			} else if res.StatusCode == http.StatusOK {
				fmt.Println("Thanks for confirming your email address! Your account is now active!")
				return
			}
		}
	}
}
