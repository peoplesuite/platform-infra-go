package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLoggingInterceptor_NilLogger(t *testing.T) {
	w := NewLoggingInterceptor(nil)
	require.NotNil(t, w)
	require.Implements(t, (*interceptor.WorkerInterceptor)(nil), w)

	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	require.NotNil(t, chain)
	out, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, next.executed)
}

func TestNewLoggingInterceptor_WithLogger_LogsStartAndComplete(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	w := NewLoggingInterceptor(logger)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)

	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.True(t, next.executed)

	logs := observed.All()
	require.Len(t, logs, 2)
	assert.Equal(t, "activity started", logs[0].Message)
	assert.Equal(t, "activity completed", logs[1].Message)
	assert.Equal(t, "unknown", logs[0].ContextMap()["activity_type"])
}

func TestNewLoggingInterceptor_WithLogger_LogsFailure(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	w := NewLoggingInterceptor(logger)
	next := &mockActivityNextError{err: assert.AnError}
	chain := w.InterceptActivity(context.Background(), next)

	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Same(t, assert.AnError, err)

	logs := observed.All()
	require.Len(t, logs, 2)
	assert.Equal(t, "activity started", logs[0].Message)
	assert.Equal(t, "activity failed", logs[1].Message)
}

type mockActivityNextError struct {
	interceptor.ActivityInboundInterceptorBase
	err error
}

func (m *mockActivityNextError) ExecuteActivity(ctx context.Context, in *interceptor.ExecuteActivityInput) (interface{}, error) {
	return nil, m.err
}
