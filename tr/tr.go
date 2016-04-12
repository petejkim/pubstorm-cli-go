package tr

var strs = map[string]map[string]string{
	"en": {
		"rise_cli_desc":    "Command line interface for PubStorm, the easiest way to publish your HTML5 websites and apps",
		"signup_desc":      "Create a new PubStorm account",
		"login_desc":       "Log in to a PubStorm account",
		"logout_desc":      "Log out from current session",
		"init_desc":        "Initialize a PubStorm project",
		"deploy_desc":      "Publish a PubStorm project",
		"domains_desc":     "List all domains for a PubStorm project",
		"domains_add_desc": "Add a new domain to a PubStorm project",
		"domains_rm_desc":  "Remove a domain from a PubStorm project",
		"projects_desc":    "List your PubStorm projects",

		"update_available":       "A PubStorm update is available.",
		"update_current_version": "Your version: %s",
		"update_latest_version":  "Latest version: %s",
		"update_instructions":    "Run `npm -g install pubstorm` to update to version %s.",

		"join_rise":           "Join PubStorm, the easiest way to publish your HTML5 websites and apps!",
		"signup_disclaimer":   "By creating an account, you agree to the following:-",
		"rise_tos":            "PubStorm Terms of Service",
		"rise_privacy_policy": "PubStorm Privacy Policy",
		"enter_email":         "Enter Email",
		"enter_password":      "Enter Password",
		"confirm_password":    "Confirm Password",
		"password_no_match":   "Passwords do not match. Please re-enter password.",
		"error_in_input":      "There were errors in your input. Please try again.",
		"account_created":     "Your account has been created. You will receive your confirmation code shortly via email.",
		"enter_confirmation":  "Enter Confirmation Code (Check your inbox!)",
		"confirmation_sucess": "Thanks for confirming your email address! Your account is now active!",
		"login_fail":          "Login failed. Please try again by running `storm login`.",
		"login_success":       "You are logged in as %s.",

		"login_rise":                "Welcome back to PubStorm, the easiest way to publish your HTML5 websites and apps!",
		"enter_credentials":         "Enter your PubStorm credentials",
		"confirmation_required":     "You have to confirm your email address to continue. Please check your inbox for the confirmation code.",
		"enter_confirmation_resend": "Enter Confirmation Code (Or enter \"resend\" if you need it sent again)",
		"confirmation_resent":       "Confirmation code has been resent. You will receive your confirmation code shortly via email.",

		"rise_config_write_failed": "Could not save PubStorm config file!",

		"logout_success":       "You are now logged out.",
		"access_token_cleared": "Access token cleared.",

		"not_logged_in":   "You are not logged in. Please login by running `storm login` or create a new account by running `storm signup`.",
		"no_rise_project": "Could not find a PubStorm project in current working directory. To initialize a new PubStorm project here, run `storm init`.",

		"something_wrong": "Something went wrong. Please try again.",

		"existing_rise_project": "A PubStorm project already exists in the current working directory; aborting.",

		"init_rise_project":   "Set up your PubStorm project",
		"enter_project_path":  "Enter Project Path",
		"enable_basic_stats":  "Enable Basic Stats",
		"force_https":         "Redirect \"http\" to \"https\" URL",
		"enter_project_name":  "Enter Project Name",
		"project_initialized": "Successfully created project \"%s\".",
		"rise_json_saved":     "Saved project settings to \"pubstorm.json\". This file should not be deleted.",

		"scanning_path":            "Scanning \"%s\"...",
		"bundling_file_count_size": "Bundling %s files (%s)...",
		"project_size_exceeded":    "Your project size cannot exceed %s!",
		"packing_bundle":           "Packing bundle \"%s\"...",
		"bundle_size_exceeded":     "Your bundle size cannot exceed %s!",
		"uploading_bundle":         "Uploading bundle \"%s\" to PubStorm Cloud...",
		"launching":                "Launching...",
		"published":                "%s published on PubStorm Cloud.",

		"ignore_file_reason":    "Ignoring \"%s\", %s...",
		"symlink_error":         "could not follow symlink",
		"symlink_to_dir":        "symlink points to a directory",
		"special_mode_bits":     "file has special mode bits set",
		"name_has_dot_prefix":   "name begins with \".\"",
		"name_has_hash_prefix":  "name begins with \"#\"",
		"name_has_tilde_suffix": "name ends with \"~\"",
		"name_in_ignore_list":   "name is in ignore list",
		"file_unreadable":       "file is not readable",

		"stat_failed":       "Could not get file info for \"%s\"; aborting.",
		"write_failed":      "Failed to write to \"%s\"; aborting.",
		"file_size_changed": "File size of \"%s\" changed while packing; aborting.",

		"domain_list":                 "List of Domains for \"%s\"",
		"enter_domain_name_to_add":    "Enter Domain Name to Add",
		"domain_limit_reached":        "You cannot add any more domains to project \"%s\".",
		"domain_added":                "Successfully added \"%s\" to project \"%s\".",
		"dns_instructions":            "Please add the following records to the DNS configuration for the domain \"%s\":-",
		"dns_more_info":               "For more information on DNS configuration, please visit %s",
		"enter_domain_name_to_remove": "Enter Domain Name to Remove",
		"domain_cannot_be_deleted":    "Domain \"%s\" cannot be deleted.",
		"domain_not_found":            "Domain \"%s\" is not found",
		"domain_removed":              "Successfully removed \"%s\" from project \"%s\".",

		"project_not_found": "Could not find a project \"%s\" that belongs to you.",

		"project_list": "Your Projects",
		"no_project":   "You do not have any PubStorm project created.",
	},
}

func T(str string) string {
	return strs["en"][str]
}
