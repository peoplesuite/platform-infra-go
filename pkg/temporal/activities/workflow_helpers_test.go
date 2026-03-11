package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"

	temporalconfig "github.com/peoplesuite/platform-infra-go/pkg/temporal/config"
	"github.com/peoplesuite/platform-infra-go/pkg/temporal/policies"
)

func TestInjectedClientProvider_GetTemporalClient_NilClient(t *testing.T) {
	p := &InjectedClientProvider{Client: nil}
	c, err := p.GetTemporalClient(context.Background())
	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "temporal client not injected")
}

func TestInjectedClientProvider_GetTemporalClient_Success(t *testing.T) {
	mockClient := mocks.NewClient(t)
	p := &InjectedClientProvider{Client: mockClient}
	c, err := p.GetTemporalClient(context.Background())
	require.NoError(t, err)
	assert.Same(t, mockClient, c)
}

func TestInjectedClientProvider_CloseClient_NoOp(t *testing.T) {
	p := &InjectedClientProvider{Client: nil}
	// CloseClient should not panic when given any client (including nil)
	assert.NotPanics(t, func() { p.CloseClient(nil) })
}

func TestFireAndForgetClientProvider_GetTemporalClient_LoaderError(t *testing.T) {
	loaderErr := errors.New("load failed")
	p := &FireAndForgetClientProvider{
		Loader: func() (temporalconfig.TemporalProfile, error) {
			return temporalconfig.TemporalProfile{}, loaderErr
		},
	}
	c, err := p.GetTemporalClient(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, c)
}

func TestFireAndForgetClientProvider_CloseClient(t *testing.T) {
	closed := false
	mockClient := mocks.NewClient(t)
	mockClient.On("Close").Run(func(args mock.Arguments) { closed = true }).Return()
	p := &FireAndForgetClientProvider{Loader: nil}
	p.CloseClient(mockClient)
	assert.True(t, closed)
}

func TestFireAndForgetClientProvider_CloseClient_Nil(t *testing.T) {
	p := &FireAndForgetClientProvider{}
	assert.NotPanics(t, func() { p.CloseClient(nil) })
}

func TestCustomIdentityClientProvider_GetTemporalClient_LoaderError(t *testing.T) {
	p := &CustomIdentityClientProvider{
		Loader: func() (temporalconfig.TemporalProfile, error) {
			return temporalconfig.TemporalProfile{}, errors.New("err")
		},
		Identity: "custom",
	}
	c, err := p.GetTemporalClient(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, c)
}

func TestCustomIdentityClientProvider_CloseClient_CloseAfterTrue(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockClient.On("Close").Return()
	p := &CustomIdentityClientProvider{CloseAfter: true}
	p.CloseClient(mockClient)
	mockClient.AssertExpectations(t)
}

func TestCustomIdentityClientProvider_CloseClient_CloseAfterFalse(t *testing.T) {
	mockClient := mocks.NewClient(t)
	p := &CustomIdentityClientProvider{CloseAfter: false}
	p.CloseClient(mockClient)
	// Close should not be called
	mockClient.AssertNotCalled(t, "Close")
}

func TestExecuteWorkflow_ProviderError(t *testing.T) {
	providerErr := errors.New("provider error")
	provider := &mockProvider{client: nil, err: providerErr}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "MyWorkflow", WorkflowInput: nil,
	}
	err := ExecuteWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Equal(t, providerErr, err)
}

func TestExecuteWorkflow_NilClient(t *testing.T) {
	provider := &mockProvider{client: nil, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "MyWorkflow", WorkflowInput: nil,
	}
	err := ExecuteWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temporal client not available")
}

func TestExecuteWorkflow_Success(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, "MyWorkflow", mock.Anything).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "MyWorkflow", WorkflowInput: "input",
	}
	err := ExecuteWorkflow(context.Background(), provider, opts)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestExecuteWorkflow_WithMemoAndSearchAttributes(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("ExecuteWorkflow", mock.Anything, mock.MatchedBy(func(opts client.StartWorkflowOptions) bool {
		return opts.Memo != nil && opts.SearchAttributes != nil //nolint:staticcheck // SA1019: SearchAttributes deprecated, use TypedSearchAttributes when migrating
	}), "WF", nil).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "id", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", WorkflowInput: nil,
		Memo:             map[string]interface{}{"k": "v"},
		SearchAttributes: map[string]interface{}{"CustomKeyword": "x"},
	}
	err := ExecuteWorkflow(context.Background(), provider, opts)
	require.NoError(t, err)
}

func TestExecuteWorkflow_ExecuteWorkflowReturnsError(t *testing.T) {
	execErr := errors.New("execute failed")
	mockClient := mocks.NewClient(t)
	mockClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, "WF", nil).Return(nil, execErr)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "id", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", WorkflowInput: nil,
	}
	err := ExecuteWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Equal(t, execErr, err)
}

func TestSignalWithStartWorkflow_ProviderError(t *testing.T) {
	providerErr := errors.New("provider error")
	provider := &mockProvider{client: nil, err: providerErr}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	err := SignalWithStartWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Equal(t, providerErr, err)
}

func TestSignalWithStartWorkflow_NilClient(t *testing.T) {
	provider := &mockProvider{client: nil, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	err := SignalWithStartWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temporal client not available")
}

func TestSignalWithStartWorkflow_Success(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("SignalWithStartWorkflow", mock.Anything, "wf-1", "sig", nil, mock.Anything, "WF", nil).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	err := SignalWithStartWorkflow(context.Background(), provider, opts)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSignalWithStartWorkflow_WithMemoAndSearchAttributes(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("SignalWithStartWorkflow", mock.Anything, "wf-1", "sig", nil, mock.MatchedBy(func(opts client.StartWorkflowOptions) bool {
		return opts.Memo != nil && opts.SearchAttributes != nil //nolint:staticcheck
	}), "WF", nil).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
		Memo:             map[string]interface{}{"k": "v"},
		SearchAttributes: map[string]interface{}{"attr": "x"},
	}
	err := SignalWithStartWorkflow(context.Background(), provider, opts)
	require.NoError(t, err)
}

func TestSignalWithStartWorkflow_ReturnsError(t *testing.T) {
	sigErr := errors.New("signal failed")
	mockClient := mocks.NewClient(t)
	mockClient.On("SignalWithStartWorkflow", mock.Anything, "wf-1", "sig", nil, mock.Anything, "WF", nil).Return(nil, sigErr)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	err := SignalWithStartWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Equal(t, sigErr, err)
}

func TestSignalWorkflow_ProviderError(t *testing.T) {
	providerErr := errors.New("provider error")
	provider := &mockProvider{client: nil, err: providerErr}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "run-1", SignalName: "sig", SignalPayload: nil}
	err := SignalWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Equal(t, providerErr, err)
}

func TestSignalWorkflow_NilClient(t *testing.T) {
	provider := &mockProvider{client: nil, err: nil}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "run-1", SignalName: "sig", SignalPayload: nil}
	err := SignalWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temporal client not available")
}

func TestSignalWorkflow_Success(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockClient.On("SignalWorkflow", mock.Anything, "wf-1", "run-1", "sig", nil).Return(nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "run-1", SignalName: "sig", SignalPayload: nil}
	err := SignalWorkflow(context.Background(), provider, opts)
	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSignalWorkflow_ReturnsError(t *testing.T) {
	sigErr := errors.New("signal failed")
	mockClient := mocks.NewClient(t)
	mockClient.On("SignalWorkflow", mock.Anything, "wf-1", "run-1", "sig", nil).Return(sigErr)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "run-1", SignalName: "sig", SignalPayload: nil}
	err := SignalWorkflow(context.Background(), provider, opts)
	require.Error(t, err)
	assert.Equal(t, sigErr, err)
}

func TestExecuteWorkflowFireAndForget_Error_LogsWarn(t *testing.T) {
	log := &recordingLogger{}
	provider := &mockProvider{client: nil, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", WorkflowInput: nil,
	}
	err := ExecuteWorkflowFireAndForget(context.Background(), provider, opts, log)
	assert.NoError(t, err)
	assert.Len(t, log.warns, 1)
	assert.Empty(t, log.infos)
	assert.Contains(t, log.warns[0].msg, "Failed to execute workflow")
}

func TestExecuteWorkflowFireAndForget_Success_LogsInfo(t *testing.T) {
	log := &recordingLogger{}
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, "WF", nil).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", WorkflowInput: nil,
	}
	err := ExecuteWorkflowFireAndForget(context.Background(), provider, opts, log)
	assert.NoError(t, err)
	assert.Len(t, log.infos, 1)
	assert.Empty(t, log.warns)
	assert.Contains(t, log.infos[0].msg, "Workflow executed successfully")
}

func TestExecuteWorkflowFireAndForget_NilLogger_NoPanic(t *testing.T) {
	provider := &mockProvider{client: nil, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", WorkflowInput: nil,
	}
	assert.NotPanics(t, func() {
		_ = ExecuteWorkflowFireAndForget(context.Background(), provider, opts, nil)
	})
}

func TestExecuteWorkflowFireAndForget_NilLogger_Success(t *testing.T) {
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, "WF", nil).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := ExecuteWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", WorkflowInput: nil,
	}
	err := ExecuteWorkflowFireAndForget(context.Background(), provider, opts, nil)
	assert.NoError(t, err)
}

func TestSignalWithStartWorkflowFireAndForget_Success_LogsInfo(t *testing.T) {
	log := &recordingLogger{}
	mockClient := mocks.NewClient(t)
	mockRun := mocks.NewWorkflowRun(t)
	mockClient.On("SignalWithStartWorkflow", mock.Anything, "wf-1", "sig", nil, mock.Anything, "WF", nil).Return(mockRun, nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	err := SignalWithStartWorkflowFireAndForget(context.Background(), provider, opts, log)
	assert.NoError(t, err)
	assert.Len(t, log.infos, 1)
	assert.Contains(t, log.infos[0].msg, "signaled/started successfully")
}

func TestSignalWithStartWorkflowFireAndForget_NilLogger(t *testing.T) {
	provider := &mockProvider{client: nil, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	assert.NotPanics(t, func() {
		_ = SignalWithStartWorkflowFireAndForget(context.Background(), provider, opts, nil)
	})
}

func TestSignalWorkflowFireAndForget_Success_LogsInfo(t *testing.T) {
	log := &recordingLogger{}
	mockClient := mocks.NewClient(t)
	mockClient.On("SignalWorkflow", mock.Anything, "wf-1", "run-1", "sig", nil).Return(nil)
	provider := &mockProvider{client: mockClient, err: nil}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "run-1", SignalName: "sig", SignalPayload: nil}
	err := SignalWorkflowFireAndForget(context.Background(), provider, opts, log)
	assert.NoError(t, err)
	assert.Len(t, log.infos, 1)
	assert.Contains(t, log.infos[0].msg, "signaled successfully")
}

func TestSignalWorkflowFireAndForget_NilLogger(t *testing.T) {
	provider := &mockProvider{client: nil, err: nil}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "run-1", SignalName: "sig", SignalPayload: nil}
	assert.NotPanics(t, func() {
		_ = SignalWorkflowFireAndForget(context.Background(), provider, opts, nil)
	})
}

func TestCustomIdentityClientProvider_CloseClient_CloseAfterTrue_NilClient(t *testing.T) {
	p := &CustomIdentityClientProvider{CloseAfter: true}
	assert.NotPanics(t, func() { p.CloseClient(nil) })
}

func TestSignalWithStartWorkflowFireAndForget_Error_LogsWarn(t *testing.T) {
	log := &recordingLogger{}
	provider := &mockProvider{client: nil, err: nil}
	opts := SignalWithStartWorkflowOptions{
		WorkflowID: "wf-1", TaskQueue: "q", Policy: policies.OneTimeBusinessProcess,
		WorkflowType: "WF", SignalName: "sig", SignalPayload: nil, WorkflowInput: nil,
	}
	err := SignalWithStartWorkflowFireAndForget(context.Background(), provider, opts, log)
	assert.NoError(t, err)
	assert.Len(t, log.warns, 1)
	assert.Contains(t, log.warns[0].msg, "Failed to signal/start workflow")
}

func TestSignalWorkflowFireAndForget_Error_LogsWarn(t *testing.T) {
	log := &recordingLogger{}
	provider := &mockProvider{client: nil, err: nil}
	opts := SignalWorkflowOptions{WorkflowID: "wf-1", RunID: "", SignalName: "sig", SignalPayload: nil}
	err := SignalWorkflowFireAndForget(context.Background(), provider, opts, log)
	assert.NoError(t, err)
	assert.Len(t, log.warns, 1)
	assert.Contains(t, log.warns[0].msg, "Failed to signal workflow")
}

type mockProvider struct {
	client client.Client
	err    error
}

func (m *mockProvider) GetTemporalClient(ctx context.Context) (client.Client, error) {
	return m.client, m.err
}

func (m *mockProvider) CloseClient(c client.Client) {}

type logEntry struct {
	msg string
	kv  []interface{}
}

type recordingLogger struct {
	warns []logEntry
	infos []logEntry
}

func (r *recordingLogger) Warn(msg string, keysAndValues ...interface{}) {
	r.warns = append(r.warns, logEntry{msg: msg, kv: keysAndValues})
}

func (r *recordingLogger) Info(msg string, keysAndValues ...interface{}) {
	r.infos = append(r.infos, logEntry{msg: msg, kv: keysAndValues})
}
