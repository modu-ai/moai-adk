package kanban

import (
	"slices"
	"testing"
)

// TestFactoryLaneLabelProducesWorkerNotation pins the canonical label a
// factory worker session launches under: `worker-<n>`.
func TestFactoryLaneLabelProducesWorkerNotation(t *testing.T) {
	t.Parallel()

	if got := FactoryLaneLabel(3); got != "worker-3" {
		t.Errorf("FactoryLaneLabel(3) = %q, want %q", got, "worker-3")
	}
}

// TestSplitFactoryLaneLabelAcceptsCanonicalAndLegacyLane: the canonical
// `worker-<n>` shape and the legacy `lane-<n>` shape both parse — the legacy
// shape is how labels written by a pre-rename launcher (or typed by an
// operator from older docs) stay readable.
func TestSplitFactoryLaneLabelAcceptsCanonicalAndLegacyLane(t *testing.T) {
	t.Parallel()

	for label, want := range map[string]int{"worker-1": 1, "worker-12": 12, "lane-3": 3} {
		if got, ok := SplitFactoryLaneLabel(label); !ok || got != want {
			t.Errorf("SplitFactoryLaneLabel(%q) = (%d, %v), want (%d, true)", label, got, ok, want)
		}
	}
	for _, label := range []string{"worker", "worker-", "worker-0", "worker-a", "Worker-3", "workers-3", "worker-3-x"} {
		if _, ok := SplitFactoryLaneLabel(label); ok {
			t.Errorf("SplitFactoryLaneLabel(%q) admitted a non-worker shape", label)
		}
	}
}

// TestLegacyFactoryLabelDetectionAndCanonicalization: the two pre-rename
// shapes (`lane-<n>`, `agent-<n>`) are recognised as legacy and map onto the
// one canonical `worker-<n>` namespace.
func TestLegacyFactoryLabelDetectionAndCanonicalization(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in        string
		legacy    bool
		canonical string
		ok        bool
	}{
		{"lane-3", true, "worker-3", true},
		{"agent-2", true, "worker-2", true},
		{"worker-5", false, "worker-5", true},
		{"plan", false, "", false},
		{"lane-0", false, "", false},
		{"", false, "", false},
	}
	for _, c := range cases {
		if got := IsLegacyFactoryLabel(c.in); got != c.legacy {
			t.Errorf("IsLegacyFactoryLabel(%q) = %v, want %v", c.in, got, c.legacy)
		}
		got, ok := CanonicalFactoryLabel(c.in)
		if ok != c.ok || got != c.canonical {
			t.Errorf("CanonicalFactoryLabel(%q) = (%q, %v), want (%q, %v)", c.in, got, ok, c.canonical, c.ok)
		}
	}
}

// TestNextFactoryWorkerNumberSpansLegacyShapes: the next free number is one
// past the highest LIVE claim across every shape, so a `-f worker` join never
// reuses a number a legacy-labelled live session still holds.
func TestNextFactoryWorkerNumberSpansLegacyShapes(t *testing.T) {
	t.Parallel()

	reg := map[string]FactoryWorkerEntry{
		"worker-1": {PID: 1},
		"lane-4":   {PID: 2},
		"agent-2":  {PID: 3},
		"worker-9": {PID: 4}, // dead — must not count
	}
	alive := func(pid int) bool { return pid != 4 }
	if got := NextFactoryWorkerNumber(reg, alive); got != 5 {
		t.Errorf("NextFactoryWorkerNumber = %d, want 5", got)
	}
}

// TestClaimFactoryWorkerNameCanonicalizesAndRespectsLegacyClaims covers the
// persisted-state compat read: factory.db rows written by a pre-rename
// launcher keep their numbers taken, and a legacy request is claimed under
// the canonical label.
func TestClaimFactoryWorkerNameCanonicalizesAndRespectsLegacyClaims(t *testing.T) {
	t.Parallel()

	alive := func(int) bool { return true }

	t.Run("legacy request claims the canonical label", func(t *testing.T) {
		t.Parallel()
		got, err := ClaimFactoryWorkerName(t.TempDir(), "lane-2", 101, alive)
		if err != nil || got != "worker-2" {
			t.Fatalf("claim lane-2 = (%q, %v), want worker-2", got, err)
		}
	})

	t.Run("legacy agent request claims the canonical label", func(t *testing.T) {
		t.Parallel()
		got, err := ClaimFactoryWorkerName(t.TempDir(), "agent-1", 101, alive)
		if err != nil || got != "worker-1" {
			t.Fatalf("claim agent-1 = (%q, %v), want worker-1", got, err)
		}
	})

	t.Run("live legacy row blocks its number", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
			"lane-3":  {PID: 201},
			"agent-4": {PID: 202},
		}); err != nil {
			t.Fatalf("seed: %v", err)
		}
		got, err := ClaimFactoryWorkerName(root, "worker-3", 203, alive)
		if err != nil || got != "worker-5" {
			t.Fatalf("claim worker-3 over live lane-3/agent-4 = (%q, %v), want worker-5", got, err)
		}
		reg := LoadFactoryRegistry(FactoryRegistryPath(root))
		if _, ok := reg["lane-3"]; !ok {
			t.Errorf("the live legacy row must survive the claim, registry = %v", reg)
		}
	})
}

// TestFactoryFreeSlotsHonoursLegacyLiveClaims: the lead loop's picker reads a
// live legacy-labelled claim as occupying its slot number.
func TestFactoryFreeSlotsHonoursLegacyLiveClaims(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
		"worker-2": {PID: 1},
		"agent-3":  {PID: 2},
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got := FactoryFreeSlots(root, 4, func(int) bool { return true })
	if !slices.Equal(got, []int{1, 4}) {
		t.Errorf("FactoryFreeSlots = %v, want [1 4]", got)
	}
}
