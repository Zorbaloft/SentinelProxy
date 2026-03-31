package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	UpstreamURL      string
	UpstreamHost     string
	RedisAddr        string
	MongoURI         string
	LogTTLDays       int
	LogBodyMaxBytes  int64
	ServerPort       string
	RateCounterTTL   time.Duration
}

func Load() *Config {
	ttlDays := 4
	if days := os.Getenv("LOG_TTL_DAYS"); days != "" {
		if d, err := strconv.Atoi(days); err == nil {
			ttlDays = d
		}
	}

	bodyMaxBytes := int64(1048576) // 1MB default
	if max := os.Getenv("LOG_BODY_MAX_BYTES"); max != "" {
		if m, err := strconv.ParseInt(max, 10, 64); err == nil {
			bodyMaxBytes = m
		}
	}

	upstreamURL := os.Getenv("UPSTREAM_URL")
	if upstreamURL == "" {
		upstreamURL = "https://3cket.com"
	}

	upstreamHost := getEnv("UPSTREAM_HOST", "3cket.local")

	return &Config{
		UpstreamURL:     upstreamURL,
		UpstreamHost:    upstreamHost,
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017/sentinel"),
		LogTTLDays:      ttlDays,
		LogBodyMaxBytes: bodyMaxBytes,
		ServerPort:      getEnv("PORT", "8080"),
		RateCounterTTL: 60 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
