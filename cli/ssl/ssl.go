package ssl

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/certs"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/pkg/spinner"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func Set(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var domainName, crtFilePath, keyFilePath string
	var err error

	if len(c.Args()) < 1 {
		for {
			domainName, err = readline.Read(tui.Bold(tr.T("ssl_enter_domain_name")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if domainName != "" {
				domainName = util.SanitizeDomain(domainName)
				break
			}
		}
	} else {
		domainName = util.SanitizeDomain(c.Args().Get(0))
	}

	if len(c.Args()) < 2 {
		for {
			crtFilePath, err = readline.Read(tui.Bold(tr.T("ssl_enter_cert_path")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			err := checkCertFile(crtFilePath)
			if err == nil {
				break
			}

			log.Errorln(err)
		}
	} else {
		crtFilePath = c.Args().Get(1)

		if err := checkCertFile(crtFilePath); err != nil {
			log.Fatalln(err)
		}
	}

	if len(c.Args()) < 3 {
		for {
			keyFilePath, err = readline.Read(tui.Bold(tr.T("ssl_enter_key_path")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			err := checkCertFile(keyFilePath)
			if err == nil {
				break
			}

			log.Errorln(err)
		}
	} else {
		keyFilePath = c.Args().Get(2)

		if err := checkCertFile(keyFilePath); err != nil {
			log.Fatalln(err)
		}
	}

	ct, appErr := certs.Create(token, proj.Name, domainName, crtFilePath, keyFilePath)
	if appErr != nil {
		switch appErr.Code {
		case certs.ErrCodeProjectNotFound:
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		case certs.ErrCodeNotFound:
			log.Fatalf(tr.T("domain_not_found"), domainName)
		case certs.ErrCodeNotAllowedDomain:
			log.Fatalf(tr.T("ssl_not_allowed_domain"), domainName)
		case certs.ErrCodeFileSizeTooLarge:
			log.Fatalln(tr.T("ssl_too_large"))
		case certs.ErrInvalidCert:
			log.Fatalln(tr.T("ssl_invalid"))
		case certs.ErrInvalidCommonName:
			log.Fatalf(tr.T("ssl_invalid_domain"), domainName)
		}

		appErr.Handle()
	}

	log.Infof(tr.T("ssl_cert_set"), tui.Undl("https://"+domainName+"/"))

	tui.Printf("\n"+tui.Undl(tui.Bold(tr.T("ssl_cert_details")+":"))+"\n", domainName)
	tui.Println(tr.T("ssl_cert_common_name") + ": " + ct.CommonName)
	tui.Println(tr.T("ssl_cert_issuer") + ": " + ct.Issuer)
	tui.Println(tr.T("ssl_cert_subject") + ": " + ct.Subject)
	tui.Println(tr.T("ssl_cert_starts_at") + ": " + ct.StartsAt.String())
	tui.Println(tr.T("ssl_cert_expires_at") + ": " + ct.ExpiresAt.String())
}

func Info(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var domainName string

	if len(c.Args()) < 1 {
		var err error
		for {
			domainName, err = readline.Read(tui.Bold(tr.T("ssl_enter_domain_name")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if domainName != "" {
				domainName = util.SanitizeDomain(domainName)
				break
			}
		}
	} else {
		domainName = util.SanitizeDomain(c.Args().Get(0))
	}

	ct, appErr := certs.Get(token, proj.Name, domainName)
	if appErr != nil {
		switch appErr.Code {
		case certs.ErrCodeProjectNotFound:
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		case certs.ErrCodeNotFound:
			log.Infof(tr.T("ssl_cert_not_found"), domainName)
			return
		}

		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("ssl_cert_details")+":"))+"\n", domainName)
	tui.Println(tr.T("ssl_cert_common_name") + ": " + ct.CommonName)
	tui.Println(tr.T("ssl_cert_issuer") + ": " + ct.Issuer)
	tui.Println(tr.T("ssl_cert_subject") + ": " + ct.Subject)
	tui.Println(tr.T("ssl_cert_starts_at") + ": " + ct.StartsAt.String())
	tui.Println(tr.T("ssl_cert_expires_at") + ": " + ct.ExpiresAt.String())
}

func Delete(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var domainName string

	if len(c.Args()) < 1 {
		var err error
		for {
			domainName, err = readline.Read(tui.Bold(tr.T("ssl_enter_domain_name")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if domainName != "" {
				domainName = util.SanitizeDomain(domainName)
				break
			}
		}
	} else {
		domainName = util.SanitizeDomain(c.Args().Get(0))
	}

	appErr := certs.Delete(token, proj.Name, domainName)
	if appErr != nil {
		switch appErr.Code {
		case certs.ErrCodeProjectNotFound:
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		case certs.ErrCodeNotFound:
			log.Fatalf(tr.T("ssl_cert_not_found"), domainName)
		}

		appErr.Handle()
	}

	log.Infof(tr.T("ssl_cert_removed"), domainName)
}

func Force(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	arg := c.Args().Get(0)
	forceHTTPS, err := strconv.ParseBool(arg)
	if err != nil {
		forceHTTPS = !(arg == "off" || arg == "disable" || arg == "disabled" || arg == "no")
	}
	proj.ForceHTTPS = forceHTTPS

	updatedProj, appErr := projects.Update(token, proj)
	if appErr != nil {
		appErr.Handle()
	}

	if updatedProj.ForceHTTPS {
		log.Info(tr.T("ssl_force_https_on"))
	} else {
		log.Info(tr.T("ssl_force_https_off"))
	}
}

func Letsencrypt(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	domainName := util.SanitizeDomain(c.Args().First())
	interactive := domainName == ""

	if interactive {
		var err error
		for {
			domainName, err = readline.Read(tui.Bold(tr.T("ssl_enter_domain_name")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if domainName != "" {
				domainName = util.SanitizeDomain(domainName)
				break
			}
		}
	}

	spin := spinner.New()
	done := make(chan struct{})
	go func() {
		tui.Printf(tr.T("ssl_letsencrypt_progress")+" "+tui.Blu("%s"), tui.Undl(domainName), string(spin.Next()))

		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				tui.Printf(tui.Blu("\b%s"), string(spin.Next()))
			case <-done:
				ticker.Stop()
				tui.Println("\b \b") // "Eat up" spinner characters.
				done <- struct{}{}
				return
			}
		}
	}()

	appErr := certs.Enable(token, proj.Name, domainName)
	done <- struct{}{} // Stop spinner.
	<-done
	if appErr != nil {
		switch appErr.Code {
		case certs.ErrCodeProjectNotFound:
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		case certs.ErrCodeAcmeServerError:
			subDn, apex := util.SplitDomain(domainName)
			log.Errorf(tr.T("ssl_letsencrypt_error"), tui.Undl(domainName))
			log.Infof(tr.T("ssl_letsencrypt_dns"), apex)
			log.Infof("  * %s: %s ---> %s", tui.Bold("CNAME (Alias)"), tui.Undl(subDn), tui.Undl(proj.DefaultDomain()))
			if subDn == "www" {
				log.Infof("\n  * %s: %s ---> %s", tui.Bold("A (Host)"), tui.Undl("@"), tui.Undl(config.RedirectorIP))
			}
			return
		}

		appErr.Handle()
	}

	log.Infof(tr.T("ssl_letsencrypt_success"), tui.Undl("https://"+domainName+"/"))
}

func checkCertFile(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(tr.T("ssl_file_not_found"), filePath)
		}
		log.Fatal(err)
	}

	if fi.Size() < 10 {
		return fmt.Errorf(tr.T("ssl_file_invalid"), filePath)
	}

	return nil
}
