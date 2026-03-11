// Package redis provides a Redis client with JSON cache helpers, key builders, and distributed locks.
// It wraps github.com/redis/go-redis/v9 and uses pkg-local Error for operation context.
//
// Use Client for low-level JSON get/set/delete and GetOrSetJSON. For per-service or namespaced
// caching (e.g. "org:", "config:"), use NewCache(client, prefix) and the Cache type so all keys
// share a prefix and avoid collision across services.
package redis

import (
	"crypto/tls"

	goredis "github.com/redis/go-redis/v9"
)

// Client is a Redis client with JSON helpers and optional TLS.
type Client struct {
	rdb *goredis.Client
}

// New creates a Redis client from Options. Caller must call Close when done.
func New(opts Options) (*Client, error) {
	rdbOpts := &goredis.Options{
		Addr:     opts.Addr,
		Username: opts.Username,
		Password: opts.Password,
		DB:       opts.DB,
	}

	if opts.TLS {
		rdbOpts.TLSConfig = &tls.Config{
			InsecureSkipVerify: !opts.VerifyTLS, //nolint:gosec
		}
	}

	rdb := goredis.NewClient(rdbOpts)

	// Skip ping during initialization for graceful degradation
	// If Redis is unavailable, the client will be created but operations will fail
	// Cache operations should treat Redis errors as cache misses
	// This allows the application to start even if Redis is temporarily unavailable

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}
