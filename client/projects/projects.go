package projects

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/project"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeValidationFailed = "validation_failed"
	ErrCodeNotFound         = "not_found"
)

func Create(token, name string) *apperror.Error {
	req := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/projects",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"name": {name},
		}.Encode(),
	}
	req.AddHeader("Authorization", "Bearer "+token)
	res, err := req.Do()
	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusCreated, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" {
			return apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), false)
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func Get(token, name string) *apperror.Error {
	uri := fmt.Sprintf("%s/projects/%s", config.Host, name)
	req := goreq.Request{
		Method:    "GET",
		Uri:       uri,
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusOK, http.StatusNotFound}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusOK {
		return nil
	} else if res.StatusCode == http.StatusNotFound {
		return apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	}

	return apperror.New(ErrCodeUnexpectedError, err, "", true)
}

func List(token string) (projects []*project.Project, appErr *apperror.Error) {
	uri := fmt.Sprintf("%s/projects", config.Host)
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

	if res.StatusCode != http.StatusOK {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j struct {
		Projects []*project.Project `json: "projects"`
	}

	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return j.Projects, nil
}
