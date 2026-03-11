package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/interceptor"
)

func TestNewMetricsInterceptor(t *testing.T) {
	w := NewMetricsInterceptor(nil)
	require.NotNil(t, w)
	require.Implements(t, (*interceptor.WorkerInterceptor)(nil), w)
}

func TestMetricsInterceptor_NilRecorder_NoPanic(t *testing.T) {
	w := NewMetricsInterceptor(nil)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	out, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.Nil(t, out)
	assert.True(t, next.executed)
}

func TestMetricsInterceptor_WithRecorder_RecordsStartedCompleted(t *testing.T) {
	rec := &fakeMetricsRecorder{}
	w := NewMetricsInterceptor(rec)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.True(t, next.executed)
	assert.Equal(t, 1, rec.started)
	assert.Equal(t, 1, rec.completed)
	assert.Equal(t, 0, rec.failed)
	assert.True(t, rec.lastDuration >= 0)
}

func TestMetricsInterceptor_WithRecorder_RecordsFailed(t *testing.T) {
	rec := &fakeMetricsRecorder{}
	w := NewMetricsInterceptor(rec)
	next := &mockActivityNextError{err: assert.AnError}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Same(t, assert.AnError, err)
	assert.Equal(t, 1, rec.started)
	assert.Equal(t, 0, rec.completed)
	assert.Equal(t, 1, rec.failed)
	assert.True(t, rec.lastDuration >= 0)
}

type fakeMetricsRecorder struct {
	started, completed, failed int
	lastDuration               time.Duration
}

func (f *fakeMetricsRecorder) RecordActivityStarted(activityType string) {
	f.started++
}

func (f *fakeMetricsRecorder) RecordActivityCompleted(activityType string, duration time.Duration) {
	f.completed++
	f.lastDuration = duration
}

func (f *fakeMetricsRecorder) RecordActivityFailed(activityType string, duration time.Duration) {
	f.failed++
	f.lastDuration = duration
}
