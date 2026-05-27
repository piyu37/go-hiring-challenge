package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPostgresHost         = "localhost"
	defaultPostgresSSLMode      = "disable"
	defaultMaxOpenConns         = 25
	defaultMaxIdleConns         = 5
	defaultConnMaxLifetime      = 5 * time.Minute
	defaultConnectRetries       = 5
	defaultConnectRetryDelay    = 2 * time.Second
	defaultRateLimitRPS float64 = 10
)

type Database struct {
	Host              string
	Port              string
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	ConnectRetries    int
	ConnectRetryDelay time.Duration
}

type Config struct {
	HTTPPort     string
	Database     Database
	RateLimitRPS float64
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort: strings.TrimSpace(os.Getenv("HTTP_PORT")),
		Database: Database{
			Host:              envOrDefault("POSTGRES_HOST", defaultPostgresHost),
			Port:              strings.TrimSpace(os.Getenv("POSTGRES_PORT")),
			User:              strings.TrimSpace(os.Getenv("POSTGRES_USER")),
			Password:          os.Getenv("POSTGRES_PASSWORD"),
			Name:              strings.TrimSpace(os.Getenv("POSTGRES_DB")),
			SSLMode:           envOrDefault("POSTGRES_SSL_MODE", defaultPostgresSSLMode),
			MaxOpenConns:      envIntOrDefault("DB_MAX_OPEN_CONNS", defaultMaxOpenConns),
			MaxIdleConns:      envIntOrDefault("DB_MAX_IDLE_CONNS", defaultMaxIdleConns),
			ConnMaxLifetime:   envDurationOrDefault("DB_CONN_MAX_LIFETIME", defaultConnMaxLifetime),
			ConnectRetries:    envIntOrDefault("DB_CONNECT_RETRIES", defaultConnectRetries),
			ConnectRetryDelay: envDurationOrDefault("DB_CONNECT_RETRY_DELAY", defaultConnectRetryDelay),
		},
		RateLimitRPS: envFloatOrDefault("RATE_LIMIT_RPS", defaultRateLimitRPS),
	}

	var missing []string
	if cfg.HTTPPort == "" {
		missing = append(missing, "HTTP_PORT")
	}
	if cfg.Database.Port == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if cfg.Database.User == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if cfg.Database.Password == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if cfg.Database.Name == "" {
		missing = append(missing, "POSTGRES_DB")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func envFloatOrDefault(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}

	return value
}

func envDurationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}
