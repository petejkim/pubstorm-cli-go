package oauth

import (
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
)

var (
	ErrCodeRequestFailed   = "request_failed"
	ErrCodeUnexpectedError = "unexpected_error"
	ErrCodeInvalidGrant    = "invalid_grant"
)

func FetchToken(email, password string) (token string, appErr *apperror.Error) {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/oauth/token",
		Accept:      "application/vnd.rise.v0+json",
		ContentType: "application/x-www-form-urlencoded",
		UserAgent:   "RiseCLI",

		BasicAuthUsername: config.ClientID,
		BasicAuthPassword: config.ClientSecret,

		Body: url.Values{
			"grant_type": {"password"},
			"username":   {email},
			"password":   {password},
		}.Encode(),
	}.Do()

	if err != nil {
		return "", apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if res.StatusCode != http.StatusBadRequest && res.StatusCode != http.StatusOK {
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusBadRequest {
		if j["error"] == "invalid_grant" {
			switch j["error_description"] {
			case "user credentials are invalid":
				return "", apperror.New(ErrCodeInvalidGrant, nil, "invalid email or password", false)
			case "user has not confirmed email address":
				return "", apperror.New(ErrCodeInvalidGrant, nil, "user has not confirmed email address", false)
			}
		}
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	token, ok := j["access_token"].(string)
	if j["token_type"] != "bearer" || token == "" || !ok {
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return token, nil
}
