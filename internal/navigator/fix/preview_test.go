package fix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =====================================================================
// UnifiedDiff — AC-NS5-008b (Go-engine half)
//
// The Go engine produces ONE *.patch unified-diff file per stale subtree at
// fix-drafts/<draft-id>/draft/<doc-surface>.patch (design.md §D.1). The
// AskUserQuestion approval-gate preview field consumes these; the orchestrator
// truncates to ~12 lines for the preview pane. The Go engine writes the FULL
// patch (no truncation).
// =====================================================================

// TestUnifiedDiff_Structure verifies a fixture old/new pair produces a valid
// unified-diff: --- / +++ headers, at least one @@ hunk header, and both
// removed (-) and added (+) lines.
func TestUnifiedDiff_Structure(t *testing.T) {
	t.Parallel()

	oldText := "line one\nline two\nline three\nline four\n"
	newText := "line one\nline TWO changed\nline three\nline four\nline five\n"

	patch := UnifiedDiff("capability-map.md", oldText, newText)

	if patch == "" {
		t.Fatal("UnifiedDiff returned empty string for a changed pair")
	}
	if !strings.Contains(patch, "--- a/capability-map.md") {
		t.Errorf("patch missing `--- a/` header; patch:\n%s", patch)
	}
	if !strings.Contains(patch, "+++ b/capability-map.md") {
		t.Errorf("patch missing `+++ b/` header; patch:\n%s", patch)
	}
	if !strings.Contains(patch, "@@") {
		t.Errorf("patch missing `@@` hunk header; patch:\n%s", patch)
	}
	// Must carry the removed old line + the added new line.
	if !strings.Contains(patch, "-line two") {
		t.Errorf("patch missing removed `-line two`; patch:\n%s", patch)
	}
	if !strings.Contains(patch, "+line TWO changed") {
		t.Errorf("patch missing added `+line TWO changed`; patch:\n%s", patch)
	}
	if !strings.Contains(patch, "+line five") {
		t.Errorf("patch missing added `+line five`; patch:\n%s", patch)
	}
	// Context lines (unchanged) carry a leading space, not -/+.
	if !strings.Contains(patch, " line one") {
		t.Errorf("patch missing context ` line one`; patch:\n%s", patch)
	}
}

// TestUnifiedDiff_IdenticalReturnsEmpty verifies the no-op case: identical old
// and new text yields an empty patch (nothing to preview).
func TestUnifiedDiff_IdenticalReturnsEmpty(t *testing.T) {
	t.Parallel()

	same := "same\ncontent\n"
	if p := UnifiedDiff("audit-report.json", same, same); p != "" {
		t.Errorf("UnifiedDiff on identical text = %q, want empty", p)
	}
}

// TestGeneratePreviews_WritesPatchFiles verifies GeneratePreviews reads the
// draft (new) + live (old) doc surfaces and writes one *.patch file per
// in-scope subtree at <draftDir>/<doc-surface>.patch (design.md §D.1 layout).
// Fail-open: a subtree whose live doc is absent is skipped (not fatal).
func TestGeneratePreviews_WritesPatchFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	draftDir := filepath.Join(root, "draft")
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		t.Fatal(err)
	}
	liveDir := filepath.Join(root, "live")
	if err := os.MkdirAll(liveDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Fixture: two in-scope subtrees, each with a live doc (old) + draft (new).
	liveA := "row-alpha-v1\nrow-beta\n"
	liveB := `{"k": "old"}` + "\n"
	draftA := "row-alpha-v2\nrow-beta\nrow-gamma\n"
	draftB := `{"k": "new", "added": true}` + "\n"

	livePathA := filepath.Join(liveDir, "capability-map.md")
	livePathB := filepath.Join(liveDir, "audit-report.json")
	draftPathA := filepath.Join(draftDir, "capability-map.md")
	draftPathB := filepath.Join(draftDir, "audit-report.json")
	for _, w := range []struct{ path, content string }{
		{livePathA, liveA}, {livePathB, liveB},
		{draftPathA, draftA}, {draftPathB, draftB},
	} {
		if err := os.WriteFile(w.path, []byte(w.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	subtrees := []PreviewSubtree{
		{DocSurface: "capability-map.md", DraftPath: draftPathA, LiveDocPath: livePathA},
		{DocSurface: "audit-report.json", DraftPath: draftPathB, LiveDocPath: livePathB},
	}

	written, err := GeneratePreviews(draftDir, subtrees)
	if err != nil {
		t.Fatalf("GeneratePreviews returned error: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("GeneratePreviews wrote %d patches, want 2; got %v", len(written), written)
	}

	// Verify each patch file exists at the design.md §D.1 path + carries valid
	// unified-diff structure.
	for _, sub := range subtrees {
		patchPath := filepath.Join(draftDir, sub.DocSurface+".patch")
		data, err := os.ReadFile(patchPath)
		if err != nil {
			t.Errorf("read patch %s: %v", patchPath, err)
			continue
		}
		body := string(data)
		if !strings.Contains(body, "--- a/"+sub.DocSurface) {
			t.Errorf("%s missing --- a/ header; body:\n%s", patchPath, body)
		}
		if !strings.Contains(body, "+++ b/"+sub.DocSurface) {
			t.Errorf("%s missing +++ b/ header; body:\n%s", patchPath, body)
		}
		if !strings.Contains(body, "@@") {
			t.Errorf("%s missing @@ hunk header; body:\n%s", patchPath, body)
		}
	}
}

// TestGeneratePreviews_MissingLiveDocSkipped verifies fail-open: a subtree
// whose live doc surface does not exist is skipped (logged), and the remaining
// subtrees still produce their patches.
func TestGeneratePreviews_MissingLiveDocSkipped(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	draftDir := filepath.Join(root, "draft")
	if err := os.MkdirAll(draftDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// One real draft + one live doc; the other live doc is absent.
	draftPath := filepath.Join(draftDir, "capability-map.md")
	livePath := filepath.Join(draftDir+"-live", "capability-map.md")
	if err := os.MkdirAll(filepath.Dir(livePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(draftPath, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(livePath, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	subtrees := []PreviewSubtree{
		{DocSurface: "capability-map.md", DraftPath: draftPath, LiveDocPath: livePath},
		{DocSurface: "audit-report.json", DraftPath: filepath.Join(draftDir, "audit-report.json"), LiveDocPath: filepath.Join(root, "nonexistent", "audit-report.json")},
	}

	written, err := GeneratePreviews(draftDir, subtrees)
	if err != nil {
		t.Fatalf("GeneratePreviews should fail-open on missing live doc, got error: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("GeneratePreviews wrote %d patches, want 1 (missing-live skipped); got %v", len(written), written)
	}
}
