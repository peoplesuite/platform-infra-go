package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
	// This test requires a running Redis instance
	// Skip if Redis is not available
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

	require.NotNil(t, client)
	assert.NotNil(t, client.rdb)
}

func TestNew_InvalidAddress(t *testing.T) {
	opts := Options{
		Addr:      "localhost:9999",
		Password:  "",
		DB:        0,
		TLS:       false,
		VerifyTLS: false,
		Timeout:   1 * time.Second,
	}

	client, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Operations against an invalid address should fail when used, even if
	// client creation itself succeeds.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, getErr := client.GetJSON(ctx, "test:key", &struct{}{})
	require.Error(t, getErr)
}

func TestNew_WithTLS_CreatesClient(t *testing.T) {
	// TLS branch: opts.TLS true sets TLSConfig (InsecureSkipVerify = !VerifyTLS)
	opts := Options{
		Addr:      "localhost:6379",
		TLS:       true,
		VerifyTLS: false,
		Timeout:   time.Second,
	}
	client, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, client)
	_ = client.Close()
}

func TestNew_WithTLS_VerifyTLSTrue(t *testing.T) {
	opts := Options{
		Addr:      "localhost:6379",
		TLS:       true,
		VerifyTLS: true,
		Timeout:   time.Second,
	}
	client, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, client)
	_ = client.Close()
}
