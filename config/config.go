package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Host          = "https://api.rise.sh"
	DefaultDomain = "rise.cloud"
	WebsiteHost   = "https://www.rise.sh"

	// these do not have to be secured
	ClientID     = "73c24fbc2eb24bbf1d3fc3749fc8ac35"
	ClientSecret = "0f3295e1b531191c0ce8ccf331421644d4c4fbab9eb179778e5172977bf0238cdbf4b3afe1ead11b9892ce8806e87cc1acc10263dfdade879a05b931809690a1"

	DotRisePath string
	AccessToken string

	MaxProjectSize = int64(1024 * 1024 * 1000) // 1 GiB
	MaxBundleSize  = int64(1024 * 1024 * 1000) // 1 GiB
)

const (
	configJSON = "config.json"
)

func init() {
	if envRiseHost := os.Getenv("RISE_HOST"); envRiseHost != "" {
		Host = envRiseHost
	}

	if runtime.GOOS == "windows" {
		DotRisePath = filepath.Join(os.Getenv("APPDATA"), "rise")
	} else {
		DotRisePath = filepath.Join(os.Getenv("HOME"), ".rise")
	}

	if err := os.MkdirAll(DotRisePath, 0700); err != nil {
		log.Fatalln("Fatal Error: Failed to make data directory!")
	}

	if err := Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln("Fatal Error: Failed to load rise config file!")
		}
	}
}

// Saves config to a json file
func Save() error {
	configPath := filepath.Join(DotRisePath, configJSON)
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(map[string]interface{}{
		"access_token": AccessToken,
	})
}

// Load config from .rise/config.json
func Load() error {
	configPath := filepath.Join(DotRisePath, configJSON)

	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var j struct {
		AccessToken string `json:"access_token"`
	}

	if err = json.NewDecoder(f).Decode(&j); err != nil {
		return err
	}

	AccessToken = j.AccessToken

	return nil
}
