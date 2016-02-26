package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/auth"
)

func main() {
	app := cli.NewApp()
	app.Name = "rise"
	app.Usage = "Command line interface for Rise.sh"

	app.Commands = []cli.Command{
		{
			Name:   "signup",
			Usage:  "Create a new Rise account",
			Action: auth.Signup,
		},
	}

	app.Run(os.Args)
}
