package session

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// t611: ResolveBlocker must only resolve a blocker whose Phase and SPECID
// match the request, and must leave every other blocker file byte-unchanged.

// resolveScopeFixture records the given blockers and returns the store, the
// state dir, and a byte snapshot of every blocker file keyed by path.
func resolveScopeFixture(t *testing.T, reports []BlockerReport) (*FileSessionStore, string, map[string][]byte) {
	t.Helper()
	dir := t.TempDir()
	s := NewFileSessionStore(dir, time.Hour)
	for _, r := range reports {
		if err := s.RecordBlocker(r); err != nil {
			t.Fatalf("RecordBlocker(%s/%s): %v", r.Phase, r.SPECID, err)
		}
	}
	return s, dir, snapshotBlockerFiles(t, dir, len(reports))
}

func snapshotBlockerFiles(t *testing.T, dir string, want int) map[string][]byte {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "blocker-*.json"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(paths) != want {
		t.Fatalf("blocker files = %d, want %d", len(paths), want)
	}
	snap := make(map[string][]byte, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		snap[p] = data
	}
	return snap
}

// assertOnlyTargetResolved checks that exactly the blocker with the given
// Timestamp was resolved with resolution, and every other file is
// byte-identical to its pre-call snapshot. A zero target asserts that no
// file changed at all.
func assertOnlyTargetResolved(t *testing.T, dir string, before map[string][]byte, target time.Time, resolution string) {
	t.Helper()
	after := snapshotBlockerFiles(t, dir, len(before))
	targetSeen := false
	for p, old := range before {
		cur, ok := after[p]
		if !ok {
			t.Errorf("blocker file disappeared: %s", filepath.Base(p))
			continue
		}
		var b BlockerReport
		if err := json.Unmarshal(cur, &b); err != nil {
			t.Fatalf("unmarshal %s: %v", p, err)
		}
		if !target.IsZero() && b.Timestamp.Equal(target) {
			targetSeen = true
			if !b.Resolved || b.Resolution != resolution {
				t.Errorf("target %s not resolved: resolved=%t resolution=%q", filepath.Base(p), b.Resolved, b.Resolution)
			}
			continue
		}
		if !bytes.Equal(old, cur) {
			t.Errorf("non-target blocker changed: %s (spec=%q phase=%q resolved=%t resolution=%q)",
				filepath.Base(p), b.SPECID, b.Phase, b.Resolved, b.Resolution)
		}
	}
	if !target.IsZero() && !targetSeen {
		t.Errorf("target blocker with timestamp %s not found", target)
	}
}

var resolveScopeBase = time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)

// Cell (i): exactly one blocker matches phase+specID; the non-matching ones
// are older, so this control holds before and after the fix.
func TestResolveBlockerScope_SingleMatchResolved(t *testing.T) {
	target := resolveScopeBase.Add(2 * time.Minute)
	s, dir, before := resolveScopeFixture(t, []BlockerReport{
		{Phase: PhasePlan, SPECID: "SPEC-B", Timestamp: resolveScopeBase},
		{Phase: PhaseRun, SPECID: "SPEC-C", Timestamp: resolveScopeBase.Add(time.Minute)},
		{Phase: PhaseRun, SPECID: "SPEC-A", Timestamp: target},
	})
	if err := s.ResolveBlocker(PhaseRun, "SPEC-A", "approved A"); err != nil {
		t.Fatalf("ResolveBlocker: %v", err)
	}
	assertOnlyTargetResolved(t, dir, before, target, "approved A")
}

// Cell (ii): no blocker matches but other unresolved blockers exist; nothing
// is resolved and an error is returned.
func TestResolveBlockerScope_NoMatchResolvesNothing(t *testing.T) {
	s, dir, before := resolveScopeFixture(t, []BlockerReport{
		{Phase: PhasePlan, SPECID: "SPEC-B", Timestamp: resolveScopeBase},
		{Phase: PhaseRun, SPECID: "SPEC-C", Timestamp: resolveScopeBase.Add(time.Minute)},
		{Phase: PhasePlan, SPECID: "SPEC-A", Timestamp: resolveScopeBase.Add(2 * time.Minute)},
	})
	if err := s.ResolveBlocker(PhaseRun, "SPEC-A", "approved A"); err == nil {
		t.Error("ResolveBlocker with no matching blocker returned nil error")
	}
	assertOnlyTargetResolved(t, dir, before, time.Time{}, "")
}

// Cell (iii): a different SPEC's blocker (same phase) is more recent than the
// matching one; only the matching one is resolved.
func TestResolveBlockerScope_NewerOtherSpecUntouched(t *testing.T) {
	target := resolveScopeBase
	s, dir, before := resolveScopeFixture(t, []BlockerReport{
		{Phase: PhaseRun, SPECID: "SPEC-A", Timestamp: target},
		{Phase: PhaseRun, SPECID: "SPEC-B", Timestamp: resolveScopeBase.Add(time.Minute)},
	})
	if err := s.ResolveBlocker(PhaseRun, "SPEC-A", "approved A"); err != nil {
		t.Fatalf("ResolveBlocker: %v", err)
	}
	assertOnlyTargetResolved(t, dir, before, target, "approved A")
}

// Cell (iv): the same SPEC in a different phase is more recent; only the
// matching phase is resolved.
func TestResolveBlockerScope_NewerOtherPhaseUntouched(t *testing.T) {
	target := resolveScopeBase
	s, dir, before := resolveScopeFixture(t, []BlockerReport{
		{Phase: PhaseRun, SPECID: "SPEC-A", Timestamp: target},
		{Phase: PhasePlan, SPECID: "SPEC-A", Timestamp: resolveScopeBase.Add(time.Minute)},
	})
	if err := s.ResolveBlocker(PhaseRun, "SPEC-A", "approved A"); err != nil {
		t.Fatalf("ResolveBlocker: %v", err)
	}
	assertOnlyTargetResolved(t, dir, before, target, "approved A")
}

// Report criterion (audit probe fixture): SPEC-A/run at T and SPEC-B/plan at
// T+1m; resolving SPEC-A/run resolves SPEC-A and leaves SPEC-B byte-unchanged.
func TestResolveBlockerScope_ReportCriterion(t *testing.T) {
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	s, dir, before := resolveScopeFixture(t, []BlockerReport{
		{Phase: PhaseRun, SPECID: "SPEC-A", Timestamp: now},
		{Phase: PhasePlan, SPECID: "SPEC-B", Timestamp: now.Add(time.Minute)},
	})
	if err := s.ResolveBlocker(PhaseRun, "SPEC-A", "approved A"); err != nil {
		t.Fatalf("ResolveBlocker: %v", err)
	}
	assertOnlyTargetResolved(t, dir, before, now, "approved A")
}

// Design edge: a record with empty Phase/SPECID (filename carries "unknown")
// matches no non-empty request, even when it is the most recent.
func TestResolveBlockerScope_EmptyFieldRecordNotMatched(t *testing.T) {
	target := resolveScopeBase
	s, dir, before := resolveScopeFixture(t, []BlockerReport{
		{Phase: PhaseRun, SPECID: "SPEC-A", Timestamp: target},
		{Timestamp: resolveScopeBase.Add(time.Minute)},
	})
	if err := s.ResolveBlocker(PhaseRun, "SPEC-A", "approved A"); err != nil {
		t.Fatalf("ResolveBlocker: %v", err)
	}
	assertOnlyTargetResolved(t, dir, before, target, "approved A")
}
