package neo4j

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewDriver_InvalidURI(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		URI:      "not-a-valid-uri",
		Username: "u",
		Password: "p",
		Database: "d",
	}
	logger := zap.NewNop()
	driver, err := NewDriver(ctx, cfg, logger)
	require.Error(t, err)
	assert.Nil(t, driver)
	assert.Contains(t, err.Error(), "neo4j new driver")
}

func TestNewDriver_VerifyConnectivityFails(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		URI:         "bolt://localhost:17687", // unlikely to be running
		Username:    "neo4j",
		Password:    "password",
		Database:    "neo4j",
		MaxConnPool: 5,
		ConnAcquire: 100 * time.Millisecond,
	}
	logger := zap.NewNop()
	driver, err := NewDriver(ctx, cfg, logger)
	if err != nil {
		assert.True(t, strings.Contains(err.Error(), "neo4j verify connectivity") || strings.Contains(err.Error(), "neo4j new driver"), "err: %s", err.Error())
		return
	}
	if driver != nil {
		_ = driver.Close(ctx)
	}
}

func TestZapNeo4jLogger_Error(t *testing.T) {
	core, observed := observer.New(zap.ErrorLevel)
	logger := zap.New(core)
	l := &zapNeo4jLogger{z: logger, level: "warn"}
	l.Error("conn", "id-1", errors.New("err"))
	logs := observed.All()
	require.Len(t, logs, 1)
	assert.Equal(t, "conn", logs[0].Message)
}

func TestZapNeo4jLogger_Warn(t *testing.T) {
	core, observed := observer.New(zap.WarnLevel)
	logger := zap.New(core)
	l := &zapNeo4jLogger{z: logger, level: "warn"}
	l.Warn("pool", "id-2", errors.New("warn err"))
	logs := observed.All()
	require.Len(t, logs, 1)
}

func TestZapNeo4jLogger_Warnf_Levels(t *testing.T) {
	core, observed := observer.New(zap.WarnLevel)
	logger := zap.New(core)

	for _, level := range []string{"debug", "info", "warn"} {
		observed.TakeAll()
		l := &zapNeo4jLogger{z: logger, level: level}
		l.Warnf("x", "id", "msg %s", "arg")
		logs := observed.All()
		require.Len(t, logs, 1, "level %s", level)
	}

	observed.TakeAll()
	l := &zapNeo4jLogger{z: logger, level: "error"}
	l.Warnf("x", "id", "msg")
	assert.Len(t, observed.All(), 0)
}

func TestZapNeo4jLogger_Infof_Levels(t *testing.T) {
	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	for _, level := range []string{"debug", "info"} {
		observed.TakeAll()
		l := &zapNeo4jLogger{z: logger, level: level}
		l.Infof("x", "id", "info %d", 1)
		logs := observed.All()
		require.Len(t, logs, 1, "level %s", level)
	}

	observed.TakeAll()
	l := &zapNeo4jLogger{z: logger, level: "warn"}
	l.Infof("x", "id", "info")
	assert.Len(t, observed.All(), 0)
}

func TestZapNeo4jLogger_Debugf_Level(t *testing.T) {
	core, observed := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	l := &zapNeo4jLogger{z: logger, level: "debug"}
	l.Debugf("x", "id", "debug %s", "v")
	logs := observed.All()
	require.Len(t, logs, 1)

	observed.TakeAll()
	l2 := &zapNeo4jLogger{z: logger, level: "info"}
	l2.Debugf("x", "id", "debug")
	assert.Len(t, observed.All(), 0)
}
