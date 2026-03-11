package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewZapLogger(t *testing.T) {
	logger := zap.NewNop()
	l := NewZapLogger(logger)

	require.NotNil(t, l)
	// All levels should not panic when called with nop logger
	assert.NotPanics(t, func() { l.Debug("msg", "k", "v") })
	assert.NotPanics(t, func() { l.Info("msg", "k", "v") })
	assert.NotPanics(t, func() { l.Warn("msg", "k", "v") })
	assert.NotPanics(t, func() { l.Error("msg", "k", "v") })
}
