package domains

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/domains"
	"github.com/nitrous-io/rise-cli-go/config"
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
