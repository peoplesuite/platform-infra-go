package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// Options configures the Temporal worker (task queue, concurrency, versioning, interceptors).
type Options struct {
	TaskQueue string

	MaxConcurrentActivities int
	MaxConcurrentWorkflows  int

	BuildID        string
	DeploymentName string
	Interceptors   []interceptor.WorkerInterceptor

	// Logger is optional; when set, deployment setup logs via zap instead of log.Printf.
	Logger *zap.Logger

	// AutoSetupDeployment enables automatic deployment setup on worker start
	// When true, the worker will automatically call SetCurrentVersion for its task queue
	// Default: false (to maintain backward compatibility)
	AutoSetupDeployment bool
}

// Builder builds and runs a Temporal worker; create with New, then RegisterWorkflow/RegisterActivity and Run or Start.
type Builder struct {
	client              client.Client
	opts                Options
	w                   worker.Worker
	deploymentSetupDone bool
}

// New creates a worker builder for the given Temporal client and options.
func New(c client.Client, opts Options) *Builder {
	if opts.MaxConcurrentActivities == 0 {
		opts.MaxConcurrentActivities = 20
	}
	if opts.MaxConcurrentWorkflows == 0 {
		opts.MaxConcurrentWorkflows = 10
	}

	workerOpts := worker.Options{
		MaxConcurrentActivityExecutionSize:     opts.MaxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: opts.MaxConcurrentWorkflows,
	}

	// Configure Worker Versioning using the new DeploymentOptions API
	if opts.BuildID != "" {
		// Use BuildID as a fallback for backward compatibility
		workerOpts.BuildID = opts.BuildID //nolint:staticcheck // SA1019: deprecated, prefer DeploymentOptions.Version when set

		// Enable Worker Versioning if deployment name is provided
		if opts.DeploymentName != "" {
			workerOpts.DeploymentOptions = worker.DeploymentOptions{
				UseVersioning: true,
				Version: worker.WorkerDeploymentVersion{
					DeploymentName: opts.DeploymentName,
					BuildID:        opts.BuildID,
				},
				// Set default versioning behavior for all workflows
				// AutoUpgrade: workflows will use the latest deployment version (better for schedules)
				DefaultVersioningBehavior: workflow.VersioningBehaviorAutoUpgrade,
			}
		}
	}

	if len(opts.Interceptors) > 0 {
		workerOpts.Interceptors = opts.Interceptors
	}

	w := worker.New(c, opts.TaskQueue, workerOpts)

	return &Builder{
		client: c,
		opts:   opts,
		w:      w,
	}
}

// RegisterWorkflow registers a workflow function with the given name.
func (b *Builder) RegisterWorkflow(wf any, name string) {
	b.w.RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{
		Name: name,
	})
}

// RegisterActivity registers an activity.
func (b *Builder) RegisterActivity(a any) {
	b.w.RegisterActivity(a)
}

// RegisterActivityWithName registers an activity with a custom name
func (b *Builder) RegisterActivityWithName(a any, name string) {
	b.w.RegisterActivityWithOptions(a, activity.RegisterOptions{
		Name: name,
	})
}

// SetupDeployment configures the deployment routing for this worker's task queue.
// This should be called before starting the worker.
// It's automatically called by Start() if AutoSetupDeployment is enabled.
func (b *Builder) SetupDeployment(ctx context.Context) error {
	if b.deploymentSetupDone {
		return nil // Already set up
	}

	if b.opts.DeploymentName == "" || b.opts.BuildID == "" {
		return fmt.Errorf("deployment name and build ID must be set to use deployment setup")
	}

	if b.opts.Logger != nil {
		b.opts.Logger.Info("setting up deployment for task queue",
			zap.String("task_queue", b.opts.TaskQueue),
			zap.String("deployment", b.opts.DeploymentName),
			zap.String("build_id", b.opts.BuildID),
		)
	} else {
		log.Printf("Setting up deployment for task queue '%s'...\n", b.opts.TaskQueue)
		log.Printf("  Deployment: %s\n", b.opts.DeploymentName)
		log.Printf("  Build ID:   %s\n", b.opts.BuildID)
	}

	setupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Get deployment client
	deploymentClient := b.client.WorkerDeploymentClient()

	// Get deployment handle
	handle := deploymentClient.GetHandle(b.opts.DeploymentName)

	// Set this version as the current version for this task queue
	_, err := handle.SetCurrentVersion(setupCtx, client.WorkerDeploymentSetCurrentVersionOptions{
		BuildID:        b.opts.BuildID,
		AllowNoPollers: true, // Allow setting even if no pollers yet (we're about to start them)
	})
	if err != nil {
		return fmt.Errorf("failed to set current version for task queue '%s': %w", b.opts.TaskQueue, err)
	}

	if b.opts.Logger != nil {
		b.opts.Logger.Info("deployment setup complete for task queue", zap.String("task_queue", b.opts.TaskQueue))
	} else {
		log.Printf("✅ Deployment setup complete for task queue '%s'\n", b.opts.TaskQueue)
	}
	b.deploymentSetupDone = true
	return nil
}

// Start starts the worker in a goroutine and returns immediately.
// Use this for multiple workers where you want to handle shutdown manually.
// The worker will run until Stop() is called or the process is interrupted.
// If AutoSetupDeployment is enabled, it will automatically configure deployments before starting.
func (b *Builder) Start() <-chan error {
	errCh := make(chan error, 1)

	// Auto-setup deployment if enabled
	if b.opts.AutoSetupDeployment && !b.deploymentSetupDone {
		if err := b.SetupDeployment(context.Background()); err != nil {
			go func() {
				errCh <- fmt.Errorf("deployment setup failed: %w", err)
			}()
			return errCh
		}
	}

	go func() {
		errCh <- b.w.Run(worker.InterruptCh())
	}()
	return errCh
}

// Run starts the worker and waits for a shutdown signal.
// Use this for single-worker scenarios.
// If AutoSetupDeployment is enabled, it will automatically configure deployments before starting.
func (b *Builder) Run(ctx context.Context) error {
	// Auto-setup deployment if enabled
	if b.opts.AutoSetupDeployment && !b.deploymentSetupDone {
		if err := b.SetupDeployment(ctx); err != nil {
			return fmt.Errorf("deployment setup failed: %w", err)
		}
	}

	// Start worker in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- b.w.Run(worker.InterruptCh())
	}()

	// Wait for shutdown signal
	sigCh := waitForShutdown()
	select {
	case sig := <-sigCh:
		_ = sig // Signal received, stop worker
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	b.w.Stop()
	return nil
}

// Stop stops the worker.
func (b *Builder) Stop() {
	b.w.Stop()
}

func waitForShutdown() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}
