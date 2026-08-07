package cli

// SPEC-WT-DOC-001 (branch-guard config surface) tests for the init-flag wiring
// (Surface 2) and the update-wizard workflow step (Surface 3). Both surfaces
// persist four workflow toggles (branch_guard.enabled + worktree.auto_create /
// auto_merge / auto_cleanup) into .moai/config/sections/workflow.yaml via
// opt-in semantics: zero-value flags MUST NOT clobber the deployed template
// default (CLAUDE.local.md §22.9 — distributed template ships default-off).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
	"github.com/spf13/cobra"
)

// TestInitCmd_WorkflowBranchGuardFlagsRegistered verifies the four
// SPEC-WT-DOC-001 workflow flags are registered on the init command.
func TestInitCmd_WorkflowBranchGuardFlagsRegistered(t *testing.T) {
	for _, name := range []string{
		"branch-guard",
		"worktree-auto-create",
		"worktree-auto-merge",
		"worktree-auto-cleanup",
	} {
		if initCmd.Flags().Lookup(name) == nil {
			t.Errorf("init command should have --%s flag", name)
		}
	}
}

// TestApplyWorkflowBranchGuardFlags_ExplicitSet flips the Set trackers only
// for flags the user passed on the command line — an explicit false must still
// mark the value Set so the writer persists it (the "explicit false vs unset"
// distinction is the contract).
func TestApplyWorkflowBranchGuardFlags_ExplicitSet(t *testing.T) {
	cmd := makeInitCmdWithFlags(map[string]string{
		"branch-guard":           "true",
		"worktree-auto-create":   "false",
		"worktree-auto-merge":    "true",
		"worktree-auto-cleanup":  "false",
	})
	opts := project.InitOptions{}
	applyWorkflowBranchGuardFlags(cmd, &opts)

	if !opts.BranchGuardSet || !opts.BranchGuardEnabled {
		t.Errorf("branch-guard true not applied: %+v", opts)
	}
	if !opts.WorktreeAutoCreateSet || opts.WorktreeAutoCreate {
		t.Errorf("worktree-auto-create false not applied: %+v", opts)
	}
	if !opts.WorktreeAutoMergeSet || !opts.WorktreeAutoMerge {
		t.Errorf("worktree-auto-merge true not applied: %+v", opts)
	}
	if !opts.WorktreeAutoCleanupSet || opts.WorktreeAutoCleanup {
		t.Errorf("worktree-auto-cleanup false not applied: %+v", opts)
	}
}

// TestApplyWorkflowBranchGuardFlags_UnsetFlagsStayUnset verifies the contract
// that powers the distributed-default preservation: when none of the four
// workflow flags is passed, every Set tracker stays false and the writer is a
// no-op (zero-value clobber defense).
func TestApplyWorkflowBranchGuardFlags_UnsetFlagsStayUnset(t *testing.T) {
	cmd := makeInitCmdWithFlags(nil)
	opts := project.InitOptions{
		BranchGuardEnabled:  true, // pretend a prior step touched this
		WorktreeAutoCreate:  true,
		WorktreeAutoMerge:   true,
		WorktreeAutoCleanup: true,
	}
	applyWorkflowBranchGuardFlags(cmd, &opts)

	if opts.BranchGuardSet || opts.WorktreeAutoCreateSet ||
		opts.WorktreeAutoMergeSet || opts.WorktreeAutoCleanupSet {
		t.Errorf("unset flags must not flip trackers: %+v", opts)
	}
}

// makeInitCmdWithFlags clones initCmd's flag set so a test can set values
// without mutating the shared command, then returns the clone. Flags whose
// Changed() must be true are set via Flags().Set; absent flags remain in their
// default (Changed() == false) state, mirroring the real CLI parsing path.
func makeInitCmdWithFlags(values map[string]string) *cobra.Command {
	clone := &cobra.Command{Use: "init"}
	// Register the four workflow flags with the same defaults as init.go.
	clone.Flags().Bool("branch-guard", false, "")
	clone.Flags().Bool("worktree-auto-create", false, "")
	clone.Flags().Bool("worktree-auto-merge", false, "")
	clone.Flags().Bool("worktree-auto-cleanup", false, "")
	for name, val := range values {
		_ = clone.Flags().Set(name, val)
	}
	return clone
}

// TestRunWorkflowConfigStep_NonTerminalSkips verifies the step is a clean
// no-op when stdin is not a TTY (the reconfigure wizard's CI / non-interactive
// path) — workflow.yaml must remain byte-identical.
func TestRunWorkflowConfigStep_NonTerminalSkips(t *testing.T) {
	root := t.TempDir()
	sectionsDir := filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowPath := filepath.Join(sectionsDir, defs.WorkflowYAML)
	before := []byte("workflow:\n    worktree:\n        auto_create: false\n")
	if err := os.WriteFile(workflowPath, before, 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runWorkflowConfigStep(&out, root); err != nil {
		t.Fatalf("runWorkflowConfigStep non-tty: %v", err)
	}
	got, _ := os.ReadFile(workflowPath)
	if !bytes.Equal(got, before) {
		t.Errorf("non-tty path mutated workflow.yaml:\nwant %q\ngot  %q", before, got)
	}
}

// TestRunWorkflowConfigStep_AbsentWorkflowYAMLSkips verifies the step is a
// no-op when workflow.yaml is absent (corrupted project) — the step must not
// synthesize one.
func TestRunWorkflowConfigStep_AbsentWorkflowYAMLSkips(t *testing.T) {
	root := t.TempDir()
	// No workflow.yaml created.
	var out bytes.Buffer
	if err := runWorkflowConfigStep(&out, root); err != nil {
		t.Fatalf("runWorkflowConfigStep absent file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML)); !os.IsNotExist(err) {
		t.Error("absent-file path unexpectedly created workflow.yaml")
	}
}

// TestBuildWorkflowToggleEdits verifies the edit-builder produces a KeyEdit
// ONLY for keys whose value the user changed. Unchanged keys are omitted so
// yamlpatch leaves their bytes (and surrounding comments) untouched.
func TestBuildWorkflowToggleEdits(t *testing.T) {
	cases := []struct {
		name                          string
		curBranch, newBranch         bool
		curCreate, newCreate         bool
		curMerge, newMerge           bool
		curCleanup, newCleanup       bool
		wantPaths                     []string
	}{
		{
			name:      "no changes -> empty",
			wantPaths: nil,
		},
		{
			name:       "only branch_guard toggled",
			newBranch:  true,
			wantPaths:  []string{"workflow.branch_guard.enabled"},
		},
		{
			name:       "all four toggled",
			curBranch:  true,
			curCreate:  true,
			curMerge:   true,
			curCleanup: true,
			wantPaths: []string{
				"workflow.branch_guard.enabled",
				"workflow.worktree.auto_create",
				"workflow.worktree.auto_merge",
				"workflow.worktree.auto_cleanup",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			edits := buildWorkflowToggleEdits(
				tc.curBranch, tc.curCreate, tc.curMerge, tc.curCleanup,
				tc.newBranch, tc.newCreate, tc.newMerge, tc.newCleanup,
			)
			gotPaths := make([]string, 0, len(edits))
			for _, e := range edits {
				gotPaths = append(gotPaths, strings.Join(e.Path, "."))
			}
			if len(gotPaths) != len(tc.wantPaths) {
				t.Fatalf("edit count = %d (%v), want %d (%v)", len(gotPaths), gotPaths, len(tc.wantPaths), tc.wantPaths)
			}
			for i, want := range tc.wantPaths {
				if gotPaths[i] != want {
					t.Errorf("edits[%d] = %q, want %q", i, gotPaths[i], want)
				}
			}
		})
	}
}

// TestReadWorkflowToggleDefaults covers the parser: present keys return their
// values, absent keys (e.g. branch_guard in the distributed template) return
// false without error.
func TestReadWorkflowToggleDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), defs.WorkflowYAML)
	content := []byte("workflow:\n    worktree:\n        auto_create: false\n        auto_merge: true\n        auto_cleanup: true\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	bg, create, merge, cleanup, err := readWorkflowToggleDefaults(path)
	if err != nil {
		t.Fatalf("readWorkflowToggleDefaults: %v", err)
	}
	if bg || create || !merge || !cleanup {
		t.Errorf("defaults = (bg=%v create=%v merge=%v cleanup=%v), want (false false true true)",
			bg, create, merge, cleanup)
	}
}

// TestApplyWorkflowConfigEditsViaYamlpatch is a focused integration test for
// Surface 3's persistence path — it exercises the yamlpatch.PatchFile seam
// directly with the edits the wizard would produce, to lock down the
// comment-preserving upsert that the wizard depends on.
func TestApplyWorkflowConfigEditsViaYamlpatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), defs.WorkflowYAML)
	before := []byte("# comment header\nworkflow:\n    # worktree comment\n    worktree:\n        auto_create: false\n        auto_merge: true\n")
	if err := os.WriteFile(path, before, 0644); err != nil {
		t.Fatal(err)
	}

	edits := []yamlpatch.KeyEdit{
		{Path: []string{"workflow", "branch_guard", "enabled"}, Value: "true"},
		{Path: []string{"workflow", "worktree", "auto_cleanup"}, Value: "false"},
	}
	if err := yamlpatch.PatchFile(path, edits); err != nil {
		t.Fatalf("PatchFile: %v", err)
	}
	got, _ := os.ReadFile(path)
	gotStr := string(got)
	for _, want := range []string{"# comment header", "# worktree comment", "auto_merge: true", "branch_guard:", "enabled: true", "auto_cleanup: false"} {
		if !strings.Contains(gotStr, want) {
			t.Errorf("workflow.yaml missing %q after patch; got:\n%s", want, gotStr)
		}
	}
}
