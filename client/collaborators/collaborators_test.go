package collaborators_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/collaborators"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "collaborators")
}

var _ = Describe("Collaborators", func() {
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
		email string

		resCode int
		resBody string

		errIsNil   bool
		errCode    string
		errDesc    string
		errIsFatal bool
		result     interface{}
	}

	DescribeTable("List",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/orient-express/collaborators"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			cols, appErr := collaborators.List("t0k3n", "orient-express")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(cols).To(Equal(e.result))
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
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("200 OK with collaborator list", expectation{
			resCode: http.StatusOK,
			resBody: `{"collaborators": [
				{ "email": "steve@apple.com", "name": "", "organization": "NeXT" },
				{ "email": "woz@apple.com", "name": "Woz", "organization": "Apple" }
			]}`,
			errIsNil: true,
			result: []*collaborators.Collaborator{
				{Email: "steve@apple.com", Name: "", Organization: "NeXT"},
				{Email: "woz@apple.com", Name: "Woz", Organization: "Apple"},
			},
		}),
	)

	DescribeTable("Add",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/orient-express/collaborators"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"email": {e.email},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := collaborators.Add("t0k3n", "orient-express", e.email)
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
			email:      "steve@apple.com",
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("409 Conflict", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusConflict,
			resBody:    `{"error": "already_exists", "error_description": "user is already a collaborator"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeAlreadyExists,
			errDesc:    "user is already a collaborator",
			errIsFatal: false,
		}),

		Entry("422 with email is not found", expectation{
			email:      "steve@apple.com",
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "email is not found"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUserNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with owner cannot be added", expectation{
			email:      "steve@apple.com",
			resCode:    422,
			resBody:    `{"error": "invalid_request", "error_description": "the owner of a project cannot be added as a collaborator"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeCannotAddOwner,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with unexpected error", expectation{
			email:      "steve@apple.com",
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "this is unexpected"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("201 Created but response body is empty", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusCreated,
			resBody:    "",
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("201 Created but added is false", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusCreated,
			resBody:    `{"added": false}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("201 Created and added is true", expectation{
			email:    "steve@apple.com",
			resCode:  http.StatusCreated,
			resBody:  `{"added": true}`,
			errIsNil: true,
		}),
	)

	DescribeTable("Remove",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/projects/orient-express/collaborators/"+e.email),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := collaborators.Remove("t0k3n", "orient-express", e.email)
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
			email:      "steve@apple.com",
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("404 with email is not found", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "email is not found"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUserNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with unexpected error", expectation{
			email:      "steve@apple.com",
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "this is unexpected"}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK but response body is empty", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusOK,
			resBody:    "",
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK but removed is false", expectation{
			email:      "steve@apple.com",
			resCode:    http.StatusOK,
			resBody:    `{"removed": false}`,
			errIsNil:   false,
			errCode:    collaborators.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK and removed is true", expectation{
			email:    "steve@apple.com",
			resCode:  http.StatusOK,
			resBody:  `{"removed": true}`,
			errIsNil: true,
		}),
	)
})
