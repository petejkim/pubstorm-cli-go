package projects_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "projects")
}

var _ = Describe("Projects", func() {
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
					ghttp.VerifyRequest("POST", "/projects"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {"application/vnd.rise.v0+json"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"User-Agent":    {"RiseCLI"},
					}),
					ghttp.VerifyForm(url.Values{
						"name": {"foo-bar-express"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := projects.Create("t0k3n", "foo-bar-express")
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

		Entry("422 with name is taken error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"name": "is taken"}}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeValidationFailed,
			errDesc:    "name is taken",
			errIsFatal: false,
		}),

		Entry("422 with validation errors", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"name": "is invalid"}}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeValidationFailed,
			errDesc:    "name is invalid",
			errIsFatal: false,
		}),

		Entry("422 with unexpected error", expectation{
			resCode:    422,
			resBody:    `{"error": "something_weng_wrong"}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful creation", expectation{
			resCode:  http.StatusCreated,
			resBody:  `{"project": { "name": "foo-bar-express" }}`,
			errIsNil: true,
		}),
	)
})
