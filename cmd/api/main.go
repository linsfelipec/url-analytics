package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/linsfelipec/url-analytics/internal/database"
	gin_handler "github.com/linsfelipec/url-analytics/internal/handler/gin"
	std_handler "github.com/linsfelipec/url-analytics/internal/handler/std"
	"github.com/linsfelipec/url-analytics/internal/url"
)

type HandlerType int

const (
	stdHandler = iota
	ginHandler
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
	dbCtx, cancel := context.WithTimeout(
		context.Background(),
		timeout,
	)
	defer cancel()

	db, err := database.Open(dbCtx, dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to PostgreSQL")

	repo := url.NewRepository(db)
	service := url.NewService(repo)

	h, err := NewHandler(ginHandler, service, os.Getenv("BASE_URL"))
	if err != nil {
		log.Fatalf("error %v", err)
	}

	server := http.Server{
		Addr:         ":8080",
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Println("HTTP server listening on :8080")

	serverCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	serverError := make(chan error, 1)
	go func() {
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
		log.Println("HTTP server stopped")

	case <-serverCtx.Done():
		log.Println("Shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			server.Close()
		}
	}

}

func NewHandler(hType HandlerType, service *url.Service, baseUrl string) (http.Handler, error) {
	switch hType {
	case stdHandler:
		stdH, err := std_handler.NewHandler(service, baseUrl)
		if err != nil {
			return nil, fmt.Errorf("Error creating handler: %v", err)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/urls", stdH.Create)
		mux.HandleFunc("GET /api/urls/{code}", stdH.GetByCode)
		mux.HandleFunc("GET /{code}", stdH.Redirect)
		return mux, nil
	case ginHandler:
		ginH, err := gin_handler.NewHandler(service, baseUrl)
		if err != nil {
			return nil, fmt.Errorf("Error creating handler: %v", err)
		}
		router := gin.Default()
		router.POST("/api/urls", ginH.Create)
		router.GET("/api/urls/:code", ginH.GetByCode)
		router.GET("/:code", ginH.Redirect)
		return router, err
	default:
		return nil, fmt.Errorf("Error creating handler: framework not defined")
	}
}
