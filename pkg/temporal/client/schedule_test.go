package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultActivityOptions(t *testing.T) {
	timeout := 15 * time.Second

	opts := DefaultActivityOptions(timeout)

	assert.Equal(t, timeout, opts.StartToCloseTimeout)
	require.NotNil(t, opts.RetryPolicy)
	assert.Equal(t, int32(3), opts.RetryPolicy.MaximumAttempts)
}

func TestRetryableActivityOptions_PositiveMaxAttempts(t *testing.T) {
	timeout := 10 * time.Second
	maxAttempts := 5

	opts := RetryableActivityOptions(timeout, maxAttempts)

	assert.Equal(t, timeout, opts.StartToCloseTimeout)
	require.NotNil(t, opts.RetryPolicy)
	assert.Equal(t, int32(5), opts.RetryPolicy.MaximumAttempts)
}

func TestRetryableActivityOptions_ZeroMaxAttempts(t *testing.T) {
	timeout := 10 * time.Second

	opts := RetryableActivityOptions(timeout, 0)

	require.NotNil(t, opts.RetryPolicy)
	assert.Equal(t, int32(3), opts.RetryPolicy.MaximumAttempts)
}

func TestRetryableActivityOptions_NegativeMaxAttempts(t *testing.T) {
	opts := RetryableActivityOptions(time.Second, -1)

	require.NotNil(t, opts.RetryPolicy)
	assert.Equal(t, int32(3), opts.RetryPolicy.MaximumAttempts)
}

func TestCronScheduleSpec(t *testing.T) {
	cron := "0 0 * * *"

	spec := CronScheduleSpec(cron)

	require.NotNil(t, spec)
	assert.Len(t, spec.CronExpressions, 1)
	assert.Equal(t, []string{cron}, spec.CronExpressions)
}
