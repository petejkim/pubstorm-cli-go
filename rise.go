package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/deploy"
	"github.com/nitrous-io/rise-cli-go/cli/domains"
	"github.com/nitrous-io/rise-cli-go/cli/initcmd"
	"github.com/nitrous-io/rise-cli-go/cli/login"
	"github.com/nitrous-io/rise-cli-go/cli/logout"
	"github.com/nitrous-io/rise-cli-go/cli/signup"
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
			Name:   "init",
			Usage:  "Create a new Rise project",
			Action: initcmd.Init,
		},
		{
			Name:   "deploy",
			Usage:  "Deploy a Rise project",
			Action: deploy.Deploy,
		},
		{
			Name:   "domains",
			Usage:  "List all domains for a Rise project",
			Action: domains.List,
		},
	}

	app.Run(os.Args)
}
