// cache demonstrates NewCache with a prefix and GetOrLoad (get-or-set pattern).
// Requires Redis running; set REDIS_ADDR (and optionally REDIS_PASSWORD, REDIS_DB) to override defaults.
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/peoplesuite/platform-infra-go/pkg/redis"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	db := 0
	if s := os.Getenv("REDIS_DB"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			db = n
		}
	}

	client, err := redis.New(redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
	if err != nil {
		log.Fatalf("Redis client failed: %v (set REDIS_ADDR if needed)", err)
	}
	defer func() { _ = client.Close() }()

	// Namespaced cache so keys are prefixed (e.g. "config:key").
	cache := redis.NewCache(client, "example:cache:")
	ctx := context.Background()
	key := "mykey"
	type Config struct {
		Value string `json:"value"`
	}

	var dest Config
	err = cache.GetOrLoad(ctx, key, &dest, 5*time.Minute, func() (any, error) {
		// Simulate loading from DB or API on cache miss.
		return Config{Value: "loaded"}, nil
	})
	if err != nil {
		log.Fatalf("GetOrLoad: %v", err)
	}
	log.Printf("GetOrLoad OK: %+v", dest)
}
