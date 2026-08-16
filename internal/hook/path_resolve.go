// path_resolve.go — sibling helpers to resolveProjectRoot for cwd-leak audit.
//
// SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 introduces two read-side / registration-time
// resolvers that complement resolveProjectRoot (post_tool_metrics.go).
//
// Distinction:
//   - resolveProjectRoot          (post_tool_metrics.go): WRITE-side resolver
//     with .moai/ existence guard. Used when a
//     handler is about to create files under
//     .moai/. Returns empty string when the
//     resolved path is NOT a MoAI project root.
//   - resolveProjectRootFromEnv          (this file): NO .moai/ guard. Used
//     for read-side or registration-time cwd
//     resolution where the .moai/ directory may
//     legitimately be absent (e.g., loading the
//     observability.yaml toggle itself from a
//     fresh project).
//   - resolveProjectRootFromInputOrEnv   (this file): NO .moai/ guard, but
//     prefers input.CWD over env var. Used by
//     handlers that have HookInput available.
//
// Both helpers emit a structured slog.Warn entry with `"cwd_fallback":true`
// whenever os.Getwd() is the last-resort fallback (REQ-HCWA-008).
//
// File placement rationale (plan.md §E): separate file rather than inline in
// post_tool_metrics.go, so that AC-HCWA-007 awk extraction of resolveProjectRoot
// does not accidentally match these siblings (the awk regex `^func resolveProjectRoot`
// would otherwise glob-match all three function names).
package hook

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/modu-ai/moai-adk/internal/config"
)

// projectRootResolver defers registration-time project-root resolution to first
// use. Handlers embed it and expose a projectRoot() accessor.
//
// Why deferred: InitDependencies (internal/cli/deps.go) constructs every handler
// for every non-trivial subcommand, long before — and usually without — a hook
// event ever being dispatched. Resolving in the constructor therefore emitted a
// cwd_fallback warning on operator-facing commands, where CLAUDE_PROJECT_DIR is
// legitimately unset and cwd is the right answer rather than a degraded one.
// Resolution semantics are unchanged: CLAUDE_PROJECT_DIR, then os.Getwd()
// (REQ-HCWA-005, REQ-HCWA-008).
type projectRootResolver struct {
	caller string
	once   sync.Once
}

// resolve fills *dir on first call and returns it. A dir already set by the
// caller (tests construct handlers with an explicit project root) is preserved.
func (r *projectRootResolver) resolve(dir *string) string {
	r.once.Do(func() {
		if *dir == "" && r.caller != "" {
			*dir = resolveProjectRootFromEnv(r.caller)
		}
	})
	return *dir
}

// resolveProjectRootFromEnv returns CLAUDE_PROJECT_DIR or os.Getwd() fallback
// without the .moai/ existence guard. Emits slog.Warn with key
// "cwd_fallback":true when os.Getwd() is used (REQ-HCWA-008).
func resolveProjectRootFromEnv(caller string) string {
	return resolveProjectRootFromEnvAt(caller, slog.LevelWarn)
}

// resolveProjectRootFromEnvAt is resolveProjectRootFromEnv with an explicit
// fallback-log level (t62). Operator CLI paths resolve with slog.LevelDebug:
// there CLAUDE_PROJECT_DIR is legitimately unset and cwd is the right answer
// rather than a degraded one, so the fallback must not reach the operator's
// terminal at the default Warn level. Hook runtime keeps the plain WARN form.
func resolveProjectRootFromEnvAt(caller string, fallbackLevel slog.Level) string {
	if root := os.Getenv(config.EnvClaudeProjectDir); root != "" {
		return root
	}
	cwd, err := os.Getwd()
	if err != nil {
		slog.Debug("cwd fallback failed", "caller", caller, "error", err)
		return ""
	}
	slog.Log(context.Background(), fallbackLevel, "cwd fallback used (CLAUDE_PROJECT_DIR not set)",
		"cwd_fallback", true,
		"caller", caller,
		"resolved_cwd", cwd,
	)
	return cwd
}

// resolveProjectRootFromInputOrEnv returns input.CWD, then CLAUDE_PROJECT_DIR,
// then os.Getwd() fallback. No .moai/ existence guard. Delegates to
// resolveProjectRootFromEnv when input is nil or input.CWD is empty.
func resolveProjectRootFromInputOrEnv(input *HookInput, caller string) string {
	if input != nil && input.CWD != "" {
		return input.CWD
	}
	return resolveProjectRootFromEnv(caller)
}
