package postgres

import (
	"context"
	"database/sql"
	"time"
)

// Health checks database connectivity with a 2s timeout.
func Health(ctx context.Context, db *sql.DB) error {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
