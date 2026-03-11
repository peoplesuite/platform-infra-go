package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	assert.Equal(t, "auth:token", Key("auth", "token"))
	assert.Equal(t, "assoc:transformed:all", Key("assoc", "transformed", "all"))
	assert.Equal(t, "single", Key("single"))
	assert.Equal(t, "", Key())
}
