// connect demonstrates creating a Temporal client with temporal/client.NewFromConfig
// or from a profile (temporal.toml). Requires a Temporal server; set TEMPORAL_ADDRESS
// and TEMPORAL_NAMESPACE to override defaults.
package main

import (
	"context"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	temporalclient "github.com/peoplesuite/platform-infra-go/pkg/temporal/client"
	temporalconfig "github.com/peoplesuite/platform-infra-go/pkg/temporal/config"
)

func main() {
	ctx := context.Background()
	logger := zap.NewNop()

	// Option 1: From env (no config file)
	addr := os.Getenv("TEMPORAL_ADDRESS")
	ns := os.Getenv("TEMPORAL_NAMESPACE")
	if addr == "" {
		addr = "localhost:7233"
	}
	if ns == "" {
		ns = "default"
	}

	var c client.Client
	var err error

	if os.Getenv("TEMPORAL_CONFIG_FILE") != "" {
		// Option 2: From temporal.toml profile
		profile, loadErr := temporalconfig.LoadTemporalProfile("")
		if loadErr != nil {
			log.Fatalf("Load profile: %v (set TEMPORAL_CONFIG_FILE or use TEMPORAL_ADDRESS/TEMPORAL_NAMESPACE)", loadErr)
		}
		c, err = temporalclient.NewFromProfile(ctx, profile)
	} else {
		cfg := temporalclient.ConnectionConfig{HostPort: addr, Namespace: ns}
		c, err = temporalclient.NewFromConfig(ctx, cfg, logger)
	}

	if err != nil {
		log.Fatalf("Temporal client failed: %v (ensure server is running; set TEMPORAL_ADDRESS/TEMPORAL_NAMESPACE if needed)", err)
	}
	defer c.Close()

	log.Printf("connected to Temporal at %s namespace %s", addr, ns)
}
