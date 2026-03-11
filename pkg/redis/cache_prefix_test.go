package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Get_Set_Delete(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	cache := NewCache(client, "prefix:")

	ctx := context.Background()
	key := "mykey"
	type V struct{ Name string }

	// Set and Get
	err := cache.Set(ctx, key, &V{Name: "alice"}, time.Minute)
	require.NoError(t, err)

	var dest V
	found, err := cache.Get(ctx, key, &dest)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "alice", dest.Name)

	// Delete
	err = cache.Delete(ctx, key)
	require.NoError(t, err)

	found, err = cache.Get(ctx, key, &dest)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestCache_DeletePattern(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	cache := NewCache(client, "ns:")

	ctx := context.Background()
	require.NoError(t, cache.Set(ctx, "user:1", map[string]int{"a": 1}, time.Minute))
	require.NoError(t, cache.Set(ctx, "user:2", map[string]int{"b": 2}, time.Minute))
	require.NoError(t, cache.Set(ctx, "other", map[string]int{"c": 3}, time.Minute))

	err := cache.DeletePattern(ctx, "user:*")
	require.NoError(t, err)

	var v map[string]int
	found1, _ := cache.Get(ctx, "user:1", &v)
	assert.False(t, found1)
	found2, _ := cache.Get(ctx, "user:2", &v)
	assert.False(t, found2)
	found3, _ := cache.Get(ctx, "other", &v)
	assert.True(t, found3)
}

func TestCache_GetOrLoad(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	cache := NewCache(client, "cache:")

	ctx := context.Background()
	key := "item"
	type V struct{ N int }

	called := false
	loadFn := func() (any, error) {
		called = true
		return &V{N: 99}, nil
	}

	var dest V
	err := cache.GetOrLoad(ctx, key, &dest, time.Minute, loadFn)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, 99, dest.N)

	// Second call uses cache
	called = false
	var dest2 V
	err = cache.GetOrLoad(ctx, key, &dest2, time.Minute, loadFn)
	require.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, 99, dest2.N)
}

func TestCache_GetOrLoad_FnError(t *testing.T) {
	client, _ := newTestClientMiniredis(t)
	cache := NewCache(client, "cache:")

	ctx := context.Background()
	loadErr := errors.New("load failed")
	loadFn := func() (any, error) { return nil, loadErr }

	var dest struct{ X int }
	err := cache.GetOrLoad(ctx, "key", &dest, time.Minute, loadFn)
	require.Error(t, err)
	assert.ErrorIs(t, err, loadErr)
}
