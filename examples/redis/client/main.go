// client demonstrates redis.New and GetJSON/SetJSON/Delete.
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
	key := "example:client:item"
	type Item struct {
		Name string `json:"name"`
	}

	// Set then get
	val := Item{Name: "hello"}
	if err := client.SetJSON(ctx, key, val, 5*time.Minute); err != nil {
		log.Fatalf("SetJSON: %v", err)
	}
	log.Println("SetJSON OK")

	var out Item
	found, err := client.GetJSON(ctx, key, &out)
	if err != nil {
		log.Fatalf("GetJSON: %v", err)
	}
	if !found {
		log.Fatal("GetJSON: key not found")
	}
	log.Printf("GetJSON OK: %+v", out)

	if err := client.Delete(ctx, key); err != nil {
		log.Fatalf("Delete: %v", err)
	}
	log.Println("Delete OK")
}
