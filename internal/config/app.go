package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Address     string
	DatabaseURL string
	JWTSecret   string
}

func NewAppConfig() (AppConfig, error) {
	if err := godotenv.Load(); err != nil {
		return AppConfig{}, fmt.Errorf("load env: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return AppConfig{}, fmt.Errorf("DATABASE_URL is empty")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return AppConfig{}, fmt.Errorf("JWT_SECRET is empty")
	}

	return AppConfig{
		Address:     os.Getenv("ADDRESS"),
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
	}, nil
}
