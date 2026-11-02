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
