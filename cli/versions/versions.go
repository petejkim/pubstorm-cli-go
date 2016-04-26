package versions

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
)

func Versions(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	depls, appErr := deployments.List(config.AccessToken, proj.Name)
	if appErr != nil {
		appErr.Handle()
	}

	tui.Printf(tui.Bold(tr.T("versions_list"))+"\n", proj.Name)
	tui.Printf(tui.Undl("%-10s %-20s %-10s\n"), "Version", "Deployed At", "State")
	for _, depl := range depls {
		if depl.Active {
			output := fmt.Sprintf("%-10s %-20s %s", fmt.Sprintf("v%d", depl.Version), humanize.Time(depl.DeployedAt), "live")
			tui.Println(tui.Bold(output))
		} else if depl.State == "deployed" {
			tui.Printf("%-10s %-20s\n", fmt.Sprintf("v%d", depl.Version), humanize.Time(depl.DeployedAt))
		}
	}
}
