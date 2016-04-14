package collab

import (
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed   = "request_failed"
	ErrCodeUnexpectedError = "unexpected_error"
	ErrCodeNotFound        = "not_found"
	ErrCodeAlreadyExists   = "already_exists"
	ErrCodeUserNotFound    = "user_not_found"
	ErrCodeCannotAddOwner  = "owner_cannot_be_collab"
)

type Collaborator struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
}

func List(token, projectName string) ([]*Collaborator, *apperror.Error) {
	req := goreq.Request{
		Method:    "GET",
		Uri:       config.Host + "/projects/" + projectName + "/collaborators",
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
		return nil, apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	}

	var j struct {
		Collaborators []*Collaborator `json:"collaborators"`
	}

	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return j.Collaborators, nil
}

func Add(token, projectName, email string) *apperror.Error {
	req := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/projects/" + projectName + "/collaborators",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"email": {email},
		}.Encode(),
	}
	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{
		http.StatusCreated,
		http.StatusNotFound,
		http.StatusConflict,
		422,
	}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusNotFound {
		return apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
	}

	if res.StatusCode == http.StatusConflict {
		return apperror.New(ErrCodeAlreadyExists, nil, "user is already a collaborator", false)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" && j["error_description"] == "email is not found" {
			return apperror.New(ErrCodeUserNotFound, nil, "", true)
		}

		if j["error"] == "invalid_request" && j["error_description"] == "the owner of a project cannot be added as a collaborator" {
			return apperror.New(ErrCodeCannotAddOwner, nil, "", true)
		}

		return apperror.New(ErrCodeUnexpectedError, nil, "", true)
	}

	if v, ok := j["added"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, nil, "", true)
	}

	return nil
}

func Remove(token, projectName, email string) *apperror.Error {
	req := goreq.Request{
		Method:    "DELETE",
		Uri:       config.Host + "/projects/" + projectName + "/collaborators/" + email,
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

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusNotFound {
		switch j["error_description"] {
		case "project could not be found":
			return apperror.New(ErrCodeNotFound, nil, "project could not be found", true)
		case "email is not found":
			return apperror.New(ErrCodeUserNotFound, nil, "", true)
		}
	}

	if v, ok := j["removed"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, nil, "", true)
	}

	return nil
}
