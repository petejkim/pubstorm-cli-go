package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/nitrous-io/rise-cli-go/bundle"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/spinner"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Deploy(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	absPath, err := filepath.Abs(proj.Path)
	util.ExitIfError(err)

	fmt.Printf("Scanning \"%s\"...\n", absPath)

	bun := bundle.New(proj.Path)
	count, size, err := bun.Assemble([]string{"rise.json", "Thumbs.db", "desktop.ini"}, false)

	fmt.Printf("Bundling %s files (%s)...\n", humanize.Comma(int64(count)), humanize.Bytes(uint64(size)))

	if size > config.MaxProjectSize {
		fmt.Printf("Error: Your project size cannot exceed %s!\n", humanize.Bytes(uint64(config.MaxProjectSize)))
		os.Exit(1)
	}

	tempDir, err := ioutil.TempDir("", "rise-deploy")
	util.ExitIfError(err)
	defer os.RemoveAll(tempDir)

	bunPath := filepath.Join(tempDir, "bundle.tar.gz")

	fmt.Printf("Packing bundle \"%s\"...\n", proj.Name)

	err = bun.Pack(bunPath, true)
	util.ExitIfError(err)

	fi, err := os.Stat(bunPath)
	util.ExitIfError(err)

	if fi.Size() > config.MaxProjectSize {
		fmt.Printf("Error: Your bundle size cannot exceed %s!\n", humanize.Bytes(uint64(config.MaxProjectSize)))
		os.Exit(1)
	}

	fmt.Printf("Uploading bundle \"%s\" to Rise Cloud...\n", proj.Name)

	deployment, appErr := deployments.Create(config.AccessToken, proj.Name, bunPath, true)
	if appErr != nil {
		appErr.Handle()
	}

	spin := spinner.New()
	fmt.Printf("\nLaunching...%s", string(spin.Next()))

	for deployment.State != deployments.DeploymentStateDeployed {
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("\b%s", string(spin.Next()))

		deployment, appErr = deployments.Get(config.AccessToken, proj.Name, deployment.ID)
		if appErr != nil {
			appErr.Handle()
		}
	}

	fmt.Printf("\b \b\n\nhttps://%s.%s/ deployed to Rise\n", proj.Name, config.DefaultDomain)
}
