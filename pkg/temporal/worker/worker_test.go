package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// newBuilderWithLazyClient creates a Builder using a LazyClient (SDK accepts only Dial/LazyClient-created clients).
func newBuilderWithLazyClient(t *testing.T, opts Options) *Builder {
	t.Helper()
	c, err := sdkclient.NewLazyClient(sdkclient.Options{
		HostPort:  "localhost:7233",
		Namespace: "default",
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	return New(c, opts)
}

func TestNew_DefaultConcurrency(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue: "queue",
		// MaxConcurrentActivities and MaxConcurrentWorkflows are 0
	})

	require.NotNil(t, b)
	b.Stop()
}

func TestNew_ExplicitConcurrency(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:               "queue",
		MaxConcurrentActivities: 15,
		MaxConcurrentWorkflows:  5,
	})

	require.NotNil(t, b)
	b.Stop()
}

func TestSetupDeployment_EmptyDeploymentName(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:      "queue",
		DeploymentName: "",
		BuildID:        "build-1",
	})

	err := b.SetupDeployment(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deployment name and build ID must be set")
}

func TestSetupDeployment_EmptyBuildID(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:      "queue",
		DeploymentName: "dep-1",
		BuildID:        "",
	})

	err := b.SetupDeployment(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deployment name and build ID must be set")
}

func TestStart_AutoSetupDeploymentFails(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:           "queue",
		AutoSetupDeployment: true,
		DeploymentName:      "",
		BuildID:             "build-1",
	})

	errCh := b.Start()
	require.NotNil(t, errCh)
	err := <-errCh
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deployment setup failed")
}

func TestRun_AutoSetupDeploymentFails(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:           "queue",
		AutoSetupDeployment: true,
		DeploymentName:      "",
		BuildID:             "build-1",
	})

	err := b.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deployment setup failed")
}

func TestStop_NoPanic(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{TaskQueue: "queue"})
	assert.NotPanics(t, func() { b.Stop() })
	assert.NotPanics(t, func() { b.Stop() })
}

func TestRegisterWorkflow_NoPanic(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{TaskQueue: "queue"})
	wf := func(ctx workflow.Context) error { return nil }
	assert.NotPanics(t, func() { b.RegisterWorkflow(wf, "TestWorkflow") })
	b.Stop()
}

func TestRegisterActivity_NoPanic(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{TaskQueue: "queue"})
	act := func(ctx context.Context) error { return nil }
	assert.NotPanics(t, func() { b.RegisterActivity(act) })
	b.Stop()
}

func TestRegisterActivityWithName_NoPanic(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{TaskQueue: "queue"})
	act := func(ctx context.Context) error { return nil }
	assert.NotPanics(t, func() { b.RegisterActivityWithName(act, "TestActivity") })
	b.Stop()
}

func TestNew_BuildIDOnly_NoDeploymentName(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:      "queue",
		BuildID:        "v1",
		DeploymentName: "", // No deployment name -> DeploymentOptions not set
	})
	require.NotNil(t, b)
	b.Stop()
}

func TestNew_BuildIDAndDeploymentName(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:      "queue",
		BuildID:        "v1",
		DeploymentName: "dep-1",
	})
	require.NotNil(t, b)
	b.Stop()
}

func TestNew_WithInterceptors(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:    "queue",
		Interceptors: DefaultInterceptors(),
	})
	require.NotNil(t, b)
	b.Stop()
}

func TestSetupDeployment_WithLogger_LogsSetup(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:      "queue",
		DeploymentName: "dep-1",
		BuildID:        "build-1",
		Logger:         logger,
	})
	err := b.SetupDeployment(context.Background())
	// Will fail because no Temporal server, but we hit the Logger.Info branch
	require.Error(t, err)
	logs := observed.All()
	require.GreaterOrEqual(t, len(logs), 1)
	assert.Equal(t, "setting up deployment for task queue", logs[0].Message)
}

func TestStart_SuccessPath_NoAutoSetup(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{
		TaskQueue:           "queue",
		AutoSetupDeployment: false,
	})
	errCh := b.Start()
	require.NotNil(t, errCh)
	time.Sleep(50 * time.Millisecond) // let Run() start before Stop()
	b.Stop()
	<-errCh // Run() returns when stopped; may be nil or connection error if no server
}

func TestStop_AfterStart(t *testing.T) {
	b := newBuilderWithLazyClient(t, Options{TaskQueue: "queue"})
	errCh := b.Start()
	time.Sleep(50 * time.Millisecond)
	b.Stop()
	<-errCh
}
