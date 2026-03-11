package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/peoplesuite/platform-infra-go/pkg/temporal/config"
)

func TestOptionsFromProfile(t *testing.T) {
	profile := config.TemporalProfile{
		Address:         "localhost:7233",
		Namespace:       "default",
		Identity:        "my-worker",
		RPCTimeout:      15 * time.Second,
		TaskQueue:       "my-queue",
		MetadataHeaders: map[string]string{"k": "v"},
	}

	opts := OptionsFromProfile(profile)
	assert.Equal(t, "localhost:7233", opts.Address)
	assert.Equal(t, "default", opts.Namespace)
	assert.Equal(t, "my-worker", opts.Identity)
	assert.Equal(t, 15*time.Second, opts.RPCTimeout)
	assert.Equal(t, map[string]string{"k": "v"}, opts.MetadataHeaders)
}

func TestOptionsFromProfile_NilMetadataHeaders(t *testing.T) {
	profile := config.TemporalProfile{
		Address:   "addr",
		Namespace: "ns",
	}

	opts := OptionsFromProfile(profile)
	requireNotNil(t, opts.MetadataHeaders)
	assert.Empty(t, opts.MetadataHeaders)
}

func requireNotNil(t *testing.T, m map[string]string) {
	t.Helper()
	if m == nil {
		t.Fatal("expected non-nil MetadataHeaders")
	}
}

func TestNewFromProfile_EmptyAddress(t *testing.T) {
	ctx := context.Background()
	profile := config.TemporalProfile{
		Address:   "",
		Namespace: "default",
	}

	c, err := NewFromProfile(ctx, profile)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "address is required")
}

func TestNewFromProfileLoader_LoaderError(t *testing.T) {
	ctx := context.Background()
	loaderErr := errors.New("load failed")
	loader := func() (config.TemporalProfile, error) {
		return config.TemporalProfile{}, loaderErr
	}

	c, err := NewFromProfileLoader(ctx, loader)

	assert.NoError(t, err)
	assert.Nil(t, c)
}

func TestNewFromProfileLoaderWithError_LoaderError(t *testing.T) {
	ctx := context.Background()
	loaderErr := errors.New("load failed")
	loader := func() (config.TemporalProfile, error) {
		return config.TemporalProfile{}, loaderErr
	}

	c, err := NewFromProfileLoaderWithError(ctx, loader)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Equal(t, loaderErr, err)
}

func TestNewFromProfileLoaderWithError_ProfileValidationError(t *testing.T) {
	ctx := context.Background()
	loader := func() (config.TemporalProfile, error) {
		return config.TemporalProfile{
			Address:   "",
			Namespace: "default",
		}, nil
	}

	c, err := NewFromProfileLoaderWithError(ctx, loader)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "address is required")
}
