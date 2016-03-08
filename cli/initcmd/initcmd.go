package initcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Init(c *cli.Context) {
	common.RequireAccessToken()

	proj, err := project.Load()
	if err != nil && !os.IsNotExist(err) {
		util.ExitIfError(err)
	}

	if proj != nil {
		fmt.Println("Error: A Rise project already exists in current path, aborting.")
		os.Exit(1)
	} else {
		proj = &project.Project{}

		fmt.Println("Set up your Rise project")

		for {
			proj.Path, err = readline.Read("Enter Project Path: [.] ", true, ".")
			util.ExitIfErrorOrEOF(err)

			if err := proj.ValidatePath(); err != nil {
				fmt.Println(err.Error())
				continue
			}

			break
		}

		for {
			enableStats, err := readline.Read("Enable Basic Stats? [Y/n] ", true, "y")
			util.ExitIfErrorOrEOF(err)

			switch strings.ToLower(enableStats) {
			case "y", "yes":
				proj.EnableStats = true
			case "n", "no":
				proj.EnableStats = false
			default:
				continue
			}

			break
		}

		for {
			forceHTTPS, err := readline.Read("Force HTTPS? [y/N] ", true, "n")
			util.ExitIfErrorOrEOF(err)

			switch strings.ToLower(forceHTTPS) {
			case "y", "yes":
				proj.ForceHTTPS = true
			case "n", "no":
				proj.ForceHTTPS = false
			default:
				continue
			}

			break
		}

		for {
			proj.Name, err = readline.Read("Enter Project Name: ", true, "")
			util.ExitIfErrorOrEOF(err)

			if err := proj.ValidateName(); err != nil {
				fmt.Println(err.Error())
				continue
			}

			appErr := projects.Create(config.AccessToken, proj.Name)
			if appErr != nil {
				appErr.Handle()
				continue
			}

			break
		}

		err = proj.Save()
		util.ExitIfError(err)
	}
}
