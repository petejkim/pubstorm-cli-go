package projects

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func List(c *cli.Context) {
	common.RequireAccessToken()

	projs, appErr := projects.Index(config.AccessToken)
	if appErr != nil {
		appErr.Handle()
	}

	if len(projs) == 0 {
		log.Info(tr.T("no_project"))
		return
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("project_list"))) + "\n")
	for _, proj := range projs {
		tui.Println("- " + proj.Name)
	}
}

func Remove(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	log.Warnf(tr.T("project_destroy_all"), proj.Name, proj.Name)
	for {
		projectName, err := readline.Read(tui.Bold(tr.T("enter")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		if projectName != proj.Name {
			log.Warn(tr.T("project_name_not_match"))
			continue
		}

		break
	}

	if appErr := projects.Delete(config.AccessToken, proj.Name); appErr != nil {
		appErr.Handle()
	}

	if err := proj.Delete(); err != nil {
		log.Fatal(tr.T("project_json_failed_to_delete"))
	}

	tui.Printf(tr.T("project_delete_success")+"\n", proj.Name)
}
