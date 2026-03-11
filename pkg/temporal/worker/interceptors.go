package worker

import (
	"go.temporal.io/sdk/interceptor"

	"github.com/peoplesuite/platform-infra-go/pkg/temporal/interceptors"
)

// DefaultInterceptors returns the default interceptor chain without health check.
// Logging and metrics use nil (no-op); tracing uses the global OpenTelemetry tracer provider.
// To enable activity logging or metrics, build the chain with interceptors.NewLoggingInterceptor(logger)
// and/or interceptors.NewMetricsInterceptor(recorder), e.g. interceptors.NewMetricsInterceptor(interceptors.NewPrometheusActivityRecorder(reg)).
// Use DefaultInterceptorsWithHealthCheck if you need health check poll marking.
func DefaultInterceptors() []interceptor.WorkerInterceptor {
	return []interceptor.WorkerInterceptor{
		interceptors.NewLoggingInterceptor(nil),
		interceptors.NewMetricsInterceptor(nil),
		interceptors.NewTracingInterceptor(),
	}
}

// DefaultInterceptorsWithHealthCheck returns the default interceptor chain
// with health check poll marking enabled.
// The health check interceptor is placed first to ensure poll marking happens
// before any other interceptor logic.
// Pass any type that implements interceptors.PollMarker (e.g. has MarkPoll()).
func DefaultInterceptorsWithHealthCheck(healthMgr interceptors.PollMarker) []interceptor.WorkerInterceptor {
	return []interceptor.WorkerInterceptor{
		interceptors.NewHealthCheckInterceptor(healthMgr),
		interceptors.NewLoggingInterceptor(nil),
		interceptors.NewMetricsInterceptor(nil),
		interceptors.NewTracingInterceptor(),
	}
}
