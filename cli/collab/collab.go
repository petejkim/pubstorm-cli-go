package collab

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/collaborators"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func List(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	cols, appErr := collaborators.List(token, proj.Name)
	if appErr != nil {
		if appErr.Code == collaborators.ErrCodeNotFound {
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		}
		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("collab_list_header")))+"\n", proj.Name)
	for _, col := range cols {
		tui.Println("- " + col.Email)
	}
}

func Add(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	collabEmail := c.Args().First()

	var err error
	interactive := collabEmail == ""

	for {
		if interactive {
			collabEmail, err = readline.Read(tui.Bold(tr.T("collab_enter_email_to_add")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		appErr := collaborators.Add(token, proj.Name, collabEmail)
		if appErr != nil {
			switch appErr.Code {
			case collaborators.ErrCodeAlreadyExists:
				if interactive {
					log.Error(appErr.Description)
					continue
				} else {
					log.Fatal(appErr.Description)
				}
			case collaborators.ErrCodeNotFound:
				log.Fatalf(tr.T("project_not_found"), proj.Name)
			case collaborators.ErrCodeUserNotFound:
				log.Fatalf(tr.T("collab_add_user_not_found"), collabEmail)
			case collaborators.ErrCodeCannotAddOwner:
				log.Fatalf(tr.T("collab_cannot_add_owner"))
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("collab_added_success"), collabEmail, proj.Name)
	tui.Println()
}

func Remove(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	collabEmail := c.Args().First()

	var err error
	interactive := collabEmail == ""

	for {
		if interactive {
			collabEmail, err = readline.Read(tui.Bold(tr.T("collab_enter_email_to_rm")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		appErr := collaborators.Remove(token, proj.Name, collabEmail)
		if appErr != nil {
			switch appErr.Code {
			case collaborators.ErrCodeNotFound:
				log.Fatalf(tr.T("project_not_found"), proj.Name)
			case collaborators.ErrCodeUserNotFound:
				log.Fatalf(tr.T("collab_rm_user_not_found"), collabEmail, proj.Name)
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("collab_removed_success"), collabEmail, proj.Name)
	tui.Println()
}
