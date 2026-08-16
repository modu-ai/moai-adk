// Package hook — observability_master.go
//
// Resolution: NEW for SPEC-V3R6-HOOK-ASYNC-EXPAND-001 REQ-HAE-003/004.
//
// IsObservabilityEnabled() implements the REQ-OBS-005 master toggle read path.
// It reads `.moai/config/sections/observability.yaml` lazily once per process
// (sync.Once) and caches the boolean result, so the dual-gate fast-path on
// TaskCreated + Notification hot paths remains cheap.
//
// COHABITATION NOTE (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3):
//
//   - This function reads `.moai/config/sections/observability.yaml`
//     top-level key `observability.enabled` (REQ-OBS-005 trace-logging master).
//
//   - observabilityOptIn() (observability.go) reads `system.yaml`
//     `hook.observability_events` (RT-006 per-event whitelist).
//
//   - hookOptInEnabled() (hook_opt_in.go) reads `system.yaml`
//     `hook.opt_in.enabled` (REQ-HOI-001 master).
//
// ALL 3 READ PATHS ARE INDEPENDENT. Do NOT unify without a fresh SPEC.
// The 3 files / 3 yaml keys are the §A.3 cohabitation invariant.
package hook

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// observabilityMasterOnce ensures the master toggle is read at most once
// per process — the dual-gate hot path on TaskCreated + Notification MUST
// stay cheap.
var (
	observabilityMasterOnce    sync.Once
	observabilityMasterEnabled bool
)

// observabilityYAMLFile is a top-level shape of `observability.yaml`
// covering only the keys this gate needs. Other keys (hook_metrics,
// trace_dir, retention_days, etc.) are owned by the registry-level
// observability subsystem and intentionally not duplicated here.
type observabilityYAMLFile struct {
	Observability struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"observability"`
}

// IsObservabilityEnabled reports whether the REQ-OBS-005 master toggle in
// `.moai/config/sections/observability.yaml` is `true`. Reads the file at
// most once per process; subsequent calls return the cached value.
//
// Returns false when:
//   - the file is missing
//   - the file is unreadable
//   - the YAML is invalid
//   - the `observability.enabled` key is absent or false
//
// Safe default (false = disabled) matches REQ-HAE-003/004 zero-overhead
// path: when the master toggle is off, the dual-gate fast-path skips the
// async goroutine entirely.
//
// The cwd-resolution heuristic: tries `$CLAUDE_PROJECT_DIR` first (the
// canonical project root from Claude Code), then falls back to the
// current working directory. Tests can override via
// SetObservabilityMasterForTesting.
func IsObservabilityEnabled() bool {
	return isObservabilityEnabled(true)
}

// IsObservabilityEnabledForCLI is the operator-CLI entry point of the same
// cached toggle (t62): identical result, but the cwd_fallback resolution
// logs at Debug instead of Warn, because a CLI process legitimately lacks
// CLAUDE_PROJECT_DIR and cwd is the right answer there. initDependencies
// (the caller on every non-trivial subcommand) uses this variant so the
// fallback line stops topping every `moai` invocation.
func IsObservabilityEnabledForCLI() bool {
	return isObservabilityEnabled(false)
}

// isObservabilityEnabled loads and caches the toggle once per process. The
// hookRuntime flag only picks the fallback-log level of the single load;
// whichever entry point runs first wins that level (in practice: CLI
// processes go through initDependencies first, hook processes resolve from
// the event handlers when the observability.yaml Stat gate skipped the
// initDependencies read).
func isObservabilityEnabled(hookRuntime bool) bool {
	observabilityMasterOnce.Do(func() {
		observabilityMasterEnabled = loadObservabilityMaster(hookRuntime)
	})
	return observabilityMasterEnabled
}

// loadObservabilityMaster does the actual file read. Separated for
// testability via SetObservabilityMasterForTesting.
//
// Resolves project root via resolveProjectRootFromEnvAt: CLAUDE_PROJECT_DIR
// env var first, then os.Getwd() fallback with a cwd_fallback marker at the
// context-chosen level (REQ-HCWA-006, REQ-HCWA-008) — WARN in the hook
// runtime, Debug on operator CLI paths (t62). The .moai/ existence guard is
// NOT applied here because the function reads from a path that may not
// include observability.yaml (the file is the toggle target itself).
func loadObservabilityMaster(hookRuntime bool) bool {
	fallbackLevel := slog.LevelWarn
	if !hookRuntime {
		fallbackLevel = slog.LevelDebug
	}
	root := resolveProjectRootFromEnvAt("loadObservabilityMaster", fallbackLevel)
	if root == "" {
		// resolveProjectRootFromEnv already logged the failure.
		return false
	}
	path := filepath.Join(root, ".moai", "config", "sections", "observability.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		// Missing file → disabled (safe default).
		slog.Debug("observability master: file missing", "path", path)
		return false
	}
	var parsed observabilityYAMLFile
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		slog.Warn("observability master: invalid YAML", "path", path, "error", err)
		return false
	}
	return parsed.Observability.Enabled
}

// SetObservabilityMasterForTesting overrides the cached observability
// master toggle. Tests MUST call this with the desired value BEFORE
// invoking handlers that consult IsObservabilityEnabled().
//
// This bypasses sync.Once by directly mutating the cached value. Calls
// outside of tests are an error (no production caller exists).
func SetObservabilityMasterForTesting(enabled bool) {
	observabilityMasterEnabled = enabled
	// Mark the sync.Once as already done so subsequent IsObservabilityEnabled
	// calls return the test-set value rather than re-reading the file.
	observabilityMasterOnce.Do(func() {})
}

// ResetObservabilityMasterForTesting clears the cached value and the
// sync.Once gate. Tests use this when they need to exercise the
// file-read path itself (rare). Subsequent IsObservabilityEnabled
// calls will re-read the file.
//
// NOT safe for concurrent use — call only from a test cleanup function
// or in a t.Cleanup() closure that runs after all goroutines complete.
func ResetObservabilityMasterForTesting() {
	observabilityMasterEnabled = false
	observabilityMasterOnce = sync.Once{}
}
