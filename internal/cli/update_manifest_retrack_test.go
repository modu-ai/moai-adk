package cli

// update_manifest_retrack_test.go — card t1275.
//
// Unit coverage for the retrack bookkeeping. The end-to-end invariant
// (init → update --force → init --force leaves drift 0 / user_modified 0,
// while a user-edited file still reads user_modified) is exercised by the
// two-build experiment recorded in .moai/reports/t1275/verdict.md; these
// tests pin the helper's own contract: drift resolved, TemplateHash
// preserved, user-owned entries untouched, absent files skipped.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

func retrackFixture(t *testing.T, root string) manifest.Manager {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load: %v", err)
	}
	return mgr
}

func retrackWriteTracked(t *testing.T, root string, mgr manifest.Manager, rel, body string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	if err := mgr.Track(rel, manifest.TemplateManaged, "sha256:templatehash"); err != nil {
		t.Fatalf("track %s: %v", rel, err)
	}
	return path
}

func TestRetrackManifestFilesResolvesDriftAndKeepsTemplateHash(t *testing.T) {
	root := t.TempDir()
	mgr := retrackFixture(t, root)

	path := retrackWriteTracked(t, root, mgr, ".claude/settings.json", "original")
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Simulate the post-deploy rewrite: the file moves, the manifest does not.
	if err := os.WriteFile(path, []byte("merged content"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	if err := retrackManifestFiles(root, mgr, os.Stderr, []string{".claude/settings.json"}); err != nil {
		t.Fatalf("retrack: %v", err)
	}

	reloaded := manifest.NewManager()
	if _, err := reloaded.Load(root); err != nil {
		t.Fatalf("reload: %v", err)
	}
	entry, ok := reloaded.GetEntry(".claude/settings.json")
	if !ok {
		t.Fatal("entry missing after retrack")
	}
	if entry.Provenance != manifest.TemplateManaged {
		t.Errorf("provenance = %q, want template_managed", entry.Provenance)
	}
	if entry.TemplateHash != "sha256:templatehash" {
		t.Errorf("TemplateHash = %q, want preserved sha256:templatehash", entry.TemplateHash)
	}
	current, err := manifest.HashFile(path)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if entry.CurrentHash != current {
		t.Errorf("CurrentHash = %q, want on-disk %q", entry.CurrentHash, current)
	}
}

func TestRetrackManifestFilesLeavesUserOwnedEntriesAlone(t *testing.T) {
	root := t.TempDir()
	mgr := retrackFixture(t, root)

	// user_modified entry whose drift must survive (the two-way invariant).
	path := retrackWriteTracked(t, root, mgr, ".claude/rules/local-only.md", "original")
	entry := mgr.Manifest().Files[".claude/rules/local-only.md"]
	entry.Provenance = manifest.UserModified
	mgr.Manifest().Files[".claude/rules/local-only.md"] = entry
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := os.WriteFile(path, []byte("user edited"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	if err := retrackManifestFiles(root, mgr, os.Stderr, []string{".claude/rules/local-only.md"}); err != nil {
		t.Fatalf("retrack: %v", err)
	}

	reloaded := manifest.NewManager()
	if _, err := reloaded.Load(root); err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := reloaded.GetEntry(".claude/rules/local-only.md")
	if !ok {
		t.Fatal("entry missing")
	}
	if got.Provenance != manifest.UserModified {
		t.Errorf("user_modified entry was reclassified to %q — retrack must not touch user-owned files", got.Provenance)
	}
}

func TestRetrackManifestFilesSkipsAbsentAndUntracked(t *testing.T) {
	root := t.TempDir()
	mgr := retrackFixture(t, root)
	retrackWriteTracked(t, root, mgr, ".moai/config/sections/quality.yaml", "q: 1\n")
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Absent file, and a path with no manifest entry at all: neither may
	// create an entry nor fail the call.
	if err := retrackManifestFiles(root, mgr, os.Stderr, []string{
		".claude/does-not-exist.md",
		".claude/never-tracked.md",
	}); err != nil {
		t.Fatalf("retrack: %v", err)
	}

	reloaded := manifest.NewManager()
	if _, err := reloaded.Load(root); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if _, ok := reloaded.GetEntry(".claude/never-tracked.md"); ok {
		t.Error("retrack created an entry for an untracked path — it is bookkeeping, not classification")
	}
	if _, ok := reloaded.GetEntry(".moai/config/sections/quality.yaml"); !ok {
		t.Error("untouched entry vanished from the manifest")
	}
}

func TestRetrackSectionFilesCoversSectionDirectory(t *testing.T) {
	root := t.TempDir()
	mgr := retrackFixture(t, root)

	path := retrackWriteTracked(t, root, mgr, ".moai/config/sections/user.yaml", "user:\n  name: a\n")
	retrackWriteTracked(t, root, mgr, ".moai/config/sections/unmanaged.yaml", "x: 1\n")
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Restore-style rewrite of one section only.
	if err := os.WriteFile(path, []byte("user:\n  name: b\n"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	retrackSectionFiles(root, os.Stderr)

	reloaded := manifest.NewManager()
	if _, err := reloaded.Load(root); err != nil {
		t.Fatalf("reload: %v", err)
	}
	entry, ok := reloaded.GetEntry(".moai/config/sections/user.yaml")
	if !ok {
		t.Fatal("section entry missing after retrackSectionFiles")
	}
	current, err := manifest.HashFile(path)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if entry.CurrentHash != current {
		t.Errorf("CurrentHash = %q, want on-disk %q", entry.CurrentHash, current)
	}
}
