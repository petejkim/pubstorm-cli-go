package password_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/password"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "password")
}

var _ = Describe("password", func() {
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
		email       string
		resetToken  string
		newPassword string

		resCode int
		resBody string

		errIsNil   bool
		errCode    string
		errDesc    string
		errIsFatal bool
	}

	DescribeTable("Forgot",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/user/password/forgot"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": {"application/x-www-form-urlencoded"},
						"Accept":       {config.ReqAccept},
						"User-Agent":   {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"email": {e.email},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := password.Forgot(e.email)
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
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"oops": }`,
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with email is required error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"email": "is required"}}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "email is required",
			errIsFatal: false,
		}),

		Entry("422 with unexpected error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params"}`,
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK but response body is empty", expectation{
			resCode:    http.StatusOK,
			resBody:    "",
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK but sent is false", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"sent": false}`,
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK and sent is true", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"sent": true}`,
			errIsNil: true,
		}),
	)

	DescribeTable("Reset",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/user/password/reset"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": {"application/x-www-form-urlencoded"},
						"Accept":       {config.ReqAccept},
						"User-Agent":   {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"email":       {e.email},
						"reset_token": {e.resetToken},
						"password":    {e.newPassword},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := password.Reset(e.email, e.resetToken, e.newPassword)
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
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"oops": }`,
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with email is required error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"email": "is required"}}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "email is required",
			errIsFatal: false,
		}),

		Entry("422 with reset_token is required error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"reset_token": "is required"}}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "reset_token is required",
			errIsFatal: false,
		}),

		Entry("422 with password is required error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"password": "is required"}}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "password is required",
			errIsFatal: false,
		}),

		Entry("422 with invalid email or reset_token", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "error_description": "invalid email or reset_token"}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "You've entered an invalid email address or password reset code",
			errIsFatal: false,
		}),

		Entry("422 with validation errors", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"password": "is too short (min. 6 characters)"}}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "Password is too short (min. 6 characters)",
			errIsFatal: false,
		}),

		Entry("422 with unexpected error", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params"}`,
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("403 Forbidden with invalid email or reset_token", expectation{
			resCode:    http.StatusForbidden,
			resBody:    `{"error": "invalid_params", "error_description": "invalid email or reset_token"}`,
			errIsNil:   false,
			errCode:    password.ErrCodeValidationFailed,
			errDesc:    "You've entered an invalid email address or password reset code",
			errIsFatal: false,
		}),

		Entry("200 OK but response body is empty", expectation{
			resCode:    http.StatusOK,
			resBody:    "",
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK but reset is false", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"reset": false}`,
			errIsNil:   false,
			errCode:    password.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("200 OK and reset is true", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"reset": true}`,
			errIsNil: true,
		}),
	)
})
