package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Struct(t *testing.T) {
	opts := Options{
		Addr:      "localhost:6379",
		Username:  "user",
		Password:  "secret",
		DB:        1,
		TLS:       true,
		VerifyTLS: true,
		Timeout:   10 * time.Second,
	}
	assert.Equal(t, "localhost:6379", opts.Addr)
	assert.Equal(t, "user", opts.Username)
	assert.Equal(t, "secret", opts.Password)
	assert.Equal(t, 1, opts.DB)
	assert.True(t, opts.TLS)
	assert.True(t, opts.VerifyTLS)
	assert.Equal(t, 10*time.Second, opts.Timeout)
}

func TestOptions_ZeroValues(t *testing.T) {
	opts := Options{}
	assert.Empty(t, opts.Addr)
	assert.Equal(t, 0, opts.DB)
	assert.False(t, opts.TLS)
	assert.False(t, opts.VerifyTLS)
}
