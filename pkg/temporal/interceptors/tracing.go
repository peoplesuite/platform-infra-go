package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/interceptor"
)

const tracerName = "github.com/peoplesuite/platform-infra-go/pkg/temporal/interceptors"

// TracingOption configures TracingInterceptor.
type TracingOption func(*TracingInterceptor)

// WithTracerProvider sets the TracerProvider. If not set, otel.GetTracerProvider() is used.
func WithTracerProvider(provider trace.TracerProvider) TracingOption {
	return func(t *TracingInterceptor) {
		t.tracerProvider = provider
	}
}

// TracingInterceptor is a worker interceptor that creates OpenTelemetry spans for activity execution.
type TracingInterceptor struct {
	interceptor.WorkerInterceptorBase
	tracerProvider trace.TracerProvider
}

// NewTracingInterceptor returns a TracingInterceptor. Options can customize the tracer provider (default: global).
func NewTracingInterceptor(opts ...TracingOption) interceptor.WorkerInterceptor {
	t := &TracingInterceptor{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// InterceptActivity implements WorkerInterceptor.InterceptActivity.
func (t *TracingInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	i := &tracingActivityInboundInterceptor{interceptor: t}
	i.Next = next
	return i
}

type tracingActivityInboundInterceptor struct {
	interceptor.ActivityInboundInterceptorBase
	interceptor *TracingInterceptor
}

func (t *tracingActivityInboundInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	provider := t.interceptor.tracerProvider
	if provider == nil {
		provider = otel.GetTracerProvider()
	}
	tracer := provider.Tracer(tracerName)
	activityType := "unknown"
	workflowID := ""
	activityID := ""
	taskQueue := ""
	if info := getActivityInfoSafe(ctx); info != nil {
		activityType = info.ActivityType.Name
		workflowID = info.WorkflowExecution.ID
		activityID = info.ActivityID
		taskQueue = info.TaskQueue
	}
	spanName := activityType
	if spanName == "" {
		spanName = "activity"
	}
	ctx, span := tracer.Start(ctx, spanName, trace.WithAttributes(
		attribute.String("activity.type", activityType),
		attribute.String("workflow.id", workflowID),
		attribute.String("activity.id", activityID),
		attribute.String("task_queue", taskQueue),
	))
	defer span.End()
	result, err := t.Next.ExecuteActivity(ctx, in)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return result, err
	}
	return result, nil
}
