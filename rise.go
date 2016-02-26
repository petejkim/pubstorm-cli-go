package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rise"
	app.Usage = "Command line interface for Rise.sh"

	app.Run(os.Args)
}
