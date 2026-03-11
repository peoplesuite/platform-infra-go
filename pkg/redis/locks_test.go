package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAcquireLock_ReleaseLock_Unit uses miniredis and does not require real Redis.
func TestAcquireLock_ReleaseLock_Unit(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	lockKey := "test:lock:unit"

	acquired, err := client.AcquireLock(ctx, lockKey, 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	acquired2, err := client.AcquireLock(ctx, lockKey, 10*time.Second)
	require.NoError(t, err)
	assert.False(t, acquired2)

	err = client.ReleaseLock(ctx, lockKey)
	require.NoError(t, err)

	acquired3, err := client.AcquireLock(ctx, lockKey, 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired3)
}

func TestAcquireLock_ReleaseLock(t *testing.T) {
	opts := Options{
		Addr:      "localhost:6379",
		Password:  "",
		DB:        0,
		TLS:       false,
		VerifyTLS: false,
		Timeout:   5 * time.Second,
	}

	client, err := New(opts)
	if err != nil {
		t.Skipf("Redis not available (client init failed): %v", err)
		return
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	lockKey := "test:lock:acquire"

	// Acquire lock
	acquired, err := client.AcquireLock(ctx, lockKey, 10*time.Second)
	if err != nil {
		t.Skipf("Redis not available (AcquireLock failed): %v", err)
		return
	}
	assert.True(t, acquired)

	// Try to acquire again (should fail)
	acquired2, err := client.AcquireLock(ctx, lockKey, 10*time.Second)
	if err != nil {
		t.Skipf("Redis not available (AcquireLock second attempt failed): %v", err)
		return
	}
	assert.False(t, acquired2)

	// Release lock
	err = client.ReleaseLock(ctx, lockKey)
	if err != nil {
		t.Skipf("Redis not available (ReleaseLock failed): %v", err)
		return
	}

	// Now should be able to acquire again
	acquired3, err := client.AcquireLock(ctx, lockKey, 10*time.Second)
	if err != nil {
		t.Skipf("Redis not available (AcquireLock after release failed): %v", err)
		return
	}
	assert.True(t, acquired3)

	// Cleanup
	_ = client.ReleaseLock(ctx, lockKey)
}
