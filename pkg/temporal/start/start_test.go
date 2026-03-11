package start

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/peoplesuite/platform-infra-go/pkg/temporal/policies"
)

func TestNewOptions(t *testing.T) {
	workflowID := "wf-1"
	taskQueue := "my-queue"
	policy := policies.OneTimeBusinessProcess

	opts := NewOptions(workflowID, taskQueue, policy)

	assert.Equal(t, workflowID, opts.ID)
	assert.Equal(t, taskQueue, opts.TaskQueue)
	assert.Equal(t, policy.ExecutionTimeout, opts.WorkflowExecutionTimeout)
	assert.Equal(t, policy.RunTimeout, opts.WorkflowRunTimeout)
	assert.Equal(t, policy.TaskTimeout, opts.WorkflowTaskTimeout)
	assert.Equal(t, policy.ReusePolicy, opts.WorkflowIDReusePolicy)
}

func TestNewOptions_WithDifferentPolicy(t *testing.T) {
	opts := NewOptions("id", "queue", policies.NotificationProcess)

	assert.Equal(t, 30*time.Minute, opts.WorkflowExecutionTimeout)
	assert.Equal(t, 30*time.Minute, opts.WorkflowRunTimeout)
	assert.Equal(t, 10*time.Second, opts.WorkflowTaskTimeout)
}

func TestWithMemo(t *testing.T) {
	opts := NewOptions("wf-1", "queue", policies.OneTimeBusinessProcess)
	memo := map[string]interface{}{
		"key": "value",
		"n":   42,
	}

	result := WithMemo(opts, memo)

	assert.Equal(t, memo, result.Memo)
}

func TestWithSearchAttributes(t *testing.T) {
	opts := NewOptions("wf-1", "queue", policies.OneTimeBusinessProcess)
	attrs := map[string]interface{}{
		"CustomKeywordField": "searchable",
		"CustomIntField":     int64(100),
	}

	result := WithSearchAttributes(opts, attrs)

	assert.Equal(t, attrs, result.SearchAttributes) //nolint:staticcheck // SA1019: deprecated field, migrate to TypedSearchAttributes later
}
