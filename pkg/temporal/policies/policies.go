package policies

import (
	"time"

	"go.temporal.io/api/enums/v1"
)

// WorkflowPolicy defines timeout and reuse policy settings for workflow starts
type WorkflowPolicy struct {
	ExecutionTimeout time.Duration
	RunTimeout       time.Duration
	TaskTimeout      time.Duration
	ReusePolicy      enums.WorkflowIdReusePolicy
}

// Predefined policy presets for common workflow types
var (
	// OneTimeBusinessProcess is for workflows that should only run once per business entity
	// Example: employee onboarding, event routing per employee
	OneTimeBusinessProcess = WorkflowPolicy{
		ExecutionTimeout: 1 * time.Hour,
		RunTimeout:       1 * time.Hour,
		TaskTimeout:      30 * time.Second,
		ReusePolicy:      enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	// RepeatableProcess is for workflows that can be retried after failure
	// Example: collector workflows, scheduled syncs
	RepeatableProcess = WorkflowPolicy{
		ExecutionTimeout: 2 * time.Hour,
		RunTimeout:       2 * time.Hour,
		TaskTimeout:      30 * time.Second,
		ReusePolicy:      enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
	}

	// NotificationProcess is for short-lived notification workflows
	// Example: email notifications, alerts
	NotificationProcess = WorkflowPolicy{
		ExecutionTimeout: 30 * time.Minute,
		RunTimeout:       30 * time.Minute,
		TaskTimeout:      10 * time.Second,
		ReusePolicy:      enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	// EventStorageProcess is for workflows that store events (idempotent)
	// Example: daily event storage, event status collection
	EventStorageProcess = WorkflowPolicy{
		ExecutionTimeout: 15 * time.Minute,
		RunTimeout:       15 * time.Minute,
		TaskTimeout:      10 * time.Second,
		ReusePolicy:      enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	// LongRunningVacationWindow is for workflows that sleep until a vacation window (e.g. long-duration vacation sync).
	// 400 days covers the maximum typical vacation window.
	LongRunningVacationWindow = WorkflowPolicy{
		ExecutionTimeout: 400 * 24 * time.Hour,
		RunTimeout:       400 * 24 * time.Hour,
		TaskTimeout:      30 * time.Second,
		ReusePolicy:      enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}

	// AccessDuringVacationWindow is for access-during-vacation workflow: max 5 working days (8 calendar days).
	AccessDuringVacationWindow = WorkflowPolicy{
		ExecutionTimeout: 8 * 24 * time.Hour,
		RunTimeout:       8 * 24 * time.Hour,
		TaskTimeout:      30 * time.Second,
		ReusePolicy:      enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
	}
)
