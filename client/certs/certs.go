package certs

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
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeValidationFailed = "validation_failed"

	ErrCodeNotFound         = "not_found"
	ErrCodeProjectNotFound  = "project_not_found"
	ErrCodeFileSizeTooLarge = "file_size_too_large"

	ErrCodeNotAllowedDomain = "domain_not_allwed"
	ErrInvalidCerts         = "invalid_certs"
	ErrCertNotMatch         = "cert_not_match"
)

func Create(token, name, domainName, crtPath, keyPath string) (appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "POST",
		Uri:       config.Host + "/projects/" + name + "/domains/" + domainName + "/cert",
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writeFileToBody(crtPath, "ssl.crt", writer); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if err := writeFileToBody(keyPath, "ssl.key", writer); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if err := writer.Close(); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	req.AddHeader("Content-Type", writer.FormDataContentType())
	bodyLen := int64(body.Len())

	req.Body = body
	req.OnBeforeRequest = func(goreq *goreq.Request, httpreq *http.Request) {
		httpreq.ContentLength = bodyLen
	}

	res, err := req.Do()
	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusForbidden, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode != http.StatusCreated {
		switch j["error_description"] {
		case "domain could not be found":
			return apperror.New(ErrCodeNotFound, nil, "domain could not be found", true)
		case "project could not be found":
			return apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
		case "Not allowed to upload certs for default domain":
			return apperror.New(ErrCodeNotAllowedDomain, nil, "Not allowed domain name", true)
		case "request body is too large":
			return apperror.New(ErrCodeFileSizeTooLarge, nil, "file size is too large", true)
		case "certificate or private key file is missing":
			return apperror.New(ErrInvalidCerts, nil, "certificate or private key file is missing", true)
		case "certificate or private key is not valid":
			return apperror.New(ErrInvalidCerts, nil, "certificate or private key is not valid", true)
		case "ssl cert is not matched domain name":
			return apperror.New(ErrCertNotMatch, nil, "ssl cert is not matched domain name", true)
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func writeFileToBody(path, paramName string, bodyWriter *multipart.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	part, err := bodyWriter.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return err
	}

	if _, err = io.Copy(part, f); err != nil {
		return err
	}

	return nil
}
