package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("pkg/neo4j")

// ExecuteRead runs a read transaction and returns the result produced by fn.
func ExecuteRead[T any](
	ctx context.Context,
	driver neo4j.DriverWithContext,
	cfg Config,
	fn func(tx neo4j.ManagedTransaction) (T, error),
) (T, error) {

	ctx, span := tracer.Start(ctx, "neo4j.ExecuteRead")
	defer span.End()

	session := ReadSession(ctx, driver, cfg)
	defer func() { _ = session.Close(ctx) }()

	res, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return fn(tx)
	})

	if err != nil {
		span.RecordError(err)
		var zero T
		return zero, err
	}

	return res.(T), nil
}

// ExecuteWrite runs a write transaction and returns the result produced by fn.
func ExecuteWrite[T any](
	ctx context.Context,
	driver neo4j.DriverWithContext,
	cfg Config,
	fn func(tx neo4j.ManagedTransaction) (T, error),
) (T, error) {

	ctx, span := tracer.Start(ctx, "neo4j.ExecuteWrite")
	defer span.End()

	session := WriteSession(ctx, driver, cfg)
	defer func() { _ = session.Close(ctx) }()

	res, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return fn(tx)
	})

	if err != nil {
		span.RecordError(err)
		var zero T
		return zero, err
	}

	return res.(T), nil
}

// Run executes a Cypher query inside a transaction without mapping.
func Run(
	ctx context.Context,
	tx neo4j.ManagedTransaction,
	cypher string,
	params map[string]any,
) error {

	result, err := tx.Run(ctx, cypher, params)
	if err != nil {
		return err
	}

	_, err = result.Consume(ctx)
	return err
}

// RunAndCollect executes a Cypher query and returns all records.
func RunAndCollect(
	ctx context.Context,
	tx neo4j.ManagedTransaction,
	cypher string,
	params map[string]any,
) ([]*neo4j.Record, error) {

	cursor, err := tx.Run(ctx, cypher, params)
	if err != nil {
		return nil, err
	}

	var records []*neo4j.Record

	for cursor.Next(ctx) {
		records = append(records, cursor.Record())
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return records, nil
}
