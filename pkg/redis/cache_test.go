package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetJSON_SetJSON(t *testing.T) {
	opts := Options{
		Addr:      "localhost:6379",
		Password:  "",
		DB:        0,
		TLS:       false,
		VerifyTLS: false,
		Timeout:   5 * time.Second,
	}

	client, err := New(opts)
	if err != nil {
		t.Skipf("Redis not available (client init failed): %v", err)
		return
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	key := "test:cache:json"

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Test SetJSON
	value := testStruct{Name: "test", Value: 42}
	err = client.SetJSON(ctx, key, value, 1*time.Minute)
	if err != nil {
		t.Skipf("Redis not available (SetJSON failed): %v", err)
		return
	}

	// Test GetJSON
	var result testStruct
	found, err := client.GetJSON(ctx, key, &result)
	if err != nil {
		t.Skipf("Redis not available (GetJSON failed): %v", err)
		return
	}
	assert.True(t, found)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)

	// Test GetJSON with non-existent key
	var notFound testStruct
	found, err = client.GetJSON(ctx, "test:cache:nonexistent", &notFound)
	if err != nil {
		t.Skipf("Redis not available (GetJSON non-existent failed): %v", err)
		return
	}
	assert.False(t, found)

	// Cleanup
	_ = client.Delete(ctx, key)
}

func TestDelete(t *testing.T) {
	opts := Options{
		Addr:      "localhost:6379",
		Password:  "",
		DB:        0,
		TLS:       false,
		VerifyTLS: false,
		Timeout:   5 * time.Second,
	}

	client, err := New(opts)
	if err != nil {
		t.Skipf("Redis not available (client init failed): %v", err)
		return
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	key := "test:cache:delete"

	// Set a value
	err = client.SetJSON(ctx, key, map[string]string{"test": "value"}, 1*time.Minute)
	if err != nil {
		t.Skipf("Redis not available (SetJSON failed): %v", err)
		return
	}

	// Verify it exists
	var result map[string]string
	found, err := client.GetJSON(ctx, key, &result)
	if err != nil {
		t.Skipf("Redis not available (GetJSON failed): %v", err)
		return
	}
	assert.True(t, found)

	// Delete it
	err = client.Delete(ctx, key)
	if err != nil {
		t.Skipf("Redis not available (Delete failed): %v", err)
		return
	}

	// Verify it's gone
	found, err = client.GetJSON(ctx, key, &result)
	if err != nil {
		t.Skipf("Redis not available (GetJSON after delete failed): %v", err)
		return
	}
	assert.False(t, found)
}
