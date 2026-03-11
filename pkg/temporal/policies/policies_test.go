package policies

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
)

func TestWorkflowPolicyPresets(t *testing.T) {
	presets := []struct {
		name     string
		policy   WorkflowPolicy
		expected struct {
			execTimeout time.Duration
			runTimeout  time.Duration
			taskTimeout time.Duration
			reusePolicy enums.WorkflowIdReusePolicy
		}
	}{
		{
			name:   "OneTimeBusinessProcess",
			policy: OneTimeBusinessProcess,
			expected: struct {
				execTimeout time.Duration
				runTimeout  time.Duration
				taskTimeout time.Duration
				reusePolicy enums.WorkflowIdReusePolicy
			}{
				execTimeout: 1 * time.Hour,
				runTimeout:  1 * time.Hour,
				taskTimeout: 30 * time.Second,
				reusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
			},
		},
		{
			name:   "RepeatableProcess",
			policy: RepeatableProcess,
			expected: struct {
				execTimeout time.Duration
				runTimeout  time.Duration
				taskTimeout time.Duration
				reusePolicy enums.WorkflowIdReusePolicy
			}{
				execTimeout: 2 * time.Hour,
				runTimeout:  2 * time.Hour,
				taskTimeout: 30 * time.Second,
				reusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
			},
		},
		{
			name:   "NotificationProcess",
			policy: NotificationProcess,
			expected: struct {
				execTimeout time.Duration
				runTimeout  time.Duration
				taskTimeout time.Duration
				reusePolicy enums.WorkflowIdReusePolicy
			}{
				execTimeout: 30 * time.Minute,
				runTimeout:  30 * time.Minute,
				taskTimeout: 10 * time.Second,
				reusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
			},
		},
		{
			name:   "EventStorageProcess",
			policy: EventStorageProcess,
			expected: struct {
				execTimeout time.Duration
				runTimeout  time.Duration
				taskTimeout time.Duration
				reusePolicy enums.WorkflowIdReusePolicy
			}{
				execTimeout: 15 * time.Minute,
				runTimeout:  15 * time.Minute,
				taskTimeout: 10 * time.Second,
				reusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
			},
		},
		{
			name:   "LongRunningVacationWindow",
			policy: LongRunningVacationWindow,
			expected: struct {
				execTimeout time.Duration
				runTimeout  time.Duration
				taskTimeout time.Duration
				reusePolicy enums.WorkflowIdReusePolicy
			}{
				execTimeout: 400 * 24 * time.Hour,
				runTimeout:  400 * 24 * time.Hour,
				taskTimeout: 30 * time.Second,
				reusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
			},
		},
		{
			name:   "AccessDuringVacationWindow",
			policy: AccessDuringVacationWindow,
			expected: struct {
				execTimeout time.Duration
				runTimeout  time.Duration
				taskTimeout time.Duration
				reusePolicy enums.WorkflowIdReusePolicy
			}{
				execTimeout: 8 * 24 * time.Hour,
				runTimeout:  8 * 24 * time.Hour,
				taskTimeout: 30 * time.Second,
				reusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
			},
		},
	}

	for _, tt := range presets {
		t.Run(tt.name, func(t *testing.T) {
			assert.Greater(t, tt.policy.ExecutionTimeout, time.Duration(0), "ExecutionTimeout should be > 0")
			assert.Greater(t, tt.policy.RunTimeout, time.Duration(0), "RunTimeout should be > 0")
			assert.Greater(t, tt.policy.TaskTimeout, time.Duration(0), "TaskTimeout should be > 0")
			assert.Equal(t, tt.expected.execTimeout, tt.policy.ExecutionTimeout)
			assert.Equal(t, tt.expected.runTimeout, tt.policy.RunTimeout)
			assert.Equal(t, tt.expected.taskTimeout, tt.policy.TaskTimeout)
			assert.Equal(t, tt.expected.reusePolicy, tt.policy.ReusePolicy)
		})
	}
}

func TestShortActivity(t *testing.T) {
	assert.Greater(t, ShortActivity.ScheduleToStartTimeout, time.Duration(0))
	assert.Greater(t, ShortActivity.StartToCloseTimeout, time.Duration(0))
	require.NotNil(t, ShortActivity.RetryPolicy)
	assert.Equal(t, int32(3), ShortActivity.RetryPolicy.MaximumAttempts)
	assert.Contains(t, ShortActivity.RetryPolicy.NonRetryableErrorTypes, "ValidationError")
	assert.Contains(t, ShortActivity.RetryPolicy.NonRetryableErrorTypes, "BusinessLogicError")
	assert.Contains(t, ShortActivity.RetryPolicy.NonRetryableErrorTypes, "IdempotencyError")
	assert.Equal(t, 1*time.Minute, ShortActivity.ScheduleToStartTimeout)
	assert.Equal(t, 5*time.Second, ShortActivity.StartToCloseTimeout)
}

func TestMediumActivity(t *testing.T) {
	assert.Greater(t, MediumActivity.ScheduleToStartTimeout, time.Duration(0))
	assert.Greater(t, MediumActivity.StartToCloseTimeout, time.Duration(0))
	require.NotNil(t, MediumActivity.RetryPolicy)
	assert.Equal(t, int32(4), MediumActivity.RetryPolicy.MaximumAttempts)
	assert.Contains(t, MediumActivity.RetryPolicy.NonRetryableErrorTypes, "ValidationError")
	assert.Contains(t, MediumActivity.RetryPolicy.NonRetryableErrorTypes, "BusinessLogicError")
	assert.Contains(t, MediumActivity.RetryPolicy.NonRetryableErrorTypes, "IdempotencyError")
	assert.Equal(t, 5*time.Minute, MediumActivity.ScheduleToStartTimeout)
	assert.Equal(t, 30*time.Second, MediumActivity.StartToCloseTimeout)
}

func TestLongActivity(t *testing.T) {
	assert.Greater(t, LongActivity.ScheduleToStartTimeout, time.Duration(0))
	assert.Greater(t, LongActivity.StartToCloseTimeout, time.Duration(0))
	require.NotNil(t, LongActivity.RetryPolicy)
	assert.Equal(t, int32(3), LongActivity.RetryPolicy.MaximumAttempts)
	assert.Contains(t, LongActivity.RetryPolicy.NonRetryableErrorTypes, "ValidationError")
	assert.Contains(t, LongActivity.RetryPolicy.NonRetryableErrorTypes, "BusinessLogicError")
	assert.Contains(t, LongActivity.RetryPolicy.NonRetryableErrorTypes, "IdempotencyError")
	assert.Equal(t, 10*time.Minute, LongActivity.ScheduleToStartTimeout)
	assert.Equal(t, 10*time.Minute, LongActivity.StartToCloseTimeout)
}
