// m0_frozen_guard_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, hook
// family: AC-009 (frozen guard denies a user-path delete).
//
// M0 discipline: observation only — no production change.
package hook

import (
	"testing"
)

// TestFrozenGuardDeniesUserPathDelete — AC-009 (ledger 9a, REQ-GRD-001/002).
// Given a managed file under the user-installed root (the plan-auditor.md
// reproduction shape — the per-user agent folder is part of the four install
// roots userassets.ResolveRoots defines at paths.go:38), a delete attempt
// must be DENIED. RED-now reason: the user install roots are absent from the
// Bash-branch protection surface — dangerousRemovalTarget (the only delete
// check) protects root/home/top-level/.git/node_modules shapes, and a deep
// user path resolves to allow.
func TestFrozenGuardDeniesUserPathDelete(t *testing.T) {
	for _, command := range []string{
		"rm -f ~/.claude/agents/moai/plan-auditor.md",
		"rm -rf ~/.claude/skills/moai-foundation-core",
		"rm -f \"$HOME\"/.codex/agents/moai/plan-auditor.toml",
		"rm -rf ~/.agents/skills/moai",
	} {
		if got := decideBash(t, command); got != DecisionDeny {
			t.Errorf("RED (intended): delete of a user-installed managed asset was allowed (decision %q, want %q) — command: %s; the user install roots are not in the protected set", got, DecisionDeny, command)
		}
	}
}
