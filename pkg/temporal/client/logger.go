package client

import (
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

// ZapTemporalLogger adapts zap.Logger to Temporal's log.Logger.
type ZapTemporalLogger struct {
	z *zap.Logger
}

// NewZapLogger returns a Temporal logger that writes to the given zap.Logger.
func NewZapLogger(logger *zap.Logger) *ZapTemporalLogger {
	return &ZapTemporalLogger{z: logger}
}

// Debug logs at debug level.
func (l *ZapTemporalLogger) Debug(msg string, keyvals ...interface{}) {
	l.z.Sugar().Debugw(msg, keyvals...)
}

// Info logs at info level.
func (l *ZapTemporalLogger) Info(msg string, keyvals ...interface{}) {
	l.z.Sugar().Infow(msg, keyvals...)
}

// Warn logs at warn level.
func (l *ZapTemporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.z.Sugar().Warnw(msg, keyvals...)
}

// Error logs at error level.
func (l *ZapTemporalLogger) Error(msg string, keyvals ...interface{}) {
	l.z.Sugar().Errorw(msg, keyvals...)
}

var _ log.Logger = (*ZapTemporalLogger)(nil)
