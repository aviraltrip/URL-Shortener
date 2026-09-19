package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	Domain      string
	DatabaseURL string
	RedisAddr   string
	RedisPass   string
	ApiQuota    int
}

func Load() *Config {
	_ = godotenv.Load()
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":3000"
	}
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost:3000"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = os.Getenv("DB_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}
	}
	redisPass := os.Getenv("REDIS_PASS")
	if redisPass == "" {
		redisPass = os.Getenv("DB_PASS")
	}
	quota := 10
	if q, err := strconv.Atoi(os.Getenv("API_QUOTA")); err == nil && q > 0 {
		quota = q
	}
	return &Config{
		AppPort:     port,
		Domain:      domain,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   redisAddr,
		RedisPass:   redisPass,
		ApiQuota:    quota,
	}
}
