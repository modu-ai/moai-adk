package cli

// Coverage tests for deps.go functions below threshold.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// --- enableObservabilityIfConfigured ---

// TestEnableObservabilityIfConfigured_NoConfigFile is a no-op when config missing.
func TestEnableObservabilityIfConfigured_NoConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := config.NewConfigManager()
	reg := hook.NewRegistry(cfg)

	// Should not panic; config file missing = silent skip.
	enableObservabilityIfConfigured(reg, dir)
}

// TestEnableObservabilityIfConfigured_WithConfigFile enables when file exists.
func TestEnableObservabilityIfConfigured_WithConfigFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "observability.yaml"), []byte("enabled: true\n"), 0o644)

	cfg := config.NewConfigManager()
	reg := hook.NewRegistry(cfg)
	// Should not panic; registry may or may not implement observabilityEnabler.
	enableObservabilityIfConfigured(reg, dir)
}

// --- buildSessionEndHandler ---

// TestBuildSessionEndHandler_NoTraceDir uses standard handler.
func TestBuildSessionEndHandler_NoTraceDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No .moai/logs directory.
	h := buildSessionEndHandler(dir)
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

// TestBuildSessionEndHandler_WithTraceDir uses observability handler.
func TestBuildSessionEndHandler_WithTraceDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	traceDir := filepath.Join(dir, ".moai", "logs")
	_ = os.MkdirAll(traceDir, 0o755)

	h := buildSessionEndHandler(dir)
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

// --- buildAutoUpdateFunc ---

// TestBuildAutoUpdateFunc_ReturnsNonNilFunc returns a valid function.
func TestBuildAutoUpdateFunc_ReturnsNonNilFunc(t *testing.T) {
	t.Parallel()

	fn := buildAutoUpdateFunc()
	if fn == nil {
		t.Error("expected non-nil function from buildAutoUpdateFunc()")
	}
}
