package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	DBMaxConns      int32
	DBMinConns      int32
	DBMaxConnLife   time.Duration
	DBMaxConnIdle   time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:             getEnv("APP_ENV", "development"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		DBMaxConns:      int32(getEnvInt("DB_MAX_CONNS", 20)),
		DBMinConns:      int32(getEnvInt("DB_MIN_CONNS", 2)),
		DBMaxConnLife:   getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
		DBMaxConnIdle:   getEnvDuration("DB_MAX_CONN_IDLE", 30*time.Minute),
		ReadTimeout:     getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 20*time.Second),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}