package neo4j

import (
	"context"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testDriver(t *testing.T) (neo4j.DriverWithContext, Config) {
	t.Helper()
	cfg := Config{
		URI:         "bolt://localhost:7687",
		Username:    "neo4j",
		Password:    "password",
		Database:    "neo4j",
		MaxConnPool: 5,
		ConnAcquire: 5 * time.Second,
	}
	logger := zap.NewNop()
	driver, err := NewDriver(context.Background(), cfg, logger)
	if err != nil {
		t.Skipf("Neo4j not available: %v", err)
		return nil, Config{}
	}
	t.Cleanup(func() { _ = driver.Close(context.Background()) })
	return driver, cfg
}

func TestExecuteRead_Integration(t *testing.T) {
	driver, cfg := testDriver(t)
	ctx := context.Background()
	out, err := ExecuteRead(ctx, driver, cfg, func(tx neo4j.ManagedTransaction) (int, error) {
		result, err := tx.Run(ctx, "RETURN 1 AS n", nil)
		if err != nil {
			return 0, err
		}
		rec, err := result.Single(ctx)
		if err != nil {
			return 0, err
		}
		n, _ := rec.Get("n")
		return int(n.(int64)), nil
	})
	require.NoError(t, err)
	assert.Equal(t, 1, out)
}

func TestExecuteRead_ErrorPath(t *testing.T) {
	driver, cfg := testDriver(t)
	ctx := context.Background()
	_, err := ExecuteRead(ctx, driver, cfg, func(tx neo4j.ManagedTransaction) (int, error) {
		return 0, assert.AnError
	})
	require.Error(t, err)
	assert.Same(t, assert.AnError, err)
}

func TestExecuteWrite_Integration(t *testing.T) {
	driver, cfg := testDriver(t)
	ctx := context.Background()
	out, err := ExecuteWrite(ctx, driver, cfg, func(tx neo4j.ManagedTransaction) (string, error) {
		_, err := tx.Run(ctx, "RETURN 'ok' AS v", nil)
		return "ok", err
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}

func TestExecuteWrite_ErrorPath(t *testing.T) {
	driver, cfg := testDriver(t)
	ctx := context.Background()
	_, err := ExecuteWrite(ctx, driver, cfg, func(tx neo4j.ManagedTransaction) (int, error) {
		return 0, assert.AnError
	})
	require.Error(t, err)
}

func TestRun_Integration(t *testing.T) {
	driver, cfg := testDriver(t)
	ctx := context.Background()
	_, err := ExecuteRead(ctx, driver, cfg, func(tx neo4j.ManagedTransaction) (struct{}, error) {
		return struct{}{}, Run(ctx, tx, "RETURN 1", nil)
	})
	require.NoError(t, err)
}

func TestRunAndCollect_Integration(t *testing.T) {
	driver, cfg := testDriver(t)
	ctx := context.Background()
	var records []*neo4j.Record
	_, err := ExecuteRead(ctx, driver, cfg, func(tx neo4j.ManagedTransaction) ([]*neo4j.Record, error) {
		recs, err := RunAndCollect(ctx, tx, "RETURN 1 AS n", nil)
		if err != nil {
			return nil, err
		}
		records = recs
		return recs, nil
	})
	require.NoError(t, err)
	require.Len(t, records, 1)
}
