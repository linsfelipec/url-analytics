package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/linsfelipec/url-analytics/internal/database"
	std_handler "github.com/linsfelipec/url-analytics/internal/handler/std"
	"github.com/linsfelipec/url-analytics/internal/url"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL env variable is required")
	}

	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(
		context.Background(),
		timeout,
	)
	defer cancel()

	db, err := database.Open(ctx, dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to PostgreSQL")

	repo := url.NewRepository(db)
	service := url.NewService(repo)

	stdHandler := std_handler.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/urls", stdHandler.Create)
	mux.HandleFunc("GET /api/urls/{code}", stdHandler.GetByCode)
	mux.HandleFunc("GET /{code}", stdHandler.Redirect)

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Println("HTTP server listening on :8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
