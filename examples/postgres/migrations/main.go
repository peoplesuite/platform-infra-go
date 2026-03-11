// migrations demonstrates postgres.RunMigrations with a local embed FS.
// Requires Postgres running; set POSTGRES_DSN to override default.
package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/peoplesuite/platform-infra-go/pkg/postgres"
	"go.uber.org/zap"
)

//go:embed sql/*.sql
var migrationsEmbed embed.FS

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

	// Use sub-FS so the root contains the migration files (e.g. 000001_example.up.sql).
	sqlFS, err := fs.Sub(migrationsEmbed, "sql")
	if err != nil {
		log.Fatalf("fs.Sub: %v", err)
	}

	if err := postgres.RunMigrations(db, sqlFS); err != nil {
		log.Fatalf("RunMigrations failed: %v", err)
	}
	log.Println("migrations applied (or already up to date)")
}
