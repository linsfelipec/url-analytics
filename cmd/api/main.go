package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/linsfelipec/url-analytics/internal/database"
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

	var dbVersion string
	err = db.QueryRowContext(ctx, "SELECT version()").Scan(&dbVersion)
	if err != nil {
		log.Fatalf("database query failed: %v", err)
	}
	log.Printf("PostgreSQL version: %s", dbVersion)
}
