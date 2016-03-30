package tr

var strs = map[string]map[string]string{
	"en": {
		"rise_cli_desc":    "Command line interface for Rise, the easiest way to publish your HTML5 websites and apps",
		"signup_desc":      "Create a new Rise account",
		"login_desc":       "Log in to a Rise account",
		"logout_desc":      "Log out from current session",
		"init_desc":        "Initialize a Rise project",
		"deploy_desc":      "Publish a Rise project",
		"domains_desc":     "List all domains for a Rise project",
		"domains_add_desc": "Add a new domain to a Rise project",
		"domains_rm_desc":  "Remove a domain from a Rise project",

		"join_rise":           "Join Rise, the easiest way to publish your HTML5 websites and apps!",
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

		"login_rise":                "Welcome back to Rise, the easiest way to publish your HTML5 websites and apps!",
		"enter_credentials":         "Enter your Rise credentials",
		"confirmation_required":     "You have to confirm your email address to continue. Please check your inbox for the confirmation code.",
		"enter_confirmation_resend": "Enter Confirmation Code (Or enter \"resend\" if you need it sent again)",
		"confirmation_resent":       "Confirmation code has been resent. You will receive your confirmation code shortly via email.",

		"rise_config_write_failed": "Could not save rise config file!",

		"logout_success":       "You are now logged out.",
		"access_token_cleared": "Access token cleared.",
	},
}

func T(str string) string {
	return strs["en"][str]
}
