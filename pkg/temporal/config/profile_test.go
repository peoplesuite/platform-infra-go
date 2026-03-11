package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTemporalProfile_WithTempFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	err := os.WriteFile(configPath, []byte(`
[profile.default]
address = "localhost:7233"
namespace = "default"
identity = "test-worker"
rpc_timeout = "10s"
task_queue = "my-queue"
`), 0644)
	require.NoError(t, err)

	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	cfg, err := LoadTemporalProfile("default")
	require.NoError(t, err)
	assert.Equal(t, "localhost:7233", cfg.Address)
	assert.Equal(t, "default", cfg.Namespace)
	assert.Equal(t, "test-worker", cfg.Identity)
	assert.Equal(t, 10*time.Second, cfg.RPCTimeout)
	assert.Equal(t, "my-queue", cfg.TaskQueue)
	assert.NotNil(t, cfg.MetadataHeaders)
}

func TestLoadTemporalProfile_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	err := os.WriteFile(configPath, []byte(`
[profile.default]
address = "original:7233"
namespace = "original"
`), 0644)
	require.NoError(t, err)

	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevAddr := os.Getenv("TEMPORAL_ADDRESS")
	prevNs := os.Getenv("TEMPORAL_NAMESPACE")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_ADDRESS", prevAddr)
		_ = os.Setenv("TEMPORAL_NAMESPACE", prevNs)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_ADDRESS", "env-addr:7233"))
	require.NoError(t, os.Setenv("TEMPORAL_NAMESPACE", "env-ns"))

	cfg, err := LoadTemporalProfile("default")
	require.NoError(t, err)
	assert.Equal(t, "env-addr:7233", cfg.Address)
	assert.Equal(t, "env-ns", cfg.Namespace)
}

func TestLoadTemporalProfile_MissingAddress(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	err := os.WriteFile(configPath, []byte(`
[profile.empty]
namespace = "default"
`), 0644)
	require.NoError(t, err)

	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	_, err = LoadTemporalProfile("empty")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing address")
}

func TestLoadTemporalWorkerConfig_WithTempFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	err := os.WriteFile(configPath, []byte(`
[profile.default]
address = "localhost:7233"
namespace = "default"

[temporal.worker]
max_concurrent_workflow_tasks = 5
max_concurrent_activities = 15
`), 0644)
	require.NoError(t, err)

	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, 5, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, 15, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalWorkerConfig_NoFile_ReturnsDefaults(t *testing.T) {
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Unsetenv("TEMPORAL_CONFIG_FILE"))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, defaultMaxConcurrentWorkflowTasks, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, defaultMaxConcurrentActivities, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalProfile_ResolveConfigFile_EnvSet(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("[profile.default]\naddress = \"x\"\n"), 0644))
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	cfg, err := LoadTemporalProfile("default")
	require.NoError(t, err)
	assert.Equal(t, "x", cfg.Address)
}

func TestLoadTemporalProfile_ResolveConfigFile_NoFile_Error(t *testing.T) {
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Unsetenv("TEMPORAL_CONFIG_FILE"))
	// Run from temp dir with no temporal.toml
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWd) })
	require.NoError(t, os.Chdir(dir))

	_, err := LoadTemporalProfile("default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TEMPORAL_CONFIG_FILE not set")
}

func TestLoadTemporalProfile_EmptyProfile_DefaultsToDefault(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[profile.default]
address = "localhost:7233"
namespace = "default"
`), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevProfile := os.Getenv("TEMPORAL_PROFILE")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_PROFILE", prevProfile)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Unsetenv("TEMPORAL_PROFILE"))

	cfg, err := LoadTemporalProfile("")
	require.NoError(t, err)
	assert.Equal(t, "localhost:7233", cfg.Address)
}

func TestLoadTemporalProfile_EmptyProfile_UsesEnv(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[profile.default]
address = "a"
[profile.other]
address = "b"
namespace = "default"
`), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevProfile := os.Getenv("TEMPORAL_PROFILE")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_PROFILE", prevProfile)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_PROFILE", "other"))

	cfg, err := LoadTemporalProfile("")
	require.NoError(t, err)
	assert.Equal(t, "b", cfg.Address)
}

func TestLoadTemporalProfile_InvalidProfileKey(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("[profile.default]\naddress = \"x\"\n"), 0644))
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	_, err := LoadTemporalProfile("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestLoadTemporalProfile_ReadInConfigError(t *testing.T) {
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.toml")))

	_, err := LoadTemporalProfile("default")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing temporal.toml")
}

func TestLoadTemporalProfile_EnvOverrides_Identity_TaskQueue_RPCTimeout(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[profile.default]
address = "localhost:7233"
namespace = "default"
identity = "file-identity"
rpc_timeout = "5s"
task_queue = "file-queue"
`), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevIdentity := os.Getenv("TEMPORAL_IDENTITY")
	prevTaskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	prevRPC := os.Getenv("TEMPORAL_RPC_TIMEOUT")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_IDENTITY", prevIdentity)
		_ = os.Setenv("TEMPORAL_TASK_QUEUE", prevTaskQueue)
		_ = os.Setenv("TEMPORAL_RPC_TIMEOUT", prevRPC)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_IDENTITY", "env-identity"))
	require.NoError(t, os.Setenv("TEMPORAL_TASK_QUEUE", "env-queue"))
	require.NoError(t, os.Setenv("TEMPORAL_RPC_TIMEOUT", "20s"))

	cfg, err := LoadTemporalProfile("default")
	require.NoError(t, err)
	assert.Equal(t, "env-identity", cfg.Identity)
	assert.Equal(t, "env-queue", cfg.TaskQueue)
	assert.Equal(t, 20*time.Second, cfg.RPCTimeout)
}

func TestLoadTemporalProfile_InvalidRPCTimeoutEnv_NoOverride(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[profile.default]
address = "localhost:7233"
namespace = "default"
rpc_timeout = "10s"
`), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevRPC := os.Getenv("TEMPORAL_RPC_TIMEOUT")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_RPC_TIMEOUT", prevRPC)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_RPC_TIMEOUT", "not-a-duration"))

	cfg, err := LoadTemporalProfile("default")
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.RPCTimeout)
}

func TestLoadTemporalWorkerConfig_FileWithNoSection_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("[profile.default]\naddress = \"x\"\n"), 0644))
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, defaultMaxConcurrentWorkflowTasks, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, defaultMaxConcurrentActivities, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalWorkerConfig_ReadInConfigError_ReturnsDefaults(t *testing.T) {
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", filepath.Join(t.TempDir(), "nonexistent.toml")))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, defaultMaxConcurrentWorkflowTasks, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, defaultMaxConcurrentActivities, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalWorkerConfig_UnmarshalKeyError_ReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("[temporal.worker]\nmax_concurrent_workflow_tasks = \"not_a_number\"\nmax_concurrent_activities = 20\n"), 0644))
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, defaultMaxConcurrentWorkflowTasks, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, defaultMaxConcurrentActivities, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalWorkerConfig_ZeroValues_Defaulted(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[temporal.worker]
max_concurrent_workflow_tasks = 0
max_concurrent_activities = -1
`), 0644))
	prev := os.Getenv("TEMPORAL_CONFIG_FILE")
	t.Cleanup(func() { _ = os.Setenv("TEMPORAL_CONFIG_FILE", prev) })
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, defaultMaxConcurrentWorkflowTasks, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, defaultMaxConcurrentActivities, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalWorkerConfig_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[temporal.worker]
max_concurrent_workflow_tasks = 2
max_concurrent_activities = 3
`), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevWf := os.Getenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS")
	prevAct := os.Getenv("TEMPORAL_WORKER_MAX_CONCURRENT_ACTIVITIES")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS", prevWf)
		_ = os.Setenv("TEMPORAL_WORKER_MAX_CONCURRENT_ACTIVITIES", prevAct)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS", "7"))
	require.NoError(t, os.Setenv("TEMPORAL_WORKER_MAX_CONCURRENT_ACTIVITIES", "8"))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, 7, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, 8, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalWorkerConfig_EnvInvalid_NoChange(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[temporal.worker]
max_concurrent_workflow_tasks = 5
max_concurrent_activities = 10
`), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevWf := os.Getenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS", prevWf)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS", "x"))

	cfg, err := LoadTemporalWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, 5, cfg.MaxConcurrentWorkflowTasks)
	assert.Equal(t, 10, cfg.MaxConcurrentActivities)
}

func TestLoadTemporalConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "temporal.toml")
	require.NoError(t, os.WriteFile(configPath, []byte("[profile.default]\naddress = \"local\"\nnamespace = \"ns\"\n"), 0644))
	prevFile := os.Getenv("TEMPORAL_CONFIG_FILE")
	prevProfile := os.Getenv("TEMPORAL_PROFILE")
	t.Cleanup(func() {
		_ = os.Setenv("TEMPORAL_CONFIG_FILE", prevFile)
		_ = os.Setenv("TEMPORAL_PROFILE", prevProfile)
	})
	require.NoError(t, os.Setenv("TEMPORAL_CONFIG_FILE", configPath))
	require.NoError(t, os.Setenv("TEMPORAL_PROFILE", "default"))

	cfg, err := LoadTemporalConfig()
	require.NoError(t, err)
	assert.Equal(t, "local", cfg.Address)
	assert.Equal(t, "ns", cfg.Namespace)
}
