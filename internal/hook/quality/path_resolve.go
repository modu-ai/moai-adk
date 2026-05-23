// path_resolve.go — package-local cwd resolver for quality-gate operations.
//
// SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 M3 introduces resolveQualityProjectDir
// as the package-quality counterpart to resolveProjectRootFromEnv in the
// parent hook package. The split is necessary because:
//
//   1. package hook/quality does NOT import package hook (avoiding a cycle).
//   2. package hook/quality has no HookInput type to consult; the
//      authoritative project-dir source here is GateConfig.ProjectDir.
//   3. package hook/quality does NOT import internal/config (kept minimal);
//      the env-var name "CLAUDE_PROJECT_DIR" is a literal string here.
//
// Resolution priority:
//
//   1. cfg.ProjectDir  — set by the caller (e.g., preToolHandler.loadGateConfig)
//   2. CLAUDE_PROJECT_DIR  — env var set by the Claude Code hook system
//   3. os.Getwd()  — last-resort fallback, emits slog.Warn cwd_fallback:true
//
// No .moai/ existence guard here — quality-gate operations (linting, vet, test)
// are contractually invoked from project roots; adding the guard would change
// semantics (linting from a subdirectory currently works as expected).
package quality

import (
	"log/slog"
	"os"
)

// resolveQualityProjectDir returns the project directory for quality-gate
// operations. Preference order: cfg.ProjectDir → CLAUDE_PROJECT_DIR env var →
// os.Getwd() fallback. Emits slog.Warn cwd_fallback:true when os.Getwd()
// fallback is used (REQ-HCWA-007, REQ-HCWA-008).
func resolveQualityProjectDir(cfg GateConfig, caller string) string {
	if cfg.ProjectDir != "" {
		return cfg.ProjectDir
	}
	if root := os.Getenv("CLAUDE_PROJECT_DIR"); root != "" {
		return root
	}
	cwd, err := os.Getwd()
	if err != nil {
		slog.Debug("cwd fallback failed", "caller", caller, "error", err)
		return ""
	}
	slog.Warn("cwd fallback used in quality gate",
		"cwd_fallback", true,
		"caller", caller,
		"resolved_cwd", cwd,
	)
	return cwd
}
