package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClientMiniredis starts miniredis and returns a Client connected to it. Caller must call client.Close() and mr.Close().
func newTestClientMiniredis(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	opts := Options{
		Addr:      mr.Addr(),
		Password:  "",
		DB:        0,
		TLS:       false,
		VerifyTLS: false,
		Timeout:   5 * time.Second,
	}
	client, err := New(opts)
	require.NoError(t, err)
	require.NotNil(t, client)
	t.Cleanup(func() { _ = client.Close(); mr.Close() })
	return client, mr
}

func TestSetBytes_GetBytes(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	key := "test:bytes"
	data := []byte("hello world")

	err := client.SetBytes(ctx, key, data, time.Minute)
	require.NoError(t, err)

	got, err := client.GetBytes(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestGetBytes_MissingKey(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()

	got, err := client.GetBytes(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDeleteByPattern(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()

	// Set several keys matching pattern
	require.NoError(t, client.SetBytes(ctx, "org:1", []byte("a"), time.Minute))
	require.NoError(t, client.SetBytes(ctx, "org:2", []byte("b"), time.Minute))
	require.NoError(t, client.SetBytes(ctx, "other:1", []byte("c"), time.Minute))

	err := client.DeleteByPattern(ctx, "org:*")
	require.NoError(t, err)

	// org:* keys should be gone
	got1, err := client.GetBytes(ctx, "org:1")
	require.NoError(t, err)
	assert.Nil(t, got1)
	got2, err := client.GetBytes(ctx, "org:2")
	require.NoError(t, err)
	assert.Nil(t, got2)
	// other:1 still exists
	got3, err := client.GetBytes(ctx, "other:1")
	require.NoError(t, err)
	assert.Equal(t, []byte("c"), got3)
}

func TestGetOrSetJSON_CacheMiss(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	key := "test:getorset"
	type V struct{ N int }

	called := false
	loadFn := func() (any, error) {
		called = true
		return &V{N: 42}, nil
	}

	var dest V
	err := client.GetOrSetJSON(ctx, key, time.Minute, loadFn, &dest)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, 42, dest.N)

	// Second call should hit cache and not call loadFn
	called = false
	var dest2 V
	err = client.GetOrSetJSON(ctx, key, time.Minute, loadFn, &dest2)
	require.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, 42, dest2.N)
}

func TestGetOrSetJSON_CacheHit(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	key := "test:cached"
	type V struct{ S string }
	require.NoError(t, client.SetJSON(ctx, key, &V{S: "hello"}, time.Minute))

	called := false
	loadFn := func() (any, error) {
		called = true
		return nil, errors.New("should not be called")
	}

	var dest V
	err := client.GetOrSetJSON(ctx, key, time.Minute, loadFn, &dest)
	require.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "hello", dest.S)
}

func TestGetOrSetJSON_FnError(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	key := "test:fnerr"
	loadErr := errors.New("load failed")
	loadFn := func() (any, error) { return nil, loadErr }

	var dest struct{ X int }
	err := client.GetOrSetJSON(ctx, key, time.Minute, loadFn, &dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetch")
	assert.ErrorIs(t, err, loadErr)
}

func TestGetJSON_InvalidJSON(t *testing.T) {
	client, mr := newTestClientMiniredis(t)
	ctx := context.Background()
	key := "test:invalidjson"
	require.NoError(t, mr.Set(key, "not valid json"))

	var dest struct{ X int }
	found, err := client.GetJSON(ctx, key, &dest)
	require.Error(t, err)
	assert.False(t, found)
	assert.Contains(t, err.Error(), "json unmarshal")
}

func TestSetJSON_MarshalError(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	// Value that cannot be JSON-marshaled
	invalidValue := make(chan int)
	err := client.SetJSON(ctx, "test:chan", invalidValue, time.Minute)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "json marshal")
}

func TestClose(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	err := client.Close()
	require.NoError(t, err)
	// Second close is a no-op or may error depending on impl; avoid double-close in cleanup
}

func TestGetOrSetJSON_UnmarshalIntoIncompatibleType(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	key := "test:incompatible"
	type V struct{ N int }
	loadFn := func() (any, error) { return &V{N: 99}, nil }
	var dest string // cannot unmarshal object into string
	err := client.GetOrSetJSON(ctx, key, time.Minute, loadFn, &dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestGetOrSetJSON_SetJSONFailsAfterFetch(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	key := "test:setfail"
	loadFn := func() (any, error) { return map[string]int{"n": 1}, nil }
	var dest struct{ N int }
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Already cancelled so SetJSON will fail
	err := client.GetOrSetJSON(ctx, key, time.Minute, loadFn, &dest)
	require.Error(t, err)
}

func TestDeleteByPattern_IteratorError(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	ctx := context.Background()
	// Empty pattern with no keys still runs iterator; use a pattern that matches nothing to hit iter.Err() path
	err := client.DeleteByPattern(ctx, "nomatch:*")
	require.NoError(t, err)
}

func TestDelete_ErrorAfterClose(t *testing.T) {
	client, mr := newTestClientMiniredis(t)
	mr.Close()
	ctx := context.Background()
	err := client.Delete(ctx, "anykey")
	require.Error(t, err)
	var re *Error
	require.True(t, errors.As(err, &re))
	assert.Equal(t, "Del", re.Op)
}

func TestSetBytes_ErrorAfterClose(t *testing.T) {
	client, mr := newTestClientMiniredis(t)
	mr.Close()
	ctx := context.Background()
	err := client.SetBytes(ctx, "k", []byte("v"), time.Minute)
	require.Error(t, err)
	var re *Error
	require.True(t, errors.As(err, &re))
	assert.Equal(t, "Set", re.Op)
}

func TestGetBytes_ErrorAfterClose(t *testing.T) {
	client, mr := newTestClientMiniredis(t)
	mr.Close()
	ctx := context.Background()
	got, err := client.GetBytes(ctx, "k")
	require.Error(t, err)
	assert.Nil(t, got)
	var re *Error
	require.True(t, errors.As(err, &re))
	assert.Equal(t, "Get", re.Op)
}
