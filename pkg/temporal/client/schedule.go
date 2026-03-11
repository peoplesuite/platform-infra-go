package client

import (
	"time"

	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DefaultActivityOptions returns activity options with the given timeout and default retry policy.
func DefaultActivityOptions(timeout time.Duration) workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
}

// RetryableActivityOptions returns activity options with timeout and custom max attempts.
func RetryableActivityOptions(timeout time.Duration, maxAttempts int) workflow.ActivityOptions {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: int32(maxAttempts),
		},
	}
}

// CronScheduleSpec returns a schedule spec for a cron expression (for use with client.ScheduleOptions).
func CronScheduleSpec(cron string) *sdkclient.ScheduleSpec {
	return &sdkclient.ScheduleSpec{
		CronExpressions: []string{cron},
	}
}
