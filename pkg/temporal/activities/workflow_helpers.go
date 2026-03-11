package activities

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"

	temporalclient "github.com/peoplesuite/platform-infra-go/pkg/temporal/client"
	temporalconfig "github.com/peoplesuite/platform-infra-go/pkg/temporal/config"
	"github.com/peoplesuite/platform-infra-go/pkg/temporal/policies"
	"github.com/peoplesuite/platform-infra-go/pkg/temporal/start"
)

// ClientProvider supplies Temporal clients for activities; call CloseClient when done.
type ClientProvider interface {
	GetTemporalClient(ctx context.Context) (client.Client, error)
	CloseClient(client.Client)
}

// FireAndForgetClientProvider creates clients on-demand via Loader; CloseClient closes them.
type FireAndForgetClientProvider struct {
	Loader func() (temporalconfig.TemporalProfile, error)
}

// GetTemporalClient returns a new Temporal client from the Loader.
func (p *FireAndForgetClientProvider) GetTemporalClient(ctx context.Context) (client.Client, error) {
	return temporalclient.NewFromProfileLoader(ctx, p.Loader)
}

// CloseClient closes the Temporal client if non-nil.
func (p *FireAndForgetClientProvider) CloseClient(c client.Client) {
	if c != nil {
		c.Close()
	}
}

// InjectedClientProvider returns a pre-injected client; CloseClient is a no-op.
type InjectedClientProvider struct {
	Client client.Client
}

// GetTemporalClient returns the injected client or an error if nil.
func (p *InjectedClientProvider) GetTemporalClient(ctx context.Context) (client.Client, error) {
	if p.Client == nil {
		return nil, fmt.Errorf("temporal client not injected")
	}
	return p.Client, nil
}

// CloseClient is a no-op; injected clients are managed externally.
func (p *InjectedClientProvider) CloseClient(c client.Client) {
	// Don't close injected clients - they're managed externally
}

// CustomIdentityClientProvider creates clients with custom Identity; optionally closes after use.
type CustomIdentityClientProvider struct {
	Loader     func() (temporalconfig.TemporalProfile, error)
	Identity   string
	CloseAfter bool // Whether to close client after use
}

// GetTemporalClient returns a Temporal client with overridden Identity.
func (p *CustomIdentityClientProvider) GetTemporalClient(ctx context.Context) (client.Client, error) {
	profile, err := p.Loader()
	if err != nil {
		return nil, nil // Fire-and-forget: return nil instead of error
	}
	opts := temporalclient.OptionsFromProfile(profile)
	opts.Identity = p.Identity // Override identity
	return temporalclient.New(ctx, opts)
}

// CloseClient closes the client if CloseAfter is true.
func (p *CustomIdentityClientProvider) CloseClient(c client.Client) {
	if p.CloseAfter && c != nil {
		c.Close()
	}
}

// ExecuteWorkflowOptions configures workflow execution
type ExecuteWorkflowOptions struct {
	WorkflowID       string
	TaskQueue        string
	Policy           policies.WorkflowPolicy
	WorkflowType     string
	WorkflowInput    interface{}
	Memo             map[string]interface{}
	SearchAttributes map[string]interface{}
}

// SignalWithStartWorkflowOptions configures signal-with-start operation
type SignalWithStartWorkflowOptions struct {
	WorkflowID       string
	TaskQueue        string
	Policy           policies.WorkflowPolicy
	WorkflowType     string
	SignalName       string
	SignalPayload    interface{}
	WorkflowInput    interface{}
	Memo             map[string]interface{}
	SearchAttributes map[string]interface{}
}

// SignalWorkflowOptions configures workflow signaling
type SignalWorkflowOptions struct {
	WorkflowID    string
	RunID         string
	SignalName    string
	SignalPayload interface{}
}

// Logger interface for fire-and-forget logging
type Logger interface {
	Warn(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
}

// ExecuteWorkflow starts a new workflow using the provided client provider
func ExecuteWorkflow(ctx context.Context, provider ClientProvider, opts ExecuteWorkflowOptions) error {
	temporalClient, err := provider.GetTemporalClient(ctx)
	if err != nil {
		return err
	}
	if temporalClient == nil {
		return fmt.Errorf("temporal client not available")
	}
	defer provider.CloseClient(temporalClient)

	startOpts := start.NewOptions(opts.WorkflowID, opts.TaskQueue, opts.Policy)
	if opts.Memo != nil {
		startOpts = start.WithMemo(startOpts, opts.Memo)
	}
	if opts.SearchAttributes != nil {
		startOpts = start.WithSearchAttributes(startOpts, opts.SearchAttributes)
	}

	_, err = temporalClient.ExecuteWorkflow(ctx, startOpts, opts.WorkflowType, opts.WorkflowInput)
	return err
}

// SignalWithStartWorkflow signals an existing workflow or starts a new one
func SignalWithStartWorkflow(ctx context.Context, provider ClientProvider, opts SignalWithStartWorkflowOptions) error {
	temporalClient, err := provider.GetTemporalClient(ctx)
	if err != nil {
		return err
	}
	if temporalClient == nil {
		return fmt.Errorf("temporal client not available")
	}
	defer provider.CloseClient(temporalClient)

	startOpts := start.NewOptions(opts.WorkflowID, opts.TaskQueue, opts.Policy)
	if opts.Memo != nil {
		startOpts = start.WithMemo(startOpts, opts.Memo)
	}
	if opts.SearchAttributes != nil {
		startOpts = start.WithSearchAttributes(startOpts, opts.SearchAttributes)
	}

	_, err = temporalClient.SignalWithStartWorkflow(
		ctx,
		opts.WorkflowID,
		opts.SignalName,
		opts.SignalPayload,
		startOpts,
		opts.WorkflowType,
		opts.WorkflowInput,
	)
	return err
}

// SignalWorkflow signals an existing workflow
func SignalWorkflow(ctx context.Context, provider ClientProvider, opts SignalWorkflowOptions) error {
	temporalClient, err := provider.GetTemporalClient(ctx)
	if err != nil {
		return err
	}
	if temporalClient == nil {
		return fmt.Errorf("temporal client not available")
	}
	defer provider.CloseClient(temporalClient)

	return temporalClient.SignalWorkflow(ctx, opts.WorkflowID, opts.RunID, opts.SignalName, opts.SignalPayload)
}

// ExecuteWorkflowFireAndForget executes workflow with fire-and-forget pattern
// Logs warnings but doesn't return errors (for security_assistant/reporter pattern)
func ExecuteWorkflowFireAndForget(ctx context.Context, provider ClientProvider, opts ExecuteWorkflowOptions, logger Logger) error {
	err := ExecuteWorkflow(ctx, provider, opts)
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to execute workflow", "error", err, "workflow_id", opts.WorkflowID)
		}
		return nil
	}
	if logger != nil {
		logger.Info("Workflow executed successfully", "workflow_id", opts.WorkflowID)
	}
	return nil
}

// SignalWithStartWorkflowFireAndForget signals/starts workflow with fire-and-forget pattern
func SignalWithStartWorkflowFireAndForget(ctx context.Context, provider ClientProvider, opts SignalWithStartWorkflowOptions, logger Logger) error {
	err := SignalWithStartWorkflow(ctx, provider, opts)
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to signal/start workflow", "error", err, "workflow_id", opts.WorkflowID)
		}
		return nil
	}
	if logger != nil {
		logger.Info("Workflow signaled/started successfully", "workflow_id", opts.WorkflowID)
	}
	return nil
}

// SignalWorkflowFireAndForget signals workflow with fire-and-forget pattern
func SignalWorkflowFireAndForget(ctx context.Context, provider ClientProvider, opts SignalWorkflowOptions, logger Logger) error {
	err := SignalWorkflow(ctx, provider, opts)
	if err != nil {
		if logger != nil {
			logger.Warn("Failed to signal workflow", "error", err, "workflow_id", opts.WorkflowID)
		}
		return nil
	}
	if logger != nil {
		logger.Info("Workflow signaled successfully", "workflow_id", opts.WorkflowID)
	}
	return nil
}
