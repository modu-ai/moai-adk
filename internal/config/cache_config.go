package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

// SPEC-V3R6-PROMPT-CACHE-001 M2 — cache.yaml config schema.
//
// CacheConfig is the typed representation of the cacheStrategy section of
// cache.yaml. It is consumed by the M1 cache_control injector
// (internal/runtime.CacheStrategy mirrors these fields).
//
// REQ-PC-005: when session_ttl == "off", no session-level breakpoint is injected.
// AC-PC-002: the four cacheStrategy keys (enabled / session_ttl / spec_ttl /
//            min_cacheable_tokens) MUST be present and validated.

// Default values for the cacheStrategy section.
const (
	// DefaultCacheSessionTTL is the session-start breakpoint TTL default.
	DefaultCacheSessionTTL = "1h"
	// DefaultCacheSpecTTL is the SPEC-body breakpoint TTL default.
	DefaultCacheSpecTTL = "5m"
	// DefaultCacheMinTokens is the conservative model-agnostic threshold
	// (haiku minimum) below which cache_control is omitted (R4 fallback).
	DefaultCacheMinTokens = 2048
)

// CacheConfig holds the parsed cacheStrategy section.
type CacheConfig struct {
	// Enabled toggles cache_control injection. Ships disabled (safe default).
	Enabled bool `yaml:"enabled"`
	// SessionTTL is the session-start breakpoint TTL: "1h" | "5m" | "off".
	SessionTTL string `yaml:"session_ttl"`
	// SpecTTL is the SPEC-body breakpoint TTL: "5m" | "off".
	SpecTTL string `yaml:"spec_ttl"`
	// MinCacheableTokens is the R4 self-contained threshold (default 2048).
	MinCacheableTokens int `yaml:"min_cacheable_tokens"`
}

// cacheFileWrapper handles the cache.yaml top-level "cacheStrategy:" key.
type cacheFileWrapper struct {
	CacheStrategy CacheConfig `yaml:"cacheStrategy"`
}

// validSessionTTLs and validSpecTTLs enumerate the accepted enum values.
var (
	validSessionTTLs = map[string]bool{"1h": true, "5m": true, "off": true}
	validSpecTTLs    = map[string]bool{"5m": true, "off": true}
)

// DefaultCacheConfig returns the built-in defaults. Caching ships disabled until
// the user explicitly opts in, but the TTL and threshold defaults are populated
// so that enabling only requires flipping `enabled: true`.
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled:            false,
		SessionTTL:         DefaultCacheSessionTTL,
		SpecTTL:            DefaultCacheSpecTTL,
		MinCacheableTokens: DefaultCacheMinTokens,
	}
}

// Validate enforces the cacheStrategy invariants:
//   - session_ttl ∈ {"1h", "5m", "off"}
//   - spec_ttl ∈ {"5m", "off"}
//   - min_cacheable_tokens >= 0
//
// An empty session_ttl/spec_ttl is treated as invalid here; callers that want
// default-filling should do so before Validate (LoadCacheConfig fills defaults).
func (c CacheConfig) Validate() error {
	if !validSessionTTLs[c.SessionTTL] {
		return fmt.Errorf("cacheStrategy.session_ttl %q invalid: want one of 1h|5m|off", c.SessionTTL)
	}
	if !validSpecTTLs[c.SpecTTL] {
		return fmt.Errorf("cacheStrategy.spec_ttl %q invalid: want one of 5m|off", c.SpecTTL)
	}
	if c.MinCacheableTokens < 0 {
		return fmt.Errorf("cacheStrategy.min_cacheable_tokens %d invalid: must be >= 0", c.MinCacheableTokens)
	}
	return nil
}

// LoadCacheConfig reads cache.yaml from the given path and returns a typed
// CacheConfig.
//
// Return semantics (plan.md M2 "Missing/invalid → log warning, fall back to
// enabled: false"):
//   - (cfg, nil) on a successful, valid parse
//   - (safeDefault, nil) on file-not-found
//   - (safeDefault, nil) on malformed YAML or invalid enum/threshold values
//
// Unlike most loaders, this function NEVER returns an error: any failure mode
// degrades to the safe default (enabled: false) so a misconfigured cache.yaml
// can never block the SDK wrapper.
//
// @MX:ANCHOR: [AUTO] LoadCacheConfig — sole cache.yaml loader; safe-default-on-any-failure is a contractual invariant
// @MX:REASON: fan_in >= 3 expected — cc.go SDK wrapper, moai doctor cache metric path, and AC-PC-002 schema test all consume this; the never-error degradation contract prevents a malformed cache.yaml from blocking launch.
func LoadCacheConfig(path string) (*CacheConfig, error) {
	defaults := DefaultCacheConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("cache.yaml read failed, using safe default (disabled)", "path", path, "error", err)
		}
		return defaults, nil
	}

	cfg := DefaultCacheConfig()
	wrapper := cacheFileWrapper{CacheStrategy: *cfg}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		slog.Warn("cache.yaml malformed, using safe default (disabled)", "path", path, "error", err)
		return DefaultCacheConfig(), nil
	}

	parsed := wrapper.CacheStrategy
	// Default-fill any empty enum fields before validation so a partial file
	// (e.g., only `enabled: true`) still validates against the defaults.
	if parsed.SessionTTL == "" {
		parsed.SessionTTL = DefaultCacheSessionTTL
	}
	if parsed.SpecTTL == "" {
		parsed.SpecTTL = DefaultCacheSpecTTL
	}

	if err := parsed.Validate(); err != nil {
		slog.Warn("cache.yaml invalid, using safe default (disabled)", "path", path, "error", err)
		return DefaultCacheConfig(), nil
	}

	return &parsed, nil
}
