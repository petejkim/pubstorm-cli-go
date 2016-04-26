package main

import (
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/franela/goreq"

	"github.com/nitrous-io/rise-cli-go/cli/collab"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/cli/deploy"
	"github.com/nitrous-io/rise-cli-go/cli/domains"
	"github.com/nitrous-io/rise-cli-go/cli/initcmd"
	"github.com/nitrous-io/rise-cli-go/cli/login"
	"github.com/nitrous-io/rise-cli-go/cli/logout"
	"github.com/nitrous-io/rise-cli-go/cli/password"
	"github.com/nitrous-io/rise-cli-go/cli/projects"
	"github.com/nitrous-io/rise-cli-go/cli/rollback"
	"github.com/nitrous-io/rise-cli-go/cli/signup"
	"github.com/nitrous-io/rise-cli-go/cli/versions"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"

	log "github.com/Sirupsen/logrus"
)

func init() {
	cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} {{if .Flags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}
   {{if .Version}}
VERSION:
   {{.Version}}
   {{end}}{{if len .Authors}}
AUTHOR(S):
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
COMMANDS:
   {{range .Commands}}{{if .Usage}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}{{end}}{{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}
`
}

func main() {
	log.SetFormatter(&tui.Formatter{})
	log.SetOutput(tui.Out)
	log.SetLevel(log.InfoLevel)
	readline.Output = tui.Out

	common.CheckForUpdates()

	// Set Goreq's connect timeout to 10s globally (its default is 1s which
	// can be too short).
	goreq.SetConnectTimeout(10 * time.Second)

	app := cli.NewApp()
	app.Name = config.AppName
	app.Version = config.Version
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
			Name: "password",
			Subcommands: []cli.Command{
				{
					Name:   "change",
					Usage:  tr.T("password_change_desc"),
					Action: password.Change,
				},
				{
					Name:   "reset",
					Usage:  tr.T("password_reset_desc"),
					Action: password.Reset,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "continue, c",
							Usage: "Use this flag if you already have a password reset token",
						},
					},
				},
			},
		},
		{
			Name:   "password.change",
			Usage:  tr.T("password_change_desc"),
			Action: password.Change,
		},
		{
			Name:   "password.reset",
			Usage:  tr.T("password_reset_desc"),
			Action: password.Reset,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "continue, c",
					Usage: "Use this flag if you already have a password reset token",
				},
			},
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
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  tr.T("domains_add_desc"),
					Action: domains.Add,
				},
				{
					Name:   "rm",
					Usage:  tr.T("domains_rm_desc"),
					Action: domains.Remove,
				},
			},
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
		{
			Name:   "projects",
			Usage:  tr.T("projects_desc"),
			Action: projects.List,
			Subcommands: []cli.Command{
				{
					Name:   "rm",
					Usage:  tr.T("projects_add_desc"),
					Action: projects.Remove,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Delete project without confirmation",
						},
					},
				},
			},
		},
		{
			Name:   "projects.rm",
			Usage:  tr.T("projects_rm_desc"),
			Action: projects.Remove,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "Delete project without confirmation",
				},
			},
		},
		{
			Name:   "collab",
			Usage:  "Lists collaborators for the current project",
			Action: collab.List,
			Subcommands: []cli.Command{
				{
					Name:   "add",
					Usage:  "Add a collaborator to the current project",
					Action: collab.Add,
				},
				{
					Name:   "rm",
					Usage:  "Remove a collaborator from the current project",
					Action: collab.Remove,
				},
			},
		},
		{
			Name:   "collab.add",
			Usage:  "Add a collaborator to the current project",
			Action: collab.Add,
		},
		{
			Name:   "collab.rm",
			Usage:  "Remove a collaborator from the current project",
			Action: collab.Remove,
		},
		{
			Name:      "rollback",
			Usage:     tr.T("rollback_desc"),
			ArgsUsage: "[RELEASE]",
			Action:    rollback.Rollback,
		},
		{
			Name:   "versions",
			Usage:  tr.T("versions_desc"),
			Action: versions.Versions,
		},
	}

	app.Run(os.Args)
}
