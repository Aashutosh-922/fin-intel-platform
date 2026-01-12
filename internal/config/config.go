package config

import (
	"log"
	"os"
)

type Config struct {
	ServiceName string

	HTTPPort string

	PostgresDSN  string
	KafkaBrokers string
}

func Load() Config {
	cfg := Config{
		ServiceName:  mustGet("SERVICE_NAME"),
		HTTPPort:     getOrDefault("HTTP_PORT", "8080"),
		PostgresDSN:  mustGet("POSTGRES_DSN"),
		KafkaBrokers: mustGet("KAFKA_BROKERS"),
	}
	return cfg
}

func mustGet(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return val
}

func getOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
