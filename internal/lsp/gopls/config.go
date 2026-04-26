package gopls

// @MX:ANCHOR: [AUTO] Config type and loader — the settings entry point depended on by bridge.go and NewBridge
// @MX:REASON: fan_in >= 3 (NewBridge, tests, external dependency injection paths)

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the gopls bridge behavior settings.
// REQ-GB-050: loaded from the gopls_bridge section of .moai/config/sections/lsp.yaml.
// REQ-GB-022: timeouts and debounce window are configurable via lsp.yaml.
type Config struct {
	// Enabled indicates whether the gopls bridge is active.
	// REQ-GB-051: when false (default), GoFeedbackGenerator retains the CLI-only path.
	Enabled bool
	// Binary is the gopls executable path or name. Default: "gopls".
	Binary string
	// Args are additional arguments passed to the gopls subprocess.
	Args []string
	// InitOptions is the initializationOptions field of the LSP initialize request.
	// REQ-GB-013: must include staticcheck: true.
	InitOptions map[string]any
	// Timeout is the per-request timeout for individual LSP requests (didOpen, diagnostics, etc.). Default: 30s.
	// REQ-GB-020: timeout for waiting on publishDiagnostics after didOpen.
	Timeout time.Duration
	// InitTimeout is the timeout for the LSP initialize handshake.
	// REQ-GB-012: initialization timeout of 30 seconds.
	InitTimeout time.Duration
	// ShutdownTimeout is the graceful shutdown timeout.
	// REQ-GB-004: SIGKILL after a 5-second timeout.
	ShutdownTimeout time.Duration
	// DebounceWindow is the publishDiagnostics debounce window.
	// REQ-GB-021: default 150ms.
	DebounceWindow time.Duration
}

// DefaultConfig returns a Config populated with sensible defaults.
// binary=gopls, timeout=30s, initTimeout=30s, shutdownTimeout=5s, debounce=150ms.
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		Binary:          "gopls",
		Args:            []string{},
		InitOptions:     map[string]any{"staticcheck": true},
		Timeout:         30 * time.Second,
		InitTimeout:     30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		DebounceWindow:  150 * time.Millisecond,
	}
}

// ─── Internal structs for YAML parsing ───────────────────────────────────────
// 1:1 mapping to the lsp.yaml structure.

type lspYAMLRoot struct {
	LSP lspYAMLConfig `yaml:"lsp"`
}

type lspYAMLConfig struct {
	GoplsBridge goplsBridgeYAML `yaml:"gopls_bridge"`
}

type goplsBridgeYAML struct {
	Enabled     bool           `yaml:"enabled"`
	Binary      string         `yaml:"binary"`
	Args        []string       `yaml:"args"`
	InitOptions map[string]any `yaml:"init_options"`
	Timeouts    timeoutsYAML   `yaml:"timeouts"`
}

type timeoutsYAML struct {
	InitializeSeconds      int `yaml:"initialize_seconds"`
	RequestSeconds         int `yaml:"request_seconds"`
	ShutdownSeconds        int `yaml:"shutdown_seconds"`
	DiagnosticsDebounceMs  int `yaml:"diagnostics_debounce_ms"`
}

// LoadConfig loads the gopls bridge settings from the YAML file at configPath.
// Returns DefaultConfig when the file does not exist (not an error).
// All other errors, such as YAML syntax errors, are returned as an error.
//
// REQ-GB-050: reads settings from lsp.yaml.
// Related to REQ-GB-002: absence of config is handled gracefully, identical to a missing gopls binary.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File not found: return defaults. This is not an error.
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("gopls: config: file read failed %q: %w", configPath, err)
	}

	var root lspYAMLRoot
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("gopls: config: YAML parsing failed %q: %w", configPath, err)
	}

	cfg := mergeWithDefaults(&root.LSP.GoplsBridge)

	// F5: binary and argument allowlist check.
	// The binary/args in lsp.yaml are a supply-chain attack vector that external editors could tamper with.
	if err := validateBinary(cfg.Binary); err != nil {
		return nil, fmt.Errorf("gopls: config: binary validation failed: %w", err)
	}
	if err := validateArgs(cfg.Args); err != nil {
		return nil, fmt.Errorf("gopls: config: argument validation failed: %w", err)
	}

	return cfg, nil
}

// mergeWithDefaults merges the gopls_bridge settings parsed from YAML with the defaults.
// Fields with zero values are filled with the defaults.
func mergeWithDefaults(y *goplsBridgeYAML) *Config {
	def := DefaultConfig()
	cfg := &Config{
		Enabled:     y.Enabled,
		Binary:      def.Binary,
		Args:        def.Args,
		InitOptions: def.InitOptions,
	}

	if y.Binary != "" {
		cfg.Binary = y.Binary
	}
	if len(y.Args) > 0 {
		cfg.Args = y.Args
	}
	if len(y.InitOptions) > 0 {
		cfg.InitOptions = y.InitOptions
	}

	// Timeouts: use defaults when the value is 0.
	cfg.InitTimeout = durationOrDefault(y.Timeouts.InitializeSeconds, 0, def.InitTimeout, time.Second)
	cfg.Timeout = durationOrDefault(y.Timeouts.RequestSeconds, 0, def.Timeout, time.Second)
	cfg.ShutdownTimeout = durationOrDefault(y.Timeouts.ShutdownSeconds, 0, def.ShutdownTimeout, time.Second)
	cfg.DebounceWindow = durationOrDefault(y.Timeouts.DiagnosticsDebounceMs, 0, def.DebounceWindow, time.Millisecond)

	return cfg
}

// durationOrDefault returns defaultVal when value equals zero, otherwise returns value * unit.
func durationOrDefault(value, zero int, defaultVal time.Duration, unit time.Duration) time.Duration {
	if value == zero {
		return defaultVal
	}
	return time.Duration(value) * unit
}

// ─── Security: binary and argument allowlist ──────────────────────────────────
//
// REQ-GB-F5: binary/args in lsp.yaml are a supply-chain attack vector.
// Reject immediately if binary points to an absolute path outside trusted paths or contains shell metacharacters.

// ErrUntrustedBinary is returned when an untrusted binary path is specified.
var ErrUntrustedBinary = errors.New("gopls: config: untrusted binary path")

// ErrUnsafeArgs is returned when an argument contains shell metacharacters.
var ErrUnsafeArgs = errors.New("gopls: config: disallowed argument characters")

// trustedPrefixes is the set of trusted directories where absolute-path binaries may reside.
// $HOME-based paths are expanded at runtime.
var trustedPrefixesStatic = []string{
	"/usr/bin",
	"/usr/local/bin",
	"/opt/homebrew/bin",
	"/opt/local/bin",
}

// binaryMetachars is the set of characters that must never appear in a binary name or path.
// Windows paths legitimately contain backslashes, so they are excluded.
// exec.Command does not go through a shell, so backslashes themselves carry no injection risk.
const binaryMetachars = ";&|`$"

// argMetachars is the set of shell metacharacters disallowed in arguments.
const argMetachars = ";&|`$\n\r"

// validateBinary checks whether cfg.Binary is an allowed binary.
//
// Allowed conditions:
//   - Bare name (no path separator): e.g. "gopls" → allowed (delegated to PATH lookup).
//   - Absolute path: must start with one of the trustedPrefixes.
//
// Rejected conditions:
//   - Path traversal (contains "..").
//   - Shell metacharacters (";", "|", "&", "`", "$", "\").
//   - Absolute path outside a trusted directory.
func validateBinary(binary string) error {
	if binary == "" {
		return fmt.Errorf("%w: empty binary name", ErrUntrustedBinary)
	}

	// Check for shell metacharacters.
	if strings.ContainsAny(binary, binaryMetachars) {
		return fmt.Errorf("%w: shell metacharacter in %q", ErrUntrustedBinary, binary)
	}

	// Bare name is allowed (e.g. "gopls", "gopls-v0.14").
	if !strings.Contains(binary, string(filepath.Separator)) && !strings.Contains(binary, "/") {
		return nil
	}

	// Absolute path check.
	// Resolve path traversal ("..") via filepath.Clean and compare with the original.
	cleaned := filepath.Clean(binary)

	// If ".." is present, filepath.Clean produces a different result. Detect this.
	if strings.Contains(binary, "..") {
		return fmt.Errorf("%w: path traversal attempt %q", ErrUntrustedBinary, binary)
	}

	// Build the trusted prefix list: static list + $HOME-based dynamic paths.
	prefixes := append([]string(nil), trustedPrefixesStatic...)
	if home, err := os.UserHomeDir(); err == nil {
		prefixes = append(prefixes,
			filepath.Join(home, "go", "bin"),
			filepath.Join(home, ".local", "bin"),
		)
	}

	// Convert both sides to forward-slash canonical form for platform-independent comparison.
	// On Windows, filepath.Clean turns "/usr/bin/x" into "\usr\bin\x", but the comparison
	// is slash-based to maintain cross-platform consistency.
	cleanedSlash := filepath.ToSlash(cleaned)
	for _, prefix := range prefixes {
		prefixSlash := filepath.ToSlash(prefix)
		if strings.HasPrefix(cleanedSlash, prefixSlash+"/") || cleanedSlash == prefixSlash {
			return nil
		}
	}

	return fmt.Errorf("%w: binary %q is outside trusted paths (allowlist: %v)", ErrUntrustedBinary, binary, prefixes)
}

// validateArgs checks whether the additional arguments to be passed to gopls are safe.
//
// Rejected conditions:
//   - Contains shell metacharacters (";", "|", "&", "`", "$", ">", "<", newline).
func validateArgs(args []string) error {
	// Dangerous character set including redirection operators.
	const dangerousChars = argMetachars + "><"
	for _, arg := range args {
		if strings.ContainsAny(arg, dangerousChars) {
			return fmt.Errorf("%w: argument %q contains disallowed characters", ErrUnsafeArgs, arg)
		}
	}
	return nil
}
