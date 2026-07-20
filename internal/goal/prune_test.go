package goal

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestOrphanPrune (AC-GLE-007) asserts a state file absent from the active
// registry (or TTL-expired) is moved to consumed/.
func TestOrphanPrune(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// active session — should be kept.
	if err := os.WriteFile(filepath.Join(srcDir, "alive.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	// TTL-expired orphan — should be moved.
	old := filepath.Join(srcDir, "dead.json")
	if err := os.WriteFile(old, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-(OrphanTTL + time.Hour))
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	moved, err := PruneOrphans(root, []string{"alive"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 1 || moved[0] != "dead.json" {
		t.Fatalf("moved: want [dead.json], got %v", moved)
	}
	// active file survives.
	if _, err := os.Stat(filepath.Join(srcDir, "alive.json")); err != nil {
		t.Errorf("active file should survive: %v", err)
	}
	// orphan moved to consumed/.
	consumed := filepath.Join(root, ConsumedDir, "dead.json")
	if _, err := os.Stat(consumed); err != nil {
		t.Errorf("orphan should be in consumed/: %v", err)
	}
}

// TestOrphanPruneRecentInactiveKept asserts a file absent from the registry but
// NOT TTL-expired is kept (the active registry may be incomplete on this host).
func TestOrphanPruneRecentInactiveKept(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, StateDir)
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	recent := filepath.Join(srcDir, "recent.json")
	if err := os.WriteFile(recent, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved, err := PruneOrphans(root, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 0 {
		t.Fatalf("recent inactive file should be kept, moved=%v", moved)
	}
}

// TestOrphanPruneNoDir asserts prune on a missing state dir is a no-op nil.
func TestOrphanPruneNoDir(t *testing.T) {
	moved, err := PruneOrphans(t.TempDir(), nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(moved) != 0 {
		t.Fatalf("expected no moves on missing dir, got %v", moved)
	}
}
