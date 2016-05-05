package cert

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/certs"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
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
			domainName, err = readline.Read(tui.Bold(tr.T("cert_enter_domain_name")+": "), true, "")
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
			crtFilePath, err = readline.Read(tui.Bold(tr.T("cert_enter_cert_path")+": "), true, "")
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
			keyFilePath, err = readline.Read(tui.Bold(tr.T("cert_enter_key_path")+": "), true, "")
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
			log.Fatalf(tr.T("cert_not_allowed_domain"), domainName)
		case certs.ErrCodeFileSizeTooLarge:
			log.Fatalln(tr.T("cert_too_large"))
		case certs.ErrInvalidCert:
			log.Fatalln(tr.T("cert_invalid"))
		case certs.ErrInvalidCommonName:
			log.Fatalf(tr.T("cert_invalid_domain"), domainName)
		}

		appErr.Handle()
	}

	log.Infof(tr.T("cert_set"), tui.Undl("https://"+domainName+"/"))

	tui.Printf("\n"+tui.Undl(tui.Bold(tr.T("cert_details")+":"))+"\n", domainName)
	tui.Println(tr.T("cert_common_name") + ": " + ct.CommonName)
	tui.Println(tr.T("cert_issuer") + ": " + ct.Issuer)
	tui.Println(tr.T("cert_subject") + ": " + ct.Subject)
	tui.Println(tr.T("cert_starts_at") + ": " + ct.StartsAt.String())
	tui.Println(tr.T("cert_expires_at") + ": " + ct.ExpiresAt.String())
}

func Info(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var domainName string

	if len(c.Args()) < 1 {
		var err error
		for {
			domainName, err = readline.Read(tui.Bold(tr.T("cert_enter_domain_name")+": "), true, "")
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
			log.Fatalf(tr.T("cert_not_found"), domainName)
		}

		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("cert_details")+":"))+"\n", domainName)
	tui.Println(tr.T("cert_common_name") + ": " + ct.CommonName)
	tui.Println(tr.T("cert_issuer") + ": " + ct.Issuer)
	tui.Println(tr.T("cert_subject") + ": " + ct.Subject)
	tui.Println(tr.T("cert_starts_at") + ": " + ct.StartsAt.String())
	tui.Println(tr.T("cert_expires_at") + ": " + ct.ExpiresAt.String())
}

func Delete(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	var domainName string

	if len(c.Args()) < 1 {
		var err error
		for {
			domainName, err = readline.Read(tui.Bold(tr.T("cert_enter_domain_name")+": "), true, "")
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
			log.Fatalf(tr.T("cert_not_found"), domainName)
		}

		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("cert_delete")+":"))+"\n", domainName)
}

func checkCertFile(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(tr.T("cert_file_not_found"), filePath)
		}
		log.Fatal(err)
	}

	if fi.Size() < 10 {
		return fmt.Errorf(tr.T("cert_file_invalid"), filePath)
	}

	return nil
}
