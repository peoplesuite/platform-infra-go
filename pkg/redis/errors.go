package redis

import "fmt"

// Error provides operation context (Op, Key) for Redis failures; use errors.As to unwrap.
type Error struct {
	Op  string
	Key string
	Err error
}

func (e *Error) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("redis %s [key=%s]: %v", e.Op, e.Key, e.Err)
	}
	return fmt.Sprintf("redis %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}
