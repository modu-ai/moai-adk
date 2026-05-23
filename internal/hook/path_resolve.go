// path_resolve.go — sibling helpers to resolveProjectRoot for cwd-leak audit.
//
// SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 introduces two read-side / registration-time
// resolvers that complement resolveProjectRoot (post_tool_metrics.go).
//
// Distinction:
//   - resolveProjectRoot          (post_tool_metrics.go): WRITE-side resolver
//                                  with .moai/ existence guard. Used when a
//                                  handler is about to create files under
//                                  .moai/. Returns empty string when the
//                                  resolved path is NOT a MoAI project root.
//   - resolveProjectRootFromEnv          (this file): NO .moai/ guard. Used
//                                  for read-side or registration-time cwd
//                                  resolution where the .moai/ directory may
//                                  legitimately be absent (e.g., loading the
//                                  observability.yaml toggle itself from a
//                                  fresh project).
//   - resolveProjectRootFromInputOrEnv   (this file): NO .moai/ guard, but
//                                  prefers input.CWD over env var. Used by
//                                  handlers that have HookInput available.
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
	"log/slog"
	"os"

	"github.com/modu-ai/moai-adk/internal/config"
)

// resolveProjectRootFromEnv returns CLAUDE_PROJECT_DIR or os.Getwd() fallback
// without the .moai/ existence guard. Emits slog.Warn with key
// "cwd_fallback":true when os.Getwd() is used (REQ-HCWA-008).
func resolveProjectRootFromEnv(caller string) string {
	if root := os.Getenv(config.EnvClaudeProjectDir); root != "" {
		return root
	}
	cwd, err := os.Getwd()
	if err != nil {
		slog.Debug("cwd fallback failed", "caller", caller, "error", err)
		return ""
	}
	slog.Warn("cwd fallback used (CLAUDE_PROJECT_DIR not set)",
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
