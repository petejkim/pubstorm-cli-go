package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/commands/deploy"
	"github.com/nitrous-io/rise-cli-go/commands/login"
	"github.com/nitrous-io/rise-cli-go/commands/logout"
	"github.com/nitrous-io/rise-cli-go/commands/signup"
)

func main() {
	app := cli.NewApp()
	app.Name = "rise"
	app.Usage = "Command line interface for Rise.sh"

	app.Commands = []cli.Command{
		{
			Name:   "signup",
			Usage:  "Create a new Rise account",
			Action: signup.Signup,
		},
		{
			Name:   "login",
			Usage:  "Log in to a Rise account",
			Action: login.Login,
		},
		{
			Name:   "logout",
			Usage:  "Log out from current session",
			Action: logout.Logout,
		},
		{
			Name:   "deploy",
			Usage:  "Deploy your project",
			Action: deploy.Deploy,
		},
	}

	app.Run(os.Args)
}
