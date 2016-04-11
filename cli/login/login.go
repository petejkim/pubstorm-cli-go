package login

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/apperror"
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

func Login(c *cli.Context) {
	var (
		email string
		token string
	)

	common.PrintLogo()
	tui.Println(tui.Bold(tr.T("login_rise")) + "\n")
	tui.Println(tr.T("enter_credentials"))
	for {
		tui.Println()
		var (
			err      error
			password string
			appErr   *apperror.Error
		)

		email, err = readline.Read(tui.Bold(tr.T("enter_email")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		password, err = readline.ReadSecurely(tui.Bold(tr.T("enter_password")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		token, appErr = oauth.FetchToken(email, password)

		if appErr != nil && appErr.Code == oauth.ErrCodeUnconfirmedEmail {
			log.Info(tr.T("confirmation_required"))

			resendUsed := false
			for {
				tui.Println()
				var prompt string
				if resendUsed {
					prompt = tr.T("enter_confirmation")
				} else {
					prompt = tr.T("enter_confirmation_resend")
				}
				confirmationCode, err := readline.Read(tui.Bold(prompt+": "), true, "")
				util.ExitIfErrorOrEOF(err)

				if confirmationCode == "resend" && !resendUsed {
					appErr = users.ResendConfirmationCode(email)
					if appErr == nil {
						resendUsed = true
						log.Info(tr.T("confirmation_resent"))
						continue
					}
				} else {
					appErr = users.Confirm(email, confirmationCode)
					if appErr == nil {
						break
					}
				}

				appErr.Handle()
			}
			log.Info(tr.T("confirmation_success"))
			tui.Println()

			token, appErr = oauth.FetchToken(email, password)
		}

		if token != "" {
			break
		}

		if appErr != nil {
			appErr.Handle()
		} else {
			util.ExitSomethingWentWrong()
		}
	}

	config.Email = email
	config.AccessToken = token
	if err := config.Save(); err != nil {
		log.Fatal(tr.T("rise_config_write_failed"))
	}
	log.Infof(tr.T("login_success"), email)
}
