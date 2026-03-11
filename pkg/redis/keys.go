package redis

import "strings"

// Key joins parts with ":" to build a Redis key (e.g. Key("user", id) => "user:123").
func Key(parts ...string) string {
	return strings.Join(parts, ":")
}
