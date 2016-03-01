package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Host = "https://api.rise.sh"

	// these do not have to be secured
	ClientID     = "73c24fbc2eb24bbf1d3fc3749fc8ac35"
	ClientSecret = "0f3295e1b531191c0ce8ccf331421644d4c4fbab9eb179778e5172977bf0238cdbf4b3afe1ead11b9892ce8806e87cc1acc10263dfdade879a05b931809690a1"

	DotRisePath string

	AccessToken string
)

const (
	configJSON = "config.json"
)

func init() {
	if envHost := os.Getenv("RISE_HOST"); envHost != "" {
		Host = envHost
	}

	if runtime.GOOS == "windows" {
		DotRisePath = filepath.Join(os.Getenv("APPDATA"), "rise")
	} else {
		DotRisePath = filepath.Join(os.Getenv("HOME"), ".rise")
	}

	if err := os.MkdirAll(DotRisePath, 0700); err != nil {
		log.Fatalln("Fatal Error: Could not make data directory!")
	}
}

// Saves config to a json file
func Save() {
	configPath := filepath.Join(DotRisePath, configJSON)
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Fatal Error: Could not write to %s!\n", configPath)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(map[string]interface{}{
		"access_token": AccessToken,
	})
}
