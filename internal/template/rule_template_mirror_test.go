// rule_template_mirror_test.go: CI guard for SPEC-V3R5-WORKFLOW-OPT-001 REQ-WO-004 /
// REQ-WO-041. Verifies that every rule file under `.claude/rules/moai/**` modified by
// this SPEC has a corresponding mirror under
// `internal/template/templates/.claude/rules/moai/**` with byte-identical content.
// The mirror is what `go:embed` ships to user projects via `moai init` / `moai update`;
// drift between source and mirror means user projects receive stale rules.
//
// Sentinel on failure: RULE_TEMPLATE_MIRROR_DRIFT
// Origin: SPEC-V3R5-WORKFLOW-OPT-001 Layer A (M1.Y task).
//
// Scope (intentional narrow allowlist):
//   - .claude/rules/moai/** files modified or created by SPEC-V3R5-WORKFLOW-OPT-001
//   - .claude/agents/moai/plan-auditor.md (Layer G target, post SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 M2a FLAT layout)
//   - .moai/config/sections/workflow.yaml (Layer B target)
//   - .moai/config/evaluator-profiles/default.md and frontend.md (Layer G target)
//
// Out of scope:
//   - Other .moai/config/sections/*.yaml files (intentionally divergent per
//     CLAUDE.local.md §22 Dev Settings Intent — user.yaml, quality.yaml, etc.
//     differ between dev environment and shipped template baseline)
//   - Other .claude/rules/moai/** files (pre-existing drift tracked separately
//     in SPEC-V3R5-LINT-DEBT-001 follow-up; see EXCL-WO-004)
//
// @MX:ANCHOR: [AUTO] fan_in=2 — guards rule mirror invariant for SPEC-V3R5-WORKFLOW-OPT-001
// modified files. Touching this test's allowlist affects mirror drift detection scope.
// @MX:REASON: Without an explicit allowlist, the test would surface pre-existing intentional
// divergence (per CLAUDE.local.md §22) and fail noisily on files outside this SPEC's scope.
package template_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// SPEC-V3R5-WORKFLOW-OPT-001 mirrored files. Each entry MUST have a byte-identical
// mirror at internal/template/templates/<path>.
//
// Allowlist is intentionally explicit (no glob) so that adding a new mirrored file
// is a deliberate code change visible in PR review.
var workflowOptMirroredPaths = []string{
	// (new entry — REQ-TMD-005 — hooks-system.md mirror parity)
	".claude/rules/moai/core/hooks-system.md",
	// Layer E — Phase Transitions skip policy
	".claude/rules/moai/workflow/spec-workflow.md",
	// SPEC-SESSION-HANDOFF-ALIGN-001 — session-handoff.md mirror parity (REQ-SHA-007).
	// Both trees are byte-identical post-neutralization (Diet/V0/`/cd` blocks ported +
	// internal SPEC-IDs stripped per CLAUDE.local.md §25). Enrolled here so future
	// single-tree edits on this always-loaded canonical rule are caught at CI.
	".claude/rules/moai/workflow/session-handoff.md",
	// Layer G — evaluator profile D7/D8 weight registration
	".moai/config/evaluator-profiles/default.md",
	".moai/config/evaluator-profiles/frontend.md",
	// per-file §25 sanitization targets — REMOVED from the byte-parity allowlist.
	// The 5 source files below retain internal-development content (SPEC-IDs, REQ/AC
	// tokens) in their .claude/ working copy, while their template mirrors are held
	// sanitized for neutral distribution (CLAUDE.local.md §25). byte-parity therefore
	// cannot hold for them; mirror cleanliness is enforced by
	// TestTemplateNoInternalContentLeak instead of byte-identity here. Ground-truth at
	// remediation time: drift∩leak = 7 files (not 1), so all leak-bearing drift files
	// move to leak-test coverage:
	//   - .claude/rules/moai/development/manager-develop-prompt-template.md (5 tokens)
	//   - .claude/rules/moai/workflow/ci-watch-protocol.md (1 token)
	//   - .claude/rules/moai/core/agent-common-protocol.md (17 tokens)
	//   - .claude/rules/moai/workflow/verification-batch-pattern.md (2 tokens)
	//   - .claude/agents/moai/plan-auditor.md (6 tokens)
	//   (the former 5+1+1 Agent Teams pattern file was deleted per
	//   SPEC-V3R6-RULES-SSOT-DEDUP-001 M6 — its content folded into
	//   team-pattern-cookbook.md 6th pattern.)
}

// SPEC-V3R5-LATE-BRANCH-001 mirrored files. Each entry MUST have a byte-identical
// mirror at internal/template/templates/<path>. Promoted from Optional (REQ-LB-008)
// to Mandatory per plan-auditor iter 1 Q2 CRITICAL recommendation (project_v3r5_late_branch_*).
//
// The 4 markdown mirrors (.md) must match byte-for-byte; the .yaml.tmpl mirror
// contains Go template variables ({{.GitMode}}, etc.) and is intentionally not
// in this list — its 4 modified keys are verified by AC-LB-005 grep instead.
//
// Sentinel on failure: RULE_TEMPLATE_MIRROR_DRIFT (shared with workflowOptMirroredPaths).
var lateBranchMirroredPaths = []string{
	// D2 — spec-assembly Phase 3 Late-branch pre-check + Phase 2.5 opt-in
	".claude/skills/moai/workflows/plan/spec-assembly.md",
	// (spec-workflow.md is already in workflowOptMirroredPaths above — Late-branch
	// closure additions are part of the same file, mirror parity verified there.)
	// per-file §21 dev-only / §25 sanitization targets — REMOVED from the byte-parity allowlist:
	//   - .claude/skills/moai/SKILL.md: source retains the dev-only `release-update` command
	//     entry + section (CLAUDE.local.md §21 — 97-series dev-only, NOT distributed to user
	//     projects) while the template mirror is held free of it. byte-parity therefore cannot
	//     hold; mirror cleanliness is covered by TestTemplateNoInternalContentLeak instead.
	//     (Origin: M2 commit 73dbdb672 intentionally stripped the release-update route from the
	//     template mirror per §21; this allowlist entry should have been removed at that time.)
	//   - manager-spec.md: source retains the SPEC ID regex pedagogical example
	//     (working copy) while the mirror is held sanitized for distribution.
	//   - manager-git.md: source retains internal SPEC-ID/token content while the
	//     mirror is held sanitized.
	// byte-parity cannot hold for any of these; mirror cleanliness is covered by
	// TestTemplateNoInternalContentLeak (leak test) instead of byte-parity here.
}

// TestRuleTemplateMirrorDrift verifies that every SPEC-V3R5-WORKFLOW-OPT-001 modified
// rule file has a byte-identical mirror under internal/template/templates/.
//
// Reports the literal sentinel RULE_TEMPLATE_MIRROR_DRIFT on every drift detected
// so CI log parsers can match the failure pattern.
//
// @MX:NOTE: [AUTO] AC-WO-003 (template mirror parity) is verified by this test
// for the subset of rule files touched by SPEC-V3R5-WORKFLOW-OPT-001.
func TestRuleTemplateMirrorDrift(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for _, rel := range workflowOptMirroredPaths {
		rel := rel // capture
		t.Run(filepath.Base(rel), func(t *testing.T) {
			t.Parallel()

			srcPath := filepath.Join(projectRoot, rel)
			mirrorPath := filepath.Join(projectRoot, "internal", "template", "templates", rel)

			srcContent, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatalf("source file unreadable %s: %v", srcPath, err)
			}

			mirrorContent, mirrorErr := os.ReadFile(mirrorPath)
			if mirrorErr != nil {
				if os.IsNotExist(mirrorErr) {
					t.Errorf(
						"RULE_TEMPLATE_MIRROR_DRIFT: source file %s has no mirror at %s; "+
							"run 'cp %s %s' and stage both files before commit",
						rel, mirrorPath, srcPath, mirrorPath,
					)
					return
				}
				t.Fatalf("mirror file unreadable %s: %v", mirrorPath, mirrorErr)
			}

			if !bytes.Equal(srcContent, mirrorContent) {
				t.Errorf(
					"RULE_TEMPLATE_MIRROR_DRIFT: source file %s differs from its mirror at %s "+
						"(source %d bytes, mirror %d bytes); run 'cp %s %s' and stage both files",
					rel, mirrorPath, len(srcContent), len(mirrorContent), srcPath, mirrorPath,
				)
			}
		})
	}
}

// TestLateBranchTemplateMirror verifies that every SPEC-V3R5-LATE-BRANCH-001 modified
// markdown file has a byte-identical mirror under internal/template/templates/.
//
// Promoted from Optional (REQ-LB-008) to Mandatory at run-phase per plan-auditor iter 1
// Q2 CRITICAL recommendation: existing workflowOptMirroredPaths allowlist did NOT cover
// the 5 LATE-BRANCH files (4 markdown + 1 yaml.tmpl), so AC-LB-005 was vacuous without
// this parallel test.
//
// Reports the literal sentinel RULE_TEMPLATE_MIRROR_DRIFT (shared with the workflow-opt
// test above) so CI log parsers match the same failure pattern.
//
// Note: the .yaml.tmpl mirror (`.moai/config/sections/git-strategy.yaml.tmpl`) contains
// Go template variables and is verified by AC-LB-005 yq+grep verification rather than
// byte equality. This test covers ONLY the 4 .md mirror pairs (M2/M3/M5 deliverables).
// spec-workflow.md is already in workflowOptMirroredPaths so it is NOT duplicated here.
//
// @MX:NOTE: [AUTO] AC-LB-005 (template mirror parity for LATE-BRANCH-001) is verified
// by this test for the 3 .md markdown files touched by D2/D3/D5.
func TestLateBranchTemplateMirror(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)

	for _, rel := range lateBranchMirroredPaths {
		rel := rel // capture
		t.Run(filepath.Base(rel), func(t *testing.T) {
			t.Parallel()

			srcPath := filepath.Join(projectRoot, rel)
			mirrorPath := filepath.Join(projectRoot, "internal", "template", "templates", rel)

			srcContent, err := os.ReadFile(srcPath)
			if err != nil {
				t.Fatalf("source file unreadable %s: %v", srcPath, err)
			}

			mirrorContent, mirrorErr := os.ReadFile(mirrorPath)
			if mirrorErr != nil {
				if os.IsNotExist(mirrorErr) {
					t.Errorf(
						"RULE_TEMPLATE_MIRROR_DRIFT: source file %s has no mirror at %s; "+
							"run 'cp %s %s' and stage both files before commit",
						rel, mirrorPath, srcPath, mirrorPath,
					)
					return
				}
				t.Fatalf("mirror file unreadable %s: %v", mirrorPath, mirrorErr)
			}

			if !bytes.Equal(srcContent, mirrorContent) {
				t.Errorf(
					"RULE_TEMPLATE_MIRROR_DRIFT: source file %s differs from its mirror at %s "+
						"(source %d bytes, mirror %d bytes); run 'cp %s %s' and stage both files",
					rel, mirrorPath, len(srcContent), len(mirrorContent), srcPath, mirrorPath,
				)
			}
		})
	}
}

// findProjectRootForMirrorTest locates the project root by walking upward from
// the current working directory until a go.mod file is found.
//
// This duplicates a small helper found in agent_askuser_audit_test.go to keep
// the mirror drift test self-contained and not coupled to the other audit's
// package-internal helpers.
func findProjectRootForMirrorTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}
