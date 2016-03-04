package common

import (
	"fmt"
	"os"

	"github.com/nitrous-io/rise-cli-go/config"
)

func RequireAccessToken() string {
	token := config.AccessToken
	if token == "" {
		fmt.Fprintln(os.Stderr, `You are not logged in. Please login with "rise login" command or create a new account with "rise signup" command.`)
		os.Exit(1)
	}
	return token
}
