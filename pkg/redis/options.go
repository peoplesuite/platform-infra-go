package redis

import "time"

// Options configures the Redis client (address, auth, TLS, timeouts).
type Options struct {
	Addr      string
	Username  string
	Password  string
	DB        int
	TLS       bool
	VerifyTLS bool
	Timeout   time.Duration
}
