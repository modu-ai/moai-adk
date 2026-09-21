package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// jev_test.go — the shared Jev persistence seam (SPEC-JEV-OPTIN-MEASURE-001
// REQ-JEVO-002, AC-JEVO-002 / AC-JEVO-003).
//
// The seam exists so the `moai init` wizard and the `moai web` settings screen
// drive ONE writer. The console reaches it through the generic schema form
// (parseSchemaForm -> ApplySchemaEdits); the wizard reaches it through
// SetJevEnabled, which is a thin wrapper over the SAME ApplySchemaEdits call.
// A second writer would mean two nested-isolation implementations and two
// places to get the byte-identity property wrong.

// writeJevWorkflowFixture lays down a workflow.yaml carrying sibling fields the
// write must not disturb, and returns the project root.
func writeJevWorkflowFixture(t *testing.T, body string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return root
}

const jevFixtureWorkflow = `workflow:
    # a comment that must survive the write
    default_mode: ""
    agentic_loop:
        max_iterations: 7
    worktree:
        auto_create: true
        auto_cleanup: false
    jev:
        enabled: false
`

// TestJevEnabledFieldName_MatchesSchema: the exported field-name constant and
// the schema FieldDef name are the same string. A drift here would give the
// wizard a write target the console does not render.
func TestJevEnabledFieldName_MatchesSchema(t *testing.T) {
	f, ok := Field(JevEnabledField)
	if !ok {
		t.Fatalf("Field(%q) not found in the schema — the console cannot render it", JevEnabledField)
	}
	if f.Section != SectionWorkflow {
		t.Errorf("Section = %v, want SectionWorkflow", f.Section)
	}
	if f.Type != TypeBool {
		t.Errorf("Type = %v, want TypeBool", f.Type)
	}
	if f.Persist.Kind != PersistSeam {
		t.Errorf("Persist.Kind = %v, want PersistSeam (the workflow.yaml yamlpatch seam)", f.Persist.Kind)
	}
}

// TestSetJevEnabled_WritesThroughSeam: the wizard entrance persists the value.
func TestSetJevEnabled_WritesThroughSeam(t *testing.T) {
	root := writeJevWorkflowFixture(t, jevFixtureWorkflow)
	if err := SetJevEnabled(root, true); err != nil {
		t.Fatalf("SetJevEnabled: %v", err)
	}
	got := readJevWorkflowYAML(t, root)
	if !strings.Contains(got, "enabled: true") {
		t.Errorf("workflow.yaml did not record the enable:\n%s", got)
	}
}

// TestSetJevEnabled_SiblingFieldsByteIdentical is AC-JEVO-003: every line other
// than the targeted jev field is byte-identical after the write.
func TestSetJevEnabled_SiblingFieldsByteIdentical(t *testing.T) {
	root := writeJevWorkflowFixture(t, jevFixtureWorkflow)
	before := readJevWorkflowYAML(t, root)
	if err := SetJevEnabled(root, true); err != nil {
		t.Fatalf("SetJevEnabled: %v", err)
	}
	after := readJevWorkflowYAML(t, root)

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")
	if len(beforeLines) != len(afterLines) {
		t.Fatalf("line count changed: %d -> %d\nbefore:\n%s\nafter:\n%s",
			len(beforeLines), len(afterLines), before, after)
	}
	changed := 0
	for i := range beforeLines {
		if beforeLines[i] != afterLines[i] {
			changed++
			if !strings.Contains(afterLines[i], "enabled:") {
				t.Errorf("non-Jev line %d changed: %q -> %q", i+1, beforeLines[i], afterLines[i])
			}
		}
	}
	if changed != 1 {
		t.Errorf("changed line count = %d, want exactly 1 (the jev enabled line)\nbefore:\n%s\nafter:\n%s",
			changed, before, after)
	}
	// Positive control: the comparison above is only meaningful if the file
	// actually carries the sibling content it claims to have preserved.
	if !strings.Contains(after, "a comment that must survive the write") {
		t.Error("positive control failed: the sibling comment is absent, so the byte-identity assertion measured nothing")
	}
	if !strings.Contains(after, "max_iterations: 7") {
		t.Error("positive control failed: the sibling scalar is absent")
	}
}

// TestSetJevEnabled_AndConsoleEditAgree: the wizard wrapper and the console's
// generic edit map produce the SAME file bytes. This is the AC-JEVO-002
// one-writer assertion expressed behaviourally — if a parallel writer existed,
// the two paths could diverge.
func TestSetJevEnabled_AndConsoleEditAgree(t *testing.T) {
	viaWrapper := writeJevWorkflowFixture(t, jevFixtureWorkflow)
	if err := SetJevEnabled(viaWrapper, true); err != nil {
		t.Fatalf("SetJevEnabled: %v", err)
	}
	viaConsole := writeJevWorkflowFixture(t, jevFixtureWorkflow)
	if err := ApplySchemaEdits(viaConsole, map[string]string{JevEnabledField: "true"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}
	if a, b := readJevWorkflowYAML(t, viaWrapper), readJevWorkflowYAML(t, viaConsole); a != b {
		t.Errorf("the two entrances produced different bytes\nwizard:\n%s\nconsole:\n%s", a, b)
	}
}

func readJevWorkflowYAML(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"))
	if err != nil {
		t.Fatalf("read workflow.yaml: %v", err)
	}
	return string(b)
}
