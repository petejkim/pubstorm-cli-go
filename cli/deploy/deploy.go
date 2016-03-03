package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/dustin/go-humanize"
	"github.com/nitrous-io/rise-cli-go/bundle"
	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/util"
)

func Deploy(c *cli.Context) {
	var (
		err  error
		proj *project.Project
	)

	proj, err = project.Load()
	if err != nil && !os.IsNotExist(err) {
		util.ExitIfError(err)
	}

	if proj == nil {
		proj = &project.Project{}

		fmt.Println("Set up your Rise project")

		for {
			proj.Name, err = readline.Read("Enter Project Name: ", true, "")
			util.ExitIfErrorOrEOF(err)

			if err := proj.ValidateName(); err != nil {
				fmt.Println(err.Error())
				continue
			}

			break
		}

		for {
			proj.Path, err = readline.Read("Enter Project Path: [.] ", true, ".")
			util.ExitIfErrorOrEOF(err)

			if err := proj.ValidatePath(); err != nil {
				fmt.Println(err.Error())
				continue
			}

			break
		}

		for {
			enableStats, err := readline.Read("Enable Basic Stats? [Y/n] ", true, "y")
			util.ExitIfErrorOrEOF(err)

			switch strings.ToLower(enableStats) {
			case "y", "yes":
				proj.EnableStats = true
			case "n", "no":
				proj.EnableStats = false
			default:
				continue
			}

			break
		}

		for {
			forceHTTPS, err := readline.Read("Force HTTPS? [y/N] ", true, "n")
			util.ExitIfErrorOrEOF(err)

			switch strings.ToLower(forceHTTPS) {
			case "y", "yes":
				proj.ForceHTTPS = true
			case "n", "no":
				proj.ForceHTTPS = false
			default:
				continue
			}

			break
		}

		err = proj.Save()
		util.ExitIfError(err)
	}

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
