package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Provider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type ProviderService struct {
	ID       int     `json:"service"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Rate     float64 `json:"rate"`
	Min      int     `json:"min"`
	Max      int     `json:"max"`
	Refill   bool    `json:"refill"`
	Cancel   bool    `json:"cancel"`
	Category string  `json:"category"`
}

func (s *ProviderService) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.ID = rawInt(raw["service"])
	s.Name = rawString(raw["name"])
	s.Type = rawString(raw["type"])
	s.Rate = rawFloat(raw["rate"])
	s.Min = rawInt(raw["min"])
	s.Max = rawInt(raw["max"])
	s.Category = rawString(raw["category"])
	s.Refill = rawBool(raw["refill"])
	s.Cancel = rawBool(raw["cancel"])
	return nil
}

func rawString(value json.RawMessage) string {
	var text string
	if json.Unmarshal(value, &text) == nil {
		return text
	}
	return strings.Trim(string(value), `"`)
}

func rawFloat(value json.RawMessage) float64 {
	var number float64
	if json.Unmarshal(value, &number) == nil {
		return number
	}
	number, _ = strconv.ParseFloat(rawString(value), 64)
	return number
}

func rawInt(value json.RawMessage) int {
	var number int
	if json.Unmarshal(value, &number) == nil {
		return number
	}
	number, _ = strconv.Atoi(rawString(value))
	return number
}

func rawBool(value json.RawMessage) bool {
	var result bool
	if json.Unmarshal(value, &result) == nil {
		return result
	}
	return strings.EqualFold(rawString(value), "true") || rawString(value) == "1"
}

func NewProvider(cfg Config) *Provider {
	return &Provider{
		baseURL: strings.TrimRight(cfg.ProviderAPIURL, "/"),
		apiKey:  cfg.ProviderAPIKey,
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *Provider) enabled() bool {
	return p.apiKey != "" && p.baseURL != ""
}

func (p *Provider) request(ctx context.Context, values url.Values, target interface{}) error {
	if !p.enabled() {
		return errors.New("provider API is not configured")
	}
	values.Set("key", p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "BabiesiQ-SMM/1.0")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("provider returned HTTP %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return errors.New("provider returned invalid JSON")
	}
	return nil
}

func (p *Provider) Services(ctx context.Context) ([]ProviderService, error) {
	var result []ProviderService
	if err := p.request(ctx, url.Values{"action": {"services"}}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *Provider) Balance(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := p.request(ctx, url.Values{"action": {"balance"}}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *Provider) AddOrder(ctx context.Context, serviceID int, data map[string]string) (map[string]interface{}, error) {
	values := url.Values{"action": {"add"}, "service": {strconv.Itoa(serviceID)}}
	for key, value := range data {
		values.Set(key, value)
	}
	var result map[string]interface{}
	if err := p.request(ctx, values, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *Provider) Action(ctx context.Context, action string, key string, value string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := p.request(ctx, url.Values{"action": {action}, key: {value}}, &result); err != nil {
		return nil, err
	}
	return result, nil
}
