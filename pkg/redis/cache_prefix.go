package redis

import (
	"context"
	"time"
)

// Cache wraps Client with a key prefix for namespaced caching (e.g. "org:", "config:").
// All keys are prefixed so multiple caches can share the same Client without collision.
// Use NewCache(client, prefix), then Get/Set/Delete/DeletePattern/GetOrLoad with logical keys.
type Cache struct {
	client *Client
	prefix string
}

// NewCache returns a Cache that uses the given prefix for every key.
// Prefix is used as-is; add a separator in prefix if desired (e.g. "org:").
func NewCache(client *Client, prefix string) *Cache {
	return &Cache{client: client, prefix: prefix}
}

func (c *Cache) key(k string) string {
	return c.prefix + k
}

// Get reads a value into dest. It returns (true, nil) if found, (false, nil) if missing, or (false, err) on error.
func (c *Cache) Get(ctx context.Context, key string, dest any) (bool, error) {
	return c.client.GetJSON(ctx, c.key(key), dest)
}

// Set writes value with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.client.SetJSON(ctx, c.key(key), value, ttl)
}

// Delete removes a single key.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Delete(ctx, c.key(key))
}

// DeletePattern removes all keys matching the logical pattern (e.g. "user:*").
// The prefix is applied, so the actual Redis pattern is prefix+pattern.
func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	return c.client.DeleteByPattern(ctx, c.key(pattern))
}

// GetOrLoad returns the cached value if present; otherwise calls loadFn, caches the result with ttl, and returns it.
// dest is populated in both cases when no error is returned.
func (c *Cache) GetOrLoad(
	ctx context.Context,
	key string,
	dest any,
	ttl time.Duration,
	loadFn func() (any, error),
) error {
	return c.client.GetOrSetJSON(ctx, c.key(key), ttl, loadFn, dest)
}
