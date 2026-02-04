package statusline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// VersionCollector reads the MoAI version from the config file.
// It implements the UpdateProvider interface for version display only.
type VersionCollector struct {
	mu     sync.RWMutex
	cached string
}

// VersionConfig represents the structure of .moai/config/config.yaml
// for parsing the version field.
type VersionConfig struct {
	Moai struct {
		Version string `yaml:"version"`
	} `yaml:"moai"`
}

// NewVersionCollector creates a VersionCollector that reads version
// from .moai/config/config.yaml. If the config file doesn't exist or
// version is not set, it returns empty string (not an error).
func NewVersionCollector() *VersionCollector {
	return &VersionCollector{}
}

// CheckUpdate reads the version from the config file and returns it.
// It implements the UpdateProvider interface but only returns the current
// version - no update checking is performed.
func (v *VersionCollector) CheckUpdate(_ context.Context) (*VersionData, error) {
	// Check cache first
	v.mu.RLock()
	if v.cached != "" {
		version := v.cached
		v.mu.RUnlock()
		return &VersionData{
			Current:   formatVersion(version),
			Available: true,
		}, nil
	}
	v.mu.RUnlock()

	// Find and read config file
	version, err := v.readVersionFromConfig()
	if err != nil {
		return &VersionData{Available: false}, nil
	}

	// Update cache
	v.mu.Lock()
	v.cached = version
	v.mu.Unlock()

	return &VersionData{
		Current:   formatVersion(version),
		Available: true,
	}, nil
}

// readVersionFromConfig searches for .moai/config/config.yaml starting
// from the current directory and working upward to find the project root.
func (v *VersionCollector) readVersionFromConfig() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Search upward for .moai/config/config.yaml
	for {
		configPath := filepath.Join(dir, ".moai", "config", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			// Found config file
			return v.parseConfigFile(configPath)
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config file not found")
}

// parseConfigFile reads and parses the config file to extract the version.
func (v *VersionCollector) parseConfigFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read config file: %w", err)
	}

	var cfg VersionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse config file: %w", err)
	}

	if cfg.Moai.Version == "" {
		return "", fmt.Errorf("version not set in config")
	}

	return cfg.Moai.Version, nil
}

// formatVersion removes the 'v' prefix from version strings if present.
// e.g., "v1.14.0" -> "1.14.0", "1.14.0" -> "1.14.0"
func formatVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}
