package bundle_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nitrous-io/rise-cli-go/bundle"
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bundle")
}

var _ = Describe("Bundle", func() {
	var (
		currDir string
		tempDir string
		err     error
	)

	BeforeEach(func() {
		currDir, err = os.Getwd()
		Expect(err).To(BeNil())
		tempDir, err = ioutil.TempDir("", "rise-test")
		Expect(err).To(BeNil())
		os.Chdir(tempDir)
	})

	AfterEach(func() {
		os.Chdir(currDir)
		os.RemoveAll(tempDir)
	})

	Describe("Assemble()", func() {
		BeforeEach(func() {
			files := []string{
				".gitignore",
				"Rakefile",
				"app/assets/javascripts/application.js",
				"app/controllers/home/home_controller.rb",
				"app/controllers/posts/posts_controller.rb",
				"app/models/.gitkeep",
				"app/models/post.rb",
				"app/models/#post.rb",
				"app/views/home/home.erb",
				"app/views/home/home.erb~",
				"config/environments/production.rb",
				"config/environments/development.rb",
				"log/development.log",
				"log/production.log",
				"public/index.html",
				"tmp/appendonly.txt",
				"vendor/assets/javascripts/jquery/jquery-2.0.js",
				"vendor/assets/javascripts/underscore.js",
				"README.rdoc",
			}

			for _, f := range files {
				if strings.Contains(f, "/") {
					dir := filepath.Dir(f)
					err = os.MkdirAll(dir, 0700)
					Expect(err).To(BeNil())
				}
				err = ioutil.WriteFile(f, []byte("foo"), 0600)
				Expect(err).To(BeNil())
			}

			// symlink to a file
			err = os.Symlink(filepath.Join(tempDir, "vendor/assets/javascripts/underscore.js"), filepath.Join(tempDir, "app/assets/javascripts/underscore.js"))
			Expect(err).To(BeNil())

			// symlink to a dir
			err = os.Symlink(filepath.Join(tempDir, "vendor/assets/javascripts/jquery"), filepath.Join(tempDir, "app/assets/javascripts/jquery"))
			Expect(err).To(BeNil())

			// unreadable file
			err = os.Chmod("tmp/appendonly.txt", 0200)
			Expect(err).To(BeNil())

			time.Sleep(100 * time.Millisecond)
		})

		It("return all files", func() {
			b := bundle.New(".")
			count, size, err := b.Assemble([]string{"log", "development.rb", "vendor/assets"}, false)
			Expect(err).To(BeNil())

			expectedFiles := []string{
				"Rakefile",
				"app/assets/javascripts/application.js",
				"app/assets/javascripts/underscore.js",
				"app/controllers/home/home_controller.rb",
				"app/controllers/posts/posts_controller.rb",
				"app/models/post.rb",
				"app/views/home/home.erb",
				"config/environments/production.rb",
				"public/index.html",
				"README.rdoc",
			}

			Expect(count).To(Equal(10))
			Expect(size).To(Equal(int64(30)))

			Expect(b.FileList()).To(ConsistOf(expectedFiles))
		})
	})
})
