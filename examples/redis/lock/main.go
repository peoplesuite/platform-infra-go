// lock demonstrates AcquireLock and ReleaseLock for distributed locking.
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

	ctx := context.Background()
	lockKey := "example:lock:resource"
	ttl := 10 * time.Second

	ok, err := client.AcquireLock(ctx, lockKey, ttl)
	if err != nil {
		log.Fatalf("AcquireLock: %v", err)
	}
	if !ok {
		log.Fatal("AcquireLock: lock already held by another client")
	}
	log.Println("lock acquired")

	// Simulate work, then release.
	time.Sleep(100 * time.Millisecond)
	if err := client.ReleaseLock(ctx, lockKey); err != nil {
		log.Fatalf("ReleaseLock: %v", err)
	}
	log.Println("lock released")
}
