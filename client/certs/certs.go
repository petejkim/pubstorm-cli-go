package certs

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
	ErrCodeNotAllowedDomain = "domain_not_allowed"
	ErrCodeAcmeServerError  = "acme_server_error"

	ErrInvalidCert       = "invalid_cert"
	ErrInvalidCommonName = "invalid_common_name"
)

type Cert struct {
	ID         uint      `json:"id"`
	StartsAt   time.Time `json:"starts_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	CommonName string    `json:"common_name"`
	Issuer     string    `json:"issuer"`
	Subject    string    `json:"subject"`
}

func Create(token, name, domainName, crtPath, keyPath string) (c *Cert, appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "POST",
		Uri:       config.Host + "/projects/" + name + "/domains/" + domainName + "/cert",
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writeFileToBody(crtPath, "cert", writer); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if err := writeFileToBody(keyPath, "key", writer); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if err := writer.Close(); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	req.AddHeader("Content-Type", writer.FormDataContentType())
	bodyLen := int64(body.Len())

	req.Body = body
	req.OnBeforeRequest = func(goreq *goreq.Request, httpreq *http.Request) {
		httpreq.ContentLength = bodyLen
	}

	res, err := req.Do()
	if err != nil {
		return nil, apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusCreated, http.StatusBadRequest, http.StatusNotFound, http.StatusForbidden, 422}, res.StatusCode) {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusCreated {
		var j struct {
			Cert *Cert `json:"cert"`
		}

		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		return j.Cert, nil
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	switch j["error_description"] {
	case "domain could not be found":
		return nil, apperror.New(ErrCodeNotFound, nil, "domain could not be found", true)
	case "project could not be found":
		return nil, apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
	case "not allowed to upload certs for default domain":
		return nil, apperror.New(ErrCodeNotAllowedDomain, nil, "not allowed to upload certs for default domain", true)
	case "request body is too large":
		return nil, apperror.New(ErrCodeFileSizeTooLarge, nil, "cert or key file is too large", true)
	case "both cert and key are required":
		return nil, apperror.New(ErrInvalidCert, nil, "certificate or private key is missing", true)
	case "invalid cert or key":
		return nil, apperror.New(ErrInvalidCert, nil, "certificate or private key is not valid", true)
	case "invalid common name (domain name mismatch)":
		return nil, apperror.New(ErrInvalidCommonName, nil, "certificate is not valid for the specified domain", true)
	}
	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}

func Get(token, name, domainName string) (c *Cert, appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "GET",
		Uri:       config.Host + "/projects/" + name + "/domains/" + domainName + "/cert",
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
			Cert *Cert `json:"cert"`
		}

		if err := res.Body.FromJsonTo(&j); err != nil {
			return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		return j.Cert, nil
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	switch j["error_description"] {
	case "cert could not be found":
		return nil, apperror.New(ErrCodeNotFound, nil, "cert could not be found", true)
	case "project could not be found":
		return nil, apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
	}
	return nil, apperror.New(ErrCodeUnexpectedError, err, "", true)
}

func Delete(token, name, domainName string) (appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "DELETE",
		Uri:       config.Host + "/projects/" + name + "/domains/" + domainName + "/cert",
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
		case "cert could not be found":
			return apperror.New(ErrCodeNotFound, nil, "cert could not be found", true)
		case "project could not be found":
			return apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
		}

		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func Enable(token, name, domainName string) *apperror.Error {
	req := goreq.Request{
		Method:    "POST",
		Uri:       config.Host + "/projects/" + name + "/domains/" + domainName + "/cert/letsencrypt",
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}
	req.AddHeader("Authorization", "Bearer "+token)

	res, err := req.Do()
	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{
		http.StatusOK,
		http.StatusNotFound,
		http.StatusForbidden,
		http.StatusServiceUnavailable}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusNotFound {
		switch j["error_description"] {
		case "project could not be found":
			return apperror.New(ErrCodeProjectNotFound, nil, "project could not be found", true)
		case "domain could not be found":
			return apperror.New(ErrCodeNotFound, nil, "domain could not be found", true)
		}

		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusForbidden {
		return apperror.New(ErrCodeNotAllowedDomain, nil, "the default domain already supports HTTPS", true)
	}

	if res.StatusCode == http.StatusServiceUnavailable {
		switch j["error_description"] {
		case "domain could not be verified":
			return apperror.New(ErrCodeAcmeServerError, nil, "domain could not be verified - have you changed its DNS configuration yet?", true)
		}
		return apperror.New(ErrCodeAcmeServerError, nil, "error communicating with Let's Encrypt", true)
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
