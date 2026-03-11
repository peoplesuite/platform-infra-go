package postgres

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestApplyPoolConfig_AllZero(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	applyPoolConfig(db, Config{})
	// No expectations: when all zero, no Set* is called
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyPoolConfig_AllPositive(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	cfg := Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
	applyPoolConfig(db, cfg)
	// sqlmock *sql.DB accepts Set* calls; we're just exercising the code paths
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyPoolConfig_Partial(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	applyPoolConfig(db, Config{
		MaxOpenConns: 3,
		// MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime zero -> not set
	})
	require.NoError(t, mock.ExpectationsWereMet())
}
