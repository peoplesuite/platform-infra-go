// worker demonstrates running a Temporal worker with temporal/worker: client from
// env or config, DefaultInterceptors, and a minimal workflow + activity.
// Requires a Temporal server. Run and stop with Ctrl+C.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	temporalclient "github.com/peoplesuite/platform-infra-go/pkg/temporal/client"
	temporalconfig "github.com/peoplesuite/platform-infra-go/pkg/temporal/config"
	temporalworker "github.com/peoplesuite/platform-infra-go/pkg/temporal/worker"
)

func main() {
	ctx := context.Background()
	logger := zap.NewNop()

	// Create Temporal client (env or temporal.toml)
	var c client.Client
	var err error
	if os.Getenv("TEMPORAL_CONFIG_FILE") != "" {
		profile, loadErr := temporalconfig.LoadTemporalProfile("")
		if loadErr != nil {
			log.Fatalf("Load profile: %v", loadErr)
		}
		c, err = temporalclient.NewFromProfile(ctx, profile)
	} else {
		addr := os.Getenv("TEMPORAL_ADDRESS")
		ns := os.Getenv("TEMPORAL_NAMESPACE")
		if addr == "" {
			addr = "localhost:7233"
		}
		if ns == "" {
			ns = "default"
		}
		c, err = temporalclient.NewFromConfig(ctx, temporalclient.ConnectionConfig{HostPort: addr, Namespace: ns}, logger)
	}
	if err != nil {
		log.Fatalf("Temporal client: %v", err)
	}
	defer c.Close()

	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "example-queue"
	}

	// Build worker with default interceptors
	opts := temporalworker.Options{
		TaskQueue:    taskQueue,
		Interceptors: temporalworker.DefaultInterceptors(),
		Logger:       logger,
	}
	b := temporalworker.New(c, opts)
	b.RegisterWorkflow(ExampleWorkflow, "ExampleWorkflow")
	b.RegisterActivity(ExampleActivity)

	log.Printf("starting worker on task queue %s (Ctrl+C to stop)", taskQueue)
	if err := b.Run(ctx); err != nil {
		log.Fatalf("worker run: %v", err)
	}
}

func ExampleWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var result string
	err := workflow.ExecuteActivity(ctx, ExampleActivity, name).Get(ctx, &result)
	return result, err
}

func ExampleActivity(ctx context.Context, name string) (string, error) {
	info := activity.GetInfo(ctx)
	log.Printf("activity run for workflow %s", info.WorkflowExecution.ID)
	return "Hello, " + name, nil
}
