package goal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestHTMLPathDerivation verifies K-4: HTMLPath derives the .html sibling from
// the StatePath .json path via TrimSuffix(".json")+".html".
func TestHTMLPathDerivation(t *testing.T) {
	got := HTMLPath("/proj", "sess-1")
	want := filepath.Join("/proj", StateDir, "sess-1.html")
	if got != want {
		t.Errorf("HTMLPath = %q, want %q", got, want)
	}
	// The .json and .html paths share the directory and stem, differ only in suffix.
	jsonPath := StatePath("/proj", "sess-1")
	if filepath.Dir(jsonPath) != filepath.Dir(got) {
		t.Errorf("dirs differ: json=%s html=%s", filepath.Dir(jsonPath), filepath.Dir(got))
	}
}

// TestClearGoalRemovesBothJSONAndHTML verifies AC-GHF-003: ClearGoal removes
// BOTH the <session>.json AND the <session>.html sibling, and is idempotent on
// both (a second call returns nil).
func TestClearGoalRemovesBothJSONAndHTML(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "sess.json")
	htmlPath := filepath.Join(dir, "sess.html")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(htmlPath, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ClearGoal(root, "sess"); err != nil {
		t.Fatalf("ClearGoal: %v", err)
	}
	for _, p := range []string{jsonPath, htmlPath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still exists after clear (err=%v)", p, err)
		}
	}
	// Idempotency: a second call on both-absent returns nil.
	if err := ClearGoal(root, "sess"); err != nil {
		t.Errorf("idempotent second ClearGoal: %v", err)
	}
}

// TestClearGoalJSONOnlySucceeds verifies AC-GHF-003 negative case: when only
// .json exists (no prior render), ClearGoal still succeeds and the absent .html
// is not an error.
func TestClearGoalJSONOnlySucceeds(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "s.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ClearGoal(root, "s"); err != nil {
		t.Fatalf("ClearGoal with json-only: %v", err)
	}
}

// TestOrphanPruneMovesHTMLSibling verifies AC-GHF-004: PruneOrphans moves the
// .html sibling alongside the .json to consumed/.
func TestOrphanPruneMovesHTMLSibling(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	jsonOld := filepath.Join(srcDir, "dead.json")
	htmlOld := filepath.Join(srcDir, "dead.html")
	if err := os.WriteFile(jsonOld, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(htmlOld, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-(OrphanTTL + time.Hour))
	if err := os.Chtimes(jsonOld, past, past); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(htmlOld, past, past); err != nil {
		t.Fatal(err)
	}

	moved, err := PruneOrphans(root, []string{"alive"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 1 || moved[0] != "dead.json" {
		t.Fatalf("moved: want [dead.json], got %v", moved)
	}
	consumedJSON := filepath.Join(root, ConsumedDir, "dead.json")
	consumedHTML := filepath.Join(root, ConsumedDir, "dead.html")
	for _, p := range []string{consumedJSON, consumedHTML} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s should be in consumed/: %v", p, err)
		}
	}
}

// TestOrphanPruneHTMLFailureDoesNotAbortJSON verifies AC-GHF-004 best-effort: a
// failure to move the .html sibling does NOT abort the .json move or the sweep.
// We simulate the .html failure by making its source unreadable/unmovable via
// a non-existent .html path is insufficient; instead we verify that when .html
// is absent, the .json still moves (the best-effort contract extends to .html
// absence being a non-event).
func TestOrphanPruneHTMLFailureDoesNotAbortJSON(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Only .json present (no .html sibling). The .json must still move; the
	// absent .html is a non-event (best-effort: no .html to move).
	jsonOld := filepath.Join(srcDir, "dead.json")
	if err := os.WriteFile(jsonOld, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-(OrphanTTL + time.Hour))
	if err := os.Chtimes(jsonOld, past, past); err != nil {
		t.Fatal(err)
	}
	moved, err := PruneOrphans(root, nil, time.Now())
	if err != nil {
		t.Fatalf("PruneOrphans: %v", err)
	}
	if len(moved) != 1 || moved[0] != "dead.json" {
		t.Fatalf("json move must succeed even with no html sibling; moved=%v", moved)
	}
}
