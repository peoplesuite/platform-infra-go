package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestWithTx_BeginTxFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	beginErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(beginErr)

	err = WithTx(ctx, db, func(*sql.Tx) error { return nil })
	require.Error(t, err)
	require.Equal(t, beginErr, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_FnReturnsError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectRollback()

	fnErr := errors.New("tx work failed")
	err = WithTx(ctx, db, func(*sql.Tx) error { return fnErr })
	require.Error(t, err)
	require.Equal(t, fnErr, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_Success_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectCommit()

	err = WithTx(ctx, db, func(tx *sql.Tx) error { return nil })
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
