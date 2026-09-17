package config

import (
	"path/filepath"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// drift_cache_fill.go — the single read point for workflow.drift_cache_fill.enabled.
//
// The key opts OUT of the out-of-band drift-cache fill: the detached child the
// SessionStart handler starts when the HEAD-SHA-keyed drift cache misses. With
// the fill disabled the handler still resolves the advisory from the cache; it
// simply never fills it, which returns the miss path to the pre-change
// behaviour of showing no drift advisory.

// DriftCacheFillEnabled reports whether the out-of-band fill may be started.
//
// Nil-receiver safe and fail-OPEN: a caller that could not load config gets the
// feature rather than silence, matching TodoEnabled's posture. The failure this
// protects against is asymmetric — a feature that vanished for an
// unreadable-config reason is far harder to diagnose than one that stayed, and
// the cost of the wrong answer here is one short-lived child process.
func (c *Config) DriftCacheFillEnabled() bool {
	if c == nil {
		return true
	}
	return c.Workflow.DriftCacheFill.Enabled
}

// DriftCacheFillEnabledForRoot is the project-root convenience form the
// SessionStart handler calls. projectRoot is the project directory (the parent
// of .moai), NOT the .moai directory itself.
//
// Every failure path returns true: an empty root, a missing .moai tree, and a
// load error all resolve to enabled.
func DriftCacheFillEnabledForRoot(projectRoot string) bool {
	if projectRoot == "" {
		return true
	}
	cfg, err := NewLoader().Load(filepath.Join(projectRoot, defs.MoAIDir))
	if err != nil {
		return true
	}
	return cfg.DriftCacheFillEnabled()
}
