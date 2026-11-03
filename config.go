package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env, Addr, PublicAPIURL, FrontendOrigin, ServicesCatalogFile string
	JWTSecret                                                    string
	JWTTTL                                                       time.Duration
	MongoURI, MongoDatabase, RedisURL                            string
	RateLimitPerMinute                                           int
	DemoEmail, DemoPassword                                      string
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
		PublicAPIURL:        get("PUBLIC_API_URL", "http://localhost:8080"),
		FrontendOrigin:      get("FRONTEND_ORIGIN", "http://localhost:5500"),
		ServicesCatalogFile: get("SERVICES_CATALOG_FILE", ""),
		JWTSecret:           get("JWT_SECRET", ""), JWTTTL: ttl,
		MongoURI:           get("MONGO_URI", "mongodb://127.0.0.1:27017"),
		MongoDatabase:      get("MONGO_DATABASE", "smm_panel"),
		RedisURL:           get("REDIS_URL", "redis://127.0.0.1:6379/0"),
		RateLimitPerMinute: rate, DemoEmail: get("DEMO_EMAIL", "demo@orbitpanel.com"),
		DemoPassword: get("DEMO_PASSWORD", "demo123"),
	}
	if cfg.Env == "production" && len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters in production")
	}
	return cfg, nil
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
