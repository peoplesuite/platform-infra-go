package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewFromConfig_EmptyHostPort(t *testing.T) {
	ctx := context.Background()
	cfg := ConnectionConfig{
		HostPort:  "",
		Namespace: "default",
	}
	logger := zap.NewNop()

	c, err := NewFromConfig(ctx, cfg, logger)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "address is required")
}

func TestNewFromConfig_EmptyNamespace(t *testing.T) {
	ctx := context.Background()
	cfg := ConnectionConfig{
		HostPort:  "localhost:7233",
		Namespace: "",
	}
	logger := zap.NewNop()

	c, err := NewFromConfig(ctx, cfg, logger)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "namespace is required")
}
