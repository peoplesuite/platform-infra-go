package postgres

import "fmt"

// Error wraps a Postgres operation failure with Op context; use errors.As to unwrap.
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("postgres %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}
