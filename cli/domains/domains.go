package domains

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/certs"
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

	// Number arguments should not be 2
	//  - No arguments
	//  [DOMAIN]
	//  [DOMAIN] [CRT_FILE] [KEY_FILE]
	if len(c.Args()) == 2 {
		cli.ShowCommandHelp(c, c.Command.Name)
		return
	}

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
				log.Infof(tr.T("default_domain_already_added"), proj.DefaultDomain(), proj.Name)
				return
			}

			// Enable the default domain.
			proj.DefaultDomainEnabled = true
			updatedProj, appErr := projects.Update(token, proj)
			if appErr != nil {
				appErr.Handle()
			}

			proj = updatedProj

			log.Infof(tr.T("default_domain_added"), proj.DefaultDomain(), proj.Name)

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

	var (
		crtFilePath, keyFilePath string
		validCert                = false
	)

	if interactive {
		crtFilePath, keyFilePath = getCertUsingInteractiveShell()
		validCert = crtFilePath != "" && keyFilePath != ""
	} else {
		if len(c.Args()) == 3 {
			crtFilePath = c.Args().Get(1)
			keyFilePath = c.Args().Get(2)

			if checkCertFile(crtFilePath) && checkCertFile(keyFilePath) {
				validCert = true
			}
		}
	}

	if validCert {
		log.Infoln(tr.T("will_upload_cert"))
		appErr := certs.Create(config.AccessToken, proj.Name, domainName, crtFilePath, keyFilePath)
		if appErr != nil {
			switch appErr.Code {
			case certs.ErrCodeProjectNotFound:
				log.Fatalf(tr.T("project_not_found"), proj.Name)
			case certs.ErrCodeFileSizeTooLarge:
				log.Fatalf(tr.T("cert_too_large"))
			case certs.ErrInvalidCerts:
				log.Fatalf(tr.T("cert_invalid"))
			}

			appErr.Handle()
		}
		log.Infof(tr.T("cert_uploaded")+"\n", proj.Name, domainName)
	}

	tui.Println()
	log.Infof(tr.T("dns_more_info"), tui.Undl(config.DNSHelpURL))
}

func Remove(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	domainName := strings.TrimSpace(c.Args().First())

	var (
		err       error
		isDefault bool
	)
	interactive := domainName == ""

	for {
		if interactive {
			domainName, err = readline.Read(tui.Bold(tr.T("enter_domain_name_to_remove")+": "), true, "")
			util.ExitIfErrorOrEOF(err)
		}

		domainName = util.SanitizeDomain(domainName)

		isDefault = isDefaultDomain(domainName, proj)
		if isDefault {
			// If default domain is already disabled, there's nothing to do.
			if !proj.DefaultDomainEnabled {
				log.Infof(tr.T("default_domain_already_removed"), proj.DefaultDomain(), proj.Name)
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

	if isDefault {
		log.Infof(tr.T("default_domain_removed"), proj.DefaultDomain(), proj.Name)
	} else {
		log.Infof(tr.T("domain_removed"), domainName, proj.Name)
	}
}

func isDefaultDomain(domain string, proj *project.Project) bool {
	return domain == defaultDomainArg || domain == proj.DefaultDomain()
}

func getCertUsingInteractiveShell() (string, string) {
	for {
		uploadCert, err := readline.Read(tui.Bold(tr.T("want_upload_cert")+"? [y/N] "), true, "n")
		util.ExitIfErrorOrEOF(err)

		switch strings.ToLower(uploadCert) {
		case "y", "yes":
		case "n", "no":
			return "", ""
		}

		break
	}

	var (
		crtFilePath, keyFilePath string
		err                      error
	)

	for {
		crtFilePath, err = readline.Read(tui.Bold(tr.T("enter_cert_path")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		if checkCertFile(crtFilePath) {
			break
		}
	}

	for {
		keyFilePath, err = readline.Read(tui.Bold(tr.T("enter_key_path")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		if checkCertFile(keyFilePath) {
			break
		}
	}

	return crtFilePath, keyFilePath
}

func checkCertFile(filePath string) bool {
	fi, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			tui.Printf(tr.T("cert_file_not_found")+"\n", filePath)
			return false
		}
		log.Fatal(err)
	}

	if fi.Size() < 10 {
		tui.Printf(tr.T("cert_file_invalid")+"\n", filePath)
		return false
	}
	return true
}
