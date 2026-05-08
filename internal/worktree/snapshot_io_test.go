package worktree

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSnapshot_JSONRoundtrip(t *testing.T) {
	snap := &Snapshot{
		SchemaVersion:  SchemaVersion,
		CapturedAt:     time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
		SnapshotID:     "snap-test-001",
		HeadSHA:        "abcdef1234567890",
		Branch:         "feat/test",
		PorcelainLines: []string{" M file1.go", "?? newfile.md"},
		UntrackedSpecs: []string{".moai/specs/SPEC-X/note.md"},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "test-snapshot.json")
	if err := SaveSnapshot(snap, path); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if !reflect.DeepEqual(snap, loaded) {
		t.Errorf("roundtrip mismatch:\norig: %+v\nload: %+v", snap, loaded)
	}
}

func TestSnapshot_LoadMissing(t *testing.T) {
	if _, err := LoadSnapshot(filepath.Join(t.TempDir(), "nonexistent.json")); err == nil {
		t.Error("expected error loading nonexistent file, got nil")
	}
}

func TestSnapshot_SaveCreatesParent(t *testing.T) {
	dir := t.TempDir()
	deepPath := filepath.Join(dir, "a", "b", "c", "snap.json")
	snap := &Snapshot{SchemaVersion: SchemaVersion, SnapshotID: "deep"}
	if err := SaveSnapshot(snap, deepPath); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	if _, err := LoadSnapshot(deepPath); err != nil {
		t.Errorf("LoadSnapshot deep path: %v", err)
	}
}

func TestSnapshot_SaveNil(t *testing.T) {
	if err := SaveSnapshot(nil, filepath.Join(t.TempDir(), "x.json")); err == nil {
		t.Error("expected error saving nil snapshot, got nil")
	}
}

func TestSnapshotPath(t *testing.T) {
	got := SnapshotPath("/proj", "snap-123")
	want := filepath.Join("/proj", ".moai/state/worktree-snapshot-snap-123.json")
	if got != want {
		t.Errorf("SnapshotPath = %q, want %q", got, want)
	}
}
