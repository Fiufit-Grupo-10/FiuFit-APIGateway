package main

import (
	"context"
	"net/url"
	"os"

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
	log.Println("=====NUEVO======")
	ctx := context.Background()
	f, err := auth.GetFirebase(ctx)
	if err != nil {
		log.Fatalf("Couldn't start firebase service: %s", err.Error())
	}
	// Users
	rawURL, found := os.LookupEnv("USERS_URL")
	log.Printf("rawURL: %s", rawURL)
	if !found || rawURL == "" {
		log.Fatal("USERS_URL enviroment variable not found")
	}

	usersURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", usersURL)
	}

	// Trainers
	rawURL, found = os.LookupEnv("TRAINERS_URL")
	log.Printf("rawURL: %s", rawURL)
	if !found || rawURL == "" {
		log.Fatal("USERS_URL enviroment variable not found")
	}

	trainersURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", trainersURL)
	}

	// Metrics
	rawURL, found = os.LookupEnv("METRICS_URL")
	log.Printf("rawURL: %s", rawURL)
	if !found || rawURL == "" {
		log.Fatal("METRICS_URL enviroment variable not found")
	}

	metricsURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", metricsURL)
	}

	// Goals
	rawURL, found = os.LookupEnv("GOALS_URL")
	log.Printf("rawURL: %s", rawURL)
	if !found || rawURL == "" {
		log.Fatal("GOALS_URL enviroment variable not found")
	}

	goalsURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", goalsURL)
	}

	tracer.Start(tracer.WithServiceName("fiufit-api-gateway"))
	defer tracer.Stop()

	gateway := gateway.New(
		gateway.Users(usersURL, f),
		gateway.Admin(usersURL, trainersURL, f),
		gateway.Trainers(trainersURL, f),
		gateway.Reviews(trainersURL, f),
		gateway.Metrics(metricsURL, f),
		gateway.Goals(goalsURL, f))

	gateway.Run("0.0.0.0:8080")
}
