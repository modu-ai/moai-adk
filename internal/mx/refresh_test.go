package mx

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newRefreshFixture creates a project with two tagged Go files and returns
// (root, stateDir).
func newRefreshFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal", "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(rel, tagBody string) {
		body := "package " + filepath.Base(filepath.Dir(rel)) + "\n\n// @MX:NOTE: [AUTO] " + tagBody + "\nfunc F() {}\n"
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("internal/a/a.go", "alpha note")
	write("internal/b/b.go", "beta note")
	stateDir := filepath.Join(root, ".moai", "state")
	return root, stateDir
}

// initialScan performs a full scan via the production stamping path.
func initialScan(t *testing.T, root, stateDir string) *Sidecar {
	t.Helper()
	s := NewScanner()
	s.SetIgnorePatterns(DefaultScanIgnore)
	tags, err := s.ScanDir(root)
	if err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(stateDir)
	sc := &Sidecar{SchemaVersion: SchemaVersion, Tags: tags, ScannedAt: time.Now(),
		Provenance: StampMXScan(root, s.ScanInventory(root))}
	if err := mgr.Write(sc); err != nil {
		t.Fatal(err)
	}
	return sc
}

// AC-GF-009 — changed-files-only refresh: with zero content changes the
// refresh re-reads (parses) 0 source files; with exactly 2 changed files it
// parses exactly those 2.
func TestRefreshIndex_ChangedFilesOnly(t *testing.T) {
	root, stateDir := newRefreshFixture(t)
	initialScan(t, root, stateDir)

	// Zero changes: a refresh re-parses nothing.
	stats, err := RefreshIndex(stateDir, root, nil)
	if err != nil {
		t.Fatalf("RefreshIndex: %v", err)
	}
	if stats.FilesParsed != 0 {
		t.Errorf("zero-change refresh parsed %d files, want 0", stats.FilesParsed)
	}

	// Exactly 2 files change (uncommitted edits).
	for _, rel := range []string{"internal/a/a.go", "internal/b/b.go"} {
		body := "package " + filepath.Base(filepath.Dir(rel)) + "\n\n// @MX:NOTE: [AUTO] edited\nfunc F() {}\n"
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	stats2, err := RefreshIndex(stateDir, root, nil)
	if err != nil {
		t.Fatalf("RefreshIndex(2): %v", err)
	}
	if stats2.FilesParsed != 2 {
		t.Errorf("two-change refresh parsed %d files, want exactly 2 (changed-files-only)", stats2.FilesParsed)
	}
	if stats2.ChangedDetected != 2 {
		t.Errorf("ChangedDetected = %d, want 2", stats2.ChangedDetected)
	}
}

// AC-GF-008 / MUTANT B — the refresh actually re-reads: after an uncommitted
// edit, the refreshed index's tags reflect the edited content and the
// provenance fingerprint matches the post-edit file hashes. A stamp-only
// mutant leaves the OLD tags in place and fails here.
func TestRefreshIndex_ReflectsUncommittedEdits(t *testing.T) {
	root, stateDir := newRefreshFixture(t)
	sc := initialScan(t, root, stateDir)
	if len(sc.Tags) != 2 {
		t.Fatalf("fixture scan = %d tags, want 2", len(sc.Tags))
	}

	// Edit one file's tag body (uncommitted).
	edited := "package b\n\n// @MX:NOTE: [AUTO] beta EDITED body\nfunc F() {}\n"
	if err := os.WriteFile(filepath.Join(root, "internal", "b", "b.go"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RefreshIndex(stateDir, root, nil); err != nil {
		t.Fatalf("RefreshIndex: %v", err)
	}

	mgr := NewManager(stateDir)
	after, err := mgr.Load()
	if err != nil {
		t.Fatal(err)
	}
	foundEdited := false
	for _, tag := range after.Tags {
		if filepath.Base(tag.File) == "b.go" {
			// The scanner's Body preserves the "[AUTO] " prefix.
			if tag.Body != "[AUTO] beta EDITED body" {
				t.Errorf("MUTANT B: refreshed index still carries the pre-edit body %q — stamp-only refresh", tag.Body)
			} else {
				foundEdited = true
			}
		}
	}
	if !foundEdited {
		t.Error("refreshed index lost the b.go tag entirely")
	}

	// The provenance inventory must match the post-edit content (fingerprint
	// freshness), not merely carry a newer timestamp.
	if after.Provenance == nil {
		t.Fatal("refreshed sidecar lost its provenance")
	}
	sum, err := HashFile(filepath.Join(root, "internal", "b", "b.go"))
	if err != nil {
		t.Fatal(err)
	}
	if got := after.Provenance.FileInventory["internal/b/b.go"]; got != sum {
		t.Errorf("inventory hash for b.go = %s, want post-edit %s", got, sum)
	}
	// And a second refresh is now a no-op (the refresh consumed the change).
	stats, err := RefreshIndex(stateDir, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesParsed != 0 {
		t.Errorf("post-refresh second pass parsed %d files, want 0 (change consumed)", stats.FilesParsed)
	}
}

// AC-GF-010 substrate — per-tree anchoring: an index stamped in a different
// tree is never incrementally trusted; the refresh treats it as a full rescan.
func TestRefreshIndex_WrongTreeFullRescan(t *testing.T) {
	root, stateDir := newRefreshFixture(t)
	initialScan(t, root, stateDir)

	// Forge a wrong-tree provenance (as if the state dir were copied from
	// another worktree).
	mgr := NewManager(stateDir)
	sc, err := mgr.Load()
	if err != nil {
		t.Fatal(err)
	}
	sc.Provenance.TreeRoot = "/some/other/tree"
	if err := mgr.Write(sc); err != nil {
		t.Fatal(err)
	}

	stats, err := RefreshIndex(stateDir, root, nil)
	if err != nil {
		t.Fatalf("RefreshIndex: %v", err)
	}
	if !stats.FullRescan {
		t.Error("wrong-tree index must force a full rescan, not an incremental one")
	}
	if stats.FilesParsed == 0 {
		t.Error("full rescan must parse the tree's files")
	}
}

// New-file pickup: a file added after the scan is parsed by the next refresh
// (walk covers the tree; read covers change).
func TestRefreshIndex_PicksUpNewFiles(t *testing.T) {
	root, stateDir := newRefreshFixture(t)
	initialScan(t, root, stateDir)

	added := "package a\n\n// @MX:NOTE: [AUTO] brand new\nfunc G() {}\n"
	if err := os.WriteFile(filepath.Join(root, "internal", "a", "g.go"), []byte(added), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RefreshIndex(stateDir, root, nil); err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(stateDir)
	after, err := mgr.Load()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tag := range after.Tags {
		if filepath.Base(tag.File) == "g.go" {
			found = true
		}
	}
	if !found {
		t.Error("new tagged file must appear in the refreshed index")
	}
	if _, ok := after.Provenance.FileInventory["internal/a/g.go"]; !ok {
		t.Error("new file must be inventoried")
	}
}

// AC-GF-013 — mx-index hash anchoring: lines inserted above a tag leave the
// tag line's content (and its ContentHash) identical, so the refreshed index
// resolves the same tag by (file, hash) at its NEW line — the anchor holds
// across drift while Line is convenience data.
func TestRefreshIndex_TagHashSurvivesLineDrift(t *testing.T) {
	root, stateDir := newRefreshFixture(t)
	initialScan(t, root, stateDir)

	mgr := NewManager(stateDir)
	before, err := mgr.Load()
	if err != nil {
		t.Fatal(err)
	}
	var beforeTag Tag
	for _, tag := range before.Tags {
		if filepath.Base(tag.File) == "a.go" {
			beforeTag = tag
		}
	}
	if beforeTag.ContentHash == "" {
		t.Fatal("scanner must stamp ContentHash on every tag (REQ-GF-011)")
	}

	// Insert 5 lines above the tag (content untouched).
	shifted := "package a\n\n\n\n\n\n// @MX:NOTE: [AUTO] alpha note\nfunc F() {}\n"
	if err := os.WriteFile(filepath.Join(root, "internal", "a", "a.go"), []byte(shifted), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RefreshIndex(stateDir, root, nil); err != nil {
		t.Fatal(err)
	}
	after, err := mgr.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range after.Tags {
		if filepath.Base(tag.File) == "a.go" {
			if tag.ContentHash != beforeTag.ContentHash {
				t.Errorf("hash anchor moved with drift: before %s after %s — anchor must be content-derived",
					beforeTag.ContentHash, tag.ContentHash)
			}
			if tag.Line != beforeTag.Line+4 {
				t.Errorf("convenience line must track the physical line: before %d after %d (want +4)",
					beforeTag.Line, tag.Line)
			}
			return
		}
	}
	t.Fatal("drifted file's tag vanished from the refreshed index")
}

// Removed-file cleanup: a vanished file's tags drop out of the refreshed index.
func TestRefreshIndex_DropsVanishedFiles(t *testing.T) {
	root, stateDir := newRefreshFixture(t)
	initialScan(t, root, stateDir)

	if err := os.Remove(filepath.Join(root, "internal", "b", "b.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := RefreshIndex(stateDir, root, nil); err != nil {
		t.Fatal(err)
	}
	mgr := NewManager(stateDir)
	after, err := mgr.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range after.Tags {
		if filepath.Base(tag.File) == "b.go" {
			t.Error("vanished file's tags must drop out of the refreshed index")
		}
	}
	if _, ok := after.Provenance.FileInventory["internal/b/b.go"]; ok {
		t.Error("vanished file must leave the inventory")
	}
}
