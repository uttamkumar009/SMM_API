package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"sort"
	"strings"
)

type Service struct {
	ID       int     `json:"service"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Rate     float64 `json:"rate"`
	Min      int     `json:"min"`
	Max      int     `json:"max"`
	Dripfeed bool    `json:"dripfeed"`
	Refill   bool    `json:"refill"`
	Cancel   bool    `json:"cancel"`
	Category string  `json:"category"`
}

var demoServices = []Service{
	{ID: 4179, Name: "Instagram Followers [Speed 20K+/Day] [Refill 30Days]", Type: "Default", Rate: 371.79, Min: 10, Max: 100000, Refill: true, Cancel: true, Category: "Instagram Followers"},
	{ID: 633, Name: "Instagram Likes [Indian] [Speed 50K/Day]", Type: "Default", Rate: 33.14, Min: 10, Max: 500000, Cancel: true, Category: "Instagram Likes"},
	{ID: 2001, Name: "YouTube Views [High Retention] [Instant Start]", Type: "Default", Rate: 18.50, Min: 100, Max: 1000000, Cancel: true, Category: "YouTube Views"},
	{ID: 2002, Name: "YouTube Subscribers [Refill 30 Days]", Type: "Default", Rate: 245.00, Min: 10, Max: 100000, Refill: true, Cancel: true, Category: "YouTube Subscribers"},
	{ID: 3001, Name: "Telegram Members [Worldwide] [Fast Delivery]", Type: "Default", Rate: 95.00, Min: 50, Max: 500000, Cancel: true, Category: "Telegram Members"},
	{ID: 3002, Name: "Telegram Post Views [Instant]", Type: "Default", Rate: 12.00, Min: 100, Max: 1000000, Cancel: true, Category: "Telegram Views"},
	{ID: 4001, Name: "Twitter/X Followers [Refill 30 Days]", Type: "Default", Rate: 180.00, Min: 10, Max: 100000, Refill: true, Cancel: true, Category: "Twitter Followers"},
	{ID: 4002, Name: "Twitter/X Likes [Fast Start]", Type: "Default", Rate: 42.00, Min: 20, Max: 100000, Cancel: true, Category: "Twitter Likes"},
	{ID: 5001, Name: "TikTok Video Views [Fast]", Type: "Default", Rate: 8.50, Min: 100, Max: 5000000, Cancel: true, Category: "TikTok Views"},
	{ID: 5002, Name: "TikTok Followers [Refill 30 Days]", Type: "Default", Rate: 210.00, Min: 10, Max: 100000, Refill: true, Cancel: true, Category: "TikTok Followers"},
}

func loadServices(path string) ([]Service, error) {
	if strings.TrimSpace(path) == "" {
		return demoServices, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Services []Service `json:"services"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if len(payload.Services) == 0 {
		return nil, errors.New("catalog must contain at least one service")
	}
	return payload.Services, nil
}

func categories(services []Service) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, service := range services {
		if service.Category != "" && !seen[service.Category] {
			seen[service.Category] = true
			result = append(result, service.Category)
		}
	}
	sort.Strings(result)
	return result
}

func filteredServices(services []Service, category string) []Service {
	if category == "" {
		return services
	}
	result := make([]Service, 0)
	for _, service := range services {
		if strings.EqualFold(service.Category, category) {
			result = append(result, service)
		}
	}
	return result
}

func serviceByID(services []Service, id int) (Service, bool) {
	for _, service := range services {
		if service.ID == id {
			return service, true
		}
	}
	return Service{}, false
}

func categoryQuery(raw string) string {
	value, err := url.QueryUnescape(raw)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return strings.TrimSpace(value)
}
