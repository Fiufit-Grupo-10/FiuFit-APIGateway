package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
)

const (
	users = iota
	trainings
	metrics
	goals
)

var enviromentVariables = map[int]string{
	users:    "USERS_URL",
	trainings: "TRAINERS_URL",
	metrics:  "METRICS_URL",
	goals:    "GOALS_URL",
}

const serviceName = "service-external-gateway"

type Config struct {
	URLS map[int]*url.URL
}

func NewConfig() (*Config, error) {
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
