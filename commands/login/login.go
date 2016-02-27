package login

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/client/oauth"
	"github.com/nitrous-io/rise-cli-go/client/users"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Login(c *cli.Context) {
	var (
		email string
		token string
	)

	fmt.Println("Enter your Rise credentials")
	for {
		var (
			err      error
			password string
			appErr   *apperror.Error
		)

		email, err = readline.Read("Enter Email: ", true)
		util.ExitIfError(err)

		password, err = readline.ReadSecurely("Enter Password: ", true)
		util.ExitIfError(err)

		token, appErr = oauth.FetchToken(email, password)

		if appErr != nil && appErr.Code == oauth.ErrCodeUnconfirmedEmail {
			fmt.Println("You have to confirm your email address to continue. Please check your inbox for the confirmation code.")
			for {
				confirmationCode, err := readline.Read("Enter Confirmation Code: ", true)
				util.ExitIfError(err)

				appErr := users.Confirm(email, confirmationCode)
				if appErr == nil {
					break
				}
				appErr.Handle()
			}

			fmt.Println("Thanks for confirming your email address! Your account is now active!")

			token, appErr = oauth.FetchToken(email, password)
		}

		if token != "" {
			break
		}

		if appErr != nil {
			appErr.Handle()
		} else {
			util.ExitSomethingWentWrong()
		}
	}

	config.AccessToken = token
	config.Save()
	fmt.Println("You are logged in as", email)
}
