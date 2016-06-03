package tr

var strs = map[string]map[string]string{
	"en": {
		"rise_cli_desc":           "Command line interface for PubStorm, the easiest way to publish your HTML5 websites and apps",
		"signup_desc":             "Create a new PubStorm account",
		"login_desc":              "Log in to a PubStorm account",
		"logout_desc":             "Log out from current session",
		"password_change_desc":    "Change your PubStorm password",
		"password_reset_desc":     "Reset your PubStorm password",
		"password_reset_continue": "Specify this flag if you already have a password reset token",
		"init_desc":               "Initialize a PubStorm project",
		"publish_desc":            "Publish a PubStorm project",
		"publish_verbose":         "Show additional information",
		"domains_desc":            "List all domains for a PubStorm project",
		"domains_add_desc":        "Add a new domain to a PubStorm project",
		"domains_add_args":        "[DOMAIN]\n\nDOMAIN: Domain to add. Specify \"default\" to enable the default .%s domain.",
		"domains_rm_desc":         "Remove a domain from a PubStorm project",
		"domains_rm_args":         "[DOMAIN]\n\nDOMAIN: Domain to remove. Specify \"default\" to disable the default .%s domain.",
		"projects_desc":           "List your PubStorm projects",
		"projects_rm_desc":        "Delete a PubStorm project",
		"projects_rm_force":       "Delete project without confirmation",
		"rollback_desc":           "Rollback to a previous or a specified version",
		"rollback_args":           "[VERSION]\n\nVERSION: Version to rollback to. Example: v1",
		"versions_desc":           "List versions of all completed deployments for a PubStorm project",
		"collab_desc":             "Lists collaborators for the current project",
		"collab_add_desc":         "Add a collaborator to the current project",
		"collab_rm_desc":          "Remove a collaborator from the current project",
		"ssl_info_desc":           "Show certificate information for a Pubstorm project",
		"ssl_info_args":           "[DOMAIN]",
		"ssl_set_desc":            "Upload an SSL certificate and a private key for a PubStorm project",
		"ssl_set_args":            "[DOMAIN] [CRT_FILE] [KEY_FILE]",
		"ssl_rm_desc":             "Remove an SSL certificate and a private key for a PubStorm project",
		"ssl_rm_args":             "[DOMAIN]",
		"ssl_force_desc":          "Enable or disable forced SSL/HTTPS",
		"ssl_force_args":          "[on/off]",
		"cert_info_desc":          "Show certificate information for a PubStorm project",
		"cert_info_args":          "[DOMAIN]",
		"cert_set_desc":           "Upload an SSL certificate and a private key for a PubStorm project",
		"cert_set_args":           "[DOMAIN] [CRT_FILE] [KEY_FILE]",
		"cert_rm_desc":            "Remove an SSL certificate and a private key for a PubStorm project",
		"cert_rm_args":            "[DOMAIN]",
		"reinit_desc":             "Re-initialize a PubStorm project",
		"reinit_args":             "[PROJECT NAME]",

		"update_available":       "A PubStorm update is available.",
		"update_current_version": "Your version: %s",
		"update_latest_version":  "Latest version: %s",
		"update_instructions":    "Run `npm -g install pubstorm` to update to version %s.",

		"join_rise":            "Join PubStorm, the easiest way to publish your HTML5 websites and apps!",
		"signup_disclaimer":    "By creating an account, you agree to the following:-",
		"rise_tos":             "PubStorm Terms of Service",
		"rise_privacy_policy":  "PubStorm Privacy Policy",
		"enter_email":          "Enter Email",
		"enter_password":       "Enter Password",
		"confirm_password":     "Confirm Password",
		"password_no_match":    "Passwords do not match. Please re-enter password.",
		"error_in_input":       "There were errors in your input. Please try again.",
		"account_created":      "Your account has been created. You will receive your confirmation code shortly via email.",
		"enter_confirmation":   "Enter Confirmation Code (Check your inbox!)",
		"confirmation_success": "Thanks for confirming your email address! Your account is now active!",
		"login_fail":           "Login failed. Please try again by running `storm login`.",
		"login_success":        "You are logged in as %s.",
		"oauth_misconfigured":  "Your version of the PubStorm CLI has expired, please update it by running `npm -g install pubstorm`.",

		"login_rise":                "Welcome back to PubStorm, the easiest way to publish your HTML5 websites and apps!",
		"enter_credentials":         "Enter your PubStorm credentials",
		"confirmation_required":     "You have to confirm your email address to continue. Please check your inbox for the confirmation code.",
		"enter_confirmation_resend": "Enter Confirmation Code (Or enter \"resend\" if you need it sent again)",
		"confirmation_resent":       "Confirmation code has been resent. You will receive your confirmation code shortly via email.",

		"reset_password": "Reset your PubStorm password",
		"reset_password_quote": `"I forgot the password for the file where I keep all my passwords"
                                              - Not you, hopefully`,
		"reset_password_email_sent": "An email with password reset instructions has been sent to %s",
		"enter_password_reset_code": "Enter Password Reset Code (Check your inbox!)",
		"password_reset_success":    "All good! Please login with your new password by running `storm login`.",

		"rise_config_write_failed": "Could not save PubStorm config file!",

		"logout_success":       "You are now logged out.",
		"access_token_cleared": "Access token cleared.",

		"not_logged_in":   "You are not logged in. Please login by running `storm login` or create a new account by running `storm signup`.",
		"login_expired":   "Your previous session has expired. Please login again by running `storm login`.",
		"no_rise_project": "Could not find a PubStorm project in current working directory. To initialize a new PubStorm project here, run `storm init`.",

		"something_wrong": "Something went wrong. Please try again.",

		"existing_rise_project": "A PubStorm project already exists in the current working directory; aborting.",

		"init_rise_project":          "Set up your PubStorm project",
		"enter_project_path":         "Enter Project Path (path to be deployed)",
		"project_path_create_ok":     "Created project directory \"%s\".",
		"project_path_create_failed": "Could not create a directory at \"%s\".",
		"enter_project_name":         "Enter Project Name",
		"project_initialized":        "Successfully created project \"%s\".",
		"rise_json_saved":            "Saved project settings to \"pubstorm.json\". This file should not be deleted.",

		"scanning_path":             "Scanning \"%s\"...",
		"bundling_file_count_size":  "Bundling %s files (%s)...",
		"bundle_root_index_missing": "Your project does not include an index.html file in the project root.",
		"project_size_exceeded":     "Your project size cannot exceed %s!",
		"packing_bundle":            "Packing bundle \"%s\"...",
		"bundle_size_exceeded":      "Your bundle size cannot exceed %s!",
		"uploading_bundle":          "Uploading bundle \"%s\" to PubStorm Cloud...",
		"optimizing":                "Optimizing...",
		"launching":                 "Launching v%d...",
		"published":                 "Successfully published \"%s\" on PubStorm Cloud.",
		"published_no_domain":       "Successfully published \"%s\" on PubStorm Cloud, but no domain name is configured. To add a domain name, run `storm domains add`.",

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

		"domain_list":                    "List of Domains for \"%s\"",
		"enter_domain_name_to_add":       "Enter Domain Name to Add",
		"domain_limit_reached":           "You cannot add any more domains to project \"%s\".",
		"domain_added":                   "Successfully added \"%s\" to project \"%s\".",
		"default_domain_added":           "Successfully enabled default domain \"%s\" for project \"%s\".",
		"default_domain_already_added":   "Default domain \"%s\" is already enabled for project \"%s\", nothing to do.",
		"dns_instructions":               "Please add the following records to the DNS configuration for the domain \"%s\":-",
		"dns_more_info":                  "For more information on DNS configuration, please visit %s",
		"enter_domain_name_to_remove":    "Enter Domain Name to Remove",
		"domain_cannot_be_deleted":       "Domain \"%s\" cannot be deleted.",
		"domain_not_found":               "Domain \"%s\" is not found",
		"domain_removed":                 "Successfully removed \"%s\" from project \"%s\".",
		"default_domain_removed":         "Successfully disabled default domain \"%s\" for project \"%s\".",
		"default_domain_already_removed": "Default domain \"%s\" is already disabled for project \"%s\", nothing to do.",

		"project_not_found": "Could not find a project \"%s\" that belongs to you.",
		"project_is_locked": "The project \"%s\" is locked by another user or process, please try again.",

		"projects_list_header":        "Your Projects",
		"shared_projects_list_header": "Projects Shared With You",
		"no_project":                  "You do not have any PubStorm project created.",

		"will_invalidate_session":     "Changing password will log you out from all other active sessions.",
		"enter_existing_password":     "Enter Existing Password",
		"password_changed":            "Your password has been changed.",
		"reenter_email":               "Please re-enter your email address to login with your new password.",
		"existing_password_incorrect": "The existing password you've entered is incorrect.",
		"new_password_same":           "You cannot reuse your previous password.",

		"project_rm_cannot_undo":        "This action cannot be undone!",
		"project_rm_permanent_delete":   "This will permanently delete \"%s\" project. To abort, press Ctrl-C.",
		"enter_project_name_to_confirm": "Enter \"%s\" (without quotes) to confirm",
		"project_name_does_not_match":   "The name you've entered does not match the project name, please try again.",
		"project_json_failed_to_delete": "Failed to delete \"pubstorm.json\".",
		"project_rm_success":            "Successfully deleted project \"%s\".",

		"collab_list_header":        "Collaborators of \"%s\"",
		"collab_add_user_not_found": "We do not know of a PubStorm user with the email address \"%s\".",
		"collab_cannot_add_owner":   "You cannot add yourself as a collaborator of a project that belongs to you.",
		"collab_enter_email_to_add": "Enter Email of Collaborator to Add",
		"collab_enter_email_to_rm":  "Enter Email of Collaborator to Remove",
		"collab_rm_user_not_found":  "User with the email address \"%s\" is not a collaborator of the project \"%s\".",
		"collab_added_success":      "Successfully added \"%s\" as a collaborator of \"%s\"",
		"collab_removed_success":    "Successfully removed \"%s\" as a collaborator of \"%s\"",

		"rollback_no_active_deployment":   "This PubStorm project does not have any completed deployment.",
		"rollback_no_previous_version":    "There is no previous version to rollback to.",
		"rollback_success":                "Successfully rolled back \"%s\" to v%d.",
		"rollback_invalid_version":        "The specified version is not valid",
		"rollback_version_not_found":      "Version v%d could not be found",
		"rollback_version_already_active": "This PubStorm project is already on v%d",

		"project_locked": "This PubStorm project is locked",

		"versions_list": "Completed deployments for \"%s\"",

		"ssl_enter_domain_name":  "Enter Domain Name",
		"ssl_cert_set":           "Successfully set an SSL certificate for %s",
		"ssl_file_not_found":     "\"%s\" could not be found",
		"ssl_file_invalid":       "\"%s\" is invalid",
		"ssl_not_allowed_domain": "You cannot set an SSL certificate for the domain \"%s\"",
		"ssl_too_large":          "Certificate or private key file is too large",
		"ssl_invalid":            "Certificate or prvate key file is not valid",
		"ssl_invalid_domain":     "Certificate's common name does not match \"%s\"",
		"ssl_enter_cert_path":    "Enter Path To Certificate",
		"ssl_enter_key_path":     "Enter Path To Private Key",
		"ssl_cert_not_found":     "The cert for \"%s\" could not be found",
		"ssl_cert_details":       "Certificate details for domain \"%s\"",
		"ssl_cert_common_name":   "Common Name",
		"ssl_cert_issuer":        "Issuer",
		"ssl_cert_subject":       "Subject",
		"ssl_cert_starts_at":     "Starts At",
		"ssl_cert_expires_at":    "Expires At",
		"ssl_cert_removed":       "Successfully removed SSL certificate for %s",
		"ssl_force_https_on":     "Forced SSL/HTTPS has been enabled.",
		"ssl_force_https_off":    "Forced SSL/HTTPS has been disabled.",

		"cert_enter_domain_name":  "Enter Domain Name",
		"cert_set":                "Successfully set an SSL certificate for %s",
		"cert_file_not_found":     "\"%s\" could not be found",
		"cert_file_invalid":       "\"%s\" is invalid",
		"cert_not_allowed_domain": "You cannot set an SSL certificate for the domain \"%s\"",
		"cert_too_large":          "Certificate or private key file is too large",
		"cert_invalid":            "Certificate or prvate key file is not valid",
		"cert_invalid_domain":     "Certificate's common name does not match \"%s\"",
		"cert_enter_cert_path":    "Enter Path To Certificate",
		"cert_enter_key_path":     "Enter Path To Private Key",
		"cert_not_found":          "No certificate for \"%s\"",
		"cert_details":            "Certificate details for domain \"%s\"",
		"cert_common_name":        "Common Name",
		"cert_issuer":             "Issuer",
		"cert_subject":            "Subject",
		"cert_starts_at":          "Starts At",
		"cert_expires_at":         "Expires At",
		"cert_removed":            "Successfully removed SSL certificate for %s",

		"existing_project":        "A PubStorm project \"%s\" already exists in the current working directory.",
		"re_init_project":         "Would you like to initialize using an existing project in current directory",
		"project_re_initialized":  "Successfully re-initialized project \"%s\".",
		"re_init_project_aborted": "Re-initialization aborted",
	},
}

func T(str string) string {
	return strs["en"][str]
}
