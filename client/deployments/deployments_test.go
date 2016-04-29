package deployments_test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "deployments")
}

var _ = Describe("Deployments", func() {
	var (
		origHost string
		server   *ghttp.Server

		deployedTime          = time.Now()
		formattedTimeBytes, _ = deployedTime.MarshalJSON()
		formattedTime         = string(formattedTimeBytes)
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
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/deployments"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),

					func(w http.ResponseWriter, req *http.Request) {
						key := http.CanonicalHeaderKey("Content-Type")
						Expect(req.Header[key][0]).To(HavePrefix("multipart/form-data"))
					},

					func(w http.ResponseWriter, req *http.Request) {
						mr, err := req.MultipartReader()
						Expect(err).To(BeNil())

						part, err := mr.NextPart()
						Expect(err).To(BeNil())

						Expect(part.FormName()).To(Equal("payload"))

						data, err := ioutil.ReadAll(part)
						Expect(err).To(BeNil())
						Expect(string(data)).To(Equal("my-bundle-yo"))
					},

					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			tempDir, err := ioutil.TempDir("", "rise-test")
			Expect(err).To(BeNil())
			defer func() {
				os.RemoveAll(tempDir)
			}()

			bunPath := filepath.Join(tempDir, "bundle.tar.gz")
			err = ioutil.WriteFile(bunPath, []byte("my-bundle-yo"), 0600)
			Expect(err).To(BeNil())

			defer func() {
				os.Remove(bunPath)
			}()

			deployment, appErr := deployments.Create("t0k3n", "foo-bar-express", bunPath, true)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				depl, ok := e.result.(*deployments.Deployment)
				Expect(ok).To(BeTrue())
				Expect(deployment.ID).To(Equal(depl.ID))
				Expect(deployment.State).To(Equal(depl.State))
				Expect(deployment.DeployedAt.Unix()).To(Equal(depl.DeployedAt.Unix()))
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
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("400 with bad request", expectation{
			resCode:    http.StatusBadRequest,
			resBody:    `{"error": "invalid_request", "error_description": "request body is too large"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeValidationFailed,
			errDesc:    "project size is too large",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeNotFound,
			errDesc:    "project could not be found",
			errIsFatal: true,
		}),

		Entry("423 with locked", expectation{
			resCode:    423,
			resBody:    `{"error": "locked", "error_description": "project is locked"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeProjectLocked,
			errDesc:    "project is locked",
			errIsFatal: true,
		}),

		Entry("successful deployment", expectation{
			resCode:  http.StatusAccepted,
			resBody:  `{"deployment": {"id": 123, "state": "deployed", "deployed_at": ` + formattedTime + `}}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 123, State: "deployed", DeployedAt: deployedTime},
		}),
	)

	DescribeTable("CreateWithChecksum",
		func(e expectation) {
			checksum := "bundl3ch3ck5um"

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/deployments"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Content-Type":  {"application/x-www-form-urlencoded"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.VerifyForm(url.Values{
						"bundle_checksum": {checksum},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			deployment, appErr := deployments.CreateWithChecksum("t0k3n", "foo-bar-express", checksum)
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
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusCreated,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeProjectNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("422 with not found", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_params", "errors": {"bundle_checksum": "the bundle could not be found"}}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeRawBundleNotFound,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("423 with locked", expectation{
			resCode:    423,
			resBody:    `{"error": "locked", "error_description": "project is locked"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeProjectLocked,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("successful deployment", expectation{
			resCode:  http.StatusAccepted,
			resBody:  `{"deployment": {"id": 123, "state": "pending_deployed"}}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 123, State: "pending_deployed"},
		}),
	)

	DescribeTable("Get",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/deployments/123"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			deployment, appErr := deployments.Get("t0k3n", "foo-bar-express", 123)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				depl, ok := e.result.(*deployments.Deployment)
				Expect(ok).To(BeTrue())
				Expect(deployment.ID).To(Equal(depl.ID))
				Expect(deployment.State).To(Equal(depl.State))
				Expect(deployment.DeployedAt.Unix()).To(Equal(depl.DeployedAt.Unix()))
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
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "deployment could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeNotFound,
			errDesc:    "deployment could not be found",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode: http.StatusOK,
			resBody: `
				{
					"deployment": {
						"id": 123,
					  "state": "deployed",
						"deployed_at": ` + formattedTime + `,
						"error_message": "index.html:Unexpected Tag\napp.json:undefined is undefined"
					}
				}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 123, State: "deployed", DeployedAt: deployedTime, ErrorMessage: "index.html:Unexpected Tag\napp.json:undefined is undefined"},
		}),
	)

	DescribeTable("Rollback",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/projects/foo-bar-express/rollback"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
						"Content-Type":  {"application/x-www-form-urlencoded"},
					}),
					ghttp.VerifyForm(url.Values{
						"version": {"12"},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			deployment, appErr := deployments.Rollback("t0k3n", "foo-bar-express", 12)
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				depl, ok := e.result.(*deployments.Deployment)
				Expect(ok).To(BeTrue())
				Expect(deployment.ID).To(Equal(depl.ID))
				Expect(deployment.State).To(Equal(depl.State))
				Expect(deployment.DeployedAt.Unix()).To(Equal(depl.DeployedAt.Unix()))
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
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with project not found error", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeProjectNotFound,
			errIsFatal: true,
		}),

		Entry("412 with active deployment not found error", expectation{
			resCode:    http.StatusPreconditionFailed,
			resBody:    `{"error": "precondition_failed", "error_description": "active deployment could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeNotFound,
			errDesc:    "does not have any completed deployment",
			errIsFatal: true,
		}),

		Entry("412 with no previous deployment found error", expectation{
			resCode:    http.StatusPreconditionFailed,
			resBody:    `{"error": "precondition_failed", "error_description": "previous completed deployment could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeNotFound,
			errDesc:    "no previous version",
			errIsFatal: true,
		}),

		Entry("422 with deployment could not be found", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_request", "error_description": "completed deployment with a given version could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeValidationFailed,
			errDesc:    "v12 could not be found",
			errIsFatal: true,
		}),

		Entry("422 with invalid_request", expectation{
			resCode:    422,
			resBody:    `{"error": "invalid_request", "error_description": "the specified deployment is already active"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeValidationFailed,
			errDesc:    "already on v12",
			errIsFatal: true,
		}),

		Entry("423 with locked", expectation{
			resCode:    423,
			resBody:    `{"error": "locked", "error_description": "project is locked"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeProjectLocked,
			errDesc:    "locked",
			errIsFatal: true,
		}),

		Entry("rollback accepted", expectation{
			resCode:  http.StatusAccepted,
			resBody:  `{"deployment": {"id": 123, "state": "deployed", "deployed_at": ` + formattedTime + `, "version": 13}}`,
			errIsNil: true,
			result:   &deployments.Deployment{ID: 123, State: "deployed", DeployedAt: deployedTime, Version: 13},
		}),
	)

	DescribeTable("List",
		func(e expectation) {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/projects/foo-bar-express/deployments"),
					ghttp.VerifyHeader(http.Header{
						"Authorization": {"Bearer t0k3n"},
						"Accept":        {config.ReqAccept},
						"User-Agent":    {config.UserAgent},
					}),
					ghttp.RespondWith(e.resCode, e.resBody),
				),
			)

			depls, appErr := deployments.List("t0k3n", "foo-bar-express")
			Expect(server.ReceivedRequests()).To(HaveLen(1))

			if e.errIsNil {
				Expect(appErr).To(BeNil())
				Expect(depls).To(HaveLen(2))
				expectedDepls, ok := e.result.([]deployments.Deployment)
				Expect(ok).To(BeTrue())

				Expect(depls[0].ID).To(Equal(expectedDepls[0].ID))
				Expect(depls[0].State).To(Equal(expectedDepls[0].State))
				Expect(depls[0].DeployedAt.Unix()).To(Equal(expectedDepls[0].DeployedAt.Unix()))
				Expect(depls[1].ID).To(Equal(expectedDepls[1].ID))
				Expect(depls[1].State).To(Equal(expectedDepls[1].State))
				Expect(depls[1].DeployedAt.Unix()).To(Equal(expectedDepls[1].DeployedAt.Unix()))
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
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("malformed json", expectation{
			resCode:    http.StatusOK,
			resBody:    `{"foo": }`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeUnexpectedError,
			errDesc:    "",
			errIsFatal: true,
		}),

		Entry("404 with not found", expectation{
			resCode:    http.StatusNotFound,
			resBody:    `{"error": "not_found", "error_description": "project could not be found"}`,
			errIsNil:   false,
			errCode:    deployments.ErrCodeProjectNotFound,
			errDesc:    "Could not find a project \"foo-bar-express\" that belongs to you.",
			errIsFatal: true,
		}),

		Entry("successfully fetched", expectation{
			resCode: http.StatusOK,
			resBody: `{"deployments": [
			  {
					"id": 123,
					"state": "deployed",
					"active": true,
					"deployed_at": ` + formattedTime + `
				},
			  {
					"id": 234,
					"state": "deployed",
					"deployed_at": ` + formattedTime + `
				}
		  ]}`,
			errIsNil: true,
			result: []deployments.Deployment{
				deployments.Deployment{ID: 123, State: "deployed", DeployedAt: deployedTime},
				deployments.Deployment{ID: 234, State: "deployed", DeployedAt: deployedTime},
			},
		}),
	)
})
