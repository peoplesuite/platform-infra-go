package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Struct(t *testing.T) {
	opts := Options{
		Address:           "localhost:7233",
		Namespace:         "default",
		Identity:          "test-worker",
		ConnectionTimeout: 5 * time.Second,
		RPCTimeout:        10 * time.Second,
		MetadataHeaders:   map[string]string{"x-custom": "value"},
	}
	assert.Equal(t, "localhost:7233", opts.Address)
	assert.Equal(t, "default", opts.Namespace)
	assert.Equal(t, "test-worker", opts.Identity)
	assert.Equal(t, 5*time.Second, opts.ConnectionTimeout)
	assert.Equal(t, 10*time.Second, opts.RPCTimeout)
	assert.Equal(t, map[string]string{"x-custom": "value"}, opts.MetadataHeaders)
}

func TestOptions_ZeroValues(t *testing.T) {
	opts := Options{}
	assert.Empty(t, opts.Address)
	assert.Empty(t, opts.Namespace)
	assert.Empty(t, opts.Identity)
	assert.Nil(t, opts.Logger)
	assert.Nil(t, opts.MetadataHeaders)
}
