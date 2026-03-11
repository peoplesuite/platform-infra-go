// connect demonstrates connecting to Postgres with postgres.New and Config.
// Requires Postgres running; set POSTGRES_DSN to override default.
package main

import (
	"context"
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

	log.Println("connected to Postgres")
}
