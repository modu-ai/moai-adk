// manifest_test.go — unit tests for the per-user manifest subsystem
// (SPEC-USER-ASSET-INSTALL-001 M1; REQ-006, REQ-021, design §2.2).
//
// Every test runs against a temp HOME (plan §D — no real-$HOME writes).
package userassets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestManifestLoadAbsentReturnsFresh covers the recovery-lattice case 1
// precondition: an absent manifest loads as a fresh, writable document —
// "absent → install".
func TestManifestLoadAbsentReturnsFresh(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	m, err := Load(ManifestPath(home))
	if err != nil {
		t.Fatalf("Load(absent): %v", err)
	}
	if m.SchemaVersion != SchemaVersion {
		t.Errorf("fresh manifest schema_version = %d, want %d", m.SchemaVersion, SchemaVersion)
	}
	if m.Files == nil {
		t.Error("fresh manifest Files map is nil — Save would write null")
	}
	if !m.SchemaKnown() {
		t.Error("fresh manifest schema must be known (install/refresh proceed)")
	}
}

// TestManifestRoundTripPreservesFields pins REQ-006's record content: every
// installed path carries sha256 + bundle + per-file moai version, and the
// recorded bundle selection survives the round trip.
func TestManifestRoundTripPreservesFields(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)
	m, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	m.Bundles = []string{"devops"}
	m.Files["claude-skills/moai-workflow-spec/SKILL.md"] = FileEntry{
		SHA256: "abc123", Bundle: "core", InstalledAt: "2026-10-06T00:00:00Z", MoaiVersion: "v3.2.0",
	}
	m.Collisions = append(m.Collisions, Collision{Path: "claude-skills/user-file.md", FirstSeenAt: "2026-10-06T00:00:00Z"})
	if err := m.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	again, err := Load(p)
	if err != nil {
		t.Fatalf("re-Load: %v", err)
	}
	fe, ok := again.Files["claude-skills/moai-workflow-spec/SKILL.md"]
	if !ok {
		t.Fatal("recorded file entry lost by round trip")
	}
	if fe.SHA256 != "abc123" || fe.Bundle != "core" || fe.MoaiVersion != "v3.2.0" || fe.InstalledAt == "" {
		t.Errorf("file entry fields incomplete: %+v", fe)
	}
	if len(again.Bundles) != 1 || again.Bundles[0] != "devops" {
		t.Errorf("bundle selection lost: %v", again.Bundles)
	}
	if len(again.Collisions) != 1 || again.Collisions[0].Path != "claude-skills/user-file.md" {
		t.Errorf("collision record lost: %+v", again.Collisions)
	}
}

// TestManifestUnknownFieldsPreserved pins REQ-021 (iter4 D27): EVERY write
// carries through fields the writing binary does not understand — top-level
// and per-file — under a KNOWN schema version.
func TestManifestUnknownFieldsPreserved(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)

	// A NEWER binary's manifest: schema 2 with fields this binary has no
	// struct for, top-level and per-file.
	newer := map[string]interface{}{
		"schema_version": 2,
		"future_top":     map[string]interface{}{"x": 1},
		"bundles":        []string{"devops"},
		"files": map[string]interface{}{
			"claude-skills/a/SKILL.md": map[string]interface{}{
				"sha256": "deadbeef", "bundle": "core",
				"installed_at": "t", "moai_version": "v9",
				"future_file_field": "keep-me",
			},
		},
	}
	raw, err := json.Marshal(newer)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := Load(p)
	if err != nil {
		t.Fatalf("Load(newer manifest): %v", err)
	}
	if m.SchemaKnown() {
		t.Error("schema 2 must read as unknown (append-only rules apply)")
	}
	// Append-only install write (REQ-021 permits it against an unknown
	// schema) must preserve the fields it does not understand.
	m.Files["claude-skills/b/SKILL.md"] = FileEntry{SHA256: "ff", Bundle: "core", InstalledAt: "t2", MoaiVersion: "v3.2.0"}
	if err := m.Save(p); err != nil {
		t.Fatalf("append-only Save against unknown schema: %v", err)
	}

	var after map[string]json.RawMessage
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &after); err != nil {
		t.Fatalf("saved manifest is not JSON: %v", err)
	}
	if string(after["future_top"]) == "" {
		t.Error("unknown top-level field future_top was dropped by Save (REQ-021)")
	}
	var fe map[string]json.RawMessage
	if err := json.Unmarshal(after["files"], &fe); err != nil {
		t.Fatal(err)
	}
	var a map[string]json.RawMessage
	if err := json.Unmarshal(fe["claude-skills/a/SKILL.md"], &a); err != nil {
		t.Fatal(err)
	}
	if string(a["future_file_field"]) == "" {
		t.Error("unknown per-file field future_file_field was dropped by Save (REQ-021)")
	}
	if _, ok := fe["claude-skills/b/SKILL.md"]; !ok {
		t.Error("the append-only install entry was not written")
	}
}

// TestManifestRemovalRefusesUnknownSchema pins REQ-021's gate: manifest-driven
// removal refuses against an unknown schema version while the append-only
// write above proceeds.
func TestManifestRemovalRefusesUnknownSchema(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)
	m, _ := Load(p)
	m.SchemaVersion = 99
	if err := m.CanRemove(); err == nil {
		t.Fatal("CanRemove() succeeded against schema_version 99 — REQ-021 refusal missing")
	} else if !strings.Contains(err.Error(), "schema") {
		t.Errorf("refusal error should name the schema gate: %v", err)
	}
}

// TestManifestCorruptJSONIsTypedError covers AC-021's second clause: a corrupt
// manifest loads as a typed corruption error — refuse removal, report, offer
// rebuild-from-scan, never auto-delete.
func TestManifestCorruptJSONIsTypedError(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(p)
	if err == nil {
		t.Fatal("Load(corrupt) returned no error")
	}
	var ce *CorruptError
	if !AsCorrupt(err, &ce) {
		t.Fatalf("Load(corrupt) error is %T, want *CorruptError", err)
	}
	if ce.Path != p {
		t.Errorf("CorruptError.Path = %q, want %q", ce.Path, p)
	}
	// Never auto-delete: the corrupt file is still on disk.
	if _, statErr := os.Stat(p); statErr != nil {
		t.Error("corrupt manifest was deleted by Load — REQ-021 forbids auto-delete")
	}
}

// TestManifestSaveIsAtomicNoResidue asserts the save pattern leaves no temp
// residue beside the manifest on success.
func TestManifestSaveIsAtomicNoResidue(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	p := ManifestPath(home)
	m, _ := Load(p)
	m.Files["x"] = FileEntry{SHA256: "1", Bundle: "core", InstalledAt: "t", MoaiVersion: "v"}
	if err := m.Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(p))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "user-assets.json" {
			t.Errorf("residue beside manifest after save: %s", e.Name())
		}
	}
}

// TestJournalRoundTrip pins the journal's completeness (directed repair R-e):
// the bundle-SELECTION delta, FULL per-entry provenance, and the
// write-completion flag all survive the round trip.
func TestJournalRoundTrip(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	j := &PendingJournal{
		SchemaVersion:    SchemaVersion,
		BundlesSelection: []string{"devops"},
		StartedAt:        "2026-10-06T00:00:00Z",
		Entries: []JournalEntry{{
			Path: "claude-skills/moai-ref-owasp-checklist/SKILL.md", ExpectedSHA256: "cafe",
			Bundle: "devops", MoaiVersion: "v3.2.0", InstalledAt: "2026-10-06T00:00:00Z",
		}},
	}
	if err := WriteJournal(JournalPath(home), j); err != nil {
		t.Fatalf("WriteJournal: %v", err)
	}
	got, err := LoadJournal(JournalPath(home))
	if err != nil {
		t.Fatalf("LoadJournal: %v", err)
	}
	if got == nil {
		t.Fatal("journal lost")
	}
	if len(got.BundlesSelection) != 1 || got.BundlesSelection[0] != "devops" {
		t.Errorf("selection delta lost: %v", got.BundlesSelection)
	}
	if len(got.Entries) != 1 {
		t.Fatalf("entries lost: %+v", got.Entries)
	}
	e := got.Entries[0]
	if e.Path != "claude-skills/moai-ref-owasp-checklist/SKILL.md" ||
		e.ExpectedSHA256 != "cafe" || e.Bundle != "devops" || e.MoaiVersion != "v3.2.0" ||
		e.InstalledAt == "" || e.WriteCompleted {
		t.Errorf("journal entry provenance incomplete: %+v", e)
	}
	if err := ClearJournal(JournalPath(home)); err != nil {
		t.Fatalf("ClearJournal: %v", err)
	}
	if again, _ := LoadJournal(JournalPath(home)); again != nil {
		t.Error("journal still loadable after ClearJournal")
	}
}

// TestJournalAbsentLoadsNil covers the no-journal fast path.
func TestJournalAbsentLoadsNil(t *testing.T) {
	t.Parallel()
	got, err := LoadJournal(JournalPath(t.TempDir()))
	if err != nil || got != nil {
		t.Errorf("LoadJournal(absent) = %v, %v; want nil, nil", got, err)
	}
}

// TestUserLockSerializes pins REQ-006's read-modify-write serialization: while
// one holder is inside the lock, a second acquire refuses (this test uses a
// short timeout — the production caller retries across its run window).
func TestUserLockSerializes(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	l1, err := AcquireUserLock(home, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if _, err := AcquireUserLock(home, 50*time.Millisecond); err == nil {
		t.Error("second acquire while held succeeded — the serialization is not real")
	}
	if err := l1.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	l2, err := AcquireUserLock(home, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	_ = l2.Release()
}

// TestRelPathValidation pins the C2 path-hygiene half: Clean normalization
// and `..` rejection before any root join.
func TestRelPathValidation(t *testing.T) {
	t.Parallel()
	if got, err := ValidateRelPath("skills/moai/SKILL.md"); err != nil || got != "skills/moai/SKILL.md" {
		t.Errorf("ValidateRelPath(clean) = %q, %v", got, err)
	}
	if _, err := ValidateRelPath("../escape.md"); err == nil {
		t.Error("`..` escape accepted")
	}
	if _, err := ValidateRelPath("skills/../../escape.md"); err == nil {
		t.Error("nested `..` escape accepted")
	}
	if got, err := ValidateRelPath("./skills//moai/./x.md"); err != nil || got != "skills/moai/x.md" {
		t.Errorf("ValidateRelPath(dot segments) = %q, %v", got, err)
	}
	if _, err := ValidateRelPath("/absolute"); err == nil {
		t.Error("absolute rel path accepted")
	}
}

// TestRootSlugs pins the backup-home root slugs (iter4 D33 layout): the four
// roots get fixed slugs so `~/.claude/skills/x` and `$HOME/.agents/skills/x`
// cannot collide on `skills/x`.
func TestRootSlugs(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	roots := ResolveRoots(home)
	want := map[string]string{
		"claude-skills": filepath.Join(home, ".claude", "skills"),
		"claude-agents": filepath.Join(home, ".claude", "agents"),
		"agents-skills": filepath.Join(home, ".agents", "skills"),
		"codex-agents":  filepath.Join(home, ".codex", "agents"),
	}
	if len(roots) != len(want) {
		t.Fatalf("ResolveRoots returned %d roots, want %d", len(roots), len(want))
	}
	for _, r := range roots {
		dir, ok := want[string(r.Slug)]
		if !ok {
			t.Errorf("unexpected root slug %q", r.Slug)
			continue
		}
		if r.Dir != dir {
			t.Errorf("root %q dir = %q, want %q", r.Slug, r.Dir, dir)
		}
	}
}
