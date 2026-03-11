package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// GetJSON gets key and unmarshals into target. Returns (false, nil) when key is missing.
func (c *Client) GetJSON(
	ctx context.Context,
	key string,
	target any,
) (bool, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, &Error{Op: "Get", Key: key, Err: err}
	}

	if err := json.Unmarshal([]byte(val), target); err != nil {
		return false, fmt.Errorf("json unmarshal: %w", err)
	}

	return true, nil
}

// SetJSON marshals value to JSON and sets key with ttl.
func (c *Client) SetJSON(
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}

	if err := c.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return &Error{Op: "Set", Key: key, Err: err}
	}

	return nil
}

// Expire sets the TTL for an existing key.
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		return &Error{Op: "Expire", Key: key, Err: err}
	}
	return nil
}

// Delete removes key.
func (c *Client) Delete(ctx context.Context, key string) error {
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return &Error{Op: "Del", Key: key, Err: err}
	}
	return nil
}

// SetBytes stores raw bytes at key with ttl. Use for large payloads that are already serialized.
func (c *Client) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return &Error{Op: "Set", Key: key, Err: err}
	}
	return nil
}

// GetBytes returns raw bytes at key. Returns (nil, nil) if key does not exist.
func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, &Error{Op: "Get", Key: key, Err: err}
	}
	return val, nil
}

// DeleteByPattern deletes all keys matching pattern (e.g. "org:*") using SCAN.
func (c *Client) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return &Error{Op: "DelByPattern", Key: iter.Val(), Err: err}
		}
	}
	return iter.Err()
}

// GetOrSetJSON attempts to get a value from cache. If not found, calls fn() to fetch the value,
// stores it in cache with the given TTL, and returns it.
func (c *Client) GetOrSetJSON(
	ctx context.Context,
	key string,
	ttl time.Duration,
	fn func() (any, error),
	target any,
) error {
	// Try to get from cache first
	found, err := c.GetJSON(ctx, key, target)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// Cache miss - call the function to fetch the value
	value, err := fn()
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	// Store in cache
	if err := c.SetJSON(ctx, key, value, ttl); err != nil {
		return err
	}

	// Unmarshal the fetched value into target
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	return nil
}
