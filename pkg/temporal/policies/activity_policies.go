package policies

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Activity timeout presets for common activity types.
// These should be used via workflow.WithActivityOptions(ctx, policies.ShortActivity)

var (
	// ShortActivity is for fast, local operations (cache lookups, simple transformations)
	// ScheduleToStartTimeout: 1 minute - time to wait for worker capacity
	// StartToCloseTimeout: 5 seconds - actual execution time
	ShortActivity = workflow.ActivityOptions{
		ScheduleToStartTimeout: 1 * time.Minute,
		StartToCloseTimeout:    5 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				"ValidationError",
				"BusinessLogicError",
				"IdempotencyError",
			},
		},
	}

	// MediumActivity is for typical API calls and database operations
	// ScheduleToStartTimeout: 5 minutes - time to wait for worker capacity
	// StartToCloseTimeout: 30 seconds - actual execution time
	MediumActivity = workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    4,
			NonRetryableErrorTypes: []string{
				"ValidationError",
				"BusinessLogicError",
				"IdempotencyError",
			},
		},
	}

	// LongActivity is for batch processing, file generation, and long-running operations
	// ScheduleToStartTimeout: 10 minutes - time to wait for worker capacity
	// StartToCloseTimeout: 10 minutes - actual execution time
	LongActivity = workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
			NonRetryableErrorTypes: []string{
				"ValidationError",
				"BusinessLogicError",
				"IdempotencyError",
			},
		},
	}
)
