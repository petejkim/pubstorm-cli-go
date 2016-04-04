package projects

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"

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
