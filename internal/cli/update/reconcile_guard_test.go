package update

// reconcile_guard_test.go — SPEC-UPDATE-MIGRATION-001 (card t1547): the
// REQ-UPM-015 protection composition and the snapshot base source, the
// pieces the cli wiring consumes for the legacy wholesale path and the
// merge-phase base.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/manifest"
)

// TestProtectFuncFor — AC-UPM-033's production composition: user-owned
// namespace (IsUserOwnedNamespace) and unresolved user-modified manifest
// records are protected; template-managed and unrecorded paths are not.
func TestProtectFuncFor(t *testing.T) {
	const modifiedRel = ".claude/agents/moai/core/edited.md"
	root := newClassifyFixture(t, map[string]string{
		modifiedRel: "operator's unabsorbed edit\n",
	})
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := mgr.Track(modifiedRel, manifest.UserModified, ""); err != nil {
		t.Fatalf("track: %v", err)
	}
	entry := mgr.Manifest().Files[modifiedRel]
	entry.CurrentHash = manifest.HashBytes([]byte("operator's unabsorbed edit\n"))
	mgr.Manifest().Files[modifiedRel] = entry

	protected := ProtectFuncFor(root, mgr.Manifest())
	for rel, want := range map[string]bool{
		// User-owned namespace — protected whatever the manifest says.
		".claude/skills/hns-worker/SKILL.md": true,
		".moai/harness/tool":                 true,
		// Unresolved user-modified (record + hash match disk) — protected.
		modifiedRel: true,
		// Unrecorded path — not protected by this predicate (the caller's
		// other machinery handles user-owned carriage routing).
		".claude/rules/moai/some-template-rule.md": false,
	} {
		if got := protected(rel); got != want {
			t.Errorf("ProtectFuncFor(%q) = %v, want %v", rel, got, want)
		}
	}
}

// TestUnresolvedUserModified — the guard's second arm: only a user_modified
// record whose recorded hash still matches the disk bytes protects. A stale
// hash, a different provenance, or an absent record does not.
func TestUnresolvedUserModified(t *testing.T) {
	const rel = ".claude/rules/moai/tuned.md"
	const onDisk = "current disk bytes\n"
	root := newClassifyFixture(t, map[string]string{rel: onDisk})
	mgr := manifest.NewManager()
	if _, err := mgr.Load(root); err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	// Absent record.
	if UnresolvedUserModified(mgr.Manifest(), root, rel) {
		t.Error("absent record resolved as unresolved user-modified")
	}
	// Template-managed provenance never protects.
	if err := mgr.Track(rel, manifest.TemplateManaged, ""); err != nil {
		t.Fatalf("track: %v", err)
	}
	if UnresolvedUserModified(mgr.Manifest(), root, rel) {
		t.Error("template-managed provenance resolved as unresolved user-modified")
	}
	// User-modified with matching hash protects.
	entry := mgr.Manifest().Files[rel]
	entry.Provenance = manifest.UserModified
	entry.CurrentHash = manifest.HashBytes([]byte(onDisk))
	mgr.Manifest().Files[rel] = entry
	if !UnresolvedUserModified(mgr.Manifest(), root, rel) {
		t.Error("matching user_modified record did not protect")
	}
	// User-modified with a stale hash does not.
	entry.CurrentHash = manifest.HashBytes([]byte("stale\n"))
	mgr.Manifest().Files[rel] = entry
	if UnresolvedUserModified(mgr.Manifest(), root, rel) {
		t.Error("stale user_modified record protected")
	}
	// Nil manifest protects nothing.
	if UnresolvedUserModified(nil, root, rel) {
		t.Error("nil manifest protected")
	}
}

// TestSafeWriteFileRefusesLinkedParentDir — gate round 10 finding 2's
// deterministic shape at the helper level: a write target whose parent chain
// holds a symlink is refused before any write.
func TestSafeWriteFileRefusesLinkedParentDir(t *testing.T) {
	external := t.TempDir()
	root := t.TempDir()
	linkAt := filepath.Join(root, ".claude", "rules", "moai")
	if err := os.MkdirAll(filepath.Dir(linkAt), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, linkAt); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if err := safeWriteFile(root, ".claude/rules/moai/file.md", []byte("x\n")); err == nil {
		t.Fatal("safeWriteFile through a linked parent must be refused")
	}
	entries, err := os.ReadDir(external)
	if err != nil || len(entries) != 0 {
		t.Errorf("external directory polluted: %v (err=%v)", entries, err)
	}
}

// TestLoadManifestReadOnly — gate round 19's loader: absent and corrupt
// manifests read as nil (the conservative no-record route) with NOTHING
// written to disk; a valid manifest parses with its entries intact.
func TestLoadManifestReadOnly(t *testing.T) {
	// Absent manifest → nil, no file created.
	absentRoot := t.TempDir()
	if mf := LoadManifestReadOnly(absentRoot); mf != nil {
		t.Errorf("absent manifest = %+v, want nil", mf)
	}
	if _, err := os.Stat(filepath.Join(absentRoot, ".moai", "manifest.json")); !os.IsNotExist(err) {
		t.Errorf("read-only load created a manifest file: %v", err)
	}

	// Corrupt manifest → nil, original bytes untouched.
	corruptRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(corruptRoot, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	const corrupt = "{ not json"
	manifestPath := filepath.Join(corruptRoot, ".moai", "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(corrupt), 0o644); err != nil {
		t.Fatalf("seed corrupt: %v", err)
	}
	if mf := LoadManifestReadOnly(corruptRoot); mf != nil {
		t.Errorf("corrupt manifest = %+v, want nil", mf)
	}
	if data, err := os.ReadFile(manifestPath); err != nil || string(data) != corrupt {
		t.Errorf("corrupt manifest altered: %q (err=%v)", data, err)
	}

	// Valid manifest parses with entries.
	validRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(validRoot, ".moai"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	valid := `{"version":"1","files":{".claude/rules/moai/rule.md":{"provenance":"template_managed","current_hash":"abc"}}}`
	if err := os.WriteFile(filepath.Join(validRoot, ".moai", "manifest.json"), []byte(valid), 0o644); err != nil {
		t.Fatalf("seed valid: %v", err)
	}
	mf := LoadManifestReadOnly(validRoot)
	if mf == nil {
		t.Fatal("valid manifest read as nil")
	}
	entry, ok := mf.Files[".claude/rules/moai/rule.md"]
	if !ok || entry.Provenance != manifest.TemplateManaged {
		t.Errorf("parsed entry = %+v (ok=%v), want the tracked rule", entry, ok)
	}
}

// TestSafeWriteFilePreservesExistingMode — the mode-preservation half: a
// restored file keeps its on-disk mode; a fresh file gets the package
// default mode.
func TestSafeWriteFilePreservesExistingMode(t *testing.T) {
	root := t.TempDir()
	rel := ".claude/rules/moai/tuned.md"
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(abs, []byte("original\n"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := safeWriteFile(root, rel, []byte("restored bytes\n")); err != nil {
		t.Fatalf("safeWriteFile: %v", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("restored file mode = %v, want the preserved 0600", info.Mode().Perm())
	}
	if data, readErr := os.ReadFile(abs); readErr != nil || string(data) != "restored bytes\n" {
		t.Errorf("restored content = %q (err=%v)", data, readErr)
	}
}

// TestExclusiveWriteFileRefusesLinkedParentDir — the exclusive-claim helper
// shares the link-free parent-chain contract (gate round 17, finding 2).
func TestExclusiveWriteFileRefusesLinkedParentDir(t *testing.T) {
	external := t.TempDir()
	root := t.TempDir()
	linkAt := filepath.Join(root, ".claude", "rules", "moai")
	if err := os.MkdirAll(filepath.Dir(linkAt), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Symlink(external, linkAt); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if err := exclusiveWriteFile(root, ".claude/rules/moai/policy.json.moai-new", []byte("x\n")); err == nil {
		t.Fatal("exclusiveWriteFile through a linked parent must be refused")
	}
	entries, err := os.ReadDir(external)
	if err != nil || len(entries) != 0 {
		t.Errorf("external directory polluted: %v (err=%v)", entries, err)
	}
}

// TestArchiveThenRemoveMissingSource — the source-side gate's first arm: a
// source that vanished before the read is a clean error, not a removal.
func TestArchiveThenRemoveMissingSource(t *testing.T) {
	root := t.TempDir()
	err := archiveThenRemove(root, ".claude/rules/moai/gone.md", filepath.Join(ReconcileArchiveFilesRoot(), "run"))
	if err == nil {
		t.Fatal("missing source must error")
	}
}

// TestArchiveThenRemoveInstallFailureKeepsSource — the install gate: when
// the rename onto the destination fails (here: the destination name is
// occupied by a DIRECTORY), the copy errors and the operator's file stays
// in place — the removal only ever follows a successful install.
func TestArchiveThenRemoveInstallFailureKeepsSource(t *testing.T) {
	const rel = ".claude/rules/moai/old-rule.md"
	root := t.TempDir()
	srcAbs := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(srcAbs), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(srcAbs, []byte("stale but mine\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// Pre-create the run dir with a DIRECTORY squatting at the destination
	// name — the claim Mkdir won't touch it (exists), and the rename onto a
	// directory fails.
	runDir := filepath.Join(root, filepath.FromSlash(ReconcileArchiveFilesRoot()), "20260102_030405")
	if err := os.MkdirAll(filepath.Join(runDir, filepath.FromSlash(rel)), 0o755); err != nil {
		t.Fatalf("mkdir squat: %v", err)
	}

	err := archiveThenRemove(root, rel, filepath.Join(ReconcileArchiveFilesRoot(), "20260102_030405"))
	if err == nil {
		t.Fatal("install onto a directory must fail")
	}
	if data, readErr := os.ReadFile(srcAbs); readErr != nil || string(data) != "stale but mine\n" {
		t.Errorf("source altered by the failed install: %q (err=%v)", data, readErr)
	}
}

// TestSnapshotBaseSource — the merge-phase base reader: config section paths
// read from the deploy-time snapshot layout; paths outside sections/ report
// no base (the conservative conflict disposition then applies).
func TestSnapshotBaseSource(t *testing.T) {
	root := t.TempDir()
	snapDir := filepath.Join(root, ".moai", "cache", "template-snapshot", "sections")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatalf("mkdir snapshot: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "git-strategy.yaml"), []byte("prior render\n"), 0o644); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	base := SnapshotBaseSource(root)
	data, ok := base(".moai/config/sections/git-strategy.yaml")
	if !ok || string(data) != "prior render\n" {
		t.Errorf("section base = %q (ok=%v), want the snapshot bytes", data, ok)
	}
	if _, ok := base(".moai/config/catalog.yaml"); ok {
		t.Error("non-section path reported a base — only sections/ have snapshots")
	}
	if _, ok := base(".claude/rules/moai/note.md"); ok {
		t.Error("non-config path reported a base")
	}
}
