package neo4j

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		got := Wrap("op", nil)
		assert.Nil(t, got)
	})

	t.Run("non-nil returns Error with Op and Unwrap", func(t *testing.T) {
		inner := errors.New("inner")
		got := Wrap("myop", inner)
		require.NotNil(t, got)

		var e *Error
		require.True(t, errors.As(got, &e))
		assert.Equal(t, "myop", e.Op)
		assert.Same(t, inner, e.Err)
		assert.Same(t, inner, e.Unwrap())
	})
}

func TestError_Error(t *testing.T) {
	inner := errors.New("connection refused")
	e := &Error{Op: "query", Err: inner}
	assert.Contains(t, e.Error(), "neo4j")
	assert.Contains(t, e.Error(), "query")
	assert.Contains(t, e.Error(), "connection refused")
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	e := &Error{Op: "op", Err: inner}
	assert.Same(t, inner, e.Unwrap())
}
