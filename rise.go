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
	"github.com/nitrous-io/rise-cli-go/tr"
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
	app.Usage = tr.T("rise_cli_desc")

	app.Commands = []cli.Command{
		{
			Name:   "signup",
			Usage:  tr.T("signup_desc"),
			Action: signup.Signup,
		},
		{
			Name:   "login",
			Usage:  tr.T("login_desc"),
			Action: login.Login,
		},
		{
			Name:   "logout",
			Usage:  tr.T("logout_desc"),
			Action: logout.Logout,
		},
		{
			Name:   "init",
			Usage:  tr.T("init_desc"),
			Action: initcmd.Init,
		},
		{
			Name:    "publish",
			Aliases: []string{"deploy"},
			Usage:   tr.T("deploy_desc"),
			Action:  deploy.Deploy,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "verbose, v",
					Usage: "Show additional information",
				},
			},
		},
		{
			Name:   "domains",
			Usage:  tr.T("domains_desc"),
			Action: domains.List,
		},
		{
			Name:   "domains.add",
			Usage:  tr.T("domains_add_desc"),
			Action: domains.Add,
		},
		{
			Name:   "domains.rm",
			Usage:  tr.T("domains_rm_desc"),
			Action: domains.Remove,
		},
	}

	app.Run(os.Args)
}
