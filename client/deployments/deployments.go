package deployments

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/progressbar"
	"github.com/nitrous-io/rise-cli-go/tr"
	"github.com/nitrous-io/rise-cli-go/tui"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed     = "request_failed"
	ErrCodeUnexpectedError   = "unexpected_error"
	ErrCodeValidationFailed  = "validation_failed"
	ErrCodeProjectNotFound   = "project_not_found"
	ErrCodeNotFound          = "not_found"
	ErrCodeProjectLocked     = "project_locked"
	ErrCodeRawBundleNotFound = "raw_bundle_not_found"

	DeploymentStateDeployed     = "deployed"
	DeploymentStateBuilding     = "pending_build"
	DeploymentStateDeploying    = "pending_deploy"
	DeploymentStateDeployFailed = "deploy_failed"
)

type Deployment struct {
	ID           uint      `json:"id"`
	State        string    `json:"state"`
	Active       bool      `json:"active,omitempty"`
	DeployedAt   time.Time `json:"deployed_at,omitempty"`
	Version      int64     `json:"version"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

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

	if !util.ContainsInt([]int{http.StatusAccepted, http.StatusBadRequest, http.StatusNotFound, 423}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	switch res.StatusCode {
	case http.StatusAccepted:
		var j struct {
			Deployment Deployment `json:"deployment"`
		}

		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		return &j.Deployment, nil
	case http.StatusBadRequest:
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
	case http.StatusNotFound:
		return nil, apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	case 423:
		return nil, apperror.New(ErrCodeProjectLocked, nil, "project is locked", true)
	}

	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}

func CreateWithChecksum(token, name, checksum string) (depl *Deployment, appErr *apperror.Error) {
	req := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/projects/" + name + "/deployments",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,
		Body: url.Values{
			"bundle_checksum": {checksum},
		}.Encode(),
	}
	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusAccepted, http.StatusNotFound, 422, 423}, res.StatusCode) {
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
	} else if res.StatusCode == http.StatusNotFound {
		return nil, apperror.New(ErrCodeProjectNotFound, nil, "", true)
	} else if res.StatusCode == 422 {
		return nil, apperror.New(ErrCodeRawBundleNotFound, nil, "", true)
	} else if res.StatusCode == 423 {
		return nil, apperror.New(ErrCodeProjectLocked, nil, "", true)
	}

	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}

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

func Rollback(token, projectName string, version int64) (depl *Deployment, appErr *apperror.Error) {
	req := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/projects/" + projectName + "/rollback",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	if version != 0 {
		req.Body = url.Values{"version": {strconv.FormatInt(version, 10)}}.Encode()
	}

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusAccepted, http.StatusNotFound, http.StatusPreconditionFailed, 422, 423}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 423 {
		return nil, apperror.New(ErrCodeProjectLocked, err, tr.T("project_locked"), true)
	}

	if res.StatusCode != http.StatusAccepted {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		if errDesc, ok := j["error_description"].(string); ok {
			switch errDesc {
			case "active deployment could not be found":
				return nil, apperror.New(ErrCodeNotFound, err, tr.T("rollback_no_active_deployment"), true)
			case "previous completed deployment could not be found":
				return nil, apperror.New(ErrCodeNotFound, err, tr.T("rollback_no_previous_version"), true)
			case "project could not be found":
				return nil, apperror.New(ErrCodeProjectNotFound, err, fmt.Sprintf(tr.T("project_not_found"), projectName), true)
			case "completed deployment with a given version could not be found":
				return nil, apperror.New(ErrCodeValidationFailed, err, fmt.Sprintf(tr.T("rollback_version_not_found"), version), true)
			case "the specified deployment is already active":
				return nil, apperror.New(ErrCodeValidationFailed, err, fmt.Sprintf(tr.T("rollback_version_already_active"), version), true)
			}
		}
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var d struct {
		Deployment *Deployment `json:"deployment"`
	}

	if err := res.Body.FromJsonTo(&d); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return d.Deployment, nil
}

func List(token, projectName string) (depls []Deployment, appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "GET",
		Uri:       config.Host + "/projects/" + projectName + "/deployments",
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
			Deployments []Deployment `json:"deployments"`
		}

		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		return j.Deployments, nil
	} else if res.StatusCode == http.StatusNotFound {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		if errDesc, ok := j["error_description"].(string); ok {
			if errDesc == "project could not be found" {
				return nil, apperror.New(ErrCodeProjectNotFound, err, fmt.Sprintf(tr.T("project_not_found"), projectName), true)
			}
		}
		return nil, apperror.New(ErrCodeProjectNotFound, err, fmt.Sprintf(tr.T("project_not_found"), projectName), true)
	}

	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}
