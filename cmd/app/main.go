package main

import (
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	router := httprouter.New()
	// routes.RegisterRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running at :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
