package redis

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	inner := errors.New("connection refused")

	t.Run("with key", func(t *testing.T) {
		e := &Error{Op: "Get", Key: "user:1", Err: inner}
		s := e.Error()
		assert.Contains(t, s, "redis")
		assert.Contains(t, s, "Get")
		assert.Contains(t, s, "user:1")
		assert.Contains(t, s, "connection refused")
	})

	t.Run("without key", func(t *testing.T) {
		e := &Error{Op: "Ping", Key: "", Err: inner}
		s := e.Error()
		assert.Contains(t, s, "redis")
		assert.Contains(t, s, "Ping")
		assert.NotContains(t, s, "key=")
		assert.Contains(t, s, "connection refused")
	})
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	e := &Error{Op: "Set", Key: "k", Err: inner}
	assert.Same(t, inner, e.Unwrap())

	var re *Error
	require.True(t, errors.As(e, &re))
	assert.Equal(t, "Set", re.Op)
	assert.Equal(t, "k", re.Key)
}
