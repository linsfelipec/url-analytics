package url

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/linsfelipec/url-analytics/internal/database"
)

func setupTestRepository(t *testing.T) (*Repository, *sql.DB) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL env variable is required")
	}

	const timeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := database.Open(ctx, dsn)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return NewRepository(db), db
}

func generateTestCode(t *testing.T) string {
	t.Helper()

	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("generating test code: %v", err)
	}

	return hex.EncodeToString(b)
}

func TestRepository_Create(t *testing.T) {
	repo, db := setupTestRepository(t)

	code := generateTestCode(t)
	originalURL := "https://example.ccom/products"

	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			"DELETE FROM urls WHERE code = $1", code)
	})

	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	iURL, err := repo.Create(ctx, code, originalURL)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if iURL.ID == 0 {
		t.Error("expected generated ID, got zero")
	}

	if iURL.Code != code {
		t.Errorf("expected code %q, got %q", code, iURL.Code)
	}

	if iURL.OriginalURL != originalURL {
		t.Errorf("expected URL %q, got %q", originalURL, iURL.OriginalURL)
	}

	if iURL.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}

	retrieved, err := repo.GetByCode(ctx, code)
	if err != nil {
		t.Fatalf("GetByCode() returned an error: %v", err)
	}

	if retrieved.ID != iURL.ID {
		t.Errorf("expected ID %d, got %d", iURL.ID, retrieved.ID)
	}

	if retrieved.OriginalURL != originalURL {
		t.Errorf("expected URL %q, got %q", originalURL, retrieved.OriginalURL)
	}
}

func TestRepository_GetByCodeNotFound(t *testing.T) {
	repo, _ := setupTestRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := repo.GetByCode(ctx, "nonexistent")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

}

func TestRepository_CreateDuplicateCode(t *testing.T) {
	repo, db := setupTestRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	code := generateTestCode(t)
	originalURL := "https://example.ccom/products"

	t.Cleanup(func() {
		db.ExecContext(context.Background(),
			"DELETE FROM urls WHERE code = $1", code)
	})

	_, err := repo.Create(ctx, code, originalURL)
	if err != nil {
		t.Fatalf("error inserting url: %v", err)
	}

	secondURL := "https://example.com/OtherProducts"
	_, err = repo.Create(ctx, code, secondURL)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected PostgreSQL error, got %T: %v", err, err)
	}
	if pgErr.Code != "23505" {
		t.Fatalf("expected SQLSTATE 23505, got %s", pgErr.Code)
	}

}
