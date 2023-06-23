package main

import (
	"context"

	"fiufit.api.gateway/cmd/gateway"
	"fiufit.api.gateway/internal/auth"

	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})
}

func main() {
	ctx := context.Background()
	f, err := auth.GetFirebase(ctx)
	if err != nil {
		log.Fatalf("Couldn't start firebase service: %s", err.Error())
	}

	config, err := NewConfig()
	if err != nil {
		log.Fatalf("Couldn't start firebase service: %s", err.Error())
	}

	usersURL := config.URLS[users]
	trainingsURL := config.URLS[trainings]
	metricsURL := config.URLS[metrics]
	goalsURL := config.URLS[goals]

	tracer.Start(tracer.WithService(serviceName))
	defer tracer.Stop()

	gateway := gateway.New(
		gateway.Users(usersURL, f),
		gateway.Admin(usersURL, trainingsURL, metricsURL, f),
		gateway.Trainings(trainingsURL, f),
		gateway.Reviews(trainingsURL, f),
		gateway.Goals(goalsURL, f))
	
	gateway.Run("0.0.0.0:8080")
}
