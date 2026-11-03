package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())
	if err = mongoClient.Ping(ctx, nil); err != nil {
		log.Printf("mongo unavailable: %v", err)
	}
	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal(err)
	}
	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("redis unavailable: %v", err)
	}
	store := NewStore(mongoClient.Database(cfg.MongoDatabase), cfg)
	services, err := loadServices(cfg.ServicesCatalogFile)
	if err != nil {
		log.Fatalf("could not load services catalog: %v", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.DemoPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	if err = store.EnsureDemo(ctx, string(hash)); err != nil {
		log.Printf("demo user seed skipped: %v", err)
	}
	api := NewAPI(cfg, store, redisClient)
	api.services = services
	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.RequestID, middleware.Recoverer, SecurityHeaders, RateLimit(redisClient, cfg.RateLimitPerMinute))
	r.Get("/health", api.health)
	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", api.login)
		r.Get("/categories", api.categories)
		r.Get("/services", api.servicesList)
		r.Group(func(r chi.Router) {
			r.Use(api.requireAuth)
			r.Get("/me", api.me)
			r.Get("/orders", api.orders)
			r.Get("/transactions", api.transactions)
			r.Post("/orders", api.createOrder)
			r.Post("/wallet/add-funds", api.addFunds)
		})
	})
	server := &http.Server{Addr: cfg.Addr, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("SMM API listening on %s", cfg.Addr)
		if e := server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			log.Fatal(e)
		}
	}()
	<-ctx.Done()
	shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	server.Shutdown(shutdown)
}
