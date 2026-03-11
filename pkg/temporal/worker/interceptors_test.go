package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/interceptor"

	"github.com/peoplesuite/platform-infra-go/pkg/temporal/interceptors"
)

func TestDefaultInterceptors(t *testing.T) {
	chain := DefaultInterceptors()
	require.NotNil(t, chain)
	assert.Len(t, chain, 3)

	for _, w := range chain {
		require.Implements(t, (*interceptor.WorkerInterceptor)(nil), w)
	}
}

func TestDefaultInterceptorsWithHealthCheck_Nil(t *testing.T) {
	chain := DefaultInterceptorsWithHealthCheck(nil)
	require.NotNil(t, chain)
	assert.Len(t, chain, 4)

	first := chain[0]
	_, ok := first.(*interceptors.HealthCheckInterceptor)
	assert.True(t, ok, "first interceptor should be HealthCheckInterceptor")
}

func TestDefaultInterceptorsWithHealthCheck_WithMarker(t *testing.T) {
	marker := &mockPollMarker{}
	chain := DefaultInterceptorsWithHealthCheck(marker)
	require.NotNil(t, chain)
	assert.Len(t, chain, 4)

	first := chain[0]
	_, ok := first.(*interceptors.HealthCheckInterceptor)
	assert.True(t, ok, "first interceptor should be HealthCheckInterceptor")
	_ = marker
}

type mockPollMarker struct{ calls int }

func (m *mockPollMarker) MarkPoll() { m.calls++ }
