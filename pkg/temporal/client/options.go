package client

import (
	"time"

	"go.temporal.io/sdk/log"
)

// Options configures Temporal client creation.
type Options struct {
	Address   string
	Namespace string

	Identity string
	Logger   log.Logger

	ConnectionTimeout time.Duration
	RPCTimeout        time.Duration
	MetadataHeaders   map[string]string
}
