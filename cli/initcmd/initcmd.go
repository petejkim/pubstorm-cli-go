package initcmd

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func Init(c *cli.Context) {
	common.RequireAccessToken()

	proj, err := project.Load()
	if err != nil && !os.IsNotExist(err) {
		util.ExitIfError(err)
	}

	if proj != nil {
		log.Fatal(tr.T("existing_rise_project"))
	} else {
		proj = &project.Project{}

		tui.Println(tr.T("init_rise_project"))

		for {
			proj.Path, err = readline.Read(tui.Bold(tr.T("enter_project_path")+": [.] "), true, ".")
			util.ExitIfErrorOrEOF(err)

			if err := proj.ValidatePath(); err != nil {
				log.Error(err.Error())
				continue
			}

			break
		}

		/*for {
			enableStats, err := readline.Read(tui.Bold(tr.T("enable_basic_stats")+"? [Y/n] "), true, "y")
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
			forceHTTPS, err := readline.Read(tui.Bold(tr.T("force_https")+"? [y/N] "), true, "n")
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
		}*/
		proj.EnableStats = true
		proj.ForceHTTPS = false

		for {
			proj.Name, err = readline.Read(tui.Bold(tr.T("enter_project_name")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if err := proj.ValidateName(); err != nil {
				log.Error(err.Error())
				continue
			}

			appErr := projects.Create(config.AccessToken, proj.Name)
			if appErr != nil {
				appErr.Handle()
				continue
			}

			break
		}

		log.Infof(tr.T("project_initialized"), proj.Name)
		if err = proj.Save(); err != nil {
			log.Fatal(err.Error())
		}
		log.Info(tr.T("rise_json_saved"))
	}
}
