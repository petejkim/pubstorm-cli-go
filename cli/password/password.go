package password

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/oauth"
	"github.com/nitrous-io/rise-cli-go/client/users"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func Change(c *cli.Context) {
	common.RequireAccessToken()

	var (
		err              error
		existingPassword string
		password         string
	)

	log.Info(tr.T("will_invalidate_session"))

	for {
		existingPassword, err = readline.ReadSecurely(tui.Bold(tr.T("enter_existing_password")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		var passwordConf string

		readPw := func() {
			password, err = readline.ReadSecurely(tui.Bold(tr.T("enter_password")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			passwordConf, err = readline.ReadSecurely(tui.Bold(tr.T("confirm_password")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		readPw()
		for password != passwordConf {
			log.Error(tr.T("password_no_match"))
			readPw()
		}

		appErr := users.ChangePassword(config.AccessToken, existingPassword, password)
		if appErr == nil {
			break
		}
		appErr.Handle()
		tui.Println(tr.T("error_in_input"))
	}

	log.Infof(tr.T("password_changed"))

	// If email is empty, ask users to input again
	email := config.Email
	if config.Email == "" {
		tui.Println(tr.T("reenter_email"))
		email, err = readline.Read(tui.Bold(tr.T("enter_email")+": "), true, "")
		util.ExitIfErrorOrEOF(err)
	}

	token, appErr := oauth.FetchToken(email, password)
	if token == "" {
		log.Error(tr.T("login_fail"))

		if appErr != nil {
			log.Fatal(appErr.Error())
		}
	}

	config.Email = email
	config.AccessToken = token
	config.Save()
	log.Infof(tr.T("login_success"), config.Email)
}
