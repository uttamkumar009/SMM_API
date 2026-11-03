package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	cfg      Config
	store    *Store
	redis    *redis.Client
	services []Service
}
type claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}
type loginRequest struct{ Email, Password string }
type orderRequest struct {
	Service   string  `json:"service"`
	Link      string  `json:"link"`
	ServiceID int     `json:"serviceId"`
	Quantity  int     `json:"quantity"`
	Rate      float64 `json:"rate"`
}

func NewAPI(cfg Config, store *Store, redis *redis.Client) *API {
	return &API{cfg: cfg, store: store, redis: redis, services: demoServices}
}
func (a *API) categories(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"categories": categories(a.services)})
}
func (a *API) servicesList(w http.ResponseWriter, r *http.Request) {
	category := categoryQuery(r.URL.Query().Get("category"))
	result := filteredServices(a.services, category)
	if category != "" && len(result) == 0 {
		errorJSON(w, http.StatusNotFound, "category not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"services": result})
}
func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]interface{}{"ok": true, "service": "smm-api", "time": time.Now().UTC()})
}
func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input) != nil {
		errorJSON(w, 400, "invalid request")
		return
	}
	user, err := a.store.User(r.Context(), strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		errorJSON(w, 401, "invalid email or password")
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{UserID: user.ID.Hex(), RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.cfg.JWTTTL)), IssuedAt: jwt.NewNumericDate(time.Now())}})
	signed, err := token.SignedString([]byte(a.cfg.JWTSecret))
	if err != nil {
		errorJSON(w, 500, "could not create session")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"token": signed, "user": user})
}
func (a *API) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.ParseWithClaims(raw, &claims{}, func(t *jwt.Token) (interface{}, error) { return []byte(a.cfg.JWTSecret), nil })
		if err != nil || !token.Valid {
			errorJSON(w, 401, "authentication required")
			return
		}
		c, ok := token.Claims.(*claims)
		if !ok {
			errorJSON(w, 401, "invalid session")
			return
		}
		id, err := primitive.ObjectIDFromHex(c.UserID)
		if err != nil {
			errorJSON(w, 401, "invalid session")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, id)))
	})
}

type userKey struct{}

func currentUserID(r *http.Request) (primitive.ObjectID, bool) {
	id, ok := r.Context().Value(userKey{}).(primitive.ObjectID)
	return id, ok
}
func (a *API) me(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		errorJSON(w, 401, "authentication required")
		return
	}
	user, err := a.store.UserByID(r.Context(), id)
	if err != nil {
		errorJSON(w, 404, "user not found")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"user": user})
}
func (a *API) orders(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		errorJSON(w, 401, "authentication required")
		return
	}
	result, err := a.store.Orders(r.Context(), id)
	if err != nil {
		errorJSON(w, 500, "could not load orders")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"orders": result})
}
func (a *API) transactions(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		errorJSON(w, 401, "authentication required")
		return
	}
	result, err := a.store.Transactions(r.Context(), id)
	if err != nil {
		errorJSON(w, 500, "could not load transactions")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"transactions": result})
}
func (a *API) createOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		errorJSON(w, 401, "authentication required")
		return
	}
	var input orderRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input) != nil || input.Service == "" || input.Link == "" || input.Quantity < 1 {
		errorJSON(w, 400, "service, link and positive quantity are required")
		return
	}
	if input.ServiceID > 0 {
		service, found := serviceByID(a.services, input.ServiceID)
		if !found {
			errorJSON(w, http.StatusBadRequest, "selected service not found")
			return
		}
		if input.Quantity < service.Min || input.Quantity > service.Max {
			errorJSON(w, http.StatusBadRequest, "quantity is outside the selected service limits")
			return
		}
		input.Service = service.Name
		input.Rate = service.Rate
	} else if input.Rate <= 0 || input.Rate > 100 {
		input.Rate = 2.4
	}
	amount := float64(input.Quantity) / 1000 * input.Rate
	order, err := a.store.CreateOrder(r.Context(), id, input.Service, input.Link, input.Quantity, amount)
	if err != nil {
		errorJSON(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]interface{}{"order": order})
}
func (a *API) addFunds(w http.ResponseWriter, r *http.Request) {
	id, ok := currentUserID(r)
	if !ok {
		errorJSON(w, 401, "authentication required")
		return
	}
	var input struct {
		Amount float64 `json:"amount"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input) != nil || input.Amount < 1 || input.Amount > 100000 {
		errorJSON(w, 400, "amount must be between 1 and 100000")
		return
	}
	user, err := a.store.AddFunds(r.Context(), id, input.Amount)
	if err != nil {
		errorJSON(w, 500, "could not add funds")
		return
	}
	writeJSON(w, 201, map[string]interface{}{"user": user})
}

var _ = chi.NewRouter
var _ = strconv.Itoa
