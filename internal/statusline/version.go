package statusline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// VersionCollector reads the MoAI version from the config file and
// compares it against the running binary version to detect when
// templates need syncing via `moai update`.
type VersionCollector struct {
	mu            sync.RWMutex
	cached        string
	binaryVersion string // running binary version (from pkg/version)
}

// VersionConfig represents the structure of .moai/config/sections/system.yaml
// for parsing version fields. It reads moai.template_version first
// (updated by moai update) and falls back to moai.version (set during init).
type VersionConfig struct {
	Moai struct {
		Version         string `yaml:"version"`
		TemplateVersion string `yaml:"template_version"`
	} `yaml:"moai"`
}

// effectiveVersion returns moai.template_version if set, otherwise
// moai.version. This ensures the statusline reflects the version updated
// by `moai update` rather than the original initialization version.
func (c *VersionConfig) effectiveVersion() string {
	if c.Moai.TemplateVersion != "" {
		return c.Moai.TemplateVersion
	}
	return c.Moai.Version
}

// NewVersionCollector creates a VersionCollector that reads the template
// version from .moai/config/sections/system.yaml and compares it against the
// running binary version. If binaryVersion is empty, no update check
// is performed.
func NewVersionCollector(binaryVersion string) *VersionCollector {
	return &VersionCollector{binaryVersion: binaryVersion}
}

// CheckUpdate reads the template version from the config file and compares
// it against the running binary version. If the binary is newer than the
// templates, UpdateAvailable is set to true with the binary version in Latest.
// When the config file is missing or has no version, the binary version is
// used as the display version so the statusline always shows something.
func (v *VersionCollector) CheckUpdate(_ context.Context) (*VersionData, error) {
	// Check cache first
	v.mu.RLock()
	if v.cached != "" {
		version := v.cached
		v.mu.RUnlock()
		return v.buildVersionData(version), nil
	}
	v.mu.RUnlock()

	// Find and read config file
	version, err := v.readVersionFromConfig()
	if err != nil {
		// Fallback: use binary version when config is unavailable
		if v.binaryVersion != "" {
			return &VersionData{
				Current:   formatVersion(v.binaryVersion),
				Available: true,
			}, nil
		}
		return &VersionData{Available: false}, nil
	}

	// Update cache
	v.mu.Lock()
	v.cached = version
	v.mu.Unlock()

	return v.buildVersionData(version), nil
}

// buildVersionData constructs VersionData from the template version,
// comparing against the binary version to detect available updates.
func (v *VersionCollector) buildVersionData(templateVersion string) *VersionData {
	data := &VersionData{
		Current:   formatVersion(templateVersion),
		Available: true,
	}

	// Compare binary version vs template version
	if v.binaryVersion != "" {
		bv := formatVersion(v.binaryVersion)
		tv := formatVersion(templateVersion)
		if bv != tv && bv != "" && tv != "" {
			data.Latest = bv
			data.UpdateAvailable = true
		}
	}

	return data
}

// readVersionFromConfig searches for .moai/config/sections/system.yaml starting
// from the current directory and working upward to find the project root.
func (v *VersionCollector) readVersionFromConfig() (string, error) {
	// Start from current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Search upward for .moai/config/sections/system.yaml
	for {
		configPath := filepath.Join(dir, defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML)
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

	version := cfg.effectiveVersion()
	if version == "" {
		return "", fmt.Errorf("version not set in config")
	}

	return version, nil
}

// formatVersion removes the 'v' prefix from version strings if present.
// e.g., "v1.14.0" -> "1.14.0", "1.14.0" -> "1.14.0"
func formatVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}
