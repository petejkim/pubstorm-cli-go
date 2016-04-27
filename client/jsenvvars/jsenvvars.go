package jsenvvars

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/client/deployments"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed            = "request_failed"
	ErrCodeUnexpectedError          = "unexpected_error"
	ErrCodeProjectNotFound          = "project_not_found"
	ErrCodeActiveDeploymentNotFound = "active_deployment_not_found"
)

func Add(token, projectName string, vars map[string]string) (d *deployments.Deployment, appErr *apperror.Error) {
	req := goreq.Request{
		Method:      "PUT",
		Uri:         config.Host + "/projects/" + projectName + "/jsenvvars/add",
		ContentType: "application/json",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,
	}

	req.AddHeader("Authorization", "Bearer "+token)

	b, err := json.Marshal(vars)
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	req.Body = b

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusAccepted, http.StatusNotFound, http.StatusPreconditionFailed}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		return nil, apperror.New(ErrCodeProjectNotFound, nil, "", true)
	case http.StatusPreconditionFailed:
		return nil, apperror.New(ErrCodeActiveDeploymentNotFound, nil, "", true)
	}

	var j struct {
		Deployment deployments.Deployment `json:"deployment"`
	}

	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return &j.Deployment, nil
}

func Delete(token, projectName string, keys []string) (d *deployments.Deployment, appErr *apperror.Error) {
	req := goreq.Request{
		Method:      "PUT",
		Uri:         config.Host + "/projects/" + projectName + "/jsenvvars/delete",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"keys": keys,
		}.Encode(),
	}

	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusAccepted, http.StatusNotFound, http.StatusPreconditionFailed}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		return nil, apperror.New(ErrCodeProjectNotFound, nil, "", true)
	case http.StatusPreconditionFailed:
		return nil, apperror.New(ErrCodeActiveDeploymentNotFound, nil, "", true)
	}

	var j struct {
		Deployment deployments.Deployment `json:"deployment"`
	}

	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return &j.Deployment, nil
}
