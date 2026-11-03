package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestCatalogCategoriesAndFiltering(t *testing.T) {
	got := categories(demoServices)
	if len(got) == 0 {
		t.Fatal("expected catalog categories")
	}
	services := filteredServices(demoServices, "Telegram Members")
	if len(services) != 1 || services[0].ID != 3001 {
		t.Fatalf("unexpected filtered services: %+v", services)
	}
}

func TestCatalogHTTPHandlers(t *testing.T) {
	api := &API{services: demoServices}

	categoryResponse := httptest.NewRecorder()
	api.categories(categoryResponse, httptest.NewRequest("GET", "/api/categories", nil))
	if categoryResponse.Code != 200 || !strings.Contains(categoryResponse.Body.String(), "Telegram Members") {
		t.Fatalf("unexpected categories response: %d %s", categoryResponse.Code, categoryResponse.Body.String())
	}

	serviceResponse := httptest.NewRecorder()
	api.servicesList(serviceResponse, httptest.NewRequest("GET", "/api/services?category=YouTube%20Views", nil))
	if serviceResponse.Code != 200 || !strings.Contains(serviceResponse.Body.String(), `"service":2001`) {
		t.Fatalf("unexpected services response: %d %s", serviceResponse.Code, serviceResponse.Body.String())
	}
}

func TestLoadServicesFallback(t *testing.T) {
	services, err := loadServices("")
	if err != nil || len(services) == 0 {
		t.Fatalf("expected fallback catalog, got %d services and error %v", len(services), err)
	}
}
