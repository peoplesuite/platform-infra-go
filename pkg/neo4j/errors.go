package neo4j

import "fmt"

// Error wraps a Neo4j operation error with additional context.
type Error struct {
	Op  string
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("neo4j %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Wrap wraps an error with operation context.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}

	return &Error{
		Op:  op,
		Err: err,
	}
}
