package bundle_test

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nitrous-io/rise-cli-go/bundle"
	. "github.com/onsi/ginkgo"
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
				"rise.json",
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
			count, size, err := b.Assemble([]string{"log", "development.rb", "vendor/assets", "rise.json"}, false)
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

	Describe("Pack", func() {
		var (
			files     map[string][]byte
			fileNames []string
		)

		BeforeEach(func() {
			files = map[string][]byte{
				"foo/foo.rb": []byte(`puts "hello"`),
				"bar.sql":    []byte(`SELECT * FROM hello;`),
				"baz/baz.js": []byte(`console.log("hello");`),
				"qux.php":    []byte(`<?php echo("hello") ?>`),
			}

			fileNames = make([]string, 0, len(files))

			for fileName, fileContent := range files {
				fileNames = append(fileNames, fileName)

				if strings.Contains(fileName, "/") {
					dir := filepath.Dir(fileName)
					err = os.MkdirAll(dir, 0700)
					Expect(err).To(BeNil())
				}
				err = ioutil.WriteFile(fileName, []byte(fileContent), 0600)
				Expect(err).To(BeNil())
			}
		})

		It("creates a compressed tarball", func() {
			b := bundle.New(".")
			_, _, err := b.Assemble(nil, false)
			Expect(err).To(BeNil())

			tarballPath := filepath.Join(tempDir, "bundle.tar.gz")
			err = b.Pack(tarballPath, false)
			Expect(err).To(BeNil())

			_, err = os.Stat(tarballPath)
			Expect(err).To(BeNil())

			f, err := os.Open(tarballPath)
			Expect(err).To(BeNil())
			defer f.Close()

			gr, err := gzip.NewReader(f)
			Expect(err).To(BeNil())
			defer gr.Close()

			tr := tar.NewReader(gr)

			filesRead := []string{}

			for i := 0; i < 4; i++ {
				hdr, err := tr.Next()
				Expect(err).To(BeNil())
				fileName := hdr.Name
				Expect(files).To(HaveKey(fileName))

				data, err := ioutil.ReadAll(tr)
				Expect(err).To(BeNil())
				Expect(data).To(Equal(files[fileName]))

				filesRead = append(filesRead, fileName)
			}

			_, err = tr.Next()
			Expect(err).To(Equal(io.EOF))
			Expect(fileNames).To(ConsistOf(filesRead))
		})
	})
})
