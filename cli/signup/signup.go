package signup

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

func Signup(c *cli.Context) {
	var (
		err      error
		email    string
		password string
	)

	common.PrintLogo()
	tui.Println(tui.Bold(tr.T("join_rise")) + "\n")
	tui.Println(tr.T("signup_disclaimer") + "\n")
	tui.Println("  * " + tr.T("rise_tos") + " - " + tui.Undl(tui.Blu(config.WebsiteHost+"/terms-of-service")))
	tui.Println("  * " + tr.T("rise_privacy_policy") + " - " + tui.Undl(tui.Blu(config.WebsiteHost+"/privacy-policy")) + "\n")

	for {
		email, err = readline.Read(tui.Bold(tr.T("enter_email")+": "), true, "")
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

		appErr := users.Create(email, password)
		if appErr == nil {
			break
		}
		appErr.Handle()
		tui.Println(tr.T("error_in_input"))
	}

	log.Info(tr.T("account_created"))
	tui.Println()

	for {
		confirmationCode, err := readline.Read(tui.Bold(tr.T("enter_confirmation")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		appErr := users.Confirm(email, confirmationCode)
		if appErr == nil {
			break
		}
		appErr.Handle()
	}

	log.Info(tr.T("confirmation_sucess"))
	tui.Println()

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
	log.Infof(tr.T("login_success"), email)
}
