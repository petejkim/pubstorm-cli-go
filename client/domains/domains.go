package domains

import (
	"net/http"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
)

const (
	ErrCodeRequestFailed   = "request_failed"
	ErrCodeUnexpectedError = "unexpected_error"
	ErrCodeNotFound        = "not_found"
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

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
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
