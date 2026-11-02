package main

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env, Addr, PublicAPIURL, FrontendOrigin string
	JWTSecret                               string
	JWTTTL                                  time.Duration
	MongoURI, MongoDatabase, RedisURL       string
	RateLimitPerMinute                      int
	DemoEmail, DemoPassword                 string
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()
	ttl, err := time.ParseDuration(get("JWT_TTL", "24h"))
	if err != nil {
		return Config{}, err
	}
	rate, err := strconv.Atoi(get("RATE_LIMIT_PER_MINUTE", "120"))
	if err != nil || rate < 1 {
		return Config{}, errors.New("RATE_LIMIT_PER_MINUTE must be positive")
	}
	cfg := Config{
		Env: get("APP_ENV", "development"), Addr: get("APP_ADDR", ":8080"),
		PublicAPIURL:   get("PUBLIC_API_URL", "http://localhost:8080"),
		FrontendOrigin: get("FRONTEND_ORIGIN", "http://localhost:5500"),
		JWTSecret:      get("JWT_SECRET", ""), JWTTTL: ttl,
		MongoURI:           get("MONGO_URI", "mongodb://127.0.0.1:27017"),
		MongoDatabase:      get("MONGO_DATABASE", "smm_panel"),
		RedisURL:           redisURL(),
		RateLimitPerMinute: rate, DemoEmail: get("DEMO_EMAIL", "demo@orbitpanel.com"),
		DemoPassword: get("DEMO_PASSWORD", "demo123"),
	}
	if cfg.Env == "production" && len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters in production")
	}
	return cfg, nil
}

func redisURL() string {
	if value := os.Getenv("REDIS_URL"); value != "" {
		return value
	}

	address := get("REDIS_ADDRESS", "127.0.0.1:6379")
	username := os.Getenv("REDIS_USERNAME")
	password := os.Getenv("REDIS_PASSWORD")
	if !strings.HasPrefix(address, "redis://") && !strings.HasPrefix(address, "rediss://") {
		address = "redis://" + address
	}
	parsed, err := url.Parse(address)
	if err != nil {
		return address
	}
	parsed.User = nil
	if username != "" || password != "" {
		parsed.User = url.UserPassword(username, password)
	}
	if parsed.Path == "" {
		parsed.Path = "/0"
	}
	return parsed.String()
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
