package temporal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/peoplesuite/platform-infra-go/pkg/temporal/client"
)

const temporalGRPCPort = "7233/tcp"

// runTemporalContainer starts a Temporal dev server container.
// Skips the test if testcontainers cannot run (e.g. no Docker).
func runTemporalContainer(t *testing.T, ctx context.Context) testcontainers.Container {
	t.Helper()
	var ctr testcontainers.Container
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("testcontainers not available (Docker required): %v", r)
		}
	}()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "temporalio/temporal:latest",
			ExposedPorts: []string{temporalGRPCPort},
			Cmd:          []string{"server", "start-dev", "--ip", "0.0.0.0"},
			WaitingFor: wait.ForListeningPort(nat.Port(temporalGRPCPort)).
				WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	}
	var err error
	ctr, err = testcontainers.GenericContainer(ctx, req)
	if err != nil {
		t.Skipf("testcontainers not available: %v", err)
	}
	return ctr
}

func temporalEndpoint(t *testing.T, ctx context.Context, ctr testcontainers.Container) string {
	t.Helper()
	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, nat.Port(temporalGRPCPort))
	require.NoError(t, err)
	return fmt.Sprintf("%s:%s", host, port.Port())
}

func TestClient_Integration_ConnectAndHealthCheck(t *testing.T) {
	ctx := context.Background()
	ctr := runTemporalContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	endpoint := temporalEndpoint(t, ctx, ctr)
	opts := client.Options{
		Address:   endpoint,
		Namespace: "default",
		Logger:    client.NewZapLogger(zap.NewNop()),
	}

	c, err := client.New(ctx, opts)
	require.NoError(t, err)
	defer c.Close()

	healthResp, err := c.CheckHealth(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, healthResp)
}

func TestClient_Integration_NewFromConfig(t *testing.T) {
	ctx := context.Background()
	ctr := runTemporalContainer(t, ctx)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			t.Logf("terminate container: %v", err)
		}
	}()

	endpoint := temporalEndpoint(t, ctx, ctr)
	cfg := client.ConnectionConfig{
		HostPort:  endpoint,
		Namespace: "default",
	}
	logger := zap.NewNop()

	c, err := client.NewFromConfig(ctx, cfg, logger)
	require.NoError(t, err)
	defer c.Close()

	_, err = c.CheckHealth(ctx, nil)
	require.NoError(t, err)
}
