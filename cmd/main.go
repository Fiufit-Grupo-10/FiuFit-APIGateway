package main

import (
	"context"
	"log"

	"fiufit.api.gateway/internal/auth"
)



func main() {
	ctx := context.Background()
	f, err := auth.GetFirebase(ctx)
	if err != nil {
		log.Fatal("nil pointer")
	}
	_, err = f.CreateUser(auth.SignUpModel{Username: "user3", Email: "user3@example.com", Password: "123456"})
	// Send the error as string in request
	if err != nil {
		log.Fatalf("%s", err)
	}
	// r := gin.Default()
	// r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, "pong!") })
	// r.Run() // listen and serve on 0.0.0.0:8080
}
