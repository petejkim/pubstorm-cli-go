package logout

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/client/oauth"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Logout(c *cli.Context) {
	token := config.AccessToken
	if token == "" {
		fmt.Fprintln(os.Stderr, "Access token not found!")
		os.Exit(1)
	}

	if appErr := oauth.InvalidateToken(token); appErr != nil {
		appErr.Handle()
	} else {
		fmt.Println("You have successfully logged out.")
	}

	config.AccessToken = ""
	err := config.Save()
	util.ExitIfError(err)

	fmt.Println("Access token cleared.")
}
