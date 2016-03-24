package domains

import (
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeNotFound         = "not_found"
	ErrCodeValidationFailed = "validation_failed"
	ErrCodeLimitReached     = "limit_reached"
)

func Index(token, projectName string) (domainNames []string, appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "GET",
		Uri:       config.Host + "/projects/" + projectName + "/domains",
		Accept:    "application/vnd.rise.v0+json",
		UserAgent: "RiseCLI",
	}
	req.AddHeader("Authorization", "Bearer "+token)
	res, err := req.Do()

	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusOK, http.StatusNotFound}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	}

	var j map[string][]string
	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return j["domains"], nil
}

func Create(token, projectName, name string) (appErr *apperror.Error) {
	req := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/projects/" + projectName + "/domains",
		Accept:      "application/vnd.rise.v0+json",
		ContentType: "application/x-www-form-urlencoded",
		UserAgent:   "RiseCLI",

		Body: url.Values{
			"name": {name},
		}.Encode(),
	}

	req.AddHeader("Authorization", "Bearer "+token)
	res, err := req.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusCreated, http.StatusNotFound, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusNotFound {
		return apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" {
			return apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), false)
		}

		if j["error"] == "invalid_request" {
			if j["error_description"] == "project cannot have more domains" {
				return apperror.New(ErrCodeLimitReached, err, "", true)
			} else if errDesc, ok := j["error_description"].(string); ok {
				return apperror.New(ErrCodeUnexpectedError, err, errDesc, true)
			} else {
				return apperror.New(ErrCodeUnexpectedError, err, "", true)
			}
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}
