package deployments

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/progressbar"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"
)

// Error codes for apperror.Error returned from this package.
const (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeValidationFailed = "validation_failed"
	ErrCodeNotFound         = "not_found"
)

// Deployment states.
const (
	DeploymentStateDeployed = "deployed"
)

// Deployment is a deployment of a project.
type Deployment struct {
	ID    uint   `json:"id"`
	State string `json:"state"`
}

// Create makes an API request to deploy a project.
func Create(token, name, bunPath string, quiet bool) (depl *Deployment, appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "POST",
		Uri:       config.Host + "/projects/" + name + "/deployments",
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	f, err := os.Open(bunPath)
	if err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("payload", filepath.Base(bunPath))
	if err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if _, err = io.Copy(part, f); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if err = writer.Close(); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	req.AddHeader("Content-Type", writer.FormDataContentType())
	bodyLen := int64(body.Len())

	if quiet {
		req.Body = body
	} else {
		pb := progressbar.NewReader(body, tui.Out, bodyLen)
		req.Body = pb
	}

	req.OnBeforeRequest = func(goreq *goreq.Request, httpreq *http.Request) {
		httpreq.ContentLength = bodyLen
	}

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusAccepted, http.StatusBadRequest, http.StatusNotFound}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusAccepted {
		var j struct {
			Deployment Deployment `json:"deployment"`
		}

		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		return &j.Deployment, nil
	} else if res.StatusCode == http.StatusBadRequest {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		if j["error"] == "invalid_request" {
			switch j["error_description"] {
			case "request body is too large":
				return nil, apperror.New(ErrCodeValidationFailed, nil, "project size is too large", true)
			}
		}
	} else if res.StatusCode == http.StatusNotFound {
		return nil, apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	}

	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}

// Get makes an API request to get a deployment's information.
func Get(token, projectName string, deploymentID uint) (depl *Deployment, appErr *apperror.Error) {
	uri := fmt.Sprintf("%s/projects/%s/deployments/%d", config.Host, projectName, deploymentID)
	req := goreq.Request{
		Method:    "GET",
		Uri:       uri,
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusOK, http.StatusNotFound}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusOK {
		var j struct {
			Deployment *Deployment `json:"deployment"`
		}

		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		return j.Deployment, nil
	} else if res.StatusCode == http.StatusNotFound {
		return nil, apperror.New(ErrCodeNotFound, nil, "deployment could not be found", true)
	}

	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}
