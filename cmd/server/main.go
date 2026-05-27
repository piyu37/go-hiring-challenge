package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"github.com/mytheresa/go-hiring-challenge/app/config"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/app/health"
	"github.com/mytheresa/go-hiring-challenge/app/middleware"
	"github.com/mytheresa/go-hiring-challenge/models"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, closeDB, err := database.Open(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Database connection failed: %s", err)
	}
	defer closeDB()

	productRepo := models.NewProductsRepository(db)
	categoryRepo := models.NewCategoriesRepository(db)

	catalogHandler := catalog.NewHandler(productRepo, categoryRepo)
	categoryHandler := categories.NewHandler(categoryRepo)
	healthHandler := health.NewHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", healthHandler.HandleLive)
	mux.HandleFunc("GET /health/ready", healthHandler.HandleReady)
	mux.HandleFunc("GET /catalog", catalogHandler.HandleList)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleGetByCode)
	mux.HandleFunc("GET /categories", categoryHandler.HandleList)
	mux.HandleFunc("POST /categories", categoryHandler.HandleCreate)

	var handler http.Handler = mux
	if rateLimiter := middleware.NewRateLimit(cfg.RateLimitRPS); rateLimiter != nil {
		handler = rateLimiter.Middleware(handler)
	}
	handler = middleware.Recovery(handler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}
		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %s", err)
	}
}
