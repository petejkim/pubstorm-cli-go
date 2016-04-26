package env

import (
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/client/jsenvvars"
	"github.com/nitrous-io/rise-cli-go/pkg/spinner"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"

	log "github.com/Sirupsen/logrus"
)

func Add(c *cli.Context) {
	if !c.Args().Present() {
		cli.ShowCommandHelp(c, c.Command.Name)
		return
	}

	pairs := make(map[string]string)
	output := ""
	for _, arg := range c.Args() {
		// Assume key does not have `=` character
		keyValue := strings.SplitN(arg, "=", 2)
		if len(keyValue) != 2 {
			log.Fatalf(tr.T("env_invalid_arg"), arg)
		}
		key := keyValue[0]
		value := keyValue[1]

		if key == "" {
			log.Fatalln(tr.T("env_key_empty"))
		}

		pairs[key] = value
		output += key + ": " + value + "\n"
	}

	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	deployment, appErr := jsenvvars.Add(token, proj.Name, pairs)
	if appErr != nil {
		appErr.Handle()
	}

	spin := spinner.New()
	tui.Printf("\n"+tr.T("launching")+" "+tui.Blu("%s"), deployment.Version, string(spin.Next()))

	for deployment.State != deployments.DeploymentStateDeployed {
		for i := 0; i < 5; i++ {
			time.Sleep(100 * time.Millisecond)
			tui.Printf(tui.Blu("\b%s"), string(spin.Next()))
		}

		deployment, appErr = deployments.Get(token, proj.Name, deployment.ID)
		if appErr != nil {
			appErr.Handle()
		}
	}

	tui.Println("\b \b")
	log.Infof(tr.T("env_set"), output)
}
