package interceptors

import (
	"context"
	"time"

	"go.temporal.io/sdk/interceptor"
)

// ActivityMetricsRecorder records activity execution metrics. Implement this interface
// (e.g. with a Prometheus implementation) and pass it to NewMetricsInterceptor.
// Pass nil to NewMetricsInterceptor for a no-op.
type ActivityMetricsRecorder interface {
	RecordActivityStarted(activityType string)
	RecordActivityCompleted(activityType string, duration time.Duration)
	RecordActivityFailed(activityType string, duration time.Duration)
}

// MetricsInterceptor is a worker interceptor that records activity metrics.
type MetricsInterceptor struct {
	interceptor.WorkerInterceptorBase
	recorder ActivityMetricsRecorder
}

// NewMetricsInterceptor returns a MetricsInterceptor. Pass nil recorder for a no-op.
func NewMetricsInterceptor(recorder ActivityMetricsRecorder) interceptor.WorkerInterceptor {
	return &MetricsInterceptor{recorder: recorder}
}

// InterceptActivity implements WorkerInterceptor.InterceptActivity.
func (m *MetricsInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	i := &metricsActivityInboundInterceptor{recorder: m.recorder}
	i.Next = next
	return i
}

type metricsActivityInboundInterceptor struct {
	interceptor.ActivityInboundInterceptorBase
	recorder ActivityMetricsRecorder
}

func (m *metricsActivityInboundInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	if m.recorder == nil {
		return m.Next.ExecuteActivity(ctx, in)
	}
	activityType := "unknown"
	if info := getActivityInfoSafe(ctx); info != nil {
		activityType = info.ActivityType.Name
	}
	m.recorder.RecordActivityStarted(activityType)
	start := time.Now()
	result, err := m.Next.ExecuteActivity(ctx, in)
	duration := time.Since(start)
	if err != nil {
		m.recorder.RecordActivityFailed(activityType, duration)
		return result, err
	}
	m.recorder.RecordActivityCompleted(activityType, duration)
	return result, nil
}
