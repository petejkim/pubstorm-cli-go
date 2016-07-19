package repos_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/repos"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "repos")
}

var _ = Describe("Repos", func() {
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
		resCode      int
		resBody      string
		errIsNil     bool
		errCode      string
		errDesc      string
		errIsFatal   bool
		expectedRepo *repos.Repo
	}

	DescribeTable("Link",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/repos"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"uri":    {"https://github.com/github/developer.github.com.git"},
						"branch": {"gh-pages"},
						"secret": {"p@55w0rd"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			repo, appErr := repos.Link("t0k3n", "foo-bar-express", "https://github.com/github/developer.github.com.git", "gh-pages", "p@55w0rd")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(repo).To(Equal(e.expectedRepo))
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(strings.ToLower(appErr.Description)).To(ContainSubstring(strings.ToLower(e.errDesc)))
				Expect(appErr.IsFatal).To(Equal(e.errIsFatal))
			}
		},

		Entry("unexpected response code", expectation{
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    repos.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("409 with conflict", expectation{
			resCode:    http.StatusConflict,
			resBody:    `{"error": "already_exists", "error_description": "project already linked to a repository"}`,
			errIsNil:   false,
			errCode:    repos.ErrCodeAlreadyLinked,
			errDesc:    "project has already been linked to a GitHub repository",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    repos.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("422 with invalid params due to missing repository URI", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": { "uri": "is required" } }`,
			errIsNil:   false,
			errCode:    repos.ErrCodeValidationFailed,
			errDesc:    "uri is required",
			errIsFatal: true,
		}),

		Entry("successful link", expectation{
			resCode: http.StatusCreated,
			resBody: `{ "repo": {
				"project_id": 1,
				"uri": "https://github.com/github/developer.github.com.git",
				"branch": "gh-pages",
				"webhook_url": "https://api.pubstorm.com/hooks/github/deadbeeffece5",
				"webhook_secret": "p@55w0rd"
			} }`,
			errIsNil: true,
			expectedRepo: &repos.Repo{
				URI:           "https://github.com/github/developer.github.com.git",
				Branch:        "gh-pages",
				WebhookURL:    "https://api.pubstorm.com/hooks/github/deadbeeffece5",
				WebhookSecret: "p@55w0rd",
			},
		}),
	)

	DescribeTable("Unlink",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/projects/foo-bar-express/repos"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := repos.Unlink("t0k3n", "foo-bar-express")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(strings.ToLower(appErr.Description)).To(ContainSubstring(strings.ToLower(e.errDesc)))
				Expect(appErr.IsFatal).To(Equal(e.errIsFatal))
			}
		},

		Entry("unexpected response code", expectation{
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    repos.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    repos.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("404 with project not linked", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project not linked to any repository"}`,
			errIsNil:   false,
			errCode:    repos.ErrCodeNotLinked,
			errDesc:    "project not linked to any repository",
			errIsFatal: false,
		}),

		Entry("successful unlink", expectation{
			resCode:  http.StatusOK,
			resBody:  `{ "deleted": true }`,
			errIsNil: true,
		}),
	)

	DescribeTable("Info",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/repos"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			repo, appErr := repos.Info("t0k3n", "foo-bar-express")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(repo).To(Equal(e.expectedRepo))
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(strings.ToLower(appErr.Description)).To(ContainSubstring(strings.ToLower(e.errDesc)))
				Expect(appErr.IsFatal).To(Equal(e.errIsFatal))
			}
		},

		Entry("unexpected response code", expectation{
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    repos.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    repos.ErrCodeProjectNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("404 with project not linked", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project not linked to any repository"}`,
			errIsNil:   false,
			errCode:    repos.ErrCodeNotLinked,
			errDesc:    "project not linked to any repository",
			errIsFatal: false,
		}),

		Entry("success", expectation{
			resCode: http.StatusOK,
			resBody: `{ "repo": {
				"project_id": 1,
				"uri": "https://github.com/github/developer.github.com.git",
				"branch": "gh-pages",
				"webhook_url": "https://api.pubstorm.com/hooks/github/deadbeeffece5",
				"webhook_secret": "p@55w0rd"
			} }`,
			errIsNil: true,
			expectedRepo: &repos.Repo{
				URI:           "https://github.com/github/developer.github.com.git",
				Branch:        "gh-pages",
				WebhookURL:    "https://api.pubstorm.com/hooks/github/deadbeeffece5",
				WebhookSecret: "p@55w0rd",
			},
		}),
	)
})
