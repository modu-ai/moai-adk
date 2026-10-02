// audit_plan_doc_surface_test.go: doc-surface guard for the audit-plan
// activation text.
//
// The plan-auditor and sync-auditor agents, the cross-model audit skill and the
// sync workflow instruct the audit path by prose. These checks pin the literals
// that prose must carry, in BOTH the source tree and the template mirror, so an
// edit to one copy cannot silently drop the instruction from the other.
//
// Scope of what these checks prove: literal checks are tests-of-record for the
// TEXT only. A literal placed in a comment would satisfy them, and no string
// check shows that a model obeys an instruction. The mechanical backstops live
// elsewhere: the audit receipt guard (the SubagentStop corroboration of an
// auditor PASS) and the `moai verify audit-plan --result-file` checker, which
// compares the plan with the audit_multi result.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	auditPlanPlanAuditor = ".claude/agents/moai/plan-auditor.md"
	auditPlanSyncAuditor = ".claude/agents/moai/sync-auditor.md"
	auditPlanSkill       = ".claude/skills/moai-ref-cross-model-audit/SKILL.md"
	auditPlanSyncWf      = ".claude/skills/moai/workflows/sync.md"
)

// auditPlanCommonLiterals is the set every one of the four documents carries:
// the verb, the flags it is run with, the plan member the auditors branch on,
// the legacy-path signature and its Gap wording, and the blocking rule.
var auditPlanCommonLiterals = []string{
	"moai verify audit-plan",
	"--project-root",
	"--result-file",
	"config_status",
	"config_status: unreadable",
	"plan surface unreachable, legacy path used",
	"Shared diagnostic snapshot contract",
	"legacy path",
	"PASS-blocking",
	"without a `gates` argument",
	"fail-closed",
}

// auditPlanSpecificLiterals adds, per document, the literals only that
// document owns.
var auditPlanSpecificLiterals = map[string][]string{
	// plan-auditor writes the digest file itself, fresh, and passes only the path.
	auditPlanPlanAuditor: {
		".moai/state/audit-plan-result.json",
		"overwrite",
		"only the path",
		"convergence_check",
	},
	// sync-auditor is read-only: it returns the members and says its verdict is
	// not final until the orchestrator's check passes.
	auditPlanSyncAuditor: {
		"read-only",
		"not final until the orchestrator",
		"convergence_check",
	},
	auditPlanSkill: {
		".moai/state/audit-plan-result.json",
		"overwrite",
		"only the path",
		"convergence_check",
	},
	// sync.md owns the post-PASS audit_multi call, the both-must-pass rule and
	// the binding statement lines.
	auditPlanSyncWf: {
		"cross_model_required",
		"audit-plan --result-file",
		"audit_multi unreachable",
		"binding: no",
		"cross_model:",
		"audit_multi:",
		"gate_unmet:",
		"audit_receipt:",
		"plan_check:",
		"both must pass",
		".moai/state/audit-plan-result.json",
		"overwrite",
		"only the path",
		"not final until the orchestrator",
	},
}

// auditPlanStaleHeading is the old token-keyed block heading. The token-keyed
// prose survives only inside the paragraph labelled as the legacy path.
const auditPlanStaleHeading = "Single-backend audit mode (per the project"

func TestAuditPlanDocSurface(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)
	trees := map[string]string{
		"source":   projectRoot,
		"template": filepath.Join(projectRoot, "internal", "template", "templates"),
	}

	for _, rel := range []string{auditPlanPlanAuditor, auditPlanSyncAuditor, auditPlanSkill, auditPlanSyncWf} {
		for tree, base := range trees {
			rel, tree, base := rel, tree, base
			t.Run(filepath.Base(rel)+"/"+tree, func(t *testing.T) {
				t.Parallel()

				path := filepath.Join(base, rel)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				body := string(data)

				want := append([]string{}, auditPlanCommonLiterals...)
				want = append(want, auditPlanSpecificLiterals[rel]...)
				for _, lit := range want {
					if !strings.Contains(body, lit) {
						t.Errorf("%s does not carry the literal %q", path, lit)
					}
				}

				if strings.Contains(body, auditPlanStaleHeading) {
					t.Errorf("%s still carries the token-keyed block heading %q", path, auditPlanStaleHeading)
				}
			})
		}
	}
}

// TestAuditPlanDocSurfaceKeepsClosureMarkers pins that the token-keyed block
// removal did not take the contract-mode second-review markers with it: the
// source sync-auditor keeps the start and end markers exactly once each.
func TestAuditPlanDocSurfaceKeepsClosureMarkers(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRootForMirrorTest(t)
	data, err := os.ReadFile(filepath.Join(projectRoot, auditPlanSyncAuditor))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, marker := range []string{
		"<!-- moai:closure-second-review:start -->",
		"<!-- moai:closure-second-review:end -->",
	} {
		if got := strings.Count(body, marker); got != 1 {
			t.Errorf("%s carries %q %d times, want exactly 1", auditPlanSyncAuditor, marker, got)
		}
	}
}
