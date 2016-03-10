package deployments_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "deployments")
}

var _ = Describe("Deployments", func() {
	var (
		origHost string
		server   *ghttp.Server
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		origHost = config.Host
		config.Host = server.URL()
	})

	AfterEach(func() {
		config.Host = origHost
		server.Close()
	})

	type expectation struct {
		resCode    int
		resBody    string
		errIsNil   bool
		errCode    string
		errDesc    string
		errIsFatal bool
	}

	DescribeTable("Create",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/deployments"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {"application/vnd.rise.v0+json"},
						"User-Agent":    {"RiseCLI"},
					}),

					func(w http.ResponseWriter, req *http.Request) {
						key := http.CanonicalHeaderKey("Content-Type")
						Expect(req.Header[key][0]).To(HavePrefix("multipart/form-data"))
					},

					func(w http.ResponseWriter, req *http.Request) {
						mr, err := req.MultipartReader()
						Expect(err).To(BeNil())

						part, err := mr.NextPart()
						Expect(err).To(BeNil())

						Expect(part.FormName()).To(Equal("payload"))

						data, err := ioutil.ReadAll(part)
						Expect(err).To(BeNil())
						Expect(string(data)).To(Equal("my-bundle-yo"))
					},

					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			tempDir, err := ioutil.TempDir("", "rise-test")
			Expect(err).To(BeNil())
			defer func() {
				os.RemoveAll(tempDir)
			}()

			bunPath := filepath.Join(tempDir, "bundle.tar.gz")
			err = ioutil.WriteFile(bunPath, []byte("my-bundle-yo"), 0600)
			Expect(err).To(BeNil())

			defer func() {
				os.Remove(bunPath)
			}()

			appErr := deployments.Create("t0k3n", "foo-bar-express", bunPath, false)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(strings.ToLower(appErr.Description)).To(ContainSubstring(e.errDesc))
				Expect(appErr.IsFatal).To(Equal(e.errIsFatal))
			}
		},

		Entry("unexpected response code", expectation{
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("400 with bad request", expectation{
			resCode:    http.StatusBadRequest,
			resBody:    `{"error": "invalid_request", "error_description": "request body is too large"}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeValidationFailed,
			errDesc:    "project size is too large",
			errIsFatal: true,
		}),

		Entry("successful deployment", expectation{
			resCode:  http.StatusAccepted,
			resBody:  `{"deployment": {"id": 10, "state": "uploaded" }}`,
			errIsNil: true,
		}),
	)
})
