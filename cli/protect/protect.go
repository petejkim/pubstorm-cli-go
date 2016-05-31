package protect

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func Protect(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var (
		username, password string
		err                error
	)

	if len(c.Args()) < 1 {
		username, err = readline.Read(tui.Bold(tr.T("enter_basic_auth_username")+": "), true, "")
		util.ExitIfErrorOrEOF(err)
	} else {
		username = c.Args().Get(0)
	}

	if len(c.Args()) < 2 {
		password, err = readline.ReadSecurely(tui.Bold(tr.T("enter_basic_auth_password")+": "), true, "")
		util.ExitIfErrorOrEOF(err)
	} else {
		password = c.Args().Get(1)
	}

	appErr := projects.Protect(token, proj.Name, username, password)
	if appErr != nil {
		appErr.Handle()
	}

	log.Infof(tr.T("protect_success"), proj.Name)
}
