package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	appMode := os.Getenv("APP_ENV")

	switch appMode {
	case "development":
		if err := godotenv.Load(); err != nil {
			return errors.New("no .env file found")
		}
		fallthrough
	case "production":
		return checkRequiredEnvVars()
	default:
		return errors.New("undefined app environment")
	}
}

func checkRequiredEnvVars() error {
	requiredVars := []string{"PORT", "GOOGLE_APPLICATION_CREDENTIALS", "FIREBASE_PROJECT_ID", "REDIS_ADDR", "REDIS_PASSWORD", "SAFE_BROWSING_API_KEY"}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			return fmt.Errorf("missing required environment variable: %s", v)
		}
	}
	return nil
}
