package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		DSN:             "postgres://localhost:5432/mydb",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
	assert.Equal(t, "postgres://localhost:5432/mydb", cfg.DSN)
	assert.Equal(t, 25, cfg.MaxOpenConns)
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, time.Hour, cfg.ConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, cfg.ConnMaxIdleTime)
}

func TestConfig_ZeroValues(t *testing.T) {
	cfg := Config{DSN: "postgres://local"}
	assert.Equal(t, "postgres://local", cfg.DSN)
	assert.Equal(t, 0, cfg.MaxOpenConns)
	assert.Equal(t, 0, cfg.MaxIdleConns)
	assert.Equal(t, time.Duration(0), cfg.ConnMaxLifetime)
	assert.Equal(t, time.Duration(0), cfg.ConnMaxIdleTime)
}
