package config

import "os"

type Config struct {
	MongoURI         string
	RedisAddr        string
	AdminToken       string
	ServerPort       string
}

func Load() *Config {
	return &Config{
		MongoURI:   getEnv("MONGO_URI", "mongodb://localhost:27017/sentinel"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		AdminToken: getEnv("SENTINEL_ADMIN_TOKEN", "changeme"),
		ServerPort: getEnv("PORT", "8090"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
