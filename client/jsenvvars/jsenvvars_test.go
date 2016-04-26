package jsenvvars_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/client/jsenvvars"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "jsenvvars")
}

var _ = Describe("JsEnvVars", func() {
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

	DescribeTable("Add",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/projects/foo-bar-express/jsenvvars/add"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/json"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.VerifyJSON(`{"foo":"bar"}`),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			deployment, appErr := jsenvvars.Add("t0k3n", "foo-bar-express", map[string]string{
				"foo": "bar",
			})
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(deployment).To(Equal(e.result))
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
			errCode:    jsenvvars.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusAccepted,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    jsenvvars.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    jsenvvars.ErrCodeProjectNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("412 with precondition failed", expectation{
			resCode:    http.StatusPreconditionFailed,
			resBody:    `{"error": "not_found", "error_description": "current active deployment could not be found"}`,
			errIsNil:   false,
			errCode:    jsenvvars.ErrCodeActiveDeploymentNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful adding a variable", expectation{
			resCode:  http.StatusAccepted,
			resBody:  `{"deployment": {"id": 10, "state": "uploaded" }}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 10, State: "uploaded"},
		}),
	)
})
