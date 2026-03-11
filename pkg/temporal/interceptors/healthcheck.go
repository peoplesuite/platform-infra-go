package interceptors

import (
	"context"

	"go.temporal.io/sdk/interceptor"
)

// PollMarker is implemented by health check managers that track worker poll activity.
// Call MarkPoll() when the worker processes a task to avoid stale health check status.
type PollMarker interface {
	MarkPoll()
}

// HealthCheckInterceptor is a worker interceptor that marks poll activity
// on every task execution to prevent stale health check status.
type HealthCheckInterceptor struct {
	interceptor.WorkerInterceptorBase
	healthMgr PollMarker
}

// NewHealthCheckInterceptor creates a new health check interceptor.
// Pass nil healthMgr to disable (no-op).
func NewHealthCheckInterceptor(healthMgr PollMarker) interceptor.WorkerInterceptor {
	return &HealthCheckInterceptor{healthMgr: healthMgr}
}

// InterceptActivity implements WorkerInterceptor.InterceptActivity.
// It wraps activity execution to mark poll on every activity invocation.
func (h *HealthCheckInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	i := &healthCheckActivityInboundInterceptor{
		healthMgr: h.healthMgr,
	}
	i.Next = next
	return i
}

// healthCheckActivityInboundInterceptor intercepts activity execution
// to update health check poll status.
type healthCheckActivityInboundInterceptor struct {
	interceptor.ActivityInboundInterceptorBase
	healthMgr PollMarker
}

// ExecuteActivity implements ActivityInboundInterceptor.ExecuteActivity.
// It marks poll before executing the activity to update health check status.
func (h *healthCheckActivityInboundInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	// Mark poll to indicate worker is actively processing tasks
	// This prevents stale health check failures when StaleThreshold is configured
	if h.healthMgr != nil {
		h.healthMgr.MarkPoll()
	}

	// Execute the activity
	return h.Next.ExecuteActivity(ctx, in)
}
