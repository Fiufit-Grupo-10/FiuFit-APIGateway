package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

const (
	Users = iota
	Trainings
	Metrics
	Goals
)

var enviromentVariables = map[int]string{
	Users:    "USERS_URL",
	Trainings: "TRAINERS_URL",
	Metrics:  "METRICS_URL",
	Goals:    "GOALS_URL",
}

const ServiceName = "service-external-gateway"

type Config struct {
	URLS map[int]*url.URL
}

func New() (*Config, error) {
	URLS := make(map[int]*url.URL)
	for key, envVar := range enviromentVariables {
		rawURL, found := os.LookupEnv(envVar)
		if !found || rawURL == "" {
			errorMsg := fmt.Sprintf("Enviroment variable %s not found", envVar)
			return nil, errors.New(errorMsg)
		}

		URL, err := url.Parse(rawURL)
		if err != nil {
			errorMsg := fmt.Sprintf("Invalid url %s", rawURL)
			return nil, errors.New(errorMsg)
		}

		URLS[key] = URL
	}

	return &Config{URLS: URLS}, nil
}
