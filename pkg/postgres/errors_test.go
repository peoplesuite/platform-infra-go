package postgres

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	inner := errors.New("connection refused")
	e := &Error{Op: "ping", Err: inner}
	assert.Contains(t, e.Error(), "postgres")
	assert.Contains(t, e.Error(), "ping")
	assert.Contains(t, e.Error(), "connection refused")
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	e := &Error{Op: "op", Err: inner}
	assert.Same(t, inner, e.Unwrap())

	var pe *Error
	require.True(t, errors.As(e, &pe))
	assert.Equal(t, "op", pe.Op)
}
