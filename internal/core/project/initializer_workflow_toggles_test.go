package project

// initializer_workflow_toggles_test.go — SPEC-INIT-WIZARD-REPAIR-001 chain ②
// (REQ-004/005/006): the tracker-gated workflow.yaml toggle writer. Only keys
// whose *Set tracker is true are written; an all-false tracker set leaves the
// deployed file byte-identical (distributed-default preservation).

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
	"gopkg.in/yaml.v3"
)

// templateLikeWorkflow mirrors the shipped workflow.yaml shape (4-space
// indent, worktree block present, NO branch_guard block).
const templateLikeWorkflow = `workflow:
    default_mode: ""
    execution_mode: auto
    worktree:
        auto_create: false
        auto_merge: false
        auto_cleanup: false
        tmux_preferred: true
    token_budget:
        plan: 30000
        run: 180000
`

// writeWorkflowFixture writes templateLikeWorkflow into a temp sectionsDir and
// returns the dir + workflow.yaml path.
func writeWorkflowFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sectionsDir, defs.WorkflowYAML)
	if err := os.WriteFile(path, []byte(templateLikeWorkflow), 0o644); err != nil {
		t.Fatal(err)
	}
	return sectionsDir, path
}

// TestWriteWorkflowTogglesYAML_AllTrackersFalseIsByteIdentical pins REQ-006's
// byte-identity clause: with every *Set tracker false the file is not touched.
func TestWriteWorkflowTogglesYAML_AllTrackersFalseIsByteIdentical(t *testing.T) {
	sectionsDir, path := writeWorkflowFixture(t)

	opts := InitOptions{ProjectRoot: filepath.Dir(filepath.Dir(sectionsDir))}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, []byte(templateLikeWorkflow)) {
		t.Errorf("all-false trackers must leave workflow.yaml byte-identical;\nwant:\n%s\ngot:\n%s", templateLikeWorkflow, got)
	}
}

// TestWriteWorkflowTogglesYAML_PatchesExistingWorktreeKeys pins REQ-006's
// persist clause for the three keys the template already carries: the patched
// leaf keeps its original indentation and every neighbor byte survives.
func TestWriteWorkflowTogglesYAML_PatchesExistingWorktreeKeys(t *testing.T) {
	sectionsDir, path := writeWorkflowFixture(t)

	root := filepath.Dir(filepath.Dir(sectionsDir))
	opts := InitOptions{
		ProjectRoot:            root,
		WorktreeAutoCreateSet:  true,
		WorktreeAutoCreate:     true,
		WorktreeAutoMergeSet:   true,
		WorktreeAutoMerge:      true,
		WorktreeAutoCleanupSet: true,
		WorktreeAutoCleanup:    true,
	}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, _ := os.ReadFile(path)
	for _, want := range []string{
		"        auto_create: true",
		"        auto_merge: true",
		"        auto_cleanup: true",
		"        tmux_preferred: true",
		"    token_budget:",
	} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("workflow.yaml missing %q; got:\n%s", want, got)
		}
	}
}

// TestWriteWorkflowTogglesYAML_InsertsBranchGuardWhenAbsent pins the
// fresh-key insert: the template ships no branch_guard block, so an explicit
// --branch-guard must insert one under workflow: without disturbing neighbors.
func TestWriteWorkflowTogglesYAML_InsertsBranchGuardWhenAbsent(t *testing.T) {
	sectionsDir, path := writeWorkflowFixture(t)

	root := filepath.Dir(filepath.Dir(sectionsDir))
	opts := InitOptions{
		ProjectRoot:       root,
		BranchGuardSet:    true,
		BranchGuardEnabled: true,
	}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Contains(got, []byte("branch_guard:")) || !bytes.Contains(got, []byte("enabled: true")) {
		t.Errorf("workflow.yaml missing inserted branch_guard block; got:\n%s", got)
	}
	// Neighbors survive the insert.
	for _, want := range []string{"    default_mode: \"\"", "        auto_create: false", "    token_budget:"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("workflow.yaml lost neighbor %q during insert; got:\n%s", want, got)
		}
	}
	// Nesting guard (boundary check): sibling keys of one mapping must share
	// indentation. Parse the result and assert branch_guard landed as a child
	// of workflow AND the pre-existing siblings kept their paths — a
	// mismatched-indent insert would re-parent every later workflow child.
	var doc map[string]any
	if err := yaml.Unmarshal(got, &doc); err != nil {
		t.Fatalf("inserted document no longer parses: %v\ngot:\n%s", err, got)
	}
	wf, _ := doc["workflow"].(map[string]any)
	if wf == nil {
		t.Fatalf("workflow mapping lost after insert; got:\n%s", got)
	}
	if bg, _ := wf["branch_guard"].(map[string]any); bg == nil || bg["enabled"] != true {
		t.Errorf("branch_guard must parse as workflow.branch_guard.enabled=true; got:\n%s", got)
	}
	if wf["default_mode"] != "" || wf["execution_mode"] != "auto" {
		t.Errorf("pre-existing workflow children changed value or nesting; got:\n%s", got)
	}
	if tb, _ := wf["token_budget"].(map[string]any); tb == nil || tb["plan"] != 30000 {
		t.Errorf("token_budget must stay a direct workflow child; got:\n%s", got)
	}
}

// TestWriteWorkflowTogglesYAML_OnlySetKeysArePatched pins REQ-004's
// tracker gating at the writer: an unset sibling's leaf is left untouched
// even while another toggle is patched.
func TestWriteWorkflowTogglesYAML_OnlySetKeysArePatched(t *testing.T) {
	sectionsDir, path := writeWorkflowFixture(t)

	root := filepath.Dir(filepath.Dir(sectionsDir))
	opts := InitOptions{
		ProjectRoot:           root,
		WorktreeAutoCreateSet: true,
		WorktreeAutoCreate:    true,
		// merge + cleanup trackers deliberately false.
	}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Contains(got, []byte("auto_create: true")) {
		t.Errorf("set key not patched; got:\n%s", got)
	}
	if !bytes.Contains(got, []byte("auto_merge: false")) || !bytes.Contains(got, []byte("auto_cleanup: false")) {
		t.Errorf("unset keys must keep the template default; got:\n%s", got)
	}
}

// TestWriteWorkflowTogglesYAML_ExplicitFalsePersisted pins acceptance.md
// §D.3: an explicit --worktree-auto-create=false is a selection, not an
// absence — the tracker fires and `false` is persisted.
func TestWriteWorkflowTogglesYAML_ExplicitFalsePersisted(t *testing.T) {
	sectionsDir, path := writeWorkflowFixture(t)

	// Start from a file where auto_create is true, so persisting false is a
	// visible change.
	changed := bytes.Replace([]byte(templateLikeWorkflow), []byte("auto_create: false"), []byte("auto_create: true"), 1)
	if err := os.WriteFile(path, changed, 0o644); err != nil {
		t.Fatal(err)
	}

	root := filepath.Dir(filepath.Dir(sectionsDir))
	opts := InitOptions{
		ProjectRoot:           root,
		WorktreeAutoCreateSet: true,
		WorktreeAutoCreate:    false,
	}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !bytes.Contains(got, []byte("auto_create: false")) {
		t.Errorf("explicit false must be persisted; got:\n%s", got)
	}
}

// TestWriteWorkflowTogglesYAML_FreshFileFallback pins the no-deployer path
// (acceptance.md §D.3): with no workflow.yaml present, a tracked selection
// creates a minimal workflow block instead of failing.
func TestWriteWorkflowTogglesYAML_FreshFileFallback(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{
		ProjectRoot:       root,
		BranchGuardSet:    true,
		BranchGuardEnabled: true,
	}
	result := &InitResult{}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, result); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(sectionsDir, defs.WorkflowYAML))
	if err != nil {
		t.Fatalf("fresh workflow.yaml not created: %v", err)
	}
	if !bytes.Contains(got, []byte("enabled: true")) {
		t.Errorf("fresh workflow.yaml missing branch_guard; got:\n%s", got)
	}
	found := false
	for _, f := range result.CreatedFiles {
		if f == filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML) {
			found = true
		}
	}
	if !found {
		t.Errorf("fresh file must be recorded in CreatedFiles; got %v", result.CreatedFiles)
	}
}

// TestWriteWorkflowTogglesYAML_FreshFileAllTrackersFalseSkips pins the
// fallback-path byte-identity: no tracker set + no file ⇒ no file synthesized.
func TestWriteWorkflowTogglesYAML_FreshFileAllTrackersFalseSkips(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	opts := InitOptions{ProjectRoot: root}
	if err := WriteWorkflowTogglesYAML(sectionsDir, opts, &InitResult{}); err != nil {
		t.Fatalf("WriteWorkflowTogglesYAML: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sectionsDir, defs.WorkflowYAML)); err == nil {
		t.Error("no tracker set: workflow.yaml must not be synthesized on the fallback path")
	}
}
