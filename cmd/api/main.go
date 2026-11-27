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
	"github.com/linsfelipec/url-analytics/internal/logger"
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

	if err := run(*framework); err != nil {
		log.Fatal(err)
	}
}

func run(framework string) error {
	logger := logger.New(os.Stdout)

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("No .env file found, %w", err)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL env variable is required")
	}

	const timeout = 5 * time.Second
	dbCtx, cancel := context.WithTimeout(
		context.Background(),
		timeout,
	)
	defer cancel()

	db, err := database.Open(dbCtx, dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	logger.Info("database connected")

	repo := url.NewRepository(db)
	service := url.NewService(repo)

	handlerType, err := parseHandlerType(framework)
	if err != nil {
		return err
	}

	h, err := NewHandler(
		handlerType,
		service,
		os.Getenv("BASE_URL"),
		db,
	)
	if err != nil {
		return err
	}

	server := http.Server{
		Addr:         ":8080",
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	logger.Info("HTTP server starting", "address", server.Addr)

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
			return fmt.Errorf("HTTP server failed: %w", err)
		}
		logger.Info("HTTP server stopped")

	case <-serverCtx.Done():
		logger.Info("Shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
			return fmt.Errorf("Graceful shutdown failed: %w", err)
		}
	}
	return nil
}

func NewHandler(hType HandlerType, service *url.Service, baseURL string, readyRepo handler.ReadyRepository) (http.Handler, error) {
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
		mux.HandleFunc("GET /readyz", handler.Readyz(readyRepo))
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
		router.GET("/readyz", gin.WrapF(handler.Readyz(readyRepo)))
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
