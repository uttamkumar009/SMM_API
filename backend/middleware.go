package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		origin := r.Header.Get("Origin")
		if origin != "" && (strings.HasPrefix(origin, "https://") || origin == "http://localhost:5500") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		next.ServeHTTP(w, r)
	})
}
func RateLimit(client *redis.Client, perMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if client == nil {
				next.ServeHTTP(w, r)
				return
			}
			key := "smm:rate:" + r.RemoteAddr
			count, err := client.Incr(r.Context(), key).Result()
			if err == nil && count == 1 {
				client.Expire(r.Context(), key, time.Minute)
			}
			if err == nil && count > int64(perMinute) {
				w.Header().Set("Retry-After", "60")
				errorJSON(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
