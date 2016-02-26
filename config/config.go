package config

import "os"

var Host = "https://api.rise.sh"

func init() {
	if envHost := os.Getenv("RISE_HOST"); envHost != "" {
		Host = envHost
	}
}
