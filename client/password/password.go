package password

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
)

func Forgot(email string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/user/password/forgot",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"email": {email},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusOK, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" && j["errors"] != nil {
			return apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), false)
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if v, ok := j["sent"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func Reset(email, resetToken, newPassword string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/user/password/reset",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"email":       {email},
			"reset_token": {resetToken},
			"password":    {newPassword},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}
	defer res.Body.Close()

	if !util.ContainsInt([]int{http.StatusOK, http.StatusForbidden, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" && j["errors"] != nil {
			return apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), false)
		}
		if j["error"] == "invalid_params" && j["error_description"] != nil {
			switch j["error_description"] {
			case "email is not found":
				return apperror.New(ErrCodeValidationFailed, nil, "You've entered an invalid email address. Please try again.", false)
			case "invalid email or reset_token":
				return apperror.New(ErrCodeValidationFailed, nil, "You've entered an invalid email address or password reset code. Please check your email inbox for the password reset instructions.", false)
			}
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == http.StatusForbidden {
		if j["error"] == "invalid_params" && j["error_description"] == "invalid email or reset_token" {
			return apperror.New(ErrCodeValidationFailed, nil, "You've entered an invalid email address or password reset code. Please check your email inbox for the password reset instructions.", false)
		}
	}

	if v, ok := j["reset"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}
