package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel"
)

// SessionConfig returns a session config targeting the configured database.
func SessionConfig(cfg Config) neo4j.SessionConfig {
	return neo4j.SessionConfig{
		DatabaseName: cfg.Database,
	}
}

// ReadSession opens a session and returns it. Caller must close it.
func ReadSession(ctx context.Context, driver neo4j.DriverWithContext, cfg Config) neo4j.SessionWithContext {
	return driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: cfg.Database,
		AccessMode:   neo4j.AccessModeRead,
	})
}

// WriteSession opens a write session. Caller must close it.
func WriteSession(ctx context.Context, driver neo4j.DriverWithContext, cfg Config) neo4j.SessionWithContext {
	return driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: cfg.Database,
		AccessMode:   neo4j.AccessModeWrite,
	})
}

// Str safely extracts a string from a neo4j.Record by key.
func Str(rec *neo4j.Record, key string) string {
	v, _ := rec.Get(key)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// Int safely extracts an int from a neo4j.Record by key.
func Int(rec *neo4j.Record, key string) int {
	v, _ := rec.Get(key)
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return int(n)
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

// Bool safely extracts a bool from a neo4j.Record by key.
func Bool(rec *neo4j.Record, key string) bool {
	v, _ := rec.Get(key)
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// StrSlice extracts a []string from a neo4j.Record by key.
func StrSlice(rec *neo4j.Record, key string) []string {
	v, _ := rec.Get(key)
	if v == nil {
		return nil
	}
	if arr, ok := v.([]any); ok {
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// MapVal extracts a map[string]any from a neo4j.Record by key.
func MapVal(rec *neo4j.Record, key string) map[string]any {
	v, _ := rec.Get(key)
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

// Query executes a read query and maps each record with scanFn.
func Query[T any](
	ctx context.Context,
	driver neo4j.DriverWithContext,
	cfg Config,
	cypher string,
	params map[string]any,
	scanFn func(*neo4j.Record) (T, error),
) ([]T, error) {
	ctx, span := otel.Tracer("pkg/neo4j").Start(ctx, "neo4j.Query")
	defer span.End()

	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: cfg.Database,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cursor, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		var items []T
		for cursor.Next(ctx) {
			item, err := scanFn(cursor.Record())
			if err != nil {
				return nil, fmt.Errorf("scan row: %w", err)
			}
			items = append(items, item)
		}
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		return items, nil
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	return result.([]T), nil
}

// QuerySingle executes a read query expecting zero or one result.
// Returns (nil, nil) if no rows found.
func QuerySingle[T any](
	ctx context.Context,
	driver neo4j.DriverWithContext,
	cfg Config,
	cypher string,
	params map[string]any,
	scanFn func(*neo4j.Record) (T, error),
) (*T, error) {
	items, err := Query(ctx, driver, cfg, cypher, params, scanFn)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}
