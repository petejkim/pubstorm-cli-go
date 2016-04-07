package oauth

import (
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed        = "request_failed"
	ErrCodeUnexpectedError      = "unexpected_error"
	ErrCodeInvalidGrant         = "invalid_grant"
	ErrCodeUnconfirmedEmail     = "unconfirmed_email"
	ErrCodeInvalidAuthorization = "invalid_authorization"
	ErrCodeOAuthMisconfigured   = "oauth_misconfigured"
)

func FetchToken(email, password string) (token string, appErr *apperror.Error) {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/oauth/token",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

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

	if !util.ContainsInt([]int{http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized}, res.StatusCode) {
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	switch res.StatusCode {
	case http.StatusBadRequest:
		if j["error"] == "invalid_grant" {
			switch j["error_description"] {
			case "user credentials are invalid":
				return "", apperror.New(ErrCodeInvalidGrant, nil, "invalid email or password", false)
			case "user has not confirmed email address":
				return "", apperror.New(ErrCodeUnconfirmedEmail, nil, "user has not confirmed email address", false)
			}
		}
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	case http.StatusUnauthorized:
		// The configured OAuth client ID and/or secret are invalid - this is a
		// build time error.
		return "", apperror.New(ErrCodeOAuthMisconfigured, nil, "OAuth Client ID/Secret misconfigured", true)
	}

	token, ok := j["access_token"].(string)
	if j["token_type"] != "bearer" || token == "" || !ok {
		return "", apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return token, nil
}

func InvalidateToken(token string) (appErr *apperror.Error) {
	req := goreq.Request{
		Method:    "DELETE",
		Uri:       config.Host + "/oauth/token",
		Accept:    config.ReqAccept,
		UserAgent: config.UserAgent,
	}

	req.AddHeader("Authorization", "Bearer "+token)
	res, err := req.Do()
	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusOK, http.StatusUnauthorized}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusUnauthorized {
		if j["error"] == "invalid_token" {
			switch j["error_description"] {
			case "access token is invalid":
				return apperror.New(ErrCodeInvalidAuthorization, nil, "invalid access token", false)
			}
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if v, ok := j["invalidated"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}
