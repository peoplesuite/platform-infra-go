package postgres

import (
	"context"
	"database/sql"
)

// WithTx runs fn inside a transaction; it commits on success and rolls back on error.
func WithTx(
	ctx context.Context,
	db *sql.DB,
	fn func(*sql.Tx) error,
) error {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
