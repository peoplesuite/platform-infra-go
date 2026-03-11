package neo4j

import (
	"testing"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

func TestSessionConfig(t *testing.T) {
	cfg := Config{
		URI:         "bolt://localhost:7687",
		Username:    "neo4j",
		Password:    "secret",
		Database:    "mydb",
		MaxConnPool: 50,
		ConnAcquire: 30 * time.Second,
		MaxConnLife: time.Hour,
		LogLevel:    "warn",
	}
	sc := SessionConfig(cfg)
	assert.Equal(t, "mydb", sc.DatabaseName)
}

// makeRecord builds a driver Record from keys and values for unit tests.
func makeRecord(keys []string, values []any) *driver.Record {
	return &driver.Record{Keys: keys, Values: values}
}

func TestStr(t *testing.T) {
	t.Run("present string", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{"hello"})
		assert.Equal(t, "hello", Str(rec, "s"))
	})
	t.Run("nil value", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{nil})
		assert.Equal(t, "", Str(rec, "s"))
	})
	t.Run("non-string coerced via Sprintf", func(t *testing.T) {
		rec := makeRecord([]string{"n"}, []any{42})
		assert.Equal(t, "42", Str(rec, "n"))
	})
	t.Run("missing key", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{"x"})
		assert.Equal(t, "", Str(rec, "missing"))
	})
}

func TestInt(t *testing.T) {
	t.Run("int64", func(t *testing.T) {
		rec := makeRecord([]string{"n"}, []any{int64(42)})
		assert.Equal(t, 42, Int(rec, "n"))
	})
	t.Run("float64", func(t *testing.T) {
		rec := makeRecord([]string{"n"}, []any{float64(99)})
		assert.Equal(t, 99, Int(rec, "n"))
	})
	t.Run("int", func(t *testing.T) {
		rec := makeRecord([]string{"n"}, []any{7})
		assert.Equal(t, 7, Int(rec, "n"))
	})
	t.Run("nil", func(t *testing.T) {
		rec := makeRecord([]string{"n"}, []any{nil})
		assert.Equal(t, 0, Int(rec, "n"))
	})
	t.Run("other type returns 0", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{"not a number"})
		assert.Equal(t, 0, Int(rec, "s"))
	})
	t.Run("missing key", func(t *testing.T) {
		rec := makeRecord([]string{"n"}, []any{1})
		assert.Equal(t, 0, Int(rec, "missing"))
	})
}

func TestBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		rec := makeRecord([]string{"b"}, []any{true})
		assert.True(t, Bool(rec, "b"))
	})
	t.Run("false", func(t *testing.T) {
		rec := makeRecord([]string{"b"}, []any{false})
		assert.False(t, Bool(rec, "b"))
	})
	t.Run("nil", func(t *testing.T) {
		rec := makeRecord([]string{"b"}, []any{nil})
		assert.False(t, Bool(rec, "b"))
	})
	t.Run("non-bool", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{"true"})
		assert.False(t, Bool(rec, "s"))
	})
	t.Run("missing key", func(t *testing.T) {
		rec := makeRecord([]string{"b"}, []any{true})
		assert.False(t, Bool(rec, "missing"))
	})
}

func TestStrSlice(t *testing.T) {
	t.Run("[]any of strings", func(t *testing.T) {
		rec := makeRecord([]string{"arr"}, []any{[]any{"a", "b", "c"}})
		assert.Equal(t, []string{"a", "b", "c"}, StrSlice(rec, "arr"))
	})
	t.Run("nil", func(t *testing.T) {
		rec := makeRecord([]string{"arr"}, []any{nil})
		assert.Nil(t, StrSlice(rec, "arr"))
	})
	t.Run("non-slice", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{"not a slice"})
		assert.Nil(t, StrSlice(rec, "s"))
	})
	t.Run("mixed types only strings appended", func(t *testing.T) {
		rec := makeRecord([]string{"arr"}, []any{[]any{"a", 42, "b"}})
		assert.Equal(t, []string{"a", "b"}, StrSlice(rec, "arr"))
	})
	t.Run("missing key", func(t *testing.T) {
		rec := makeRecord([]string{"arr"}, []any{[]any{"x"}})
		assert.Nil(t, StrSlice(rec, "missing"))
	})
}

func TestMapVal(t *testing.T) {
	t.Run("map[string]any", func(t *testing.T) {
		m := map[string]any{"x": 1, "y": "two"}
		rec := makeRecord([]string{"m"}, []any{m})
		assert.Equal(t, m, MapVal(rec, "m"))
	})
	t.Run("nil", func(t *testing.T) {
		rec := makeRecord([]string{"m"}, []any{nil})
		assert.Nil(t, MapVal(rec, "m"))
	})
	t.Run("non-map", func(t *testing.T) {
		rec := makeRecord([]string{"s"}, []any{"not a map"})
		assert.Nil(t, MapVal(rec, "s"))
	})
	t.Run("missing key", func(t *testing.T) {
		rec := makeRecord([]string{"m"}, []any{map[string]any{"k": 1}})
		assert.Nil(t, MapVal(rec, "missing"))
	})
}
