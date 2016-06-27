package certs_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nitrous-io/rise-cli-go/client/certs"
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

		startsAt              = time.Now().Add(-365 * 24 * time.Hour)
		expiresAt             = time.Now().Add(+365 * 24 * time.Hour)
		formattedStartsAt, _  = startsAt.MarshalJSON()
		formattedExpiresAt, _ = expiresAt.MarshalJSON()
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
		result     interface{}
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

						Expect(part.FormName()).To(Equal("cert"))

						data, err := ioutil.ReadAll(part)
						Expect(err).To(BeNil())
						Expect(string(data)).To(Equal("certificate"))

						part, err = mr.NextPart()
						Expect(err).To(BeNil())

						Expect(part.FormName()).To(Equal("key"))

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

			ct, appErr := certs.Create("t0k3n", "foo-bar-express", "foo-bar-express.com", crtPath, keyPath)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(ct).NotTo(BeNil())
				r, ok := e.result.(*certs.Cert)
				Expect(ok).To(BeTrue())
				Expect(r.ID).To(Equal(ct.ID))
				Expect(r.CommonName).To(Equal(ct.CommonName))
				Expect(r.StartsAt.Unix()).To(Equal(ct.StartsAt.Unix()))
				Expect(r.ExpiresAt.Unix()).To(Equal(ct.ExpiresAt.Unix()))
				Expect(r.Issuer).To(Equal(ct.Issuer))
				Expect(r.Subject).To(Equal(ct.Subject))
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
			errDesc:    "cert or key file is too large",
			errIsFatal: true,
		}),

		Entry("403 with forbidden", expectation{
			resCode:    http.StatusForbidden,
			resBody:    `{"error": "invalid_request", "error_description": "not allowed to upload certs for default domain"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeNotAllowedDomain,
			errDesc:    "not allowed to upload certs for default domain",
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

		Entry("422 with invalid params due to missing cert or key", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "both cert and key are required"}`,
			errIsNil:   false,
			errCode:    certs.ErrInvalidCert,
			errDesc:    "certificate or private key is missing",
			errIsFatal: true,
		}),

		Entry("422 with invalid params due to invalid certs", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "invalid cert or key"}`,
			errIsNil:   false,
			errCode:    certs.ErrInvalidCert,
			errDesc:    "certificate or private key is not valid",
			errIsFatal: true,
		}),

		Entry("422 with invalid params due to not matched domain name", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "invalid common name (domain name mismatch)"}`,
			errIsNil:   false,
			errCode:    certs.ErrInvalidCommonName,
			errDesc:    "certificate is not valid for the specified domain",
			errIsFatal: true,
		}),

		Entry("successful creation", expectation{
			resCode: http.StatusCreated,
			resBody: fmt.Sprintf(`
				{
					"cert": {
						"id": 10,
						"starts_at": %s,
						"expires_at": %s,
						"common_name": "*.foo-bar-express.com",
						"issuer": "/C=SG/OU=NitrousCA/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com",
						"subject": "/C=SG/O=Nitrous/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com"
					}
				}
			`, formattedStartsAt, formattedExpiresAt),
			errIsNil: true,
			result: &certs.Cert{
				ID:         10,
				StartsAt:   startsAt,
				ExpiresAt:  expiresAt,
				CommonName: "*.foo-bar-express.com",
				Issuer:     "/C=SG/OU=NitrousCA/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com",
				Subject:    "/C=SG/O=Nitrous/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com",
			},
		}),
	)

	DescribeTable("Get",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/domains/foo-bar-express.com/cert"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			ct, appErr := certs.Get("t0k3n", "foo-bar-express", "foo-bar-express.com")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(ct).NotTo(BeNil())
				r, ok := e.result.(*certs.Cert)
				Expect(ok).To(BeTrue())
				Expect(r.ID).To(Equal(ct.ID))
				Expect(r.CommonName).To(Equal(ct.CommonName))
				Expect(r.StartsAt.Unix()).To(Equal(ct.StartsAt.Unix()))
				Expect(r.ExpiresAt.Unix()).To(Equal(ct.ExpiresAt.Unix()))
				Expect(r.Issuer).To(Equal(ct.Issuer))
				Expect(r.Subject).To(Equal(ct.Subject))
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
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    certs.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with cert not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "cert could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeNotFound,
			errDesc:    "cert could not be found",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("successful fetch", expectation{
			resCode: http.StatusOK,
			resBody: fmt.Sprintf(`
					{
						"cert": {
							"id": 10,
							"starts_at": %s,
							"expires_at": %s,
							"common_name": "*.foo-bar-express.com",
							"issuer": "/C=SG/OU=NitrousCA/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com",
							"subject": "/C=SG/O=Nitrous/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com"
						}
					}
				`, formattedStartsAt, formattedExpiresAt),
			errIsNil: true,
			result: &certs.Cert{
				ID:         10,
				StartsAt:   startsAt,
				ExpiresAt:  expiresAt,
				CommonName: "*.foo-bar-express.com",
				Issuer:     "/C=SG/OU=NitrousCA/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com",
				Subject:    "/C=SG/O=Nitrous/L=Singapore/ST=Singapore/CN=*.foo-bar-express.com",
			},
		}),
	)

	DescribeTable("Delete",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/projects/foo-bar-express/domains/foo-bar-express.com/cert"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := certs.Delete("t0k3n", "foo-bar-express", "foo-bar-express.com")
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
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    certs.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with cert not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "cert could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeNotFound,
			errDesc:    "cert could not be found",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("successful fetch", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"deleted": true}`,
			errIsNil: true,
		}),
	)

	DescribeTable("Enable",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/domains/foo-bar-express.com/cert/letsencrypt"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := certs.Enable("t0k3n", "foo-bar-express", "foo-bar-express.com")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(appErr.Description).To(ContainSubstring(e.errDesc))
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

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("404 with domain not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "domain could not be found"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeNotFound,
			errDesc:    "domain could not be found",
			errIsFatal: true,
		}),

		Entry("403 with default domain is already secure", expectation{
			resCode:    http.StatusForbidden,
			resBody:    `{"error": "forbidden", "error_description": "the default domain is already secure"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeNotAllowedDomain,
			errDesc:    "the default domain already supports HTTPS",
			errIsFatal: true,
		}),

		Entry("409 with certificate has already been setup", expectation{
			resCode:    http.StatusConflict,
			resBody:    `{"error": "already_exists", "error_description": "a certificate from Let's Encrypt has already been setup"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeCertExists,
			errDesc:    "a Let's Encrypt certificate has already been setup for this domain",
			errIsFatal: true,
		}),

		Entry("503 with domain could not be verified", expectation{
			resCode:    http.StatusServiceUnavailable,
			resBody:    `{"error": "service_unavailable", "error_description": "domain could not be verified"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeAcmeServerError,
			errDesc:    "domain could not be verified - have you changed its DNS configuration yet?",
			errIsFatal: true,
		}),

		Entry("503 with service unavailable", expectation{
			resCode:    http.StatusServiceUnavailable,
			resBody:    `{"error": "service_unavailable"}`,
			errIsNil:   false,
			errCode:    certs.ErrCodeAcmeServerError,
			errDesc:    "error communicating with Let's Encrypt",
			errIsFatal: true,
		}),

		Entry("successful enable", expectation{
			resCode:  http.StatusOK,
			resBody:  `{ "cert": {} }`,
			errIsNil: true,
		}),
	)
})
