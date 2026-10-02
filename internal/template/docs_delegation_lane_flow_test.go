package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// docs_delegation_lane_flow_test.go: content guard for the reconciliation
// procedure of an isolated specialist spawn.
//
// The procedure section in kanban-dispatch-mechanics.md and the duty
// sentence in agent-common-protocol.md are documentation-only artifacts —
// exactly the kind of content a rule diet, split, or compression deletes
// silently. This guard fails when the section or any of its mandatory step
// markers disappears, from either copy of each rule (the local .claude/rules
// tree and the embedded template tree).
//
// The negative scan (forbidden cross-tree git redirect forms) is scoped to
// the extracted section: the file as a whole quotes measured guard-refusal
// shapes elsewhere, so a whole-file scan could never reach zero and would be
// an impossible-direction check. The positive assertions above it are its
// control — a file missing the section fails them before the negative scan
// could pass vacuously on empty input.

const reconciliationSectionHeading = "### Reconciling an isolated specialist spawn"

// One marker per mandatory step of the procedure, plus the exception
// trigger/record structure and the close contract the section must name.
var reconciliationMechanicsMarkers = []string{
	"Verify the landing",                   // step 1 — post-spawn landing verification
	"Name the branch",                      // step 2 — the WT- rename duty
	"WT-<slug>",                            // step 2 — the prefix itself
	"git merge --ff-only",                  // step 3 — the sanctioned merge form
	"Harvest the evidence before disposal", // step 4 — gitignored evidence
	"single sync commit",                   // step 5 — merged as-is
	"3-phase close",                        // step 5 — the close contract named
	"ownership exception",                  // exception named
	"structurally",                         // trigger — structurally twice + refusal evidence
}

// The duty sentence in the always-loaded rule: the runtime-decision
// obligation and the pointer into the procedure body.
var reconciliationProtocolMarkers = []string{
	"runtime decision",
	"Reconciling an isolated specialist spawn",
}

// Forbidden inside the procedure section: the cross-tree git redirect
// forms. The section must teach the plain-git-in-the-lane-tree form only —
// the worktree guard refuses these forms, and repeating them to an isolated
// agent is the failure the procedure exists to repair.
var reconciliationForbiddenForms = []string{
	"git -C",
	"--git-dir",
}

func TestReconciliationProcedureDocumented(t *testing.T) {
	t.Parallel()

	// Test binaries run with cwd = internal/template (the package dir), so
	// the template copy is "templates/..." and the local rules tree is
	// two levels up — the same relative anchors templatesRoot uses.
	files := []struct {
		label string
		path  string
		kind  string // "mechanics" | "protocol"
	}{
		{"local", filepath.Join("..", "..", ".claude", "rules", "moai", "workflow", "kanban-dispatch-mechanics.md"), "mechanics"},
		{"template", filepath.Join("templates", ".claude", "rules", "moai", "workflow", "kanban-dispatch-mechanics.md"), "mechanics"},
		{"local", filepath.Join("..", "..", ".claude", "rules", "moai", "core", "agent-common-protocol.md"), "protocol"},
		{"template", filepath.Join("templates", ".claude", "rules", "moai", "core", "agent-common-protocol.md"), "protocol"},
	}

	for _, f := range files {
		data, err := os.ReadFile(f.path)
		if err != nil {
			t.Errorf("%s: ReadFile(%q) error: %v — rule copy missing", f.label, f.path, err)
			continue
		}
		switch f.kind {
		case "mechanics":
			checkMechanicsSection(t, f.label, string(data))
		case "protocol":
			checkProtocolDutySentence(t, f.label, string(data))
		}
	}
}

// checkMechanicsSection asserts the procedure section exists, carries every
// mandatory step marker, and teaches no forbidden cross-tree form.
func checkMechanicsSection(t *testing.T, label, content string) {
	t.Helper()

	start := strings.Index(content, reconciliationSectionHeading)
	if start < 0 {
		t.Errorf("%s: reconciliation procedure section %q missing from kanban-dispatch-mechanics.md — the mandatory steps it carries (landing verification, WT- rename, ff-only merge, evidence harvest, run-to-sync adjacency) have nowhere to live", label, reconciliationSectionHeading)
		return
	}

	// The section body runs from its heading to the next markdown heading.
	rest := content[start+len(reconciliationSectionHeading):]
	body := rest
	if end := strings.Index(rest, "\n## "); end >= 0 {
		body = rest[:end]
	}

	for _, marker := range reconciliationMechanicsMarkers {
		if !strings.Contains(body, marker) {
			t.Errorf("%s: reconciliation procedure section lost a mandatory step marker %q (kanban-dispatch-mechanics.md)", label, marker)
		}
	}

	for _, forbidden := range reconciliationForbiddenForms {
		if strings.Contains(body, forbidden) {
			t.Errorf("%s: reconciliation procedure section teaches a forbidden cross-tree form %q — the procedure must instruct plain git inside the lane tree only", label, forbidden)
		}
	}
}

// checkProtocolDutySentence asserts the always-loaded rule carries the
// runtime-decision obligation and the pointer into the procedure body.
func checkProtocolDutySentence(t *testing.T, label, content string) {
	t.Helper()

	for _, marker := range reconciliationProtocolMarkers {
		if !strings.Contains(content, marker) {
			t.Errorf("%s: agent-common-protocol.md lost the reconciliation duty marker %q — the lane-session paragraph must carry the runtime-decision obligation and the pointer into the procedure body", label, marker)
		}
	}
}
