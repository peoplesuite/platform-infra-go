package neo4j

import (
	"context"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcneo4j "github.com/testcontainers/testcontainers-go/modules/neo4j"
	"go.uber.org/zap"
)

const neo4jTestPassword = "testcontainers-neo4j"

// runNeo4jContainer starts a Neo4j container. Skips the test if testcontainers cannot run.
func runNeo4jContainer(t *testing.T, ctx context.Context) *tcneo4j.Neo4jContainer {
	t.Helper()
	var ctr *tcneo4j.Neo4jContainer
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("testcontainers not available (Docker required): %v", r)
		}
	}()
	var err error
	ctr, err = tcneo4j.Run(ctx, "neo4j:5-community",
		tcneo4j.WithAdminPassword(neo4jTestPassword),
	)
	if err != nil {
		t.Skipf("testcontainers not available: %v", err)
	}
	return ctr
}

func TestNewDriver_Integration(t *testing.T) {
	ctx := context.Background()
	ctr := runNeo4jContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	boltURL, err := ctr.BoltUrl(ctx)
	require.NoError(t, err)

	logger := zap.NewNop()
	cfg := Config{
		URI:         boltURL,
		Username:    "neo4j",
		Password:    neo4jTestPassword,
		Database:    "neo4j",
		MaxConnPool: 5,
		ConnAcquire: 5 * time.Second,
		MaxConnLife: 10 * time.Minute,
		LogLevel:    "warn",
	}
	driver, err := NewDriver(ctx, cfg, logger)
	require.NoError(t, err)
	defer func() { _ = driver.Close(ctx) }()
	assert.NotNil(t, driver)
}

func TestQuery_Integration(t *testing.T) {
	ctx := context.Background()
	ctr := runNeo4jContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	boltURL, err := ctr.BoltUrl(ctx)
	require.NoError(t, err)

	logger := zap.NewNop()
	cfg := Config{
		URI:         boltURL,
		Username:    "neo4j",
		Password:    neo4jTestPassword,
		Database:    "neo4j",
		MaxConnPool: 5,
	}
	driver, err := NewDriver(ctx, cfg, logger)
	require.NoError(t, err)
	defer func() { _ = driver.Close(ctx) }()

	items, err := Query(ctx, driver, cfg, "RETURN 1 AS n", nil, func(rec *neo4j.Record) (int, error) {
		return Int(rec, "n"), nil
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, 1, items[0])
}

func TestQuerySingle_Integration(t *testing.T) {
	ctx := context.Background()
	ctr := runNeo4jContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	boltURL, err := ctr.BoltUrl(ctx)
	require.NoError(t, err)

	logger := zap.NewNop()
	cfg := Config{
		URI:         boltURL,
		Username:    "neo4j",
		Password:    neo4jTestPassword,
		Database:    "neo4j",
		MaxConnPool: 5,
	}
	driver, err := NewDriver(ctx, cfg, logger)
	require.NoError(t, err)
	defer func() { _ = driver.Close(ctx) }()

	one, err := QuerySingle(ctx, driver, cfg, "RETURN 42 AS v", nil, func(rec *neo4j.Record) (int, error) {
		return Int(rec, "v"), nil
	})
	require.NoError(t, err)
	require.NotNil(t, one)
	assert.Equal(t, 42, *one)
}
