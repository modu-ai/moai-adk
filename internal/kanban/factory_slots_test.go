package kanban

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestFactoryFreeSlots is the shared-cluster AC for the t85 lead loop's
// picker: an empty or missing registry reads as all-free (fail-open), a live
// claim makes its slot busy, a dead claim is pruned so its slot reads free,
// and claims outside 1..workers do not widen the result.
func TestFactoryFreeSlots(t *testing.T) {
	t.Parallel()

	alwaysAlive := func(int) bool { return true }
	neverAlive := func(int) bool { return false }

	t.Run("missing registry means all free", func(t *testing.T) {
		t.Parallel()
		got := FactoryFreeSlots(t.TempDir(), 3, alwaysAlive)
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("FactoryFreeSlots(missing, 3) = %v, want [1 2 3]", got)
		}
	})

	t.Run("live claim makes its slot busy", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
			"lane-1": {PID: 11100},
			"lane-3": {PID: 11101},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		got := FactoryFreeSlots(root, 4, alwaysAlive)
		if !slices.Equal(got, []int{2, 4}) {
			t.Errorf("FactoryFreeSlots with live 1,3 = %v, want [2 4]", got)
		}
	})

	t.Run("dead claim is pruned and reads free", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
			"lane-2": {PID: 11100},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		got := FactoryFreeSlots(root, 3, neverAlive)
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Errorf("FactoryFreeSlots with dead 2 = %v, want [1 2 3]", got)
		}
	})

	t.Run("claims beyond workers do not widen the result", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
			"lane-9": {PID: 11100},
		}); err != nil {
			t.Fatalf("seed registry: %v", err)
		}
		got := FactoryFreeSlots(root, 2, alwaysAlive)
		if !slices.Equal(got, []int{1, 2}) {
			t.Errorf("FactoryFreeSlots with out-of-range claim = %v, want [1 2]", got)
		}
	})
}

// TestPruneFactoryDeadClaims pins the shared prune rule on its own: only
// live, positively-numbered claims survive.
func TestPruneFactoryDeadClaims(t *testing.T) {
	t.Parallel()

	reg := map[string]FactoryWorkerEntry{
		"lane-1": {PID: 11100}, // live
		"lane-2": {PID: 11101}, // dead
		"lane-3": {PID: 0},     // non-positive
		"lane-4": {PID: -5},    // negative
	}
	got := PruneFactoryDeadClaims(reg, func(pid int) bool { return pid == 11100 })
	if len(got) != 1 {
		t.Fatalf("pruned registry = %v, want only lane-1", got)
	}
	if _, ok := got["lane-1"]; !ok {
		t.Errorf("live claim lane-1 must survive, got %v", got)
	}
}

// TestBacklogQueuedCountSharedShape covers the one-call queued-count both the
// kanban notice and the factory lead loop render from: missing file reads 0,
// only state queued counts, and the path helper lands where the store reads.
func TestBacklogQueuedCountSharedShape(t *testing.T) {
	t.Parallel()

	t.Run("missing file and empty root read as zero", func(t *testing.T) {
		t.Parallel()
		if got := QueuedBacklogCountForRoot(t.TempDir()); got != 0 {
			t.Errorf("QueuedBacklogCountForRoot(missing) = %d, want 0", got)
		}
		if got := QueuedBacklogCountForRoot(""); got != 0 {
			t.Errorf("QueuedBacklogCountForRoot(\"\") = %d, want 0", got)
		}
	})

	t.Run("only queued items count", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := NewBacklogStore(BacklogPathForRoot(root)).Mutate(func(rec *BacklogRecord) error {
			rec.Items = []BacklogItem{
				{ID: "t1", State: BacklogStateQueued},
				{ID: "t2", State: BacklogStatePicked},
				{ID: "t3", State: BacklogStateQueued},
				{ID: "t4", State: BacklogStateDropped},
				{ID: "t5", State: BacklogStateQueued},
			}
			return nil
		}); err != nil {
			t.Fatalf("seed backlog: %v", err)
		}
		if got := QueuedBacklogCountForRoot(root); got != 3 {
			t.Errorf("QueuedBacklogCountForRoot = %d, want 3 (queued only)", got)
		}
	})

	t.Run("path helper lands under .moai/state/kanban", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		want := filepath.Join(root, ".moai", "state", "kanban", "backlog.json")
		if got := BacklogPathForRoot(root); got != want {
			t.Errorf("BacklogPathForRoot = %q, want %q", got, want)
		}
	})
}

// TestFactoryRegistryRoundTrip pins the moved cluster's load/save shape: the
// file the v1 cli code wrote is the file the shared cluster reads.
func TestFactoryRegistryRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	seed := map[string]FactoryWorkerEntry{
		"lane-1": {PID: os.Getpid(), RegisteredAt: "2026-08-17T00:00:00Z"},
	}
	if err := SaveFactoryRegistry(FactoryRegistryPath(root), seed); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := LoadFactoryRegistry(FactoryRegistryPath(root))
	if len(got) != 1 || got["lane-1"].PID != os.Getpid() {
		t.Errorf("round trip = %v, want lane-1 at this pid", got)
	}

	// A malformed file fails open to an empty registry (never an error).
	if err := os.WriteFile(FactoryRegistryPath(root), []byte("not json"), 0o600); err != nil {
		t.Fatalf("plant malformed file: %v", err)
	}
	if got := LoadFactoryRegistry(FactoryRegistryPath(root)); len(got) != 0 {
		t.Errorf("malformed registry = %v, want empty (fail-open)", got)
	}
}
