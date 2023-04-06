package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"fiufit.api.gateway/cmd/gateway"
	"fiufit.api.gateway/internal/auth"
)



func main() {
	ctx := context.Background()
	f, err := auth.GetFirebase(ctx)
	if err != nil {
		log.Fatal("Couldn't start firebase service")
	}
	rawURL, found := os.LookupEnv("USERS_URL")
	if !found {
		log.Fatal("USERS_URL enviroment variable not found")
	}

	usersURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %s", usersURL)
	}
	gateway := gateway.New(gateway.Users(usersURL, f))
	gateway.Run()
}
