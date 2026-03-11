package redis

import (
	"context"
	"time"
)

// AcquireLock attempts to set key with ttl using SetNX; returns true if acquired.
func (c *Client) AcquireLock(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (bool, error) {
	return c.rdb.SetNX(ctx, key, "1", ttl).Result()
}

// ReleaseLock deletes the lock key.
func (c *Client) ReleaseLock(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// Expire sets the TTL on an existing key.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Expire(ctx, key, ttl).Err()
}
