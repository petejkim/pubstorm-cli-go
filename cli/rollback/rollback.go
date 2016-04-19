package rollback

import (
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/spinner"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"

	log "github.com/Sirupsen/logrus"
)

func Rollback(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var (
		verStr = c.Args().First()
		ver    int64
		err    error
	)

	if len(verStr) > 1 {
		ver, err = strconv.ParseInt(verStr[1:], 10, 64)
	}

	if verStr != "" && (verStr[0] != 'v' || err != nil) {
		log.Error(tr.T("rollback_invalid_version"))
		return
	}

	deployment, appErr := deployments.Rollback(config.AccessToken, proj.Name, ver)
	if appErr != nil {
		appErr.Handle()
	}

	spin := spinner.New()
	tui.Printf(tr.T("launching")+" "+tui.Blu("%s"), deployment.Version, string(spin.Next()))

	for deployment.State != deployments.DeploymentStateDeployed {
		for i := 0; i < 5; i++ {
			time.Sleep(100 * time.Millisecond)
			tui.Printf(tui.Blu("\b%s"), string(spin.Next()))
		}

		deployment, appErr = deployments.Get(config.AccessToken, proj.Name, deployment.ID)
		if appErr != nil {
			appErr.Handle()
		}
	}

	tui.Println("\b \b")
	log.Infof(tr.T("rollback_success"), proj.Name, deployment.Version)
}
