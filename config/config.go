package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	AppName  = "PubStorm"
	Version  = "0.0.0"
	BuildEnv = "development"

	Host          = "http://localhost:3000"
	DefaultDomain = "risecloud.dev"
	WebsiteHost   = "https://www.rise.dev"

	LatestVersionURL = "https://s3-us-west-2.amazonaws.com/rise-development-usw2/versions/latest.json"

	RedirectorIP = "52.38.113.95" // don't ever change this
	DNSHelpURL   = "https://help.pubstorm.com/custom-domains"

	// these do not have to be secured
	ClientID     = "73c24fbc2eb24bbf1d3fc3749fc8ac35"
	ClientSecret = "0f3295e1b531191c0ce8ccf331421644d4c4fbab9eb179778e5172977bf0238cdbf4b3afe1ead11b9892ce8806e87cc1acc10263dfdade879a05b931809690a1"

	UserAgent = "PubStormCLI"
	ReqAccept = "application/vnd.pubstorm.v0+json"

	ProjectJSON = "pubstorm.json"

	DotRisePath string
	AccessToken string
	Email       string

	MaxProjectSize = int64(1024 * 1024 * 1000) // 1 GiB
	MaxBundleSize  = int64(1024 * 1024 * 1000) // 1 GiB
)

const (
	configJSONPath = "config.json"
)

func init() {
	if envRiseHost := os.Getenv("RISE_HOST"); envRiseHost != "" {
		Host = envRiseHost
	}

	if runtime.GOOS == "windows" {
		DotRisePath = filepath.Join(os.Getenv("APPDATA"), "PubStorm")
	} else {
		DotRisePath = filepath.Join(os.Getenv("HOME"), ".pubstorm")
	}

	if BuildEnv != "production" {
		DotRisePath += "-" + BuildEnv
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
	configPath := filepath.Join(DotRisePath, configJSONPath)
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(map[string]interface{}{
		"email":        Email,
		"access_token": AccessToken,
	})
}

// Load config from .pubstorm/config.json
func Load() error {
	configPath := filepath.Join(DotRisePath, configJSONPath)

	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var j struct {
		Email       string `json:"email"`
		AccessToken string `json:"access_token"`
	}

	if err = json.NewDecoder(f).Decode(&j); err != nil {
		return err
	}

	Email = j.Email
	AccessToken = j.AccessToken

	return nil
}
