package users_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/users"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "users")
}

var _ = Describe("Users", func() {
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
					ghttp.VerifyRequest("POST", "/users"),
					ghttp.VerifyHeader(http.Header{
						"Accept":       {"application/vnd.rise.v0+json"},
						"Content-Type": {"application/x-www-form-urlencoded"},
						"User-Agent":   {"RiseCLI"},
					}),
					ghttp.VerifyForm(url.Values{
						"email":    {"foo@example.com"},
						"password": {"p@55w0rd"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := users.Create("foo@example.com", "p@55w0rd")
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
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with email is taken error)", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"email": "is taken"}}`,
			errIsNil:   false,
			errCode:    users.ErrCodeValidationFailed,
			errDesc:    "email is taken",
			errIsFatal: false,
		}),

		Entry("422 with validation errors", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"password": "is invalid"}}`,
			errIsNil:   false,
			errCode:    users.ErrCodeValidationFailed,
			errDesc:    "password is invalid",
			errIsFatal: false,
		}),

		Entry("422 with unexpected error", expectation{
			resCode:    422,
			resBody:    `{"error": "something_weng_wrong"}`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful creation", expectation{
			resCode:  http.StatusCreated,
			resBody:  `{"user": { "email": "foo@example.com", "name": "", "organization": ""}}`,
			errIsNil: true,
		}),
	)

	DescribeTable("Confirm",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/user/confirm"),
					ghttp.VerifyHeader(http.Header{
						"Accept":       {"application/vnd.rise.v0+json"},
						"Content-Type": {"application/x-www-form-urlencoded"},
						"User-Agent":   {"RiseCLI"},
					}),
					ghttp.VerifyForm(url.Values{
						"email":             {"foo@example.com"},
						"confirmation_code": {"123456"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := users.Confirm("foo@example.com", "123456")
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
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with invalid code error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "invalid email or confirmation_code"}`,
			errIsNil:   false,
			errCode:    users.ErrCodeValidationFailed,
			errDesc:    "incorrect confirmation code",
			errIsFatal: false,
		}),

		Entry("422 with unexpected error", expectation{
			resCode:    422,
			resBody:    `{"error": "something_weng_wrong"}`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("status is OK, but somehow confirmed is false", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"confirmed": false}`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful confirmation", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"confirmed": true}`,
			errIsNil: true,
		}),
	)

	DescribeTable("ResendConfirmationCode",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/user/confirm/resend"),
					ghttp.VerifyHeader(http.Header{
						"Accept":       {"application/vnd.rise.v0+json"},
						"Content-Type": {"application/x-www-form-urlencoded"},
						"User-Agent":   {"RiseCLI"},
					}),
					ghttp.VerifyForm(url.Values{
						"email": {"foo@example.com"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := users.ResendConfirmationCode("foo@example.com")
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
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("invalid params", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "email is not found or already confirmed"}`,
			errIsNil:   false,
			errCode:    users.ErrCodeValidationFailed,
			errDesc:    "already confirmed",
			errIsFatal: true,
		}),

		Entry("status is OK, but somehow sent is false", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"sent": false}`,
			errIsNil:   false,
			errCode:    users.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful confirmation", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"sent": true}`,
			errIsNil: true,
		}),
	)
})
