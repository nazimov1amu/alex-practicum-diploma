package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Address              string
	DatabaseURL          string
	AccrualSystemAddress string
	JWTSecret            string
}

func NewAppConfig() (AppConfig, error) {
	_ = godotenv.Load()

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cfg := AppConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
	fs.StringVar(&cfg.Address, "a", ":8080", "run address")
	fs.StringVar(&cfg.DatabaseURL, "d", "", "database uri")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return AppConfig{}, fmt.Errorf("parse flags: %w", err)
	}

	if v := os.Getenv("RUN_ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("DATABASE_URI"); v != "" {
		cfg.DatabaseURL = v
	}
	if v := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); v != "" {
		cfg.AccrualSystemAddress = v
	}

	if cfg.DatabaseURL == "" {
		return AppConfig{}, fmt.Errorf("DATABASE_URI is empty")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "change-me"
	}
	if err := os.Setenv("JWT_SECRET", cfg.JWTSecret); err != nil {
		return AppConfig{}, fmt.Errorf("set JWT_SECRET: %w", err)
	}

	return cfg, nil
}
