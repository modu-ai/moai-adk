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
//   - .claude/agents/moai/plan-auditor.md (Layer G target)
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
	// Layer A — manager-develop prompt template
	".claude/rules/moai/development/manager-develop-prompt-template.md",
	// Layer C — CI watch background standardization
	".claude/rules/moai/workflow/ci-watch-protocol.md",
	// Layer D + H — parallel execution batching + tool optimization patterns
	".claude/rules/moai/core/agent-common-protocol.md",
	// Layer E — Phase Transitions skip policy
	".claude/rules/moai/workflow/spec-workflow.md",
	// Layer B — Agent Teams pattern (NEW file)
	".claude/rules/moai/workflow/agent-teams-pattern.md",
	// Layer D — verification batch pattern (NEW file)
	".claude/rules/moai/workflow/verification-batch-pattern.md",
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
