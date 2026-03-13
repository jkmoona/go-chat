package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port               string
	DatabaseURL        string
	ClientURL          string
	AccessTokenSecret  string
	RefreshTokenSecret string
	SecureCookies      bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		ClientURL:          getEnv("CLIENT_URL", "http://localhost:5173"),
		AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
		RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
		SecureCookies:      os.Getenv("SECURE_COOKIES") != "false",
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.AccessTokenSecret == "" {
		return nil, fmt.Errorf("ACCESS_TOKEN_SECRET is required")
	}
	if cfg.RefreshTokenSecret == "" {
		return nil, fmt.Errorf("REFRESH_TOKEN_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
