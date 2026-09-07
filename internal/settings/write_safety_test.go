package settings

// SPEC-WEB-WRITE-SAFETY-001 — reduced unit tests for the two first-measurement
// conclusions (M-a/M-b), serving as RED-first evidence for AC-WWS-003/004/005
// and the AC-WWS-004 positive control.
//
// Measured defect signatures (M1(d), worktree t517, pre-fix tree):
//   - feedback.yaml lost its blank line after a VALUE-IDENTICAL save (seam
//     PatchFile re-encodes the whole document; yaml.v3 normalizes blank lines).
//   - git-strategy.yaml was re-marshaled (key reorder + zero-value keys added)
//     after a VALUE-IDENTICAL save (applyTypedEdits calls SetSection
//     unconditionally, which sets gitStrategyDirty and forces Save() to rewrite).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/settings/yamlpatch"
)

// TestPatchFileValueInvariantPreservesBytes carries AC-WWS-005 (RED-first): a
// seam patch whose new value equals the persisted value must leave the file
// byte-identical — blank lines included. RED on the defective tree: PatchFile
// re-encodes the document through the yaml.v3 encoder, which normalizes blank
// lines away (the documented limitation in the yamlpatch package header).
func TestPatchFileValueInvariantPreservesBytes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "feedback")
	path := filepath.Join(root, ".moai", "config", "sections", "feedback.yaml")

	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
		{Path: []string{"feedback", "repository"}, Value: "modu-ai/moai-adk"}, // identical value
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}

	after := readSection(t, root, "feedback")
	if after != before {
		t.Errorf("value-invariant seam patch rewrote the file — presentation elements lost\n--- before ---\n%s\n--- after ---\n%s", before, after)
	}
}

// TestPatchFileScalarChangePreservesPresentation carries AC-WWS-005 (RED-first)
// for the value-CHANGING case: when a save does replace a scalar, the untouched
// presentation elements — blank lines, comments, key order, unknown keys — must
// survive byte-for-byte around the one changed line.
func TestPatchFileScalarChangePreservesPresentation(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	before := seedSectionFixture(t, root, "feedback")
	path := filepath.Join(root, ".moai", "config", "sections", "feedback.yaml")

	err := yamlpatch.PatchFile(path, []yamlpatch.KeyEdit{
		{Path: []string{"feedback", "repository"}, Value: "someone/else-fork"},
	})
	if err != nil {
		t.Fatalf("PatchFile: %v", err)
	}

	after := readSection(t, root, "feedback")
	if !strings.Contains(after, "repository: someone/else-fork") {
		t.Errorf("scalar not updated:\n%s", after)
	}
	if got, want := blankLineCount(after), blankLineCount(before); got != want {
		t.Errorf("blank lines not preserved: before=%d after=%d\n%s", want, got, after)
	}
	if got, want := sectionCommentLines(after), sectionCommentLines(before); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Error("comments not preserved by seam routing")
	}
}

// TestApplySchemaEditsValueInvariantTouchesNothing carries AC-WWS-003 and
// AC-WWS-004 (RED-first): submitting values that equal the persisted values for
// BOTH a typed section (git_strategy.mode) and a seam section
// (feedback.repository) must rewrite NEITHER file. RED on the defective tree:
// applyTypedEdits calls SetSection("git_strategy") unconditionally (raising
// gitStrategyDirty, so Save() re-marshals git-strategy.yaml), and
// ApplySchemaEdits routes the feedback edit to PatchFile (blank-line
// normalization).
func TestApplySchemaEditsValueInvariantTouchesNothing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	feedbackBefore := seedSectionFixture(t, root, "feedback")

	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	gsBefore, err := os.ReadFile(filepath.Join(sectionsDir, "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	err = ApplySchemaEdits(root, map[string]string{
		"git_strategy.mode":    "team",               // fixture's persisted value — value-invariant
		"feedback.repository":  "modu-ai/moai-adk",   // fixture's persisted value — value-invariant
	})
	if err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}

	gsAfter, err := os.ReadFile(filepath.Join(sectionsDir, "git-strategy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(gsBefore) != string(gsAfter) {
		t.Errorf("value-invariant save rewrote git-strategy.yaml (dirty-gate bypassed):\n--- before ---\n%s\n--- after ---\n%s", gsBefore, gsAfter)
	}
	feedbackAfter := readSection(t, root, "feedback")
	if feedbackAfter != feedbackBefore {
		t.Errorf("value-invariant save rewrote feedback.yaml (seam presentation lost)\n--- before ---\n%s\n--- after ---\n%s", feedbackBefore, feedbackAfter)
	}
}

// TestApplySchemaEditsGitStrategyRealChangeStillRewrites is the AC-WWS-004
// POSITIVE CONTROL (mandatory): the same environment where the value-invariant
// save above must hold — with a genuinely changed value — must still rewrite
// git-strategy.yaml. Without this control a green on the value-invariant test
// is indistinguishable from "the gate never ran at all".
func TestApplySchemaEditsGitStrategyRealChangeStillRewrites(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	seedTypedFixtures(t, root, "git-strategy", "llm", "quality")
	gsPath := filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")
	before, err := os.ReadFile(gsPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ApplySchemaEdits(root, map[string]string{"git_strategy.mode": "personal"}); err != nil {
		t.Fatalf("ApplySchemaEdits: %v", err)
	}

	after, err := os.ReadFile(gsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Error("positive control failed: a real git_strategy change did NOT rewrite git-strategy.yaml (gate is inert)")
	}
	if !strings.Contains(string(after), "mode: personal") {
		t.Errorf("changed value not persisted:\n%s", after)
	}
}

func blankLineCount(s string) int {
	count := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "" {
			count++
		}
	}
	return count
}
