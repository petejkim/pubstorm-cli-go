package projects

import (
	"fmt"

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
	force := c.Bool("force")

	common.RequireAccessToken()
	proj := common.RequireProject()

	if !force {
		log.Warnf(tui.Undl(tui.Bold(tr.T("project_rm_cannot_undo")))+" "+tr.T("project_rm_permanent_delete"), proj.Name)
		for {
			projectName, err := readline.Read(tui.Bold(fmt.Sprintf(tr.T("enter_project_name_to_confirm"), proj.Name)+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if projectName != proj.Name {
				log.Error(tr.T("project_name_does_not_match"))
				continue
			}

			break
		}
	}

	if appErr := projects.Delete(config.AccessToken, proj.Name); appErr != nil {
		appErr.Handle()
	}

	if err := proj.Delete(); err != nil {
		log.Fatal(tr.T("project_json_failed_to_delete"))
	}

	log.Infof(tr.T("project_rm_success"), proj.Name)
}
