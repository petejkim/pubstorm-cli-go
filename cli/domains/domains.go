package domains

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/domains"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/util"
)

func List(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	domainNames, appErr := domains.Index(config.AccessToken, proj.Name)
	if appErr != nil {
		appErr.Handle()
	}

	fmt.Printf("Domains for \"%s\"\n", proj.Name)
	for _, domainName := range domainNames {
		fmt.Println(domainName)
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
			domainName, err = readline.Read("Enter Domain Name to Add: ", true, "")
			util.ExitIfErrorOrEOF(err)
		}

		appErr := domains.Create(config.AccessToken, proj.Name, domainName)
		if appErr != nil {
			if appErr.Code == domains.ErrCodeValidationFailed {
				fmt.Println(appErr.Description)
				if interactive {
					continue
				} else {
					os.Exit(1)
				}
			} else if appErr.Code == domains.ErrCodeLimitReached {
				log.Fatalf("You cannot add any more domains to project \"%s\"!", proj.Name)
			}
			appErr.Handle()
		}

		if appErr == nil || !interactive {
			break
		}
	}

	fmt.Printf("Successfully added \"%s\" to project \"%s\"\n", domainName, proj.Name)
}
