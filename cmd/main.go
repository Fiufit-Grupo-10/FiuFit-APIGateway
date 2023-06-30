package main

import (
	"context"

	"fiufit.api.gateway/cmd/gateway"
	"fiufit.api.gateway/internal/auth"
	"fiufit.api.gateway/internal/config"

	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	c, err := config.New()
	if err != nil {
		log.Fatalf("Invalid gateway configuration: %s", err.Error())
	}

	config.InitLogger(c)

	ctx := context.Background()
	f, err := auth.GetFirebase(ctx)
	if err != nil {
		log.Fatalf("Couldn't start firebase service: %s", err.Error())
	}

	usersURL := c.URLS[config.Users]
	trainingsURL := c.URLS[config.Trainings]
	metricsURL := c.URLS[config.Metrics]
	goalsURL := c.URLS[config.Goals]

	tracer.Start(tracer.WithService(config.ServiceName))
	defer tracer.Stop()

	gateway := gateway.New(c,
		gateway.Users(usersURL, f),
		gateway.Admin(usersURL, trainingsURL, metricsURL, f),
		gateway.Trainings(trainingsURL, f),
		gateway.Reviews(trainingsURL, f),
		gateway.Goals(goalsURL, f),
		gateway.Metrics(metricsURL, f))

	gateway.Run("0.0.0.0:8080")
}
