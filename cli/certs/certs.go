package certs

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

func Create(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	// Number arguments should be 0 or 3
	//  - No arguments
	//  [DOMAIN] [CRT_FILE] [KEY_FILE]
	if len(c.Args()) != 0 && len(c.Args()) != 3 {
		cli.ShowCommandHelp(c, c.Command.Name)
		return
	}

	var domainName, crtFilePath, keyFilePath string

	if len(c.Args()) == 0 {
		var err error

		for {
			domainName, err = readline.Read(tui.Bold(tr.T("enter_domain_name_to_add")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			if domainName != "" {
				domainName = util.SanitizeDomain(domainName)
				break
			}
		}

		for {
			crtFilePath, err = readline.Read(tui.Bold(tr.T("enter_cert_path")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			err := checkCertFile(crtFilePath)
			if err == nil {
				break
			}

			log.Errorln(err)
		}

		for {
			keyFilePath, err = readline.Read(tui.Bold(tr.T("enter_key_path")+": "), true, "")
			util.ExitIfErrorOrEOF(err)

			err := checkCertFile(keyFilePath)
			if err == nil {
				break
			}

			log.Errorln(err)
		}
	} else {
		domainName = util.SanitizeDomain(c.Args().Get(0))
		crtFilePath = c.Args().Get(1)
		keyFilePath = c.Args().Get(2)

		if err := checkCertFile(crtFilePath); err != nil {
			log.Fatalln(err)
		}
		if err := checkCertFile(keyFilePath); err != nil {
			log.Fatalln(err)
		}
	}

	appErr := certs.Create(token, proj.Name, domainName, crtFilePath, keyFilePath)
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
		case certs.ErrInvalidCerts:
			log.Fatalln(tr.T("cert_invalid"))
		case certs.ErrCertNotMatch:
			log.Fatalf(tr.T("cert_not_match"), domainName)
		}

		appErr.Handle()
	}

	tui.Printf(tui.Undl(tui.Bold(tr.T("cert_uploaded")))+"\n", proj.Name, domainName)
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
