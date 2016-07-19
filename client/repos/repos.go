package repos

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
	ErrCodeValidationFailed = "validation_failed"
	ErrCodeAlreadyLinked    = "already_linked"
	ErrCodeNotLinked        = "not_linked"
	ErrCodeProjectNotFound  = "project_not_found"
)

type Repo struct {
	URI           string `json:"uri"`
	Branch        string `json:"branch"`
	WebhookURL    string `json:"webhook_url"`
	WebhookSecret string `json:"webhook_secret"`
}

func Link(token, projectName, repoURL, branch, secret string) (*Repo, *apperror.Error) {
	req := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/projects/" + projectName + "/repos",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"uri":    {repoURL},
			"branch": {branch},
			"secret": {secret},
		}.Encode(),
	}
	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusCreated, http.StatusConflict, http.StatusNotFound, 422}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusConflict {
		return nil, apperror.New(ErrCodeAlreadyLinked, err, "project has already been linked to a GitHub repository", true)
	}

	if res.StatusCode == http.StatusNotFound {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		switch j["error_description"] {
		case "project could not be found":
			return nil, apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
		}

		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		if j["error"] == "invalid_params" {
			return nil, apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), true)
		}

		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j struct {
		Repo *Repo `json:"repo"`
	}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return j.Repo, nil
}

func Unlink(token, projectName string) *apperror.Error {
	req := goreq.Request{
		Method:    "DELETE",
		Uri:       config.Host + "/projects/" + projectName + "/repos",
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

	if res.StatusCode == http.StatusNotFound {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		switch j["error_description"] {
		case "project not linked to any repository":
			return apperror.New(ErrCodeNotLinked, err, "project not linked to any repository", false)
		case "project could not be found":
			return apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
		}

		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func Info(token, projectName string) (*Repo, *apperror.Error) {
	req := goreq.Request{
		Method:    "GET",
		Uri:       config.Host + "/projects/" + projectName + "/repos",
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

	if res.StatusCode == http.StatusNotFound {
		var j map[string]interface{}
		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		switch j["error_description"] {
		case "project not linked to any repository":
			return nil, apperror.New(ErrCodeNotLinked, err, "project not linked to any repository", false)
		case "project could not be found":
			return nil, apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
		}

		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j struct {
		Repo *Repo `json:"repo"`
	}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return j.Repo, nil
}
