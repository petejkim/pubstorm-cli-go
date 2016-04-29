package rawbundles_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/client/rawbundles"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "rawbundles")
}

var _ = Describe("RawBundles", func() {
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

	DescribeTable("Get",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/raw_bundles/bundl3ch3ck5um"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			bun, appErr := rawbundles.Get("t0k3n", "foo-bar-express", "bundl3ch3ck5um")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(bun).To(Equal(e.result))
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
			errCode:    rawbundles.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    rawbundles.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    rawbundles.ErrCodeProjectNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "raw bundle could not be found"}`,
			errIsNil:   false,
			errCode:    rawbundles.ErrCodeNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode:  http.StatusOK,
			resBody:  `{"raw_bundle": {"id": 123, "checksum": "bundl3ch3ck5um", "uploaded_path": "/foo/bar"}}`,
			errIsNil: true,
			result:   &rawbundles.RawBundle{ID: 123, Checksum: "bundl3ch3ck5um", UploadedPath: "/foo/bar"},
		}),
	)
})
