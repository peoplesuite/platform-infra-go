// query demonstrates Query, QuerySingle, and record helpers (Str, Int).
// Requires Neo4j running; set NEO4J_URI, NEO4J_PASSWORD (and optionally NEO4J_DATABASE) to override defaults.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	n4j "github.com/peoplesuite/platform-infra-go/pkg/neo4j"
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
	cfg := n4j.Config{
		URI:         uri,
		Username:    getEnv("NEO4J_USERNAME", "neo4j"),
		Password:    password,
		Database:    getEnv("NEO4J_DATABASE", "neo4j"),
		ConnAcquire: 30 * time.Second,
		MaxConnLife: time.Hour,
	}

	driver, err := n4j.NewDriver(ctx, cfg, logger)
	if err != nil {
		log.Fatalf("Neo4j not available: %v", err)
	}
	defer func() { _ = driver.Close(ctx) }()

	// QuerySingle: expect zero or one row
	node, err := n4j.QuerySingle(ctx, driver, cfg,
		"RETURN 1 AS id, 'hello' AS name",
		nil,
		func(rec *neo4j.Record) (string, error) {
			return n4j.Str(rec, "name"), nil
		},
	)
	if err != nil {
		log.Fatalf("QuerySingle: %v", err)
	}
	if node != nil {
		log.Printf("QuerySingle result: %q", *node)
	} else {
		log.Println("QuerySingle: no row")
	}

	// Query: multiple rows
	ids, err := n4j.Query(ctx, driver, cfg,
		"UNWIND [1, 2, 3] AS x RETURN x",
		nil,
		func(rec *neo4j.Record) (int, error) {
			return n4j.Int(rec, "x"), nil
		},
	)
	if err != nil {
		log.Fatalf("Query: %v", err)
	}
	log.Printf("Query result: %v", ids)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
