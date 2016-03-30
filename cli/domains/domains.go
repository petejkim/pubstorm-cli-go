package domains

import (
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/domains"
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

	domainNames, appErr := domains.Index(config.AccessToken, proj.Name)
	if appErr != nil {
		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("domains_for")))+"\n", proj.Name)
	for _, domainName := range domainNames {
		tui.Println(domainName)
	}
}

func Add(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	domainName := c.Args().First()

	var err error
	interactive := domainName == ""

	for {
		if interactive {
			domainName, err = readline.Read(tui.Bold(tr.T("enter_domain_name_to_add")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		domainName = util.SanitizeDomain(domainName)

		appErr := domains.Create(config.AccessToken, proj.Name, domainName)
		if appErr != nil {
			if appErr.Code == domains.ErrCodeValidationFailed {
				if interactive {
					log.Error(appErr.Description)
					continue
				} else {
					log.Fatal(appErr.Description)
				}
			} else if appErr.Code == domains.ErrCodeLimitReached {
				log.Fatalf(tr.T("domain_limit_reached"), proj.Name)
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("domain_added"), domainName, proj.Name)
}

func Remove(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	domainName := c.Args().First()

	var err error
	interactive := domainName == ""

	for {
		if interactive {
			domainName, err = readline.Read(tui.Bold(tr.T("enter_domain_name_to_remove")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		domainName = util.SanitizeDomain(domainName)

		if domainName == proj.Name+"."+config.DefaultDomain {
			if interactive {
				log.Errorf(tr.T("domain_cannot_be_deleted"), domainName)
				continue
			} else {
				log.Fatalf(tr.T("domain_cannot_be_deleted"), domainName)
			}
		}

		appErr := domains.Delete(config.AccessToken, proj.Name, domainName)
		if appErr != nil {
			if appErr.Code == domains.ErrCodeNotFound {
				if interactive {
					log.Errorf(tr.T("domain_not_found"), domainName)
					continue
				} else {
					log.Fatalf(tr.T("domain_not_found"), domainName)
				}
			}

			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("domain_removed"), domainName, proj.Name)
}
