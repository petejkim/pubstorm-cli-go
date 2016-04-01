package common

import (
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
	build := config.BuildEnv
	if build == "production" {
		build = ""
	}
	tui.Printf(
		"%s\n%s\n%s\n%s\n%s%s\n\n",
		tui.Ylo(`     ____        __   _____ __`),
		tui.Ylo(`    / __ \__  __/ /_ / ___// /_____  _________ ___`),
		tui.Red(`   / /_/ / / / / __ \\__ \/ __/ __ \/ ___/ __ `+"`"+`__ \`),
		tui.Mag(`  / ____/ /_/ / /_/ /__/ / /_/ /_/ / /  / / / / / /`),
		tui.Blu(` /_/    \__,_/_.___/____/\__/\____/_/  /_/ /_/ /_/ `),
		build,
	)
}
