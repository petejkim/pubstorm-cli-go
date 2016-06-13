package configuration

import (
	"fmt"
	"strings"

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

func Update(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	tui.Println(fmt.Sprintf(tr.T("configure_project"), tui.Undl(tui.Bold(proj.Name))))

	for {
		optimize, err := readline.Read(tui.Bold(tr.T("optimized_project")+": "), true, "y")
		util.ExitIfErrorOrEOF(err)

		switch strings.ToLower(optimize) {
		case "yes", "y":
			proj.SkipBuild = false
			break
		case "no", "n":
			proj.SkipBuild = true
			break
		default:
			log.Error(tr.T("optimized_project_invalid_response"))
			continue
		}

		updatedProj, appErr := projects.Update(config.AccessToken, proj)
		if appErr != nil {
			appErr.Handle()
			continue
		}
		proj = updatedProj

		break
	}

	if proj.SkipBuild {
		log.Info(tr.T("project_disabled_optimizer"))
	} else {
		log.Info(tr.T("project_enabled_optimizer"))
	}

	tui.Println(fmt.Sprintf(tr.T("configuration_updated"), tui.Undl(tui.Bold(proj.Name))))
}
