package config

// session_worktree.go — SPEC-SESSION-WORKTREE-001 M1 activation decision.
//
// SessionWorktreeEnabled is the SINGLE activation decision point consumed by
// the auto-entry wrappers in M2 (moai init), M3 (moai web), and M6
// (moai profile). It resolves the MOAI_SESSION_WORKTREE env override first
// (REQ-SW-003: env wins over config), then falls through to the
// workflow.session_worktree.enabled config flag. The default (both unset) is
// OFF, so moai init / moai profile / moai web behave byte-identically to the
// shared-checkout baseline when the feature is not opted into (REQ-SW-001).

import "os"

// SessionWorktreeEnabled reports whether the session-worktree auto-entry
// feature is active for the given config, honoring the MOAI_SESSION_WORKTREE
// env override.
//
// Resolution order (REQ-SW-003):
//  1. EnvSessionWorktree == "1" → forced ON (overrides config either way).
//  2. EnvSessionWorktree == "0" → forced OFF (overrides config either way).
//  3. otherwise (unset, "", or any unrecognized value) → fall through to the
//     workflow.session_worktree.enabled config flag.
//
// A nil cfg is treated as OFF (a caller that failed to load config MUST NOT
// silently enable worktree isolation). This nil-safety is load-bearing for
// fail-open behavior on missing config.
//
// @MX:ANCHOR: [AUTO] single activation decision site for session-worktree auto-entry
// @MX:REASON: REQ-SW-001/002/003 — every M2/M3/M6 entry wrapper MUST route through this one helper; a second decision site would diverge from the env-wins contract
func SessionWorktreeEnabled(cfg *Config) bool {
	switch os.Getenv(EnvSessionWorktree) {
	case "1":
		return true
	case "0":
		return false
	default:
		// Unset, empty, or unrecognized value — fall through to config.
		if cfg == nil {
			return false
		}
		return cfg.Workflow.SessionWorktree.Enabled
	}
}
