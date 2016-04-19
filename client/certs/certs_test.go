package certs_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/certs"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "certs")
}

var _ = Describe("Certs", func() {
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
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/domains/foo-bar-express.com/cert"),
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

						Expect(part.FormName()).To(Equal("ssl.crt"))

						data, err := ioutil.ReadAll(part)
						Expect(err).To(BeNil())
						Expect(string(data)).To(Equal("certificate"))

						part, err = mr.NextPart()
						Expect(err).To(BeNil())

						Expect(part.FormName()).To(Equal("ssl.key"))

						data, err = ioutil.ReadAll(part)
						Expect(err).To(BeNil())
						Expect(string(data)).To(Equal("private key"))
					},

					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			tempDir, err := ioutil.TempDir("", "pubstorm-certs-test")
			Expect(err).To(BeNil())
			defer func() {
				os.RemoveAll(tempDir)
			}()

			crtPath := filepath.Join(tempDir, "ssl.crt")
			err = ioutil.WriteFile(crtPath, []byte("certificate"), 0600)
			Expect(err).To(BeNil())

			keyPath := filepath.Join(tempDir, "ssl.key")
			err = ioutil.WriteFile(keyPath, []byte("private key"), 0600)
			Expect(err).To(BeNil())

			defer func() {
				os.Remove(crtPath)
				os.Remove(keyPath)
			}()

			appErr := certs.Create("t0k3n", "foo-bar-express", "foo-bar-express.com", crtPath, keyPath)
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
			errCode:    certs.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    certs.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("400 with bad request", expectation{
			resCode:    http.StatusBadRequest,
			resBody:    `{"error": "invalid_request", "error_description": "request body is too large"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeFileSizeTooLarge,
			errDesc:    "file size is too large",
			errIsFatal: true,
		}),

		Entry("403 with forbidden", expectation{
			resCode:    http.StatusForbidden,
			resBody:    `{"error": "invalid_request", "error_description": "Not allowed to upload certs for default domain"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeNotAllowedDomain,
			errDesc:    "not allowed domain name",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("422 with invalid params due to missing certs", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "certificate or private key file is missing"}`,
			errIsNil:   false,
			errCode:    certs.ErrInvalidCerts,
			errDesc:    "certificate or private key file is missing",
			errIsFatal: true,
		}),

		Entry("422 with invalid params due to invalid certs", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "certificate or private key is not valid"}`,
			errIsNil:   false,
			errCode:    certs.ErrInvalidCerts,
			errDesc:    "certificate or private key is not valid",
			errIsFatal: true,
		}),

		Entry("422 with invalid params due to not matched domain name", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "ssl cert is not matched domain name"}`,
			errIsNil:   false,
			errCode:    certs.ErrCertNotMatch,
			errDesc:    "ssl cert is not matched domain name",
			errIsFatal: true,
		}),

		Entry("successful creation", expectation{
			resCode:  http.StatusCreated,
			resBody:  `{"cert": {"id": 10 }}`,
			errIsNil: true,
		}),
	)
})
