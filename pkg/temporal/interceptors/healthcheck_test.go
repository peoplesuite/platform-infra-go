package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/interceptor"
)

type mockPollMarker struct {
	markPollCalls int
}

func (m *mockPollMarker) MarkPoll() {
	m.markPollCalls++
}

// mockActivityNext records ExecuteActivity calls and returns (nil, nil).
type mockActivityNext struct {
	interceptor.ActivityInboundInterceptorBase
	executed bool
}

func (m *mockActivityNext) ExecuteActivity(ctx context.Context, in *interceptor.ExecuteActivityInput) (interface{}, error) {
	m.executed = true
	return nil, nil
}

func TestNewHealthCheckInterceptor_NilHealthMgr(t *testing.T) {
	w := NewHealthCheckInterceptor(nil)
	require.NotNil(t, w)

	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	require.NotNil(t, chain)

	out, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, next.executed)
}

func TestNewHealthCheckInterceptor_WithHealthMgr(t *testing.T) {
	marker := &mockPollMarker{}
	w := NewHealthCheckInterceptor(marker)
	require.NotNil(t, w)

	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	require.NotNil(t, chain)

	out, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, next.executed)
	assert.Equal(t, 1, marker.markPollCalls)
}

func TestHealthCheckInterceptor_MarkPollBeforeNext(t *testing.T) {
	marker := &mockPollMarker{}
	w := NewHealthCheckInterceptor(marker)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)

	_, _ = chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Equal(t, 1, marker.markPollCalls)
	_, _ = chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Equal(t, 2, marker.markPollCalls)
}
