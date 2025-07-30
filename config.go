package main

import (
	"errors"
	"os"
)

// Config struct to hold the application configuration
type UserInput struct {
	Upstream struct {
		URL       string `mapstructure:"url"`
		BasicAuth struct {
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		} `mapstructure:"basicAuth"`
		TLS struct {
			InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
			CAFile             string `mapstructure:"caFile"`
		} `mapstructure:"tls"`
	} `mapstructure:"upstream"`
	InjectLabels map[string]string `mapstructure:"injectLabels"`
}

// input validation
func (config UserInput) Validate() error {
	// upstream.url is mandatory
	if config.Upstream.URL == "" {
		return errors.New("--upstream-url is a mandatory argument")
	}

	// if username or password is defined
	if (config.Upstream.BasicAuth.Username == "") != (config.Upstream.BasicAuth.Password == "") {
		return errors.New("Must define neither or both of arguments --upstream-basicAuth-username and --upstream-basicAuth-password")
	}

	// verify that tls caFile exists on-disk
	if config.Upstream.TLS.CAFile != "" {
		if _, err := os.Stat(config.Upstream.TLS.CAFile); err != nil {
			return err
		}
	}

	return nil
}
