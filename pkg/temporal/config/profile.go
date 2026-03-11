package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// TemporalProfile represents a Temporal connection profile from temporal.toml
type TemporalProfile struct {
	Address         string            `mapstructure:"address"`
	Namespace       string            `mapstructure:"namespace"`
	Identity        string            `mapstructure:"identity"`
	RPCTimeout      time.Duration     `mapstructure:"rpc_timeout"`
	TaskQueue       string            `mapstructure:"task_queue"`
	MetadataHeaders map[string]string `mapstructure:"metadata_headers"`
}

// resolveTemporalConfigFile returns the path to temporal.toml.
// Uses TEMPORAL_CONFIG_FILE if set; otherwise looks for configs/temporal.toml then temporal.toml.
func resolveTemporalConfigFile() (string, error) {
	if configFile := os.Getenv("TEMPORAL_CONFIG_FILE"); configFile != "" {
		return configFile, nil
	}
	for _, path := range []string{"configs/temporal.toml", "temporal.toml"} {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath, nil
			}
		}
	}
	return "", fmt.Errorf("TEMPORAL_CONFIG_FILE not set and temporal.toml not found in standard locations (configs/temporal.toml, temporal.toml)")
}

// LoadTemporalProfile loads a Temporal profile from temporal.toml
// Profile is selected via TEMPORAL_PROFILE env var, defaults to "default"
// Config file location can be specified via TEMPORAL_CONFIG_FILE env var
// Environment variables can override: TEMPORAL_ADDRESS, TEMPORAL_NAMESPACE, TEMPORAL_IDENTITY, TEMPORAL_TASK_QUEUE
func LoadTemporalProfile(profile string) (TemporalProfile, error) {
	configFile, err := resolveTemporalConfigFile()
	if err != nil {
		return TemporalProfile{}, err
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		return TemporalProfile{}, fmt.Errorf("missing temporal.toml at %s: %w", configFile, err)
	}

	// Fallback to default if profile is empty
	if profile == "" {
		profile = os.Getenv("TEMPORAL_PROFILE")
		if profile == "" {
			profile = "default"
		}
	}

	key := fmt.Sprintf("profile.%s", profile)

	var cfg TemporalProfile
	if err := v.UnmarshalKey(key, &cfg); err != nil {
		return TemporalProfile{}, fmt.Errorf("profile '%s' invalid: %w", profile, err)
	}

	// Ensure metadata_headers is initialized if empty
	if cfg.MetadataHeaders == nil {
		cfg.MetadataHeaders = make(map[string]string)
	}

	// Apply environment variable overrides (env vars always win)
	if addr := os.Getenv("TEMPORAL_ADDRESS"); addr != "" {
		cfg.Address = addr
	}
	if ns := os.Getenv("TEMPORAL_NAMESPACE"); ns != "" {
		cfg.Namespace = ns
	}
	if identity := os.Getenv("TEMPORAL_IDENTITY"); identity != "" {
		cfg.Identity = identity
	}
	if timeoutStr := os.Getenv("TEMPORAL_RPC_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			cfg.RPCTimeout = timeout
		}
	}
	if taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE"); taskQueue != "" {
		cfg.TaskQueue = taskQueue
	}

	if cfg.Address == "" {
		return TemporalProfile{}, fmt.Errorf("profile '%s' missing address", profile)
	}

	return cfg, nil
}

// LoadTemporalConfig loads Temporal profile using TEMPORAL_PROFILE env var
func LoadTemporalConfig() (TemporalProfile, error) {
	return LoadTemporalProfile("")
}

// TemporalWorkerConfig is worker runtime capacity from temporal.toml [temporal.worker].
// It controls how many tasks the process executes simultaneously (worker concern, not workflow semantics).
type TemporalWorkerConfig struct {
	MaxConcurrentWorkflowTasks int `mapstructure:"max_concurrent_workflow_tasks"`
	MaxConcurrentActivities    int `mapstructure:"max_concurrent_activities"`
}

const (
	defaultMaxConcurrentWorkflowTasks = 10
	defaultMaxConcurrentActivities    = 20
)

// LoadTemporalWorkerConfig loads [temporal.worker] from the same temporal.toml used for profiles.
// Uses TEMPORAL_CONFIG_FILE or standard locations. If the section is missing, returns sensible defaults.
// Env overrides: TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS, TEMPORAL_WORKER_MAX_CONCURRENT_ACTIVITIES.
func LoadTemporalWorkerConfig() (TemporalWorkerConfig, error) {
	configFile, err := resolveTemporalConfigFile()
	if err != nil {
		return TemporalWorkerConfig{
			MaxConcurrentWorkflowTasks: defaultMaxConcurrentWorkflowTasks,
			MaxConcurrentActivities:    defaultMaxConcurrentActivities,
		}, nil
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return TemporalWorkerConfig{
			MaxConcurrentWorkflowTasks: defaultMaxConcurrentWorkflowTasks,
			MaxConcurrentActivities:    defaultMaxConcurrentActivities,
		}, nil
	}

	var cfg TemporalWorkerConfig
	if err := v.UnmarshalKey("temporal.worker", &cfg); err != nil {
		return TemporalWorkerConfig{
			MaxConcurrentWorkflowTasks: defaultMaxConcurrentWorkflowTasks,
			MaxConcurrentActivities:    defaultMaxConcurrentActivities,
		}, nil
	}

	if cfg.MaxConcurrentWorkflowTasks <= 0 {
		cfg.MaxConcurrentWorkflowTasks = defaultMaxConcurrentWorkflowTasks
	}
	if cfg.MaxConcurrentActivities <= 0 {
		cfg.MaxConcurrentActivities = defaultMaxConcurrentActivities
	}

	if n := os.Getenv("TEMPORAL_WORKER_MAX_CONCURRENT_WORKFLOW_TASKS"); n != "" {
		if v, err := parseInt(n); err == nil && v > 0 {
			cfg.MaxConcurrentWorkflowTasks = v
		}
	}
	if n := os.Getenv("TEMPORAL_WORKER_MAX_CONCURRENT_ACTIVITIES"); n != "" {
		if v, err := parseInt(n); err == nil && v > 0 {
			cfg.MaxConcurrentActivities = v
		}
	}

	return cfg, nil
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
