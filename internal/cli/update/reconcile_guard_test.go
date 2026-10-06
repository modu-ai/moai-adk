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
