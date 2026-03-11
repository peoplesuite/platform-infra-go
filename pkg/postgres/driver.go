package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// New opens a Postgres connection, applies pool config, and pings. Caller must close the DB.
func New(ctx context.Context, cfg Config, logger *zap.Logger) (*sql.DB, error) {

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}

	applyPoolConfig(db, cfg)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	logger.Info("postgres connected")

	return db, nil
}
