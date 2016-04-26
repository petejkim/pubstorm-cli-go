package rollback

import (
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/tr"

	log "github.com/Sirupsen/logrus"
)

func Rollback(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

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

	var d *deployments.Deployment

	d, appErr := deployments.Rollback(config.AccessToken, proj.Name, ver)
	if appErr != nil {
		appErr.Handle()
	}

	log.Infof(tr.T("rollback_success"), d.Version)
}
