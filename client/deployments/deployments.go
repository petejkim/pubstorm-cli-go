package deployments

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/pkg/progressbar"
)

const (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeValidationFailed = "validation_failed"
)

func Create(token, name, bunPath string, verbose bool) *apperror.Error {
	req := goreq.Request{
		Method:    "POST",
		Uri:       config.Host + "/projects/" + name + "/deployments",
		Accept:    "application/vnd.rise.v0+json",
		UserAgent: "RiseCLI",
	}
	req.AddHeader("Authorization", "Bearer "+token)

	f, err := os.Open(bunPath)
	if err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("payload", filepath.Base(bunPath))
	if err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if _, err = io.Copy(part, f); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if err = writer.Close(); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	req.AddHeader("Content-Type", writer.FormDataContentType())
	bodyLen := int64(body.Len())

	if verbose {
		pb := progressbar.NewReader(body, os.Stdout, bodyLen)
		req.Body = pb
	} else {
		req.Body = body
	}

	req.OnBeforeRequest = func(goreq *goreq.Request, httpreq *http.Request) {
		httpreq.ContentLength = bodyLen
	}
	res, err := req.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if res.StatusCode != http.StatusBadRequest && res.StatusCode != http.StatusAccepted {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusBadRequest {
		if j["error"] == "invalid_request" {
			switch j["error_description"] {
			case "request body is too large":
				return apperror.New(ErrCodeValidationFailed, nil, "project size is too large", true)
			}
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}
