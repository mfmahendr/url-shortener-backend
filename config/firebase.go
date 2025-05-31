package config

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func InitFirebase(ctx context.Context) *firebase.App {
	log.Println("Initializing Firebase...")
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	config := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}

	app, err := firebase.NewApp(ctx, config, opt) 
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	return app
}