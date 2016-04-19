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

// If the user types in this value, they're referring to the default domain.
const defaultDomainArg = "default"

func List(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	domainNames, appErr := domains.Index(token, proj.Name)
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

		if isDefaultDomain(domainName, proj) {
			// If default domain is already enabled, there's nothing to do.
			if proj.DefaultDomainEnabled {
				log.Infof(tr.T("default_domain_already_added"),
					tui.Undl(tui.Blu(proj.DefaultDomain())))
				return
			}

			// Enable the default domain.
			proj.DefaultDomainEnabled = true
			updatedProj, appErr := projects.Update(token, proj)
			if appErr != nil {
				appErr.Handle()
			}

			proj = updatedProj

			log.Infof("Successfully enabled default domain %s.",
				tui.Undl(tui.Blu(proj.DefaultDomain())))

			return
		}

		// Add a "regular" custom domain.
		appErr := domains.Create(token, proj.Name, domainName)
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

	dns_inst := fmt.Sprintf(tr.T("dns_instructions"), Dn) + "\n\n"
	dns_inst += fmt.Sprintf("  * %s: %s ---> %s", tui.Bold("CNAME (Alias)"), tui.Undl(subDn), tui.Undl(proj.DefaultDomain()))
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

		if isDefaultDomain(domainName, proj) {
			// If default domain is already disabled, there's nothing to do.
			if !proj.DefaultDomainEnabled {
				log.Infof(tr.T("default_domain_already_removed"),
					tui.Undl(tui.Blu(proj.DefaultDomain())))
				return
			}

			// Disable the default domain.
			proj.DefaultDomainEnabled = false
			updatedProj, appErr := projects.Update(token, proj)
			if appErr != nil {
				appErr.Handle()
			}

			proj = updatedProj
			break
		}

		// Delete a "regular" custom domain.
		appErr := domains.Delete(token, proj.Name, domainName)
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

	log.Infof(tr.T("domain_removed"), tui.Undl(tui.Blu(domainName)), proj.Name)
}

func isDefaultDomain(domain string, proj *project.Project) bool {
	return domain == defaultDomainArg || domain == proj.DefaultDomain()
}
