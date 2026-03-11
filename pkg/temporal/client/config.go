package client

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// ConnectionConfig holds minimal connection settings (e.g. from env or app config).
// Use NewFromConfig when you have HostPort/Namespace from another config source.
type ConnectionConfig struct {
	HostPort  string
	Namespace string
}

// NewFromConfig creates a Temporal client from connection config and zap logger.
func NewFromConfig(ctx context.Context, cfg ConnectionConfig, logger *zap.Logger) (client.Client, error) {
	opts := Options{
		Address:   cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    NewZapLogger(logger),
	}
	return New(ctx, opts)
}
