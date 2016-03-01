package project_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/project"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "project")
}

var _ = Describe("Project", func() {
	DescribeTable("ValidateName()",
		func(name string, expectedError error) {
			proj := &project.Project{Name: name}
			err := proj.ValidateName()

			if expectedError == nil {
				Expect(err).To(BeNil())
			} else {
				Expect(err).To(Equal(expectedError))
			}
		},

		Entry("normal", "abc", nil),
		Entry("allows hyphens", "good-one", nil),
		Entry("allows multiple hyphens", "hello-world--foobar", nil),
		Entry("disallows starting with a hyphen", "-abc", project.ErrNameInvalid),
		Entry("disallows ending with a hyphen", "abc-", project.ErrNameInvalid),
		Entry("disallows spaces", "good one", project.ErrNameInvalid),
		Entry("disallows names shorter than 3 characters", "aa", project.ErrNameInvalidLength),
		Entry("disallows names longer than 63 characters", strings.Repeat("a", 64), project.ErrNameInvalidLength),
		Entry("disallows special characters", "good&one", project.ErrNameInvalid),
	)

	Describe("file system dependent tests", func() {
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

		Describe("ValidatePath()", func() {
			var tempDeployDir string

			BeforeEach(func() {
				tempDeployDir = filepath.Join(tempDir, "public")
				err = os.Mkdir(tempDeployDir, 0700)
				Expect(err).To(BeNil())
			})

			Context("when path is absolute", func() {
				It("returns error", func() {
					proj := &project.Project{Path: tempDeployDir}
					Expect(proj.ValidatePath()).To(Equal(project.ErrPathNotRelative))
				})
			})

			Context("when path does not exist", func() {
				It("returns error", func() {
					proj := &project.Project{Path: "./public2"}
					Expect(proj.ValidatePath()).To(Equal(project.ErrPathNotExist))
				})
			})

			Context("when path is not a directory", func() {
				It("returns error", func() {
					err = ioutil.WriteFile(filepath.Join(tempDir, "public2"), []byte{'a'}, 0600)
					Expect(err).To(BeNil())

					proj := &project.Project{Path: "./public2"}
					Expect(proj.ValidatePath()).To(Equal(project.ErrPathNotDir))
				})
			})

			Context("when path is relative and exists", func() {
				It("returns nil", func() {
					proj := &project.Project{Path: "./public"}
					Expect(proj.ValidatePath()).To(BeNil())
				})
			})
		})

		Describe("Save()", func() {
			It("persists settings in rise.json file in the current working directory", func() {
				proj := &project.Project{
					Name:        "foo-bar-express",
					Path:        "./build",
					EnableStats: true,
					ForceHTTPS:  false,
				}

				err = proj.Save()
				Expect(err).To(BeNil())

				f, err := os.Open(filepath.Join(tempDir, "rise.json"))
				Expect(err).To(BeNil())
				defer f.Close()

				var j map[string]interface{}
				err = json.NewDecoder(f).Decode(&j)
				Expect(err).To(BeNil())

				Expect(j).NotTo(BeNil())
				Expect(j["name"]).To(Equal("foo-bar-express"))
				Expect(j["path"]).To(Equal("./build"))
				Expect(j["enable_stats"]).To(BeTrue())
				Expect(j["force_https"]).To(BeFalse())
			})
		})

		Describe("Load()", func() {
			Context("when rise.json does not exist", func() {
				It("returns error", func() {
					proj, err := project.Load()
					Expect(err).NotTo(BeNil())
					Expect(os.IsNotExist(err)).To(BeTrue())
					Expect(proj).To(BeNil())
				})
			})

			Context("when rise.json exists", func() {
				BeforeEach(func() {
					err = ioutil.WriteFile("rise.json", []byte(`
						{
							"name": "good-beer-company",
							"path": "./output",
							"enable_stats": false,
							"force_https": true
						}
					`), 0600)
					Expect(err).To(BeNil())
				})

				It("loads rise.json and returns a project", func() {
					proj, err := project.Load()
					Expect(err).To(BeNil())

					Expect(proj).NotTo(BeNil())
					Expect(proj.Name).To(Equal("good-beer-company"))
					Expect(proj.Path).To(Equal("./output"))
					Expect(proj.EnableStats).To(BeFalse())
					Expect(proj.ForceHTTPS).To(BeTrue())
				})
			})
		})
	})
})
