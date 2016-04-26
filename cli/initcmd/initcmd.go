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
	}

	// Check for existence of a pubstorm.default.json file. If it's present, use
	// its values as defaults.
	proj, err = project.LoadDefault()
	if err != nil {
		common.DebugLog().Warnf("error trying to load default project config, err: %v", err)
		proj = &project.Project{}
	}

	tui.Println(tr.T("init_rise_project"))

	for {
		defaultPath := "."
		if proj.Path != "" {
			defaultPath = proj.Path
		}
		proj.Path, err = readline.Read(tui.Bold(tr.T("enter_project_path")+": ["+defaultPath+"] "), true, defaultPath)
		util.ExitIfErrorOrEOF(err)

		if err := proj.ValidatePath(); err != nil {
			if err == project.ErrPathNotExist {
				if err := os.MkdirAll(proj.Path, 0700); err != nil {
					log.Infof(tr.T("project_path_create_failed"), proj.Path)
					common.DebugLog().Warnf("failed to create project directory at %q, err: %v", proj.Path, err)
					continue
				}

				log.Infof(tr.T("project_path_create_ok"), proj.Path)
			} else {
				log.Error(err.Error())
				continue
			}
		}

		break
	}

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
