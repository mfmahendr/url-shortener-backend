package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mfmahendr/url-shortener-backend/config"
	"github.com/mfmahendr/url-shortener-backend/internal/di"
	"github.com/mfmahendr/url-shortener-backend/internal/utils/validators"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	
	ctx := context.Background()

	validators.Init()
	firebaseApp := config.InitFirebase(ctx)
	
	// initialize the controller and services using dependency injection
	controller, err := di.InitializeController(ctx, firebaseApp, os.Getenv("SAFE_BROWSING_API_KEY"))
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	// middleware and routes setup
	controller.RateLimiter.SetLimit(5, 30 * time.Second)		// 5 request per 30 seconds
	authMiddleware, err := di.InitializeAuthMiddleware(firebaseApp)
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