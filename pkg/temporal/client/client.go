package client

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
)

// DefaultMaxPayloadSize is the gRPC max receive/send message size (16 MB).
// FetchAssociates can return ~7 MB when Redis is down (full associates in result); default 4 MB would cause ResourceExhausted.
const DefaultMaxPayloadSize = 16 * 1024 * 1024

// New creates a Temporal client from Options. Caller must close it when done.
func New(ctx context.Context, opts Options) (client.Client, error) {
	if opts.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if opts.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	clientOpts := client.Options{
		HostPort:  opts.Address,
		Namespace: opts.Namespace,
		ConnectionOptions: client.ConnectionOptions{
			MaxPayloadSize: DefaultMaxPayloadSize,
		},
	}

	if opts.Identity != "" {
		clientOpts.Identity = opts.Identity
	}

	if opts.Logger != nil {
		clientOpts.Logger = opts.Logger
	}

	// ConnectionTimeout is handled by the SDK internally
	// No need to set it explicitly in client.Options

	// MetadataHeaders and RPCTimeout are stored in opts for future use
	// Metadata headers may be implemented via gRPC interceptors if needed
	// RPCTimeout is stored but Temporal SDK handles timeouts internally

	c, err := client.Dial(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to dial temporal client: %w", err)
	}

	return c, nil
}
