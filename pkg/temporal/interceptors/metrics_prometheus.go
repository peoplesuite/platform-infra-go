package interceptors

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	activityStartedTotalName    = "temporal_activity_started_total"
	activityCompletedTotalName  = "temporal_activity_completed_total"
	activityFailedTotalName     = "temporal_activity_failed_total"
	activityDurationSecondsName = "temporal_activity_duration_seconds"
	activityTypeLabel           = "activity_type"
)

// PrometheusActivityRecorder implements ActivityMetricsRecorder with Prometheus counters and histogram.
// Pass a custom registry or nil to use prometheus.DefaultRegisterer.
type PrometheusActivityRecorder struct {
	started   *prometheus.CounterVec
	completed *prometheus.CounterVec
	failed    *prometheus.CounterVec
	duration  *prometheus.HistogramVec
}

// NewPrometheusActivityRecorder creates a Prometheus-backed ActivityMetricsRecorder and registers
// its metrics with the given registry. If registry is nil, prometheus.DefaultRegisterer is used.
func NewPrometheusActivityRecorder(registry prometheus.Registerer) ActivityMetricsRecorder {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}
	p := &PrometheusActivityRecorder{
		started: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: activityStartedTotalName,
			Help: "Total number of activity executions started.",
		}, []string{activityTypeLabel}),
		completed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: activityCompletedTotalName,
			Help: "Total number of activity executions completed successfully.",
		}, []string{activityTypeLabel}),
		failed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: activityFailedTotalName,
			Help: "Total number of activity executions that failed.",
		}, []string{activityTypeLabel}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    activityDurationSecondsName,
			Help:    "Activity execution duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{activityTypeLabel}),
	}
	registry.MustRegister(p.started, p.completed, p.failed, p.duration)
	return p
}

// RecordActivityStarted implements ActivityMetricsRecorder.
func (p *PrometheusActivityRecorder) RecordActivityStarted(activityType string) {
	p.started.WithLabelValues(activityType).Inc()
}

// RecordActivityCompleted implements ActivityMetricsRecorder.
func (p *PrometheusActivityRecorder) RecordActivityCompleted(activityType string, duration time.Duration) {
	p.completed.WithLabelValues(activityType).Inc()
	p.duration.WithLabelValues(activityType).Observe(duration.Seconds())
}

// RecordActivityFailed implements ActivityMetricsRecorder.
func (p *PrometheusActivityRecorder) RecordActivityFailed(activityType string, duration time.Duration) {
	p.failed.WithLabelValues(activityType).Inc()
	p.duration.WithLabelValues(activityType).Observe(duration.Seconds())
}
