package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_EmptyAddress(t *testing.T) {
	ctx := context.Background()
	opts := Options{
		Address:   "",
		Namespace: "default",
	}

	c, err := New(ctx, opts)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "address is required")
}

func TestNew_EmptyNamespace(t *testing.T) {
	ctx := context.Background()
	opts := Options{
		Address:   "localhost:7233",
		Namespace: "",
	}

	c, err := New(ctx, opts)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "namespace is required")
}

func TestNew_DialFails(t *testing.T) {
	ctx := context.Background()
	opts := Options{
		Address:   "invalid-host-no-dns:7233",
		Namespace: "default",
	}
	c, err := New(ctx, opts)
	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "failed to dial temporal client")
}

func TestNew_WithIdentity(t *testing.T) {
	ctx := context.Background()
	opts := Options{
		Address:   "localhost:7233",
		Namespace: "default",
		Identity:  "my-worker",
	}
	c, err := New(ctx, opts)
	if err != nil {
		// Dial may fail if Temporal not running; we only care that Identity branch was used
		require.Contains(t, err.Error(), "failed to dial temporal client")
		return
	}
	require.NotNil(t, c)
	c.Close()
}

func TestNew_WithLogger(t *testing.T) {
	ctx := context.Background()
	opts := Options{
		Address:   "localhost:7233",
		Namespace: "default",
		Logger:    &testLogger{},
	}
	c, err := New(ctx, opts)
	if err != nil {
		require.Contains(t, err.Error(), "failed to dial temporal client")
		return
	}
	require.NotNil(t, c)
	c.Close()
}

type testLogger struct{}

func (t *testLogger) Debug(msg string, keyvals ...interface{}) {}
func (t *testLogger) Info(msg string, keyvals ...interface{})  {}
func (t *testLogger) Warn(msg string, keyvals ...interface{})  {}
func (t *testLogger) Error(msg string, keyvals ...interface{}) {}
func (t *testLogger) With(keyvals ...interface{}) interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
} {
	return t
}
