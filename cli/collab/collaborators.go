package collab

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/collab"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func List(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	collaborators, appErr := collab.List(config.AccessToken, proj.Name)
	if appErr != nil {
		if appErr.Code == collab.ErrCodeNotFound {
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		}
		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("collab_list_header")))+"\n", proj.Name)
	for _, collab := range collaborators {
		tui.Println("- " + collab.Email)
	}
}

func Add(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	collabEmail := c.Args().First()

	var err error
	interactive := collabEmail == ""

	for {
		if interactive {
			collabEmail, err = readline.Read(tui.Bold(tr.T("collab_enter_email_to_add")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		appErr := collab.Add(config.AccessToken, proj.Name, collabEmail)
		if appErr != nil {
			switch appErr.Code {
			case collab.ErrCodeAlreadyExists:
				if interactive {
					log.Error(appErr.Description)
					continue
				} else {
					log.Fatal(appErr.Description)
				}
			case collab.ErrCodeNotFound:
				log.Fatalf(tr.T("project_not_found"), proj.Name)
			case collab.ErrCodeUserNotFound:
				log.Fatalf(tr.T("collab_user_not_found"), tui.Undl(collabEmail))
			case collab.ErrCodeCannotAddOwner:
				log.Fatalf(tr.T("collab_cannot_add_owner"))
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("collab_added_success"), tui.Undl(collabEmail), proj.Name)
	tui.Println()
}

func Remove(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	collabEmail := c.Args().First()

	var err error
	interactive := collabEmail == ""

	for {
		if interactive {
			collabEmail, err = readline.Read(tui.Bold(tr.T("collab_enter_email_to_rm")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		appErr := collab.Remove(config.AccessToken, proj.Name, collabEmail)
		if appErr != nil {
			switch appErr.Code {
			case collab.ErrCodeNotFound:
				log.Fatalf(tr.T("project_not_found"), proj.Name)
			case collab.ErrCodeUserNotFound:
				log.Fatalf(tr.T("collab_user_not_found"), tui.Undl(collabEmail))
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("collab_removed_success"), tui.Undl(collabEmail), proj.Name)
	tui.Println()
}
