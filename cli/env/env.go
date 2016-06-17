package env

import (
	"strings"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/cli/deploy"
	"github.com/nitrous-io/rise-cli-go/client/jsenvvars"
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

	log.Infof(tr.T("env_updating"), proj.Name)

	deployment, appErr := jsenvvars.Add(token, proj.Name, pairs)
	if appErr != nil {
		appErr.Handle()
	}

	deploy.ShowDeploymentProcess(token, proj.Name, deployment)
	log.Infof(tr.T("env_set"), output)
}

func Delete(c *cli.Context) {
	if !c.Args().Present() {
		cli.ShowCommandHelp(c, c.Command.Name)
		return
	}

	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	log.Infof(tr.T("env_updating"), proj.Name)

	deployment, appErr := jsenvvars.Delete(token, proj.Name, c.Args())
	if appErr != nil {
		appErr.Handle()
	}

	deploy.ShowDeploymentProcess(token, proj.Name, deployment)
	log.Infof(tr.T("env_deleted"), strings.Join(c.Args(), ", "))
}

func List(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	envvars, appErr := jsenvvars.List(token, proj.Name)
	if appErr != nil {
		appErr.Handle()
	}

	if len(*envvars) > 0 {
		tui.Printf(tui.Undl(tui.Bold(tr.T("env_list_header")))+"\n", proj.Name)
		for key, value := range *envvars {
			tui.Println(tui.Bold(key) + ": " + value)
		}
	} else {
		log.Infof(tr.T("no_env")+"\n", proj.Name)
	}
}
