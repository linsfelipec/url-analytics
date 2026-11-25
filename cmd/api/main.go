package main

import (
	"context"
	"flag"
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
	"github.com/linsfelipec/url-analytics/internal/handler"
	gin_handler "github.com/linsfelipec/url-analytics/internal/handler/gin"
	std_handler "github.com/linsfelipec/url-analytics/internal/handler/std"
	"github.com/linsfelipec/url-analytics/internal/url"
)

type HandlerType string

const (
	stdHandler = "std"
	ginHandler = "gin"
)

func main() {
	framework := flag.String(
		"f",
		"gin",
		"HTTP framework to use: gin or std",
	)

	flag.Parse()

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

	handlerType, err := parseHandlerType(*framework)
	if err != nil {
		log.Fatal(err)
	}

	h, err := NewHandler(
		handlerType,
		service,
		os.Getenv("BASE_URL"),
		db,
	)
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

func NewHandler(hType HandlerType, service *url.Service, baseURL string, db handler.ReadyRepository) (http.Handler, error) {
	switch hType {
	case stdHandler:
		stdH, err := std_handler.NewHandler(service, baseURL)
		if err != nil {
			return nil, fmt.Errorf("error creating handler: %w", err)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/urls", stdH.Create)
		mux.HandleFunc("GET /api/urls/{code}", stdH.GetByCode)
		mux.HandleFunc("GET /{code}", stdH.Redirect)
		mux.HandleFunc("GET /healthz", handler.Healthz)
		mux.HandleFunc("GET /readyz", handler.Readyz(db))
		return mux, nil
	case ginHandler:
		ginH, err := gin_handler.NewHandler(service, baseURL)
		if err != nil {
			return nil, fmt.Errorf("error creating handler: %w", err)
		}
		router := gin.Default()
		router.POST("/api/urls", ginH.Create)
		router.GET("/api/urls/:code", ginH.GetByCode)
		router.GET("/:code", ginH.Redirect)
		router.GET("/healthz", gin.WrapF(handler.Healthz))
		router.GET("GET /readyz", gin.WrapF(handler.Readyz(db)))
		return router, nil
	default:
		return nil, fmt.Errorf("Error creating handler: framework not defined")
	}
}

func parseHandlerType(framework string) (HandlerType, error) {
	switch framework {
	case "gin":
		return ginHandler, nil
	case "std":
		return stdHandler, nil
	default:
		return "", fmt.Errorf("unsupported framework: %s", framework)
	}
}
