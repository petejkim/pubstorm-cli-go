package rawbundles

import (
	"net/http"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed   = "request_failed"
	ErrCodeUnexpectedError = "unexpected_error"
	ErrCodeProjectNotFound = "project_not_found"
	ErrCodeNotFound        = "not_found"
)

type RawBundle struct {
	ID           uint   `json:"id"`
	Checksum     string `json:"checksum"`
	UploadedPath string `json:"uploaded_path"`
}

func Get(token, name, checksum string) (*RawBundle, *apperror.Error) {
	req := goreq.Request{
		Method:    "GET",
		Uri:       config.Host + "/projects/" + name + "/raw_bundles/" + checksum,
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
		case "project could not be found":
			return nil, apperror.New(ErrCodeProjectNotFound, err, "", true)
		case "raw bundle could not be found":
			return nil, apperror.New(ErrCodeNotFound, err, "", true)
		}
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j struct {
		RawBundle *RawBundle `json:"raw_bundle"`
	}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return j.RawBundle, nil
}
