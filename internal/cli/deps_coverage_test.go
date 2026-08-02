package cli

// Coverage tests for deps.go functions below threshold.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// stubObservabilityRegistry is a minimal hook.Registry stub whose
// EnableObservability call-count lets the deps.go gate tests assert whether the
// master toggle was consulted. It exists because the gate under test performs a
// local `observabilityEnabler` type assertion, so the stub must satisfy both the
// hook.Registry interface AND expose EnableObservability.
type stubObservabilityRegistry struct {
	enableCalls int
	lastLogDir  string
}

func (s *stubObservabilityRegistry) Register(hook.Handler) {}
func (s *stubObservabilityRegistry) Dispatch(context.Context, hook.EventType, *hook.HookInput) (*hook.HookOutput, error) {
	return nil, nil
}
func (s *stubObservabilityRegistry) Handlers(hook.EventType) []hook.Handler { return nil }
func (s *stubObservabilityRegistry) EnableObservability(logDir string) {
	s.enableCalls++
	s.lastLogDir = logDir
}

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

// TestEnableObservabilityIfConfigured_WithConfigFile enables when file exists
// AND the observability master toggle is on (AC-OEG-002).
//
// After SPEC-OBS-ENABLED-GATE-001 the gate consults hook.IsObservabilityEnabled()
// when the file exists; the seam (SetObservabilityMasterForTesting) is the only
// honest way to drive the enabled-decision from a tempdir because
// IsObservabilityEnabled resolves cwd via CLAUDE_PROJECT_DIR → os.Getwd().
//
// Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state
// (internal/hook/notification_test.go:87 precedent).
func TestEnableObservabilityIfConfigured_WithConfigFile(t *testing.T) {
	hook.ResetObservabilityMasterForTesting()
	hook.SetObservabilityMasterForTesting(true)
	t.Cleanup(hook.ResetObservabilityMasterForTesting)

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "observability.yaml"), []byte("observability:\n  enabled: true\n"), 0o644)

	stub := &stubObservabilityRegistry{}
	enableObservabilityIfConfigured(stub, dir)

	if stub.enableCalls != 1 {
		t.Fatalf("EnableObservability call-count = %d, want 1 (enabled:true + file present)", stub.enableCalls)
	}
	wantLogDir := filepath.Join(dir, ".moai", "logs")
	if stub.lastLogDir != wantLogDir {
		t.Errorf("EnableObservability logDir = %q, want %q", stub.lastLogDir, wantLogDir)
	}
	if _, err := os.Stat(wantLogDir); err != nil {
		t.Errorf("log directory %q was not created: %v", wantLogDir, err)
	}
}

// TestEnableObservabilityIfConfigured_DisabledByEnabledKey verifies AC-OEG-001:
// when observability.yaml exists but the master toggle resolves to false, the
// gate MUST NOT call EnableObservability and MUST NOT create the log directory.
//
// The tempdir satisfies the os.Stat existence gate; the master-toggle value
// (false) is injected via the seam because IsObservabilityEnabled cannot read a
// tempdir.
//
// Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state
// (internal/hook/notification_test.go:87 precedent).
func TestEnableObservabilityIfConfigured_DisabledByEnabledKey(t *testing.T) {
	hook.ResetObservabilityMasterForTesting()
	hook.SetObservabilityMasterForTesting(false)
	t.Cleanup(hook.ResetObservabilityMasterForTesting)

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "observability.yaml"), []byte("observability:\n  enabled: false\n"), 0o644)

	stub := &stubObservabilityRegistry{}
	enableObservabilityIfConfigured(stub, dir)

	if stub.enableCalls != 0 {
		t.Errorf("EnableObservability call-count = %d, want 0 (enabled:false must NOT enable observability); logDir=%q",
			stub.enableCalls, stub.lastLogDir)
	}
	logDir := filepath.Join(dir, ".moai", "logs")
	if _, err := os.Stat(logDir); err == nil {
		t.Errorf("log directory %q was created; expected it NOT to be created under enabled:false", logDir)
	}
}

// TestEnableObservabilityIfConfigured_AbsentKeyDefaultsDisabled verifies AC-OEG-003:
// when observability.yaml exists but the enabled key is absent, the gate treats
// the config as disabled (safe-default false), aligning with
// IsObservabilityEnabled's absent-key semantics.
//
// Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state
// (internal/hook/notification_test.go:87 precedent).
func TestEnableObservabilityIfConfigured_AbsentKeyDefaultsDisabled(t *testing.T) {
	hook.ResetObservabilityMasterForTesting()
	hook.SetObservabilityMasterForTesting(false)
	t.Cleanup(hook.ResetObservabilityMasterForTesting)

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".moai", "config", "sections")
	_ = os.MkdirAll(cfgDir, 0o755)
	// File present, no enabled key (IsObservabilityEnabled safe-defaults to false).
	_ = os.WriteFile(filepath.Join(cfgDir, "observability.yaml"), []byte("observability:\n"), 0o644)

	stub := &stubObservabilityRegistry{}
	enableObservabilityIfConfigured(stub, dir)

	if stub.enableCalls != 0 {
		t.Errorf("EnableObservability call-count = %d, want 0 (absent enabled-key must safe-default to disabled); logDir=%q",
			stub.enableCalls, stub.lastLogDir)
	}
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
