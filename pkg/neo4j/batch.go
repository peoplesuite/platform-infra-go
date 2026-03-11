package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// DefaultBatchSize is the default number of items per batch for BatchWriter.Write.
const DefaultBatchSize = 500

// BatchWriter executes Cypher statements in batches using UNWIND $batch.
type BatchWriter struct {
	driver    neo4j.DriverWithContext
	sessCfg   neo4j.SessionConfig
	batchSize int
	logger    *zap.Logger
	tracer    trace.Tracer
}

// NewBatchWriter returns a BatchWriter for executing Cypher in batches (UNWIND $batch).
func NewBatchWriter(driver neo4j.DriverWithContext, cfg Config, logger *zap.Logger) *BatchWriter {
	return &BatchWriter{
		driver:    driver,
		sessCfg:   SessionConfig(cfg),
		batchSize: DefaultBatchSize,
		logger:    logger,
		tracer:    otel.Tracer("pkg/neo4j"),
	}
}

// WithBatchSize overrides the default batch size.
func (w *BatchWriter) WithBatchSize(size int) *BatchWriter {
	if size > 0 {
		w.batchSize = size
	}
	return w
}

// Write executes a Cypher query against successive batches of items.
// The query must use UNWIND $batch AS row.
func (w *BatchWriter) Write(ctx context.Context, cypher string, items []map[string]any) error {
	if len(items) == 0 {
		return nil
	}

	ctx, span := w.tracer.Start(ctx, "neo4j.BatchWriter.Write",
		trace.WithAttributes(
			attribute.Int("batch.total_items", len(items)),
			attribute.Int("batch.size", w.batchSize),
		),
	)
	defer span.End()

	for offset := 0; offset < len(items); offset += w.batchSize {
		end := offset + w.batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[offset:end]

		session := w.driver.NewSession(ctx, w.sessCfg)
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, cypher, map[string]any{"batch": batch})
			if err != nil {
				return nil, err
			}
			_, err = result.Consume(ctx)
			return nil, err
		})
		_ = session.Close(ctx)

		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("batch write at offset %d: %w", offset, err)
		}
	}

	return nil
}

// RunSchema executes a list of DDL statements (constraints, indexes).
// Each statement runs independently; failures are collected.
func (w *BatchWriter) RunSchema(ctx context.Context, statements []string) error {
	session := w.driver.NewSession(ctx, w.sessCfg)
	defer func() { _ = session.Close(ctx) }()

	for _, stmt := range statements {
		_, err := session.Run(ctx, stmt, nil)
		if err != nil {
			w.logger.Warn("schema statement failed (may already exist)",
				zap.String("stmt", truncate(stmt, 80)),
				zap.Error(err),
			)
		}
	}
	return nil
}

// DeleteByIDs runs a DETACH DELETE query for a list of integer IDs.
// The query must use UNWIND $ids AS id.
func (w *BatchWriter) DeleteByIDs(ctx context.Context, cypher string, ids []int) error {
	if len(ids) == 0 {
		return nil
	}

	session := w.driver.NewSession(ctx, w.sessCfg)
	defer func() { _ = session.Close(ctx) }()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, cypher, map[string]any{"ids": ids})
		if err != nil {
			return nil, err
		}
		_, err = result.Consume(ctx)
		return nil, err
	})
	return err
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
