package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime settings loaded from the environment.
type Config struct {
	AppEnv       string
	HTTPPort     string
	DatabaseURL  string
	RedisURL     string
	KafkaBrokers string
	KafkaTopic   string
	KafkaDLQTopic string
	KafkaGroupID string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	MaxDeliveryAttempts int
	RetryBackoff        time.Duration
}

// Load reads configuration from environment variables with sensible local defaults.
// A local .env file is loaded when present (does not override existing env vars).
func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://notifyhub:notifyhub@localhost:5432/notifyhub?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers:  getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:    getEnv("KAFKA_TOPIC", "notifications.send"),
		KafkaDLQTopic: getEnv("KAFKA_DLQ_TOPIC", "notifications.dlq"),
		KafkaGroupID:  getEnv("KAFKA_GROUP_ID", "notifyhub-worker"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),

		MaxDeliveryAttempts: GetEnvInt("MAX_DELIVERY_ATTEMPTS", 3),
		RetryBackoff:        time.Duration(GetEnvInt("RETRY_BACKOFF_SECONDS", 2)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetEnvInt is a small helper for integer env vars (used later for retries, etc.).
func GetEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}
