package projects_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/projects"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/project"
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
		result     interface{}
	}

	DescribeTable("Create",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
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

	DescribeTable("Get",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := projects.Get("t0k3n", "foo-bar-express")
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

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"project": {"name": "foo-bar-express" }}`,
			errIsNil: true,
		}),
	)

	DescribeTable("List",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			result, appErr := projects.List("t0k3n")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				expectedResult, ok := e.result.([]*project.Project)
				Expect(ok).To(BeTrue())
				Expect(len(result)).To(Equal(len(expectedResult)))
				for i, r := range expectedResult {
					Expect(result[i].Name).To(Equal(r.Name))
				}
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
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode: http.StatusOK,
			resBody: `{"projects": [{"name": "foo-bar-express"}, {"name": "baz-qux-entertainment"}]}`,
			result: []*project.Project{
				&project.Project{Name: "foo-bar-express"},
				&project.Project{Name: "baz-qux-entertainment"},
			},
			errIsNil: true,
		}),
	)
})
