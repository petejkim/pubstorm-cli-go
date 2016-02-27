package config

import "os"

var (
	Host = "https://api.rise.sh"

	// these do not have to be secured
	ClientID     = "73c24fbc2eb24bbf1d3fc3749fc8ac35"
	ClientSecret = "0f3295e1b531191c0ce8ccf331421644d4c4fbab9eb179778e5172977bf0238cdbf4b3afe1ead11b9892ce8806e87cc1acc10263dfdade879a05b931809690a1"
)

func init() {
	if envHost := os.Getenv("RISE_HOST"); envHost != "" {
		Host = envHost
	}
}
