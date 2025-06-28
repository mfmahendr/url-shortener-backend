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
	config := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}

	var (
		err error
		app *firebase.App
	)
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		opt := option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
		app, err = firebase.NewApp(ctx, config, opt) 
	} else {
		app, err = firebase.NewApp(ctx, config) 
	}

	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	return app
}