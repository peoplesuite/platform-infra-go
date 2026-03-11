// basic demonstrates connect, health check, and WithTx.
// Requires Postgres running; set POSTGRES_DSN to override default.
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/peoplesuite/platform-infra-go/pkg/postgres"
	"go.uber.org/zap"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://localhost:5432/postgres?sslmode=disable"
	}

	ctx := context.Background()
	logger := zap.NewNop()
	cfg := postgres.Config{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := postgres.New(ctx, cfg, logger)
	if err != nil {
		log.Fatalf("Postgres not available: %v (set POSTGRES_DSN if needed)", err)
	}
	defer func() { _ = db.Close() }()

	if err := postgres.Health(ctx, db); err != nil {
		log.Fatalf("health check failed: %v", err)
	}
	log.Println("health check OK")

	err = postgres.WithTx(ctx, db, func(tx *sql.Tx) error {
		// Example: run a no-op or simple query inside a transaction
		_, err := tx.ExecContext(ctx, "SELECT 1")
		return err
	})
	if err != nil {
		log.Fatalf("WithTx failed: %v", err)
	}
	log.Println("transaction committed")
}
