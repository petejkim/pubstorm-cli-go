package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/nitrous-io/rise-cli-go/bundle"
	"github.com/nitrous-io/rise-cli-go/cli/common"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Deploy(c *cli.Context) {
	common.RequireAccessToken()
	proj := common.RequireProject()

	absPath, err := filepath.Abs(proj.Path)
	util.ExitIfError(err)

	fmt.Printf("Scanning \"%s\"...\n", absPath)

	bun := bundle.New(proj.Path)
	count, size, err := bun.Assemble(nil, true)

	fmt.Printf("Bundling %s files (%s)...\n", humanize.Comma(int64(count)), humanize.Bytes(uint64(size)))

	tempDir, err := ioutil.TempDir("", "rise-deploy")
	util.ExitIfError(err)
	defer os.RemoveAll(tempDir)

	bunPath := filepath.Join(tempDir, "bundle.tar.gz")

	fmt.Printf("Packing bundle \"%s\"...\n", proj.Name)

	err = bun.Pack(bunPath, true)
	util.ExitIfError(err)
}
