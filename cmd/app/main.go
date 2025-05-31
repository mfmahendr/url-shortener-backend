package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/mfmahendr/url-shortener-backend/config"
	"github.com/mfmahendr/url-shortener-backend/internal/di"
	val "github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	val.Init()

	// initialize the controller and services using dependency injection
	ctx := context.Background()
	controller, err := di.InitializeController(ctx)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	// middleware and routes setup
	authMiddleware, err := di.InitializeAuthMiddleware(ctx)
	if err != nil {
		log.Fatalf("failed to initialize auth middleware: %v", err)
	}
	controller.RegisterRoutes(*authMiddleware)


	// start the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, controller.Router))
}