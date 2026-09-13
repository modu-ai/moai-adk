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
	"strings"
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
	// INTENTIONAL UPDATE (t621): the original incident this guards is "adopted
	// cards were moved to a path no consumer reads", and the assertion below
	// states exactly that. The home-fallback ROOT that used to be asserted here
	// first turned out to be one such path itself — the layer below re-keys it,
	// landing the queue at ~/.moai/db/<key>-<hash>/todo while the statusline and
	// the console read ~/.moai/db/<key>/todo. What is kept is the home
	// property, on the queue rather than on the root.
	queue := BacklogPathForRoot(root)
	if !strings.HasPrefix(queue, home+string(filepath.Separator)) {
		t.Fatalf("adopted queue = %q, want it under the home directory %q — this copy must walk the non-git branch",
			queue, home)
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

// TestARefusedRelocationStillServesTheCards_NonTemp — C-row copy of
// TestBothEntryPointsServeTheCardsOnATemporaryBase (REQ-WTQ-005), on the
// home-based half.
//
// INTENTIONAL UPDATE (t621): the predecessor blocked adoptLocalTodoQueue's
// write. That function is gone — the one-time relocation now belongs to
// resolveStateDir — so the successor blocks THAT relocation instead, which is
// the same claim one layer down and the only place it can still fail: a
// relocation that cannot complete must leave the resolution serving the cards
// from where they already are, never reporting an empty queue beside them.
func TestARefusedRelocationStillServesTheCards_NonTemp(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)
	assertSeamHeard(t, proj)

	// A queue at a legacy location, which the resolution wants to relocate into
	// the home database on the adopting path.
	legacyDir := LegacyStateDirForRoot(proj)
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, backlogFileName),
		[]byte(`{"version":1,"last_seq":1,"items":[`+
			`{"id":"t1","text":"card 1","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"}]}`),
		0o600); err != nil {
		t.Fatal(err)
	}

	// Make the relocation target unbuildable by occupying its parent with a
	// file: MkdirAll then fails and the relocation returns having moved nothing.
	target := StateDirForRoot(proj)
	blocker := filepath.Dir(target)
	if err := os.MkdirAll(filepath.Dir(blocker), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}

	root := ResolveTodoQueueRootAdopting(proj)
	path := BacklogPathForRootAdopting(root)
	t.Logf("legacy=%s target=%s served=%s", legacyDir, target, path)
	if _, err := os.Stat(target); err == nil {
		t.Fatalf("precondition: the relocation target %q was built, so nothing was refused", target)
	}
	rec, err := NewBacklogStore(path).LoadPure()
	if err != nil || rec == nil {
		t.Fatalf("a refused relocation left no readable queue at %s (err %v)", path, err)
	}
	if len(rec.Items) != 1 {
		t.Errorf("a refused relocation serves %d cards from %s, want the 1 already queued", len(rec.Items), path)
	}
}
