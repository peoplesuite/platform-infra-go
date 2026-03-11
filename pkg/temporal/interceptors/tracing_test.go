package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.temporal.io/sdk/interceptor"
)

func TestNewTracingInterceptor(t *testing.T) {
	w := NewTracingInterceptor()
	require.NotNil(t, w)
	require.Implements(t, (*interceptor.WorkerInterceptor)(nil), w)
}

func TestTracingInterceptor_ChainRuns(t *testing.T) {
	w := NewTracingInterceptor()
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	require.NotNil(t, chain)
	out, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, next.executed)
}

func TestTracingInterceptor_PropagatesError(t *testing.T) {
	w := NewTracingInterceptor()
	next := &mockActivityNextError{err: assert.AnError}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Same(t, assert.AnError, err)
}

func TestNewTracingInterceptor_WithTracerProvider(t *testing.T) {
	// WithTracerProvider(nil) leaves provider nil so ExecuteActivity uses global; chain should still run.
	w := NewTracingInterceptor(WithTracerProvider(nil))
	require.NotNil(t, w)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
}

func TestTracingInterceptor_WithNonNilTracerProvider(t *testing.T) {
	provider := noop.NewTracerProvider()
	w := NewTracingInterceptor(WithTracerProvider(provider))
	require.NotNil(t, w)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.True(t, next.executed)
}

func TestTracingInterceptor_WithTracerProvider_ErrorPath(t *testing.T) {
	provider := noop.NewTracerProvider()
	w := NewTracingInterceptor(WithTracerProvider(provider))
	next := &mockActivityNextError{err: assert.AnError}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Same(t, assert.AnError, err)
}
