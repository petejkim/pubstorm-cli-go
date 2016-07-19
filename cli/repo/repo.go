package repo

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/repos"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"
)

const defaultBranch = "master"

func Link(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	tui.Println(tr.T("link_repo_explanation"))
	log.Infof(tr.T("link_repo_caveat"))

	var (
		err           error
		repoURL       string
		branch        string
		webhookSecret string
	)

	for {
		repoURL, err = readline.Read(tui.Bold(tr.T("enter_repo_to_link")+": "), true, "")
		util.ExitIfErrorOrEOF(err)

		if repoURL != "" {
			break
		}
	}

	branch, err = readline.Read(tui.Bold(tr.T("enter_branch")+": "), false, "")
	util.ExitIfErrorOrEOF(err)

	branch = strings.TrimSpace(branch)

	if branch == "" {
		branch = defaultBranch
	}

	// Generate a random secret to use for the webhook.
	// TODO We should allow users to regenerate the secret (e.g. when they want to
	// cycle it).
	webhookSecret, err = generateRandomString(32)
	util.ExitIfError(err)

	repo, appErr := repos.Link(token, proj.Name, repoURL, branch, webhookSecret)
	if appErr != nil {
		appErr.Handle()
	}

	log.Infof(tr.T("link_repo_success"), tui.Undl(repo.URI), repo.Branch)
	printWebhookInstructions(repo)
}

func Unlink(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	appErr := repos.Unlink(token, proj.Name)
	if appErr != nil {
		if appErr.Code == repos.ErrCodeNotLinked {
			log.Errorf(tr.T("project_not_linked"), proj.Name)
			return
		}
		appErr.Handle()
	}

	log.Infof(tr.T("unlink_repo_success"), proj.Name)
}

func Info(c *cli.Context) {
	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	repo, appErr := repos.Info(token, proj.Name)
	if appErr != nil {
		if appErr.Code == repos.ErrCodeNotLinked {
			log.Errorf(tr.T("project_not_linked"), proj.Name)
			return
		}

		appErr.Handle()
	}

	log.Infof(tr.T("linked_repo_info_repo"), proj.Name, tui.Undl(repo.URI), repo.Branch)
	printWebhookInstructions(repo)
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func printWebhookInstructions(repo *repos.Repo) {
	tui.Println(tui.Undl(tui.Bold(tr.T("link_repo_instructions_header"))))
	tui.Printf(tr.T("link_repo_instructions_line_1"), tui.Undl(tui.Bold(tui.Blu(repo.URI))), tui.Bold(tui.Blu("Settings")))
	tui.Println()
	tui.Printf(tr.T("link_repo_instructions_line_2"), tui.Bold(tui.Blu("Webhooks & services")), tui.Bold(tui.Blu("Add webhook")))
	tui.Println()
	tui.Printf(tr.T("link_repo_instructions_line_3"), tui.Undl(tui.Bold(tui.Blu(repo.WebhookURL))), tui.Bold(tui.Blu("Payload URL")), tui.Bold(tui.Blu("Content type")), tui.Bold(tui.Blu("application/json")))
	tui.Println()
	tui.Printf(tr.T("link_repo_instructions_line_4"), tui.Ylo(repo.WebhookSecret), tui.Bold(tui.Blu("Secret")))
	tui.Println()
	tui.Printf(tr.T("link_repo_instructions_line_5"), tui.Bold(tui.Blu("Just the push event")), tui.Bold(tui.Blu("Add webhook")))
	tui.Println()
}
