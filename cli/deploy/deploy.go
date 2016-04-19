package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"
	humanize "github.com/dustin/go-humanize"
	"github.com/nitrous-io/rise-cli-go/bundle"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/spinner"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"

	log "github.com/Sirupsen/logrus"
)

func Deploy(c *cli.Context) {
	verbose := c.Bool("verbose")

	token := common.RequireAccessToken()
	proj := common.RequireProject(token)

	absPath, err := filepath.Abs(proj.Path)
	util.ExitIfError(err)

	tui.Printf(tr.T("scanning_path")+"\n", absPath)

	bun := bundle.New(proj.Path)
	count, size, err := bun.Assemble([]string{config.ProjectJSON, "Thumbs.db", "desktop.ini"}, verbose)

	log.Infof(tr.T("bundling_file_count_size"), humanize.Comma(int64(count)), humanize.Bytes(uint64(size)))

	if size > config.MaxProjectSize {
		log.Fatalf(tr.T("project_size_exceeded"), humanize.Bytes(uint64(config.MaxProjectSize)))
	}

	indexHTMLPath := filepath.Join(absPath, "index.html")
	if _, err := os.Stat(indexHTMLPath); os.IsNotExist(err) {
		log.Warnf(tr.T("bundle_root_index_missing"))
	}

	tempDir, err := ioutil.TempDir("", "rise-deploy")
	util.ExitIfError(err)
	defer os.RemoveAll(tempDir)

	bunPath := filepath.Join(tempDir, "bundle.tar.gz")

	tui.Printf("\n"+tr.T("packing_bundle")+"\n", proj.Name)

	err = bun.Pack(bunPath, true, true)
	util.ExitIfError(err)

	fi, err := os.Stat(bunPath)
	util.ExitIfError(err)

	if fi.Size() > config.MaxBundleSize {
		log.Fatalf(tr.T("bundle_size_exceeded"), humanize.Bytes(uint64(config.MaxBundleSize)))
	}

	tui.Printf("\n"+tr.T("uploading_bundle")+"\n", proj.Name)

	deployment, appErr := deployments.Create(config.AccessToken, proj.Name, bunPath, false)
	if appErr != nil {
		if appErr.Code == projects.ErrCodeNotFound {
			log.Fatalf(tr.T("project_not_found"), proj.Name)
		}
		appErr.Handle()
	}

	spin := spinner.New()
	tui.Printf("\n"+tr.T("launching")+" "+tui.Blu("%s"), string(spin.Next()))

	for deployment.State != deployments.DeploymentStateDeployed {
		time.Sleep(500 * time.Millisecond)

		tui.Printf(tui.Blu("\b%s"), string(spin.Next()))

		deployment, appErr = deployments.Get(config.AccessToken, proj.Name, deployment.ID)
		if appErr != nil {
			appErr.Handle()
		}
	}

	projUrl := fmt.Sprintf("https://%s.%s/", proj.Name, config.DefaultDomain)
	tui.Println("\b \b")
	log.Infof(tr.T("published"), tui.Undl(tui.Blu(projUrl)))
}
