package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type response struct {
	OK      bool          `json:"ok,omitempty"`
	Message string        `json:"message,omitempty"`
	User    interface{}   `json:"user,omitempty"`
	Token   string        `json:"token,omitempty"`
	Orders  []interface{} `json:"orders,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type user struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Balance float64 `json:"balance"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && (strings.HasPrefix(origin, "https://") || origin == "http://localhost:5500") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	write := func(status int, value interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(value)
	}

	switch {
	case r.URL.Path == "/api/health" && r.Method == http.MethodGet:
		write(http.StatusOK, response{OK: true, Message: "smm-api-vercel"})
	case r.URL.Path == "/api/auth/login" && r.Method == http.MethodPost:
		var input loginRequest
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input) != nil ||
			input.Email != "demo@orbitpanel.com" || input.Password != "demo123" {
			write(http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
			return
		}
		write(http.StatusOK, response{
			Token: "vercel-demo-token-" + time.Now().UTC().Format("20060102150405"),
			User:  user{ID: "usr_demo", Name: "Alex Kumar", Email: input.Email, Balance: 248.90},
		})
	case r.URL.Path == "/api/me" && r.Method == http.MethodGet:
		write(http.StatusOK, response{User: user{ID: "usr_demo", Name: "Alex Kumar", Email: "demo@orbitpanel.com", Balance: 248.90}})
	case r.URL.Path == "/api/orders" && r.Method == http.MethodGet:
		write(http.StatusOK, response{Orders: []interface{}{}})
	default:
		write(http.StatusNotFound, map[string]string{"error": "route not found"})
	}
}
