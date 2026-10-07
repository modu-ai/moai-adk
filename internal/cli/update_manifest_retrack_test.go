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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

	if err := retrackSectionFiles(root, os.Stderr); err != nil {
		t.Fatalf("retrackSectionFiles on a writable tree: %v", err)
	}

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

// TestRetrackSectionFiles_FailureReturnsErrorAndPrintsNothing pins the
// card t1527 repair-round-3 contract on the FAILURE path: the save error is
// RETURNED for the caller's collector (the init tail routes it into the
// warning summary panel) and the helper itself prints nothing on its writer —
// a regression that printed its own diagnostic would break the init
// ordering contract (nothing after the completion card but the panel) and
// double-surface the failure.
func TestRetrackSectionFiles_FailureReturnsErrorAndPrintsNothing(t *testing.T) {
	root := t.TempDir()
	mgr := retrackFixture(t, root)
	retrackWriteTracked(t, root, mgr, ".moai/config/sections/user.yaml", "user:\n  name: a\n")
	if err := mgr.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	// Dirty the tracked file so the retrack is dirty and Save actually runs.
	sections := filepath.Join(root, ".moai", "config", "sections", "user.yaml")
	if err := os.WriteFile(sections, []byte("user:\n  name: b\n"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}

	// Break the SAVE, not the load: .moai goes read-only so the helper's own
	// Load still succeeds (a broken load is the documented no-op escape) but
	// mgr.Save's write into .moai fails.
	if runtime.GOOS == "windows" {
		t.Skip("a 0500 directory does not deny writes on Windows; the failure injection cannot reproduce there")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses directory permission checks")
	}
	moaiDir := filepath.Join(root, ".moai")
	if err := os.Chmod(moaiDir, 0o500); err != nil {
		t.Fatalf("chmod .moai read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(moaiDir, 0o755) })

	var errOut bytes.Buffer
	err := retrackSectionFiles(root, &errOut)

	if err == nil {
		t.Fatal("a failed manifest save must return an error for the caller's collector")
	}
	if errOut.Len() != 0 {
		t.Errorf("the helper must not print the failure (one-surface rule), writer=%q", errOut.String())
	}
}

// --- card t1276 F3: the user-invocable restore entry point retracks the
// sections it rewrote from the backup.

func TestRunUpdateRestoreRetracksSections(t *testing.T) {
	root, backupDir := newRestoreFixture(t)

	// Reproduce the failure window: a run whose deploy saved the NEW hashes,
	// then the restore puts the backup's older content back on disk.
	sections := filepath.Join(root, ".moai", "config", "sections", "user.yaml")
	if err := os.WriteFile(sections, []byte("user:\n  name: new-version-value\n"), 0o644); err != nil {
		t.Fatalf("rewrite section: %v", err)
	}
	mgr := retrackFixture(t, root)
	if err := mgr.Track(".moai/config/sections/user.yaml", manifest.TemplateManaged, "sha256:tpl"); err != nil {
		t.Fatalf("track v2: %v", err)
	}
	if err := mgr.Save(); err != nil {
		t.Fatalf("save v2 manifest: %v", err)
	}

	if err := runUpdateRestore(root, backupDir, io.Discard); err != nil {
		t.Fatalf("restore: %v", err)
	}

	reloaded := manifest.NewManager()
	if _, err := reloaded.Load(root); err != nil {
		t.Fatalf("reload manifest: %v", err)
	}
	entry, ok := reloaded.GetEntry(".moai/config/sections/user.yaml")
	if !ok {
		t.Fatal("entry missing after restore")
	}
	current, err := manifest.HashFile(sections)
	if err != nil {
		t.Fatalf("hash restored section: %v", err)
	}
	if entry.CurrentHash != current {
		t.Errorf("CurrentHash = %q after restore, want the restored on-disk %q — restore must retrack what it rewrote (t1276 F3)", entry.CurrentHash, current)
	}
}
