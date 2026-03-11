package client

import (
	"context"

	"go.temporal.io/sdk/client"

	temporalconfig "github.com/peoplesuite/platform-infra-go/pkg/temporal/config"
)

// OptionsFromProfile converts a config TemporalProfile to client Options.
func OptionsFromProfile(profile temporalconfig.TemporalProfile) Options {
	opts := Options{
		Address:         profile.Address,
		Namespace:       profile.Namespace,
		Identity:        profile.Identity,
		RPCTimeout:      profile.RPCTimeout,
		MetadataHeaders: profile.MetadataHeaders,
	}
	if opts.MetadataHeaders == nil {
		opts.MetadataHeaders = make(map[string]string)
	}
	return opts
}

// NewFromProfile creates a Temporal client from a config TemporalProfile (e.g. from temporal.toml).
func NewFromProfile(ctx context.Context, profile temporalconfig.TemporalProfile) (client.Client, error) {
	return New(ctx, OptionsFromProfile(profile))
}

// NewFromProfileLoader creates a Temporal client using a profile loader function.
// If the loader returns an error, it returns (nil, nil) to support fire-and-forget patterns.
// This is useful for activities that need to create clients on-demand but shouldn't fail workflows.
func NewFromProfileLoader(ctx context.Context, loader func() (temporalconfig.TemporalProfile, error)) (client.Client, error) {
	profile, err := loader()
	if err != nil {
		return nil, nil // Fire-and-forget: return nil instead of error
	}
	return NewFromProfile(ctx, profile)
}

// NewFromProfileLoaderWithError creates a Temporal client using a profile loader function.
// Unlike NewFromProfileLoader, this function returns errors from the loader.
// Use this when you need explicit error handling (e.g., in main functions).
func NewFromProfileLoaderWithError(ctx context.Context, loader func() (temporalconfig.TemporalProfile, error)) (client.Client, error) {
	profile, err := loader()
	if err != nil {
		return nil, err
	}
	return NewFromProfile(ctx, profile)
}
