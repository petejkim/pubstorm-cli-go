package main

import (
	"os"

	"github.com/codegangsta/cli"
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
	}

	app.Run(os.Args)
}
