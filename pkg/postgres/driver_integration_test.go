package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNew_Health_WithTx_Integration requires a running Postgres at localhost with DSN.
// Skip if Postgres is not available.
func TestNew_Health_WithTx_Integration(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()
	cfg := Config{
		DSN:             "postgres://localhost:5432/postgres?sslmode=disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := New(ctx, cfg, logger)
	if err != nil {
		t.Skipf("Postgres not available (skip integration): %v", err)
		return
	}
	defer func() { _ = db.Close() }()

	err = Health(ctx, db)
	require.NoError(t, err)

	err = WithTx(ctx, db, func(tx *sql.Tx) error { return nil })
	require.NoError(t, err)

	assert.NotNil(t, db)
}
