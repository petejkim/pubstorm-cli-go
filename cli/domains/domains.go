package domains

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/domains"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func List(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	domainNames, appErr := domains.Index(config.AccessToken, proj.Name)
	if appErr != nil {
		if appErr.Code == projects.ErrCodeNotFound {
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		}
		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("domain_list")))+"\n", proj.Name)
	for _, domainName := range domainNames {
		tui.Println("- " + domainName)
	}
}

func Add(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	domainName := strings.TrimSpace(c.Args().First())

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
			} else if appErr.Code == domains.ErrCodeNotFound {
				log.Fatalf(tr.T("project_not_found"), proj.Name)
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	log.Infof(tr.T("domain_added"), domainName, proj.Name)
	tui.Println()

	subDn, Dn := util.SplitDomain(domainName)
	riseDn := fmt.Sprintf("%s.%s", proj.Name, config.DefaultDomain)

	dns_inst := fmt.Sprintf(tr.T("dns_instructions"), Dn) + "\n\n"
	dns_inst += fmt.Sprintf("  * %s: %s ---> %s", tui.Bold("CNAME (Alias)"), tui.Undl(subDn), tui.Undl(riseDn))
	if subDn == "www" {
		dns_inst += fmt.Sprintf("\n  * %s: %s ---> %s", tui.Bold("A (Host)"), tui.Undl("@"), tui.Undl(config.RedirectorIP))
	}
	log.Info(dns_inst)
	tui.Println()
	log.Infof(tr.T("dns_more_info"), tui.Undl(tui.Blu(config.DNSHelpURL)))
}

func Remove(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	domainName := strings.TrimSpace(c.Args().First())

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
				if strings.Contains(appErr.Description, "project") {
					log.Fatalf(tr.T("project_not_found"), proj.Name)
				}
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
