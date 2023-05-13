package main

import (
	"context"
	"fiufit.api.gateway/cmd/gateway"
	"fiufit.api.gateway/internal/auth"
	"log"
	"net/url"
	"os"
)

func main() {
	log.Println("=====NUEVO======")
	ctx := context.Background()
	f, err := auth.GetFirebase(ctx)
	if err != nil {
		log.Fatalf("Couldn't start firebase service: %s", err.Error())
	}
	rawURL, found := os.LookupEnv("USERS_URL")
	log.Printf("rawURL: %s", rawURL)
	if !found && rawURL != "" {
		log.Fatal("USERS_URL enviroment variable not found")
	}

	usersURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", usersURL)
	}

	rawURL, found = os.LookupEnv("TRAINERS_URL")
	log.Printf("rawURL: %s", rawURL)
	if !found {
		log.Fatal("USERS_URL enviroment variable not found")
	}

	trainersURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", trainersURL)
	}

	gateway := gateway.New(gateway.Users(usersURL, f), gateway.Admin(usersURL, f), gateway.Trainers(trainersURL, f), gateway.Reviews(trainersURL, f))

	gateway.Run("0.0.0.0:8080")
}
