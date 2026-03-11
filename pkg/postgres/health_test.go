package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestHealth_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	mock.ExpectPing()

	err = Health(ctx, db)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHealth_PingFails(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	pingErr := errors.New("ping failed")
	mock.ExpectPing().WillReturnError(pingErr)

	err = Health(ctx, db)
	require.Error(t, err)
	require.Equal(t, pingErr, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
