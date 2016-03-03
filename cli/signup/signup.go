package signup

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/client/oauth"
	"github.com/nitrous-io/rise-cli-go/client/users"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Signup(c *cli.Context) {
	var (
		err      error
		email    string
		password string
	)

	fmt.Println("Create a Rise account")
	for {
		email, err = readline.Read("Enter Email: ", true, "")
		util.ExitIfErrorOrEOF(err)

		var passwordConf string

		readPw := func() {
			password, err = readline.ReadSecurely("Enter Password: ", true, "")
			util.ExitIfErrorOrEOF(err)

			passwordConf, err = readline.ReadSecurely("Confirm Password: ", true, "")
			util.ExitIfErrorOrEOF(err)
		}

		readPw()
		for password != passwordConf {
			fmt.Println("Passwords do not match. Please re-enter password.")
			readPw()
		}

		appErr := users.Create(email, password)
		if appErr == nil {
			break
		}
		appErr.Handle()
		fmt.Println("There were errors in your input. Please try again.")
	}

	fmt.Println("Your account has been created. You will receive your confirmation code shortly.")

	for {
		confirmationCode, err := readline.Read("Enter Confirmation Code (check your inbox): ", true, "")
		util.ExitIfErrorOrEOF(err)

		appErr := users.Confirm(email, confirmationCode)
		if appErr == nil {
			break
		}
		appErr.Handle()
	}

	fmt.Println("Thanks for confirming your email address! Your account is now active!")

	token, appErr := oauth.FetchToken(email, password)
	if token == "" {
		fmt.Println("Error: Could not login to Rise. Use `rise login` command to try again.")

		if appErr != nil {
			log.Fatalln(appErr.Error())
		}
	}

	config.AccessToken = token
	config.Save()
	fmt.Println("You are logged in as", email)
}
