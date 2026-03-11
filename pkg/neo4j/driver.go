package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

// Config holds Neo4j connection and pool settings.
type Config struct {
	URI         string        `envconfig:"NEO4J_URI" default:"bolt://localhost:7687"`
	Username    string        `envconfig:"NEO4J_USERNAME" default:"neo4j"`
	Password    string        `envconfig:"NEO4J_PASSWORD" required:"true"`
	Database    string        `envconfig:"NEO4J_DATABASE" default:"neo4j"`
	MaxConnPool int           `envconfig:"NEO4J_MAX_CONN_POOL" default:"50"`
	ConnAcquire time.Duration `envconfig:"NEO4J_CONN_ACQUIRE_TIMEOUT" default:"30s"`
	MaxConnLife time.Duration `envconfig:"NEO4J_MAX_CONN_LIFETIME" default:"1h"`
	LogLevel    string        `envconfig:"NEO4J_LOG_LEVEL" default:"warn"` // debug, info, warn, error
}

// NewDriver creates a Neo4j driver, verifies connectivity, and returns it.
// Caller is responsible for closing: defer driver.Close(ctx)
func NewDriver(ctx context.Context, cfg Config, logger *zap.Logger) (neo4j.DriverWithContext, error) {
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
		func(c *neo4j.Config) { //nolint:staticcheck // SA1019: neo4j.Config deprecated in v6; migrate when upgrading driver
			c.MaxConnectionPoolSize = cfg.MaxConnPool
			c.ConnectionAcquisitionTimeout = cfg.ConnAcquire
			c.MaxConnectionLifetime = cfg.MaxConnLife
			c.Log = &zapNeo4jLogger{z: logger.Named("neo4j-driver"), level: cfg.LogLevel}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("neo4j new driver: %w", err)
	}

	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("neo4j verify connectivity: %w", err)
	}

	logger.Info("neo4j driver connected",
		zap.String("uri", cfg.URI),
		zap.String("database", cfg.Database),
		zap.Int("maxPool", cfg.MaxConnPool),
	)

	return driver, nil
}

type zapNeo4jLogger struct {
	z     *zap.Logger
	level string
}

func (l *zapNeo4jLogger) Error(name string, id string, err error) {
	l.z.Error(name, zap.String("id", id), zap.Error(err))
}

func (l *zapNeo4jLogger) Warn(name string, id string, err error) {
	l.z.Warn(name, zap.String("id", id), zap.Error(err))
}

func (l *zapNeo4jLogger) Warnf(name string, id string, msg string, args ...any) {
	if l.level == "debug" || l.level == "info" || l.level == "warn" {
		l.z.Warn(fmt.Sprintf(msg, args...), zap.String("name", name), zap.String("id", id))
	}
}

func (l *zapNeo4jLogger) Infof(name string, id string, msg string, args ...any) {
	if l.level == "debug" || l.level == "info" {
		l.z.Info(fmt.Sprintf(msg, args...), zap.String("name", name), zap.String("id", id))
	}
}

func (l *zapNeo4jLogger) Debugf(name string, id string, msg string, args ...any) {
	if l.level == "debug" {
		l.z.Debug(fmt.Sprintf(msg, args...), zap.String("name", name), zap.String("id", id))
	}
}
