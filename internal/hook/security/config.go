package security

import (
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ExtraSecurityConfig holds user-defined extra security patterns
// that extend (never replace) the built-in DefaultSecurityPolicy.
type ExtraSecurityConfig struct {
	Security struct {
		ExtraDangerousBashPatterns    []string `yaml:"extra_dangerous_bash_patterns"`
		ExtraDenyPatterns             []string `yaml:"extra_deny_patterns"`
		ExtraAskPatterns              []string `yaml:"extra_ask_patterns"`
		ExtraSensitiveContentPatterns []string `yaml:"extra_sensitive_content_patterns"`
	} `yaml:"security"`
}

// LoadExtraSecurityConfig reads security.yaml from the project config directory.
// Returns nil if the file does not exist or cannot be parsed (graceful degradation).
func LoadExtraSecurityConfig(projectDir string) *ExtraSecurityConfig {
	cfgPath := filepath.Join(projectDir, ".moai", "config", "sections", "security.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		slog.Debug("security config not found, using defaults only", "path", cfgPath)
		return nil
	}

	var cfg ExtraSecurityConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		slog.Warn("failed to parse security config, using defaults only", "error", err)
		return nil
	}

	return &cfg
}
