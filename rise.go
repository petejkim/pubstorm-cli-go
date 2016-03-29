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
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tui"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.SetFormatter(&tui.Formatter{})
	log.SetOutput(tui.Out)
	log.SetLevel(log.InfoLevel)
	readline.Output = tui.Out

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
		{
			Name:   "domains.add",
			Usage:  "Add a new domain to a Rise project",
			Action: domains.Add,
		},
		{
			Name:   "domains.rm",
			Usage:  "Remove a domain from a Rise project",
			Action: domains.Remove,
		},
	}

	app.Run(os.Args)
}
