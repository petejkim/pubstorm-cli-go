package domains_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/domains"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "domains")
}

var _ = Describe("Domains", func() {
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
		result     []string
	}

	DescribeTable("Index",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/domains"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			domainNames, appErr := domains.Index("t0k3n", "foo-bar-express")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(domainNames).To(Equal(e.result))
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
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"domains": [ "foo-bar-express.pubstorm.site", "foo-bar-express.com" ]}`,
			errIsNil: true,
			result:   []string{"foo-bar-express.pubstorm.site", "foo-bar-express.com"},
		}),
	)

	DescribeTable("Create",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/domains"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"name": {"foo-bar-express.com"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := domains.Create("t0k3n", "foo-bar-express", "foo-bar-express.com")
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
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("422 with name is invalid", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"name": "is invalid"}}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeValidationFailed,
			errDesc:    "name is invalid",
			errIsFatal: false,
		}),

		Entry("422 with no more domains error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_request", "error_description": "project cannot have more domains"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeLimitReached,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with unexpected error", expectation{
			resCode:    422,
			resBody:    `{"error": "something_weng_wrong"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("423 with locked", expectation{
			resCode:    423,
			resBody:    `{"error": "locked", "error_description": "project is locked"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeProjectLocked,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful creation", expectation{
			resCode:  http.StatusCreated,
			resBody:  `{"domain": { "name": "www.foo-bar-express.com" }}`,
			errIsNil: true,
		}),
	)

	DescribeTable("Delete",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/projects/foo-bar-express/domains/foo-bar-express.com"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := domains.Delete("t0k3n", "foo-bar-express", "foo-bar-express.com")
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
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    domains.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("423 with locked", expectation{
			resCode:    423,
			resBody:    `{"error": "locked", "error_description": "project is locked"}`,
			errIsNil:   false,
			errCode:    domains.ErrCodeProjectLocked,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successfully deleted", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"deleted": true}`,
			errIsNil: true,
		}),
	)
})
