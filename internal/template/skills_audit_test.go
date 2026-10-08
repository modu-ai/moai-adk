package template

import (
	"io/fs"
	"strings"
	"testing"
)

// TestSkillsContainPlanAuditGateMarkers verifies that the workflow skill files
// modified by SPEC-WF-AUDIT-GATE-001 contain the required Plan Audit Gate markers.
// (The team run.md case was removed when the Agent Teams static-layer skills were
// retired by SPEC-AGENT-TEAM-RETIRE-001.)
//
// Required patterns per tasks.md T-06:
// (a) "Phase 1: Plan Audit Gate" header
// (b) "plan-auditor" keyword
// (c) "--skip-audit" keyword
// (d) "INCONCLUSIVE" keyword
// (e) ".moai/reports/plan-audit/" path
//
// Card t1586 split note: for spec-workflow.md the audit intent (the deployed
// ruleset documents gate + verdict routing + report streams) is preserved
// across the core+companion pair — patterns (a)(b)(d) stay in the core stub,
// pattern (e) moved to the lazy companion with the Report Persistence body.
//
// Source: SPEC-WF-AUDIT-GATE-001 T-06
func TestSkillsContainPlanAuditGateMarkers(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Table-driven: each entry specifies a skill file and the required patterns.
	tests := []struct {
		// name is a human-readable label for the test.
		name     string
		filePath string
		// requiredPatterns is the list of strings that MUST appear in the file.
		requiredPatterns []string
	}{
		{
			// AC-WAG-01: solo run.md must invoke plan-auditor before Phase 1.
			// AC-WAG-02, 03, 04, 06, 07, 08, 09: gate body must document full verdict routing.
			name:     "solo run/phase-execution.md — plan audit gate markers",
			filePath: ".claude/skills/moai/workflows/run/phase-execution.md",
			requiredPatterns: []string{
				"Phase 1: Plan Audit Gate",
				"plan-auditor",
				"--skip-audit",
				"INCONCLUSIVE",
				".moai/reports/plan-audit/",
			},
		},
		{
			// plan.md must declare audit-ready signal at workflow completion.
			name:     "plan.md — audit-ready signal",
			filePath: ".claude/skills/moai/workflows/plan.md",
			requiredPatterns: []string{
				"plan_complete_at",
				"audit-ready",
			},
		},
		{
			// spec-workflow.md must document Phase 1 and its verdicts.
			// Card t1586 core+detail split: the gate header, the plan-auditor
			// invocation, and the INCONCLUSIVE verdict routing stay in the core
			// stub; the run-gate report-stream path moved to the detail companion
			// (next entry) with the Report Persistence body it belongs to.
			name:     "spec-workflow.md — Phase 1 documentation",
			filePath: ".claude/rules/moai/workflow/spec-workflow.md",
			requiredPatterns: []string{
				"Phase 1: Plan Audit Gate",
				"plan-auditor",
				"INCONCLUSIVE",
			},
		},
		{
			// spec-workflow-detail.md (card t1586 lazy companion of
			// spec-workflow.md) owns Report Persistence — the two plan-audit
			// report streams and the run-gate record directory.
			name:     "spec-workflow-detail.md — Report Persistence stream path",
			filePath: ".claude/rules/moai/workflow/spec-workflow-detail.md",
			requiredPatterns: []string{
				".moai/reports/plan-audit/",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, tt.filePath)
			if readErr != nil {
				t.Fatalf("ReadFile(%q): %v", tt.filePath, readErr)
			}

			content := string(data)

			for _, pattern := range tt.requiredPatterns {
				if !strings.Contains(content, pattern) {
					t.Errorf("file %q missing required pattern: %q", tt.filePath, pattern)
				}
			}
		})
	}

	t.Logf("audited %d skill files for Plan Audit Gate markers", len(tests))
}

// TestReportsDirGitkeepExists verifies that the plan-audit report directory
// is tracked in the template tree via a .gitkeep file.
//
// Source: SPEC-WF-AUDIT-GATE-001 T-01, AC-WAG-10
func TestReportsDirGitkeepExists(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const gitkeepPath = ".moai/reports/plan-audit/.gitkeep"

	data, readErr := fs.ReadFile(fsys, gitkeepPath)
	if readErr != nil {
		t.Fatalf("missing %q in embedded templates: %v", gitkeepPath, readErr)
	}

	// Verify .gitkeep has the SPEC reference comment.
	if !strings.Contains(string(data), "SPEC-WF-AUDIT-GATE-001") {
		t.Errorf(".gitkeep at %q should reference SPEC-WF-AUDIT-GATE-001", gitkeepPath)
	}
}
