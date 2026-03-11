package interceptors

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/interceptor"
)

func TestNewPrometheusActivityRecorder_ImplementsInterface(t *testing.T) {
	reg := prometheus.NewRegistry()
	rec := NewPrometheusActivityRecorder(reg)
	require.NotNil(t, rec)
	var _ = rec // assert returns ActivityMetricsRecorder
}

func TestNewPrometheusActivityRecorder_NilRegistry(t *testing.T) {
	rec := NewPrometheusActivityRecorder(nil)
	require.NotNil(t, rec)
	rec.RecordActivityStarted("test")
	rec.RecordActivityCompleted("test", 0)
	rec.RecordActivityFailed("test", 0)
}

func TestPrometheusActivityRecorder_RecordsFailed(t *testing.T) {
	reg := prometheus.NewRegistry()
	rec := NewPrometheusActivityRecorder(reg)
	w := NewMetricsInterceptor(rec)
	next := &mockActivityNextError{err: assert.AnError}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.Error(t, err)

	metrics, err := reg.Gather()
	require.NoError(t, err)
	var failed float64
	for _, m := range metrics {
		if m.GetName() == activityFailedTotalName {
			failed = getCounterValue(m)
			break
		}
	}
	assert.Equal(t, 1.0, failed)
}

func TestPrometheusActivityRecorder_RecordsMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	rec := NewPrometheusActivityRecorder(reg)
	w := NewMetricsInterceptor(rec)
	next := &mockActivityNext{}
	chain := w.InterceptActivity(context.Background(), next)
	_, err := chain.ExecuteActivity(context.Background(), &interceptor.ExecuteActivityInput{})
	assert.NoError(t, err)
	assert.True(t, next.executed)

	metrics, err := reg.Gather()
	require.NoError(t, err)
	var started, completed, failed float64
	for _, m := range metrics {
		if m.GetName() == activityStartedTotalName {
			started = getCounterValue(m)
		}
		if m.GetName() == activityCompletedTotalName {
			completed = getCounterValue(m)
		}
		if m.GetName() == activityFailedTotalName {
			failed = getCounterValue(m)
		}
	}
	assert.Equal(t, 1.0, started)
	assert.Equal(t, 1.0, completed)
	assert.Equal(t, 0.0, failed)
}

func getCounterValue(m *dto.MetricFamily) float64 {
	if m.GetType() != dto.MetricType_COUNTER {
		return 0
	}
	var sum float64
	for _, metric := range m.GetMetric() {
		if metric.Counter != nil {
			sum += metric.Counter.GetValue()
		}
	}
	return sum
}
