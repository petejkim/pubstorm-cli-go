package tr

var strs = map[string]map[string]string{
	"en": {
		"join_rise":           "Join Rise, the easiest way to publish your HTML5 websites and apps.",
		"signup_disclaimer":   "By creating an account, you agree to the following:-",
		"rise_tos":            "Rise Terms of Service",
		"rise_privacy_policy": "Rise Privacy Policy",
		"enter_email":         "Enter Email",
		"enter_password":      "Enter Password",
		"confirm_password":    "Confirm Password",
		"password_no_match":   "Passwords do not match. Please re-enter password.",
		"error_in_input":      "There were errors in your input. Please try again.",
		"account_created":     "Your account has been created. You will receive your confirmation code shortly via email.",
		"enter_confirmation":  "Enter Confirmation Code (Check your inbox!)",
		"confirmation_sucess": "Thanks for confirming your email address! Your account is now active!",
		"login_fail":          "Login failed. Please try again by running `rise login` command.",
		"login_success":       "You are logged in as %s.",
	},
}

func T(str string) string {
	return strs["en"][str]
}
