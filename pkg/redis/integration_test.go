package redis

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// runRedisContainer starts a Redis container. Skips the test if testcontainers cannot run.
func runRedisContainer(t *testing.T, ctx context.Context) *tcredis.RedisContainer {
	t.Helper()
	var ctr *tcredis.RedisContainer
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("testcontainers not available (Docker required): %v", r)
		}
	}()
	var err error
	ctr, err = tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Skipf("testcontainers not available: %v", err)
	}
	return ctr
}

func optsFromContainer(t *testing.T, ctx context.Context, ctr *tcredis.RedisContainer) Options {
	t.Helper()
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	u, err := url.Parse(connStr)
	require.NoError(t, err)
	return Options{
		Addr:      u.Host,
		Password:  "",
		DB:        0,
		TLS:       false,
		VerifyTLS: false,
		Timeout:   5 * time.Second,
	}
}

func TestGetJSON_SetJSON_Integration(t *testing.T) {
	ctx := context.Background()
	ctr := runRedisContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	client, err := New(optsFromContainer(t, ctx, ctr))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	key := "test:integration:json"
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	value := testStruct{Name: "integration", Value: 99}
	err = client.SetJSON(ctx, key, value, time.Minute)
	require.NoError(t, err)

	var result testStruct
	found, err := client.GetJSON(ctx, key, &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "integration", result.Name)
	assert.Equal(t, 99, result.Value)

	var notFound testStruct
	found, err = client.GetJSON(ctx, "test:integration:nonexistent", &notFound)
	require.NoError(t, err)
	assert.False(t, found)

	_ = client.Delete(ctx, key)
}

func TestDelete_Integration(t *testing.T) {
	ctx := context.Background()
	ctr := runRedisContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	client, err := New(optsFromContainer(t, ctx, ctr))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	key := "test:integration:delete"
	err = client.SetJSON(ctx, key, map[string]string{"k": "v"}, time.Minute)
	require.NoError(t, err)

	var result map[string]string
	found, err := client.GetJSON(ctx, key, &result)
	require.NoError(t, err)
	assert.True(t, found)

	err = client.Delete(ctx, key)
	require.NoError(t, err)

	found, err = client.GetJSON(ctx, key, &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAcquireLock_ReleaseLock_Integration(t *testing.T) {
	ctx := context.Background()
	ctr := runRedisContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	client, err := New(optsFromContainer(t, ctx, ctr))
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	lockKey := "test:integration:lock"
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

	_ = client.ReleaseLock(ctx, lockKey)
}
