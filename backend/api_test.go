package main

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_TTL", "24h")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "120")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" || cfg.MongoDatabase != "smm_panel" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
