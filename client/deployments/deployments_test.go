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
		result     *deployments.Deployment
	}

	DescribeTable("Create",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/deployments"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
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

			deployment, appErr := deployments.Create("t0k3n", "foo-bar-express", bunPath, true)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(deployment.ID).To(Equal(e.result.ID))
				Expect(deployment.State).To(Equal(e.result.State))
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

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("successful deployment", expectation{
			resCode:  http.StatusAccepted,
			resBody:  `{"deployment": {"id": 10, "state": "uploaded" }}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 10, State: "uploaded"},
		}),
	)

	DescribeTable("Get",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/deployments/123"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			deployment, appErr := deployments.Get("t0k3n", "foo-bar-express", 123)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(deployment.ID).To(Equal(e.result.ID))
				Expect(deployment.State).To(Equal(e.result.State))
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
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "deployment could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeNotFound,
			errDesc:    "deployment could not be found",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"deployment": {"id": 123, "state": "uploaded" }}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 123, State: "uploaded"},
		}),
	)
})
