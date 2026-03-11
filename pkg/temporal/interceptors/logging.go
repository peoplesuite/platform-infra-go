package interceptors

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"
)

// LoggingInterceptor is a worker interceptor that logs activity start, completion, and failure.
type LoggingInterceptor struct {
	interceptor.WorkerInterceptorBase
	logger *zap.Logger
}

// NewLoggingInterceptor returns a LoggingInterceptor. Pass nil logger for a no-op (no logs).
func NewLoggingInterceptor(logger *zap.Logger) interceptor.WorkerInterceptor {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LoggingInterceptor{logger: logger}
}

// InterceptActivity implements WorkerInterceptor.InterceptActivity.
func (l *LoggingInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	i := &loggingActivityInboundInterceptor{logger: l.logger}
	i.Next = next
	return i
}

type loggingActivityInboundInterceptor struct {
	interceptor.ActivityInboundInterceptorBase
	logger *zap.Logger
}

func (l *loggingActivityInboundInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	start := time.Now()
	activityType := "unknown"
	workflowID := ""
	activityID := ""
	if info := getActivityInfoSafe(ctx); info != nil {
		activityType = info.ActivityType.Name
		workflowID = info.WorkflowExecution.ID
		activityID = info.ActivityID
	}
	fields := []zap.Field{
		zap.String("activity_type", activityType),
		zap.String("workflow_id", workflowID),
		zap.String("activity_id", activityID),
	}
	l.logger.Info("activity started", fields...)
	result, err := l.Next.ExecuteActivity(ctx, in)
	duration := time.Since(start)
	fields = append(fields, zap.Duration("duration", duration))
	if err != nil {
		l.logger.Error("activity failed", append(fields, zap.Error(err))...)
		return result, err
	}
	l.logger.Info("activity completed", fields...)
	return result, nil
}

// getActivityInfoSafe returns activity info from ctx, or nil if not an activity context (e.g. in tests).
func getActivityInfoSafe(ctx context.Context) *activity.Info {
	var info *activity.Info
	func() {
		defer func() { _ = recover() }()
		i := activity.GetInfo(ctx)
		info = &i
	}()
	return info
}
