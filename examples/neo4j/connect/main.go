// connect demonstrates neo4j.NewDriver and Config; caller must close the driver.
// Requires Neo4j running; set NEO4J_URI, NEO4J_USERNAME, NEO4J_PASSWORD (and optionally NEO4J_DATABASE) to override defaults.
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
		log.Fatal("NEO4J_PASSWORD is required (set it to your Neo4j password)")
	}

	ctx := context.Background()
	logger := zap.NewNop()
	cfg := neo4j.Config{
		URI:         uri,
		Username:    getEnv("NEO4J_USERNAME", "neo4j"),
		Password:    password,
		Database:    getEnv("NEO4J_DATABASE", "neo4j"),
		MaxConnPool: 50,
		ConnAcquire: 30 * time.Second,
		MaxConnLife: time.Hour,
		LogLevel:    "warn",
	}

	driver, err := neo4j.NewDriver(ctx, cfg, logger)
	if err != nil {
		log.Fatalf("Neo4j not available: %v (set NEO4J_URI, NEO4J_PASSWORD if needed)", err)
	}
	defer func() { _ = driver.Close(ctx) }()

	log.Println("connected to Neo4j")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
