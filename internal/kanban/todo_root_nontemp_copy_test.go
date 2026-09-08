package kanban

// todo_root_nontemp_copy_test.go — SPEC-TODO-HOME-TEMP-GUARD-001 M2, the §C.1
// C-row disposition for this package: four tests that keep PASSING after the
// temporary-origin guard lands but stop EXERCISING the branch they were written
// for, because their t.TempDir() base is now a temporary origin and the guard
// answers before the home-fallback branch is reached.
//
// A test that passes vacuously does not fail, so "no unrelated test broke" says
// nothing about them. The disposition is a non-temporary COPY per original: the
// original stays exactly where it is (it remains a valid regression guard for
// behaviour UNDER the guard), and the copy re-reaches the original branch by
// declaring its base non-temporary through the REQ-THG-009 seam.
//
// Each copy carries the ORIGINAL assertion unchanged, plus one addition: a
// direct assertion on the discriminant's verdict for its own base
// (TempOriginReason(base) reports isTemp == false). That addition is what makes
// the copy non-vacuous — remove the seam stub and the copy goes RED on that
// line rather than sailing through. An earlier formulation asserted only "the
// copy's subject is the home root", which held for exactly one of these four:
// two assert an ABSENCE (which the guard leaves true) and two assert a value
// that EQUALS what the guard returns, so PASS could not separate the two states.
//
// Every test here overrides HomeDirFn and TempRootsFn, both process-global, so
// none of them run in parallel.

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// assertSeamHeard is the shared C-row discriminant assertion: this base must
// classify NON-temporary, which is only true when the injected root set was
// actually read. It is a positive assertion about the discriminant itself, not
// about the resolver's return value — the resolver's return is precisely what
// cannot separate the guarded and unguarded states in most of these fixtures.
func assertSeamHeard(t *testing.T, base string) {
	t.Helper()
	if reason, isTemp := TempOriginReason(base); isTemp {
		t.Fatalf("the injected temp-root set was not read: base %q still classifies temporary (reason %q) — "+
			"this copy would exercise the guard's refusal rather than the home-fallback branch it was written for",
			base, reason)
	}
}

// TestResolveTodoQueueRoot_PureFallbackWritesNothing_NonTemp — C-row copy of
// TestResolveTodoQueueRoot_PureFallbackWritesNothing (AC-WTQ-006).
//
// The original asserts an ABSENCE — that the fallback root was not created —
// which the guard leaves true for a different reason (no fallback root is ever
// computed). This copy keeps that assertion on a base where the fallback root
// IS computed, so the absence is again the pure resolver's restraint.
func TestResolveTodoQueueRoot_PureFallbackWritesNothing_NonTemp(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)
	assertSeamHeard(t, dir)

	local := seedLocalQueue(t, dir, 2)
	before, err := os.Stat(local)
	if err != nil {
		t.Fatalf("stat seeded local queue: %v", err)
	}
	fallbackRoot := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))

	time.Sleep(10 * time.Millisecond)
	_ = ResolveTodoQueueRoot(dir)

	after, err := os.Stat(local)
	if err != nil {
		t.Fatalf("local queue moved or removed by the pure resolver: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("local queue mtime changed: %v -> %v", before.ModTime(), after.ModTime())
	}
	if _, err := os.Stat(fallbackRoot); !os.IsNotExist(err) {
		t.Fatalf("pure resolver created the fallback root %q (stat err = %v)", fallbackRoot, err)
	}
}

// TestResolveTodoQueueRoot_ReadThroughToProjectLocal_NonTemp — C-row copy of
// TestResolveTodoQueueRoot_ReadThroughToProjectLocal (AC-WTQ-007 / decision D-2).
//
// The original asserts `got == dir`, and the guard's substitute root is ALSO
// dir — so the original's PASS cannot distinguish "read-through fired" from
// "the guard refused". This copy restores read-through as the reason.
func TestResolveTodoQueueRoot_ReadThroughToProjectLocal_NonTemp(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)
	assertSeamHeard(t, dir)
	seedLocalQueue(t, dir, 3)

	got := ResolveTodoQueueRoot(dir)
	if got != dir {
		t.Fatalf("read-through root = %q, want project-local root %q", got, dir)
	}
	rec, err := NewBacklogStore(BacklogPathForRoot(got)).Load()
	if err != nil {
		t.Fatalf("load through resolved root: %v", err)
	}
	if len(rec.Items) != 3 {
		t.Fatalf("read-through load holds %d items, want 3", len(rec.Items))
	}
}

// TestAdoptionLandsWhereConsumersRead_NonTemp — C-row copy of
// TestAdoptionLandsWhereConsumersRead, the regression guard for the incident
// where adopted cards were moved to a path no consumer read.
//
// This is the one original whose subject (the adopted queue under the home
// root) does differ from the guard's return, so its PASS was already
// discriminating; the copy is kept for uniformity of disposition and carries
// the seam assertion for the same reason as the others.
func TestAdoptionLandsWhereConsumersRead_NonTemp(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)
	assertSeamHeard(t, proj)

	seedLocalQueue(t, proj, 1)

	root := ResolveTodoQueueRootAdopting(proj)
	if want := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(proj)); root != want {
		t.Fatalf("adopting root = %q, want the home fallback %q — this copy must walk the adopt branch", root, want)
	}
	if _, err := os.Stat(BacklogPathForRoot(root)); err != nil {
		t.Fatalf("the queue is not readable where every consumer looks: %s (%v)",
			BacklogPathForRoot(root), err)
	}
	// bare-join-intentional: this asserts the old path is EMPTY, so it must name
	// it. The convention guard skips lines carrying this marker.
	stale := filepath.Join(root, "backlog.json") // bare-join-intentional
	if _, err := os.Stat(stale); err == nil {
		t.Errorf("a queue was left at the bare-join path %s, which no consumer reads", stale)
	}
}

// TestAdoptingAndPureResolversAgreeWhenAdoptionFails_NonTemp — C-row copy of
// TestAdoptingAndPureResolversAgreeWhenAdoptionFails (REQ-WTQ-005).
//
// The original deliberately makes adoption FAIL, after which both resolvers
// read through to `proj` — the same value the guard substitutes. Its PASS
// therefore could not tell the two states apart either.
func TestAdoptingAndPureResolversAgreeWhenAdoptionFails_NonTemp(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)
	assertSeamHeard(t, proj)

	seedLocalQueue(t, proj, 1)

	// Make the fallback root unwritable by occupying its parent with a file:
	// MkdirAll then fails and adoption returns having moved nothing.
	fallback, _ := homeTodoQueueRoot(proj)
	blocker := filepath.Dir(fallback)
	if err := os.MkdirAll(filepath.Dir(blocker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}

	adopting := ResolveTodoQueueRootAdopting(proj)
	pure := ResolveTodoQueueRoot(proj)
	if adopting != pure {
		t.Errorf("resolvers diverge when adoption fails: adopting=%s pure=%s", adopting, pure)
	}
	if _, err := os.Stat(BacklogPathForRoot(adopting)); err != nil {
		t.Errorf("the adopting resolver points at no readable queue: %s (%v)",
			BacklogPathForRoot(adopting), err)
	}
}
