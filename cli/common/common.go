package common

import (
	"fmt"
	"os"

	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"
)

func RequireAccessToken() string {
	token := config.AccessToken
	if token == "" {
		fmt.Fprintln(os.Stderr, `You are not logged in. Please login with "rise login" command or create a new account with "rise signup" command.`)
		os.Exit(1)
	}
	return token
}

func RequireProject() *project.Project {
	proj, err := project.Load()
	if os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Error: Could not find a Rise project in current path. Run `rise init` to create a Rise project.")
		os.Exit(1)
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
