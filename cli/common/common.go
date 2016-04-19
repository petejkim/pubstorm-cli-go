package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"
)

var (
	sharedDebugLogger *log.Logger
	once              sync.Once
)

func DebugLog() *log.Logger {
	p := filepath.Join(config.DotRisePath, "debug.log")

	once.Do(func() {
		sharedDebugLogger = log.New()
		sharedDebugLogger.Level = log.DebugLevel
		sharedDebugLogger.Out = ioutil.Discard

		if f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); err == nil {
			sharedDebugLogger.Out = f
		}
	})

	return sharedDebugLogger
}

func RequireAccessToken() string {
	token := config.AccessToken
	if token == "" {
		log.Fatal(tr.T("not_logged_in"))
	}
	return token
}

func RequireProject(accessToken string) *project.Project {
	proj, err := project.Load()
	if os.IsNotExist(err) {
		log.Fatal(tr.T("no_rise_project"))
	}
	util.ExitIfError(err)

	apiProj, appErr := projects.Get(accessToken, proj.Name)
	if appErr != nil {
		if appErr.Code == projects.ErrCodeNotFound {
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		}
		appErr.Handle()
	}

	proj.DefaultDomainEnabled = apiProj.DefaultDomainEnabled

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

func CheckForUpdates() {
	p := filepath.Join(config.DotRisePath, "last-update")

	lastChecked, err := readUpdateCheckTime(p)
	if err != nil {
		DebugLog().Errorf("failed to read last update check time, err: %v", err)
	}
	if lastChecked.Add(1 * time.Hour).After(time.Now()) {
		DebugLog().Debugf("update check last ran at %v, skipping", lastChecked)
		return
	}
	if err := writeUpdateCheckTime(p, time.Now()); err != nil {
		DebugLog().Errorf("failed to save last update check time, err: %v", err)
	}

	res, err := goreq.Request{
		Method:    "GET",
		Uri:       config.LatestVersionURL,
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
		Timeout:   1 * time.Second,
	}.Do()

	if err != nil {
		DebugLog().Errorf("failed in HTTP request for version file, err: %v", err)
		return
	}
	defer res.Body.Close()

	var j struct {
		Version string `json:"version"`
	}
	if err := res.Body.FromJsonTo(&j); err != nil {
		DebugLog().Errorf("update check version file cannot be read, err: %v", err)
		return
	}

	if j.Version == "" || j.Version == config.Version {
		DebugLog().Debugf("update not necessary - current: %s, latest: %s",
			config.Version, j.Version)
		return
	}

	DebugLog().Debugf("update available - current: %s, latest: %s",
		config.Version, j.Version)

	tui.Println(tui.Bold(strings.Repeat("-", 72)))
	tui.Println(tr.T("update_available"))
	tui.Println()
	tui.Printf(tr.T("update_current_version"), tui.Ylo(config.Version))
	tui.Println()
	tui.Printf(tr.T("update_latest_version"), tui.Grn(tui.Bold(j.Version)))
	tui.Println()
	tui.Println()
	tui.Printf(tr.T("update_instructions"), tui.Grn(tui.Bold(j.Version)))
	tui.Println()
	tui.Println(tui.Bold(strings.Repeat("-", 72)))
	tui.Println()
}

func readUpdateCheckTime(path string) (time.Time, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil // i.e. update check never ran.
		}

		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(b)))
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func writeUpdateCheckTime(path string, t time.Time) error {
	return ioutil.WriteFile(path, []byte(t.Format(time.RFC3339)), 0644)
}
