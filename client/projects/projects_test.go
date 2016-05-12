package projects_test

import (
	"net/http"
	"net/url"
	"strconv"
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
		proj    *project.Project
		resCode int
		resBody string

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

			result, appErr := projects.Get("t0k3n", "foo-bar-express")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(result).To(Equal(e.result))
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
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeNotFound,
			errDesc:    `Could not find a project "foo-bar-express" that belongs to you.`,
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"project": {"name": "foo-bar-express", "default_domain_enabled": true }}`,
			errIsNil: true,
			result:   &project.Project{Name: "foo-bar-express", DefaultDomainEnabled: true},
		}),
	)

	DescribeTable("Index",
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

			projs, sharedProjs, appErr := projects.Index("t0k3n")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				expectedResult, ok := e.result.(map[string][]*project.Project)
				Expect(ok).To(BeTrue())

				expectedProjs, ok := expectedResult["projects"]
				Expect(ok).To(BeTrue())
				Expect(len(projs)).To(Equal(len(expectedProjs)))
				for i, p := range expectedProjs {
					Expect(projs[i].Name).To(Equal(p.Name))
				}

				expectedSharedProjs, ok := expectedResult["shared_projects"]
				Expect(ok).To(BeTrue())
				Expect(len(sharedProjs)).To(Equal(len(expectedSharedProjs)))
				for i, p := range expectedSharedProjs {
					Expect(sharedProjs[i].Name).To(Equal(p.Name))
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
			result: map[string][]*project.Project{
				"projects": []*project.Project{
					&project.Project{Name: "foo-bar-express"},
					&project.Project{Name: "baz-qux-entertainment"},
				},
				"shared_projects": nil,
			},
			errIsNil: true,
		}),

		Entry("successfully fetched with shared projects", expectation{
			resCode: http.StatusOK,
			resBody: `{
				"projects": [
					{"name": "foo-bar-express"},
					{"name": "baz-qux-entertainment"}
				],
				"shared_projects": [
					{"name": "nestorbot"}
				]
			}`,
			result: map[string][]*project.Project{
				"projects": []*project.Project{
					&project.Project{Name: "foo-bar-express"},
					&project.Project{Name: "baz-qux-entertainment"},
				},
				"shared_projects": []*project.Project{
					&project.Project{Name: "nestorbot"},
				},
			},
			errIsNil: true,
		}),
	)

	DescribeTable("Update",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/projects/"+e.proj.Name),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"default_domain_enabled": {
							strconv.FormatBool(e.proj.DefaultDomainEnabled),
						},
						"force_https": {
							strconv.FormatBool(e.proj.ForceHTTPS),
						},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			result, appErr := projects.Update("t0k3n", e.proj)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(result).To(Equal(e.result))
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(strings.ToLower(appErr.Description)).To(ContainSubstring(strings.ToLower(e.errDesc)))
				Expect(appErr.IsFatal).To(Equal(e.errIsFatal))
			}
		},

		Entry("unexpected response code", expectation{
			proj:       &project.Project{Name: "foo-bar-express"},
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			proj:       &project.Project{Name: "foo-bar-express"},
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    projects.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful update", expectation{
			proj:     &project.Project{Name: "foo-bar-express"},
			resCode:  http.StatusOK,
			resBody:  `{"project": { "name": "foo-bar-express", "default_domain_enabled": true, "force_https": false }}`,
			errIsNil: true,
			result:   &project.Project{Name: "foo-bar-express", DefaultDomainEnabled: true},
		}),

		Entry("successful update to set default domain enabled to true", expectation{
			proj:     &project.Project{Name: "foo-bar-express", DefaultDomainEnabled: false},
			resCode:  http.StatusOK,
			resBody:  `{"project": { "name": "foo-bar-express", "default_domain_enabled": true, "force_https": false }}`,
			errIsNil: true,
			result:   &project.Project{Name: "foo-bar-express", DefaultDomainEnabled: true},
		}),

		Entry("successful update to set default domain enabled to false", expectation{
			proj:     &project.Project{Name: "foo-bar-express", DefaultDomainEnabled: true},
			resCode:  http.StatusOK,
			resBody:  `{"project": { "name": "foo-bar-express", "default_domain_enabled": false, "force_https": false }}`,
			errIsNil: true,
			result:   &project.Project{Name: "foo-bar-express", DefaultDomainEnabled: false},
		}),

		Entry("successful update to set force https to true", expectation{
			proj:     &project.Project{Name: "foo-bar-express", ForceHTTPS: false},
			resCode:  http.StatusOK,
			resBody:  `{"project": { "name": "foo-bar-express", "default_domain_enabled": false, "force_https": true }}`,
			errIsNil: true,
			result:   &project.Project{Name: "foo-bar-express", ForceHTTPS: true},
		}),

		Entry("successful update to set force https to false", expectation{
			proj:     &project.Project{Name: "foo-bar-express", ForceHTTPS: true},
			resCode:  http.StatusOK,
			resBody:  `{"project": { "name": "foo-bar-express", "default_domain_enabled": false, "force_https": false }}`,
			errIsNil: true,
			result:   &project.Project{Name: "foo-bar-express", ForceHTTPS: false},
		}),
	)

	DescribeTable("Delete",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/projects/foo-bar-express"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := projects.Delete("t0k3n", "foo-bar-express")
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

		Entry("404 with not found error", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "errors_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    projects.ErrCodeNotFound,
			errDesc:    `Could not find a project "foo-bar-express" that belongs to you.`,
			errIsFatal: true,
		}),

		Entry("successful deletion", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"deleted": true}`,
			errIsNil: true,
		}),
	)
})
