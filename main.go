package main

import (
	"log"
	"os"
	"subteez/router"
	"subteez/subscene"

	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router.InitializeAndRun(subscene.SubsceneApi{}, port)
}
