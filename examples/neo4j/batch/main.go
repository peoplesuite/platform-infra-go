// batch demonstrates NewBatchWriter and Write with UNWIND $batch.
// Requires Neo4j running; set NEO4J_URI, NEO4J_PASSWORD (and optionally NEO4J_DATABASE) to override defaults.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/peoplesuite/platform-infra-go/pkg/neo4j"
	"go.uber.org/zap"
)

func main() {
	uri := os.Getenv("NEO4J_URI")
	if uri == "" {
		uri = "bolt://localhost:7687"
	}
	password := os.Getenv("NEO4J_PASSWORD")
	if password == "" {
		log.Fatal("NEO4J_PASSWORD is required")
	}

	ctx := context.Background()
	logger := zap.NewNop()
	cfg := neo4j.Config{
		URI:         uri,
		Username:    getEnv("NEO4J_USERNAME", "neo4j"),
		Password:    password,
		Database:    getEnv("NEO4J_DATABASE", "neo4j"),
		ConnAcquire: 30 * time.Second,
		MaxConnLife: time.Hour,
	}

	driver, err := neo4j.NewDriver(ctx, cfg, logger)
	if err != nil {
		log.Fatalf("Neo4j not available: %v", err)
	}
	defer func() { _ = driver.Close(ctx) }()

	writer := neo4j.NewBatchWriter(driver, cfg, logger).WithBatchSize(10)

	// Example: create nodes in batch; query must use UNWIND $batch AS row.
	cypher := `
		UNWIND $batch AS row
		CREATE (n:ExampleBatch {id: row.id, name: row.name})
	`
	items := []map[string]any{
		{"id": 1, "name": "a"},
		{"id": 2, "name": "b"},
		{"id": 3, "name": "c"},
	}
	if err := writer.Write(ctx, cypher, items); err != nil {
		log.Fatalf("BatchWriter Write: %v", err)
	}
	log.Printf("batch write OK (%d items)", len(items))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
