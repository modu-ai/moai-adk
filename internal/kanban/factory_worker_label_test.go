package kanban

// factory_worker_label_test.go — the factory lane label contract, updated to
// the SPEC-ROLE-NAMING-CODE-001 vocabulary: `lane-<n>` is canonical, the
// legacy `worker-<n>` / `agent-<n>` shapes are detection values only, and the
// shared legacy number space is replaced by the same-run refusal
// (REQ-RNC-004, REQ-RNC-009, REQ-RNC-022).

import (
	"slices"
	"strings"
	"testing"
)

// TestFactoryLaneLabelProducesLaneNotation pins the canonical label a
// factory lane session launches under: `lane-<n>`.
func TestFactoryLaneLabelProducesLaneNotation(t *testing.T) {
	t.Parallel()

	if got := FactoryLaneLabel(3); got != "lane-3" {
		t.Errorf("FactoryLaneLabel(3) = %q, want %q", got, "lane-3")
	}
}

// TestSplitFactoryLaneLabelCanonicalOnly: only the canonical `lane-<n>` shape
// parses as a lane — the legacy shapes map to no role (REQ-RNC-009).
func TestSplitFactoryLaneLabelCanonicalOnly(t *testing.T) {
	t.Parallel()

	for label, want := range map[string]int{"lane-1": 1, "lane-12": 12} {
		if got, ok := SplitFactoryLaneLabel(label); !ok || got != want {
			t.Errorf("SplitFactoryLaneLabel(%q) = (%d, %v), want (%d, true)", label, got, ok, want)
		}
	}
	for _, label := range []string{"worker-1", "agent-1", "lane", "lane-", "lane-0", "lane-a", "Lane-3", "lanes-3", "lane-3-x"} {
		if _, ok := SplitFactoryLaneLabel(label); ok {
			t.Errorf("SplitFactoryLaneLabel(%q) admitted a non-lane shape", label)
		}
	}
}

// TestLegacyFactoryLabelDetection: the pre-rename shapes (`worker-<n>`,
// `agent-<n>`) are recognised as legacy — detection only, never a role mapping
// (REQ-RNC-009).
func TestLegacyFactoryLabelDetection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in        string
		legacy    bool
		canonical bool
	}{
		{"worker-3", true, false},
		{"agent-2", true, false},
		{"lane-3", false, true}, // canonical now
		{"plan", false, false},
		{"lane-0", false, false},
		{"", false, false},
	}
	for _, c := range cases {
		if got := IsLegacyFactoryLabel(c.in); got != c.legacy {
			t.Errorf("IsLegacyFactoryLabel(%q) = %v, want %v", c.in, got, c.legacy)
		}
		if _, ok := CanonicalFactoryLabel(c.in); ok != c.canonical {
			t.Errorf("CanonicalFactoryLabel(%q) ok = %v, want %v (canonical only)", c.in, ok, c.canonical)
		}
	}
}

// TestNextFactoryWorkerNumberCountsCanonicalOnly: the next free number is one
// past the highest LIVE canonical claim. Legacy claims hold no number — a
// live legacy record refuses the join instead (design §4).
func TestNextFactoryWorkerNumberCountsCanonicalOnly(t *testing.T) {
	t.Parallel()

	reg := map[string]FactoryWorkerEntry{
		"lane-1":   {PID: 1},
		"worker-4": {PID: 2}, // legacy — holds no number
		"agent-2":  {PID: 3}, // legacy
		"lane-9":   {PID: 4}, // dead — must not count
	}
	alive := func(pid int) bool { return pid != 4 }
	if got := NextFactoryWorkerNumber(reg, alive); got != 2 {
		t.Errorf("NextFactoryWorkerNumber = %d, want 2", got)
	}
}

// TestClaimFactoryWorkerNameCanonicalOnly covers the claim input surface: a
// canonical `lane-<n>` request claims that label; a legacy request is refused
// with the canonical name (REQ-RNC-005 arrives in full at M2).
func TestClaimFactoryWorkerNameCanonicalOnly(t *testing.T) {
	t.Parallel()

	alive := func(int) bool { return true }

	t.Run("canonical request claims its label", func(t *testing.T) {
		t.Parallel()
		got, err := ClaimFactoryWorkerName(t.TempDir(), "lane-2", 101, "testrun", alive)
		if err != nil || got != "lane-2" {
			t.Fatalf("claim lane-2 = (%q, %v), want lane-2", got, err)
		}
	})

	for _, legacy := range []string{"worker-2", "agent-1"} {
		want := "lane-" + strings.TrimPrefix(strings.TrimPrefix(legacy, "worker-"), "agent-")
		t.Run("legacy request "+legacy+" refused naming "+want, func(t *testing.T) {
			t.Parallel()
			_, err := ClaimFactoryWorkerName(t.TempDir(), legacy, 101, "testrun", alive)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("claim %s = %v, want an error naming %s", legacy, err, want)
			}
		})
	}

	// A live legacy row with no run id (the legacy import shape) belongs to
	// no run: the claim proceeds past it and the row is never rewritten (P3).
	t.Run("empty-run legacy row is ignored, not rewritten", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
			"worker-3": {PID: 201},
		}); err != nil {
			t.Fatalf("seed: %v", err)
		}
		got, err := ClaimFactoryWorkerName(root, "lane-3", 203, "run7", alive)
		if err != nil || got != "lane-3" {
			t.Fatalf("claim lane-3 past empty-run legacy worker-3 = (%q, %v), want lane-3", got, err)
		}
		reg := LoadFactoryRegistry(FactoryRegistryPath(root))
		if _, ok := reg["worker-3"]; !ok {
			t.Errorf("the legacy row must survive untouched, registry = %v", reg)
		}
		if _, ok := reg["lane-3"]; !ok {
			t.Errorf("the new claim must be recorded, registry = %v", reg)
		}
	})
}

// TestFactoryFreeSlotsCountsCanonicalOnly: the lead loop's picker counts a
// live canonical `lane-<n>` claim as occupying its slot; legacy claims hold
// no slot number (REQ-RNC-022's refusal replaced number-space sharing).
func TestFactoryFreeSlotsCountsCanonicalOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := SaveFactoryRegistry(FactoryRegistryPath(root), map[string]FactoryWorkerEntry{
		"lane-2":   {PID: 1},
		"worker-3": {PID: 2}, // legacy — no slot number
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got := FactoryFreeSlots(root, 4, func(int) bool { return true })
	if !slices.Equal(got, []int{1, 3, 4}) {
		t.Errorf("FactoryFreeSlots = %v, want [1 3 4]", got)
	}
}
