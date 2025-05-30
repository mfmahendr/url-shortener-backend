package config

import (
	"errors"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return errors.New("No .env file found")
	}

	return nil
}