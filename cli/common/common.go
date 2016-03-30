package common

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"
)

func RequireAccessToken() string {
	token := config.AccessToken
	if token == "" {
		log.Fatal(tr.T("not_logged_in"))
	}
	return token
}

func RequireProject() *project.Project {
	proj, err := project.Load()
	if os.IsNotExist(err) {
		log.Fatal(tr.T("no_rise_project"))
	}
	util.ExitIfError(err)
	return proj
}

func PrintLogo() {
	fmt.Printf(
		"%s\n%s\n%s\n%s\n%s\n\n",
		tui.Ylo(`        _`),
		tui.Ylo(`  _ __ (_) ___   ___`),
		tui.Red(` | '__|| |/ __| / _ \`),
		tui.Mag(` | |   | |\__ \|  __/`),
		tui.Blu(` |_|   |_||___/ \___|`),
	)
}
