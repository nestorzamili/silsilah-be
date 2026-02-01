package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL string

	RedisURL string

	JWTSecret           string
	JWTAccessExpiry     time.Duration
	JWTRefreshExpiry    time.Duration

	MinIOEndpoint       string
	MinIOPublicEndpoint string
	MinIOAccessKey      string
	MinIOSecretKey      string
	MinIOBucket         string
	MinIOUseSSL         bool
	MinIOPublicUseSSL   bool

	CORSOrigins string

	ResendAPIKey string
	FromEmail    string
	Domain       string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),

		MinIOEndpoint:       getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOPublicEndpoint: getEnv("MINIO_PUBLIC_ENDPOINT", getEnv("MINIO_ENDPOINT", "localhost:9000")),
		MinIOAccessKey:      getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey:      getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:         getEnv("MINIO_BUCKET", "silsilah-media"),
		MinIOUseSSL:         getBoolEnv("MINIO_USE_SSL", false),
		MinIOPublicUseSSL:   getBoolEnv("MINIO_PUBLIC_USE_SSL", true),

		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173"),

		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@example.com"),
		Domain:       getEnv("DOMAIN", "localhost:5173"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
