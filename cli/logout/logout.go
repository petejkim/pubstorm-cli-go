package logout

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/oauth"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/tr"

	log "github.com/Sirupsen/logrus"
)

func Logout(c *cli.Context) {
	token := common.RequireAccessToken()

	if appErr := oauth.InvalidateToken(token); appErr != nil {
		appErr.Handle()
	} else {
		log.Info(tr.T("logout_success"))
	}

	config.AccessToken = ""
	if err := config.Save(); err != nil {
		log.Fatalln(tr.T("rise_config_write_failed"))
	}
	log.Info(tr.T("access_token_cleared"))
}
