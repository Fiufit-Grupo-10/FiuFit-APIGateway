package config

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"strconv"
)

const (
	Users = iota
	Trainings
	Metrics
	Goals
)

var urlEnvVariables = map[int]string{
	Users:     "USERS_URL",
	Trainings: "TRAINERS_URL",
	Metrics:   "METRICS_URL",
	Goals:     "GOALS_URL",
}

const ServiceName = "service-external-gateway"

type Services map[int]*url.URL

type Config struct {
	URLS            Services
	LogLevel        log.Level
	IsDevEnviroment bool
}

func getServices() (Services, error) {
	URLS := make(map[int]*url.URL)
	for key, envVar := range urlEnvVariables {
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
	return URLS, nil
}

func getLogLevel() log.Level {
	lvl, found := os.LookupEnv("LOG_LEVEL")
	if !found {
		return log.InfoLevel
	}
	level, err := log.ParseLevel(lvl)
	if err != nil {
		return log.InfoLevel
	}

	return level
}

func isDevEnviroment() bool {
	value, found := os.LookupEnv("DEVLOG")
	if !found {
		return false
	}

	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parsedValue
}

func New() (*Config, error) {
	services, err := getServices()
	if err != nil {
		return nil, err
	}

	return &Config{
		URLS:            services,
		LogLevel:        getLogLevel(),
		IsDevEnviroment: isDevEnviroment(),
	}, nil
}

func InitLogger(config *Config) {
	log.SetLevel(config.LogLevel)
	if !config.IsDevEnviroment {
		log.SetFormatter(&log.JSONFormatter{})
	}
}
