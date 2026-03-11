package postgres

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations_ErrNoChange_ReturnsNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Postgres migrate driver: lock, check information_schema, create version table, read version; then unlock.
	mock.ExpectQuery("SELECT CURRENT_(SCHEMA|DATABASE)\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow("public"))
	mock.ExpectQuery("SELECT CURRENT_(SCHEMA|DATABASE)\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow("test"))
	mock.ExpectExec("SELECT pg_advisory_lock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema.tables WHERE table_schema = \\$1 AND table_name = \\$2 LIMIT 1").WithArgs("test", "schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS.*schema_migrations.*").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_advisory_unlock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_advisory_lock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT .+ FROM .*schema_migrations.*").WillReturnRows(
		sqlmock.NewRows([]string{"version", "dirty"}).AddRow(0, false),
	)
	mock.ExpectExec("SELECT pg_advisory_unlock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))

	// FS with only a placeholder migration at version 0; DB already at 0 so Up() returns ErrNoChange.
	migrationsFS := fstest.MapFS{
		"000000000000_placeholder.up.sql":   {Data: []byte("-- up\n")},
		"000000000000_placeholder.down.sql": {Data: []byte("-- down\n")},
	}
	err = RunMigrations(db, migrationsFS)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunMigrations_UpError_Wrapped(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT CURRENT_(SCHEMA|DATABASE)\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow("public"))
	mock.ExpectQuery("SELECT CURRENT_(SCHEMA|DATABASE)\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow("test"))
	mock.ExpectExec("SELECT pg_advisory_lock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema.tables WHERE table_schema = \\$1 AND table_name = \\$2 LIMIT 1").WithArgs("test", "schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS.*schema_migrations.*").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_advisory_unlock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_advisory_lock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT .+ FROM .*schema_migrations.*").WillReturnError(errors.New("db error"))
	mock.ExpectExec("SELECT pg_advisory_unlock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))

	migrationsFS := fstest.MapFS{
		"000000000000_placeholder.up.sql":   {Data: []byte("-- up\n")},
		"000000000000_placeholder.down.sql": {Data: []byte("-- down\n")},
	}
	err = RunMigrations(db, migrationsFS)
	require.Error(t, err)
	require.Contains(t, err.Error(), "migrate up")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunMigrations_WithInstanceError(t *testing.T) {
	// WithInstance with nil db would panic; test with invalid config by passing closed db
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	_ = db.Close() // closed db may cause WithInstance or later to fail
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").WillReturnError(errors.New("connection closed"))

	migrationsFS := fstest.MapFS{}
	err = RunMigrations(db, migrationsFS)
	require.Error(t, err)
}

func TestRunMigrations_ErrNoChange_IsNotWrapped(t *testing.T) {
	// Ensure we don't wrap ErrNoChange
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectQuery("SELECT CURRENT_(SCHEMA|DATABASE)\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow("public"))
	mock.ExpectQuery("SELECT CURRENT_(SCHEMA|DATABASE)\\(\\)").WillReturnRows(sqlmock.NewRows([]string{"current"}).AddRow("test"))
	mock.ExpectExec("SELECT pg_advisory_lock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM information_schema.tables WHERE table_schema = \\$1 AND table_name = \\$2 LIMIT 1").WithArgs("test", "schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS.*schema_migrations.*").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_advisory_unlock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SELECT pg_advisory_lock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT .+ FROM .*schema_migrations.*").WillReturnRows(
		sqlmock.NewRows([]string{"version", "dirty"}).AddRow(0, false),
	)
	mock.ExpectExec("SELECT pg_advisory_unlock\\(.+\\)").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))

	migrationsFS := fstest.MapFS{
		"000000000000_placeholder.up.sql":   {Data: []byte("-- up\n")},
		"000000000000_placeholder.down.sql": {Data: []byte("-- down\n")},
	}
	err = RunMigrations(db, migrationsFS)
	require.NoError(t, err)
	require.False(t, errors.Is(err, migrate.ErrNoChange), "ErrNoChange should be normalized to nil, not returned")
}
