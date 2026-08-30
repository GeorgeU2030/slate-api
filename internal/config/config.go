package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	AppEnv  string
	DBURL   string

	JWTSecret                   string
	JWTAccessTokenExpireMinutes int
	RefreshTokenExpireDays      int

	TMDBAPIKey string

	GoogleClientID     string
	GoogleClientSecret string

	FrontendURL string
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func Load() (*Config, error) {
	// Load environment variables from .env file
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		DBURL:       os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TMDBAPIKey:  os.Getenv("TMDB_API_KEY"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}

	accessMin, err := getEnvInt("JWT_ACCESS_TOKEN_EXPIRE_MINUTES", 15)
	if err != nil {
		return nil, err
	}
	cfg.JWTAccessTokenExpireMinutes = accessMin

	refreshDays, err := getEnvInt("REFRESH_TOKEN_EXPIRE_DAYS", 30)
	if err != nil {
		return nil, err
	}
	cfg.RefreshTokenExpireDays = refreshDays

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	missing := []string{}

	if c.DBURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.TMDBAPIKey == "" {
		missing = append(missing, "TMDB_API_KEY")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s: %w", key, err)
	}
	return n, nil
}