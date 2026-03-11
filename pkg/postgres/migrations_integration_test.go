package postgres

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"

	_ "github.com/lib/pq"
	"github.com/peoplesuite/platform-infra-go/pkg/postgres/migrations"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// runPostgresContainer starts a Postgres container. Skips the test if testcontainers
// cannot run (e.g. no Docker or rootless on Windows).
func runPostgresContainer(t *testing.T, ctx context.Context, dbName string) *postgres.PostgresContainer {
	t.Helper()
	var ctr *postgres.PostgresContainer
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("testcontainers not available (Docker required): %v", r)
		}
	}()
	var err error
	ctr, err = postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Skipf("testcontainers not available: %v", err)
	}
	return ctr
}

func TestRunMigrations_Integration_AppliesEmbeddedMigrations(t *testing.T) {
	ctx := context.Background()
	ctr := runPostgresContainer(t, ctx, "migrations_test")
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = RunMigrations(db, migrations.FS)
	require.NoError(t, err)

	var version int
	var dirty bool
	err = db.QueryRowContext(ctx, "SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
	require.NoError(t, err)
	require.Equal(t, 1, version)
	require.False(t, dirty)
}

func TestRunMigrations_Integration_Idempotent_ReturnsNilWhenUpToDate(t *testing.T) {
	ctx := context.Background()
	ctr := runPostgresContainer(t, ctx, "migrations_idempotent_test")
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = RunMigrations(db, migrations.FS)
	require.NoError(t, err)

	err = RunMigrations(db, migrations.FS)
	require.NoError(t, err)
}

func TestRunMigrations_Integration_CustomFS_AppliesAndVerifies(t *testing.T) {
	ctx := context.Background()
	ctr := runPostgresContainer(t, ctx, "migrations_custom_test")
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	migrationsFS := fstest.MapFS{
		"000000000001_create_widgets.up.sql":   {Data: []byte("CREATE TABLE IF NOT EXISTS widgets (id SERIAL PRIMARY KEY, name TEXT);\n")},
		"000000000001_create_widgets.down.sql": {Data: []byte("DROP TABLE IF EXISTS widgets;\n")},
	}

	err = RunMigrations(db, migrationsFS)
	require.NoError(t, err)

	var exists bool
	err = db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'widgets')",
	).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists)
}
