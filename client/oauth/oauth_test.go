package oauth_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/oauth"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "oauth")
}

var _ = Describe("OAuth", func() {
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
		tokenRecvd string
	}

	DescribeTable("FetchToken",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.VerifyBasicAuth(config.ClientID, config.ClientSecret),
					ghttp.VerifyHeader(http.Header{
						"Accept":       {"application/vnd.rise.v0+json"},
						"Content-Type": {"application/x-www-form-urlencoded"},
						"User-Agent":   {"RiseCLI"},
					}),
					ghttp.VerifyForm(url.Values{
						"grant_type": {"password"},
						"username":   {"foo@example.com"},
						"password":   {"p@55w0rd"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			token, appErr := oauth.FetchToken("foo@example.com", "p@55w0rd")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
			} else {
				Expect(appErr).NotTo(BeNil())
				Expect(appErr.Code).To(Equal(e.errCode))
				Expect(appErr.Description).To(ContainSubstring(e.errDesc))
				Expect(appErr.IsFatal).To(Equal(e.errIsFatal))
			}

			Expect(token).To(Equal(e.tokenRecvd))
		},

		Entry("unexpected response code", expectation{
			resCode:    http.StatusInternalServerError,
			resBody:    "",
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("400 with invalid user credentials", expectation{
			resCode:    http.StatusBadRequest,
			resBody:    `{"error": "invalid_grant", "error_description": "user credentials are invalid"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeInvalidGrant,
			errDesc:    "invalid email or password",
			errIsFatal: false,
		}),

		Entry("400 with unconfirmed user error", expectation{
			resCode:    http.StatusBadRequest,
			resBody:    `{"error": "invalid_grant", "error_description": "user has not confirmed email address"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnconfirmedEmail,
			errDesc:    "user has not confirmed email address",
			errIsFatal: false,
		}),

		Entry("400 with unexpected error", expectation{
			resCode:    400,
			resBody:    `{"error": "something_weng_wrong"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("status is ok but no token", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"site": "is_hacked", "result": "DROP TABLE *;"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("unexpected token type", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"access_token": "myawes0met0ken", "token_type": "nonsense", "client_id": "cafebabe"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("bearer token granted", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"access_token": "myawes0met0ken", "token_type": "bearer", "client_id": "cafebabe"}`,
			errIsNil:   true,
			tokenRecvd: "myawes0met0ken",
		}),
	)

	DescribeTable("InvalidateToken",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/oauth/token"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {"application/vnd.rise.v0+json"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"User-Agent":    {"RiseCLI"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			appErr := oauth.InvalidateToken("t0k3n")
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
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"invalidated": }`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("401 with token required", expectation{
			resCode:    http.StatusUnauthorized,
			resBody:    `{"error": "invalid_token", "error_description": "access token is required"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("401 with token invalid", expectation{
			resCode:    http.StatusUnauthorized,
			resBody:    `{"error": "invalid_token", "error_description": "access token is invalid"}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeInvalidAuthorization,
			errDesc:    "invalid access token",
			errIsFatal: false,
		}),

		Entry("status is ok but token is not invalidated", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"invalidated": false}`,
			errIsNil:   false,
			errCode:    oauth.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("token invalidated", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"invalidated": true}`,
			errIsNil: true,
		}),
	)
})
