package start

import (
	"go.temporal.io/sdk/client"

	"github.com/peoplesuite/platform-infra-go/pkg/temporal/policies"
)

// NewOptions creates a StartWorkflowOptions with the given workflow ID, task queue, and policy
func NewOptions(workflowID, taskQueue string, policy policies.WorkflowPolicy) client.StartWorkflowOptions {
	return client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                taskQueue,
		WorkflowExecutionTimeout: policy.ExecutionTimeout,
		WorkflowRunTimeout:       policy.RunTimeout,
		WorkflowTaskTimeout:      policy.TaskTimeout,
		WorkflowIDReusePolicy:    policy.ReusePolicy,
	}
}

// WithMemo adds memo metadata to the workflow options
func WithMemo(opts client.StartWorkflowOptions, memo map[string]interface{}) client.StartWorkflowOptions {
	opts.Memo = memo
	return opts
}

// WithSearchAttributes adds searchable attributes to the workflow options.
func WithSearchAttributes(opts client.StartWorkflowOptions, attrs map[string]interface{}) client.StartWorkflowOptions {
	opts.SearchAttributes = attrs //nolint:staticcheck // SA1019: deprecated, use TypedSearchAttributes when migrating
	return opts
}
