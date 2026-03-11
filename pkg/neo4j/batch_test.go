package neo4j

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBatchWriter(t *testing.T) {
	// NewBatchWriter with nil driver is allowed (caller's responsibility)
	logger := zap.NewNop()
	w := NewBatchWriter(nil, Config{Database: "neo4j"}, logger)
	require.NotNil(t, w)
	// Default batch size
	w2 := w.WithBatchSize(0)
	assert.Same(t, w, w2)
}

func TestWithBatchSize(t *testing.T) {
	logger := zap.NewNop()
	w := NewBatchWriter(nil, Config{Database: "neo4j"}, logger)

	// Positive size: returns same receiver (chainable)
	w2 := w.WithBatchSize(100)
	assert.Same(t, w, w2)

	// Zero and negative: no change (still chainable)
	w3 := w.WithBatchSize(0)
	assert.Same(t, w, w3)
	w4 := w.WithBatchSize(-1)
	assert.Same(t, w, w4)
}

func TestBatchWriter_Write_EmptyItems(t *testing.T) {
	logger := zap.NewNop()
	w := NewBatchWriter(nil, Config{Database: "neo4j"}, logger)
	ctx := context.Background()

	// Empty slice: returns nil without calling driver
	err := w.Write(ctx, "UNWIND $batch AS row CREATE (n:Node) SET n = row", nil)
	assert.NoError(t, err)

	err = w.Write(ctx, "UNWIND $batch AS row RETURN row", []map[string]any{})
	assert.NoError(t, err)
}

func TestBatchWriter_DeleteByIDs_EmptyIds(t *testing.T) {
	logger := zap.NewNop()
	w := NewBatchWriter(nil, Config{Database: "neo4j"}, logger)
	ctx := context.Background()

	err := w.DeleteByIDs(ctx, "UNWIND $ids AS id MATCH (n) WHERE id(n) = id DETACH DELETE n", nil)
	assert.NoError(t, err)

	err = w.DeleteByIDs(ctx, "UNWIND $ids AS id DELETE n", []int{})
	assert.NoError(t, err)
}

func TestDefaultBatchSize(t *testing.T) {
	assert.Equal(t, 500, DefaultBatchSize)
}
