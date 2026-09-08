package kanban

// todo_root_temp_guard_test.go — SPEC-TODO-HOME-TEMP-GUARD-001 M2: the
// temporary-origin guard wired into the home-fallback branch.
//
// Producing acceptance criteria: AC-THG-001 (branches (a), (b), (c)),
// AC-THG-003 (a non-temporary non-git base KEEPS its home fallback — the
// scope guard), AC-THG-005 (the pure path stays silent and writes nothing),
// AC-THG-008 (the git branch never reaches the guard; key derivation is
// unchanged).
//
// Every test here overrides HomeDirFn and/or TempRootsFn, both process-global,
// so none of them run in parallel.

import (
	"os"
	"path/filepath"
	"testing"
)

// countHomeQueueDirs reports how many entries exist under
// <home>/.moai/todo — the canary measurement for "no home queue was created".
// An absent directory counts as zero, which is the point: the guard's success
// shape is that nothing was ever made there.
func countHomeQueueDirs(t *testing.T, home string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, ".moai", "todo"))
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("read canary home queue dir: %v", err)
	}
	return len(entries)
}

// declareNonTemporary points the REQ-THG-009 temp-root seam at a set that
// contains nothing the calling test uses, so a t.TempDir() base classifies
// NON-temporary and the home-fallback branch stays reachable.
//
// This is the only sanctioned way to build a "non-temporary" fixture here:
// the repository's [HARD] isolation discipline puts every fixture under
// t.TempDir(), which is inside os.TempDir() by definition. Without the seam
// the guard would ship unexercisable on its scope-preserving side.
func declareNonTemporary(t *testing.T) {
	t.Helper()
	stubTempRoots(t, filepath.Join(t.TempDir(), "a-root-that-contains-nothing"))
}

// TestTodoQueueRoot_TempOriginRefusesHomeQueue — AC-THG-001.
//
// The Given splits in two because the contamination path is only REACHED when
// the base carries a project-local queue: adoptLocalTodoQueue returns early
// when there is no local file, so branch (a) alone could report "no home
// directory was created" without the guard having done anything. Branch (b)
// walks the path that actually creates one, which is what makes the zero a
// result rather than a coincidence.
func TestTodoQueueRoot_TempOriginRefusesHomeQueue(t *testing.T) {
	t.Run("no local queue: both resolvers return the launch base", func(t *testing.T) {
		dir := t.TempDir() // not a git repository, and under os.TempDir() by construction
		home := t.TempDir()
		stubHome(t, home)

		if reason, isTemp := TempOriginReason(dir); !isTemp {
			t.Fatalf("precondition: base %q must classify temporary under the production root set (reason %q)", dir, reason)
		}
		homeRoot := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))

		for _, tc := range []struct {
			name string
			got  string
		}{
			{"pure", ResolveTodoQueueRoot(dir)},
			{"adopting", ResolveTodoQueueRootAdopting(dir)},
		} {
			if tc.got == homeRoot {
				t.Errorf("%s resolver returned the home queue root %q for a temporary origin", tc.name, tc.got)
			}
			if tc.got != dir {
				t.Errorf("%s resolver = %q, want the launch base %q", tc.name, tc.got, dir)
			}
			// The root-layer property REQ-THG-001 fixes: what a consumer READS
			// through the returned root must be the project's canonical local
			// queue path. String equality alone would accept a value one layer
			// off (base/.moai/state/todo), which consumers would then extend
			// into a path nothing writes.
			if got, want := BacklogPathForRoot(tc.got), BacklogPathForRoot(dir); got != want {
				t.Errorf("%s resolver: BacklogPathForRoot(returned) = %q, want the local canonical %q", tc.name, got, want)
			}
		}

		if n := countHomeQueueDirs(t, home); n != 0 {
			t.Errorf("canary HOME polluted: %d entr(ies) under %s/.moai/todo", n, home)
		}
	})

	t.Run("local queue present: the returned root is where that queue is read", func(t *testing.T) {
		dir := t.TempDir()
		home := t.TempDir()
		stubHome(t, home)
		local := seedLocalQueue(t, dir, 3)

		root := ResolveTodoQueueRootAdopting(dir)
		if root != dir {
			t.Fatalf("adopting root = %q, want the launch base %q", root, dir)
		}

		// [the decisive assertion] the queue is actually reachable through the
		// returned root, and it is the fixture's own file.
		queue := BacklogPathForRoot(root)
		if _, err := os.Stat(queue); err != nil {
			t.Fatalf("the queue is not readable where every consumer looks: %s (%v)", queue, err)
		}
		if queue != local {
			t.Errorf("BacklogPathForRoot(returned) = %q, want the seeded local queue %q", queue, local)
		}
		// The local queue stays put — a refusal migrates nothing. Asserted
		// BEFORE the load below: BacklogStore.Load performs the storage
		// cutover, so a stat taken after it measures that migration rather
		// than the guard's refusal.
		if _, err := os.Stat(local); err != nil {
			t.Errorf("local queue left its original path %q: %v", local, err)
		}
		rec, err := NewBacklogStore(queue).Load()
		if err != nil {
			t.Fatalf("load through the returned root: %v", err)
		}
		if len(rec.Items) != 3 {
			t.Errorf("queue read through the returned root holds %d items, want 3", len(rec.Items))
		}
		if n := countHomeQueueDirs(t, home); n != 0 {
			t.Errorf("canary HOME polluted: %d entr(ies) under %s/.moai/todo", n, home)
		}
	})

	t.Run("temporary origin AND home unresolvable: the base still wins", func(t *testing.T) {
		// The two conditions are not exclusive, and REQ-THG-001 fixes the
		// return on their intersection. This is what pins the discriminant
		// AHEAD of the home-resolution outcome: placed after it, the return
		// would be homeTodoQueueRoot's no-home value (resolveStateDir(base,
		// false) = base/.moai/state/todo) — one layer off, by ordering alone.
		dir := t.TempDir()
		orig := HomeDirFn
		HomeDirFn = func() (string, error) { return "", os.ErrNotExist }
		t.Cleanup(func() { HomeDirFn = orig })

		got := ResolveTodoQueueRoot(dir)
		if got != dir {
			t.Fatalf("temp-origin ∧ home-unresolvable root = %q, want the launch base %q", got, dir)
		}
		layerMisaligned := filepath.Join(dir, ".moai", "state", "todo")
		if got == layerMisaligned {
			t.Errorf("returned the state directory %q rather than the base", layerMisaligned)
		}
		if p, want := BacklogPathForRoot(got), BacklogPathForRoot(dir); p != want {
			t.Errorf("BacklogPathForRoot(returned) = %q, want the local canonical %q", p, want)
		}
		if _, err := os.Stat(layerMisaligned); !os.IsNotExist(err) {
			t.Errorf("the branch created %q (stat err = %v)", layerMisaligned, err)
		}
	})
}

// TestTodoQueueRoot_NonTempNonGitKeepsHomeFallback — AC-THG-003, the scope
// guard. The rejected wide reading ("refuse the home queue for ANY non-git
// base") fails here, which is the whole point of the criterion: the guard's
// trigger is a temporary origin, never the absence of git.
//
// The fixture is only constructible because of the REQ-THG-009 seam. This
// repository's [HARD] isolation discipline puts every fixture under
// t.TempDir(), which is inside os.TempDir() by definition, so "just use a
// non-temporary directory" is not an available move — the seam is the
// sanctioned way to declare a base non-temporary while staying inside the
// discipline.
func TestTodoQueueRoot_NonTempNonGitKeepsHomeFallback(t *testing.T) {
	dir := t.TempDir() // not a git repository
	home := t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)

	// The positive assertion that the seam was actually heard. Without it this
	// test passes vacuously when the stub is ignored — and a stub that is not
	// read is exactly the shape that would make a whole family of assertions
	// meaningless.
	if reason, isTemp := TempOriginReason(dir); isTemp {
		t.Fatalf("the injected root set did not take: base %q still classifies temporary (reason %q)", dir, reason)
	}

	want := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))
	if got := ResolveTodoQueueRoot(dir); got != want {
		t.Errorf("pure resolver = %q, want the home fallback %q — the documented \"exactly one queue\" availability was withdrawn", got, want)
	}

	// The adopting path's adopt-not-shadow migration is unchanged too.
	local := seedLocalQueue(t, dir, 2)
	if got := ResolveTodoQueueRootAdopting(dir); got != want {
		t.Fatalf("adopting resolver = %q, want the home fallback %q", got, want)
	}
	if _, err := os.Stat(BacklogPathForRoot(want)); err != nil {
		t.Errorf("adoption did not land where consumers read: %s (%v)", BacklogPathForRoot(want), err)
	}
	if _, err := os.Stat(local); !os.IsNotExist(err) {
		t.Errorf("local queue still at %q after adoption (stat err = %v)", local, err)
	}
}

// TestTodoQueueRoot_PureGuardIsSilent — AC-THG-005, the console half.
//
// The command path surfaces guidance (internal/cli); the PURE path must not.
// It returns a string and nothing else: no write, no error channel, no
// message — which is what keeps a console page render from touching the
// operator's backlog (REQ-THG-007).
func TestTodoQueueRoot_PureGuardIsSilent(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	stubHome(t, home)
	local := seedLocalQueue(t, dir, 2)
	before, err := os.Stat(local)
	if err != nil {
		t.Fatalf("stat seeded local queue: %v", err)
	}

	got := ResolveTodoQueueRoot(dir)
	if got != dir {
		t.Fatalf("pure resolver on a temporary origin = %q, want %q", got, dir)
	}

	after, err := os.Stat(local)
	if err != nil {
		t.Fatalf("the pure resolver moved or removed the local queue: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("the pure resolver changed the local queue's mtime: %v -> %v", before.ModTime(), after.ModTime())
	}
	if n := countHomeQueueDirs(t, home); n != 0 {
		t.Errorf("the pure resolver created %d entr(ies) under the canary HOME queue dir", n)
	}
	// TempOriginRefusal is the command path's read of the same decision, and it
	// is read-only for the same reason.
	sub, matched, refused := TempOriginRefusal(dir)
	if !refused {
		t.Fatalf("TempOriginRefusal reported no refusal for temporary origin %q", dir)
	}
	if sub != dir {
		t.Errorf("TempOriginRefusal substitute = %q, want %q", sub, dir)
	}
	if matched == "" {
		t.Error("TempOriginRefusal named no matched temp root; guidance could not name what it matched")
	}
	if n := countHomeQueueDirs(t, home); n != 0 {
		t.Errorf("TempOriginRefusal created %d entr(ies) under the canary HOME queue dir", n)
	}
}

// TestTodoQueueRoot_GitBranchUnreachedByGuard — AC-THG-008, the git-branch
// invariant plus key-derivation invariance.
//
// A repository living UNDER a temporary root is still resolved by
// primaryCheckoutRoot, which answers first: the guard is never consulted, so
// the guard cannot withdraw a real repository's queue. And
// TodoQueueProjectKey is untouched by the guard — changing its derivation
// would make every existing home queue unreachable.
func TestTodoQueueRoot_GitBranchUnreachedByGuard(t *testing.T) {
	primary := t.TempDir() // under os.TempDir(): a temporary origin BY LOCATION
	initTodoRootGitRepo(t, primary)
	home := t.TempDir()
	stubHome(t, home)

	if _, isTemp := TempOriginReason(primary); !isTemp {
		t.Fatalf("precondition: %q must classify temporary, or this test proves nothing", primary)
	}
	if got := ResolveTodoQueueRootAdopting(primary); !sameTodoRootDir(got, primary) {
		t.Errorf("git-resolvable base under a temp root = %q, want the primary checkout %q", got, primary)
	}
	if n := countHomeQueueDirs(t, home); n != 0 {
		t.Errorf("the git branch touched the canary home queue dir: %d entr(ies)", n)
	}

	// Key derivation is a pure function of the path and the guard does not
	// touch it: same base, same key, before and after.
	base := t.TempDir()
	first := TodoQueueProjectKey(base)
	_ = ResolveTodoQueueRoot(base)
	if second := TodoQueueProjectKey(base); second != first {
		t.Errorf("TodoQueueProjectKey drifted across a guarded resolution: %q -> %q", first, second)
	}
}

// TestTempOrigin_LexicalResembler — the FALSE-POSITIVE direction, pinned.
//
// The discriminant matches each temp root against two anchors (its normalized
// form and its lexical Clean form). Widening the anchor set makes a
// genuinely-inside path easier to recognize; the risk it must not carry is the
// other direction — a legitimate project misread as temporary, which withdraws
// a real operator's queue. That misclassification is the expensive one, and it
// is silent.
//
// The fixture: a base whose spelling RESEMBLES a temp root (it carries the
// root's basename as its own leading component) but which resolves outside
// every root in the injected set. It must classify non-temporary.
func TestTempOrigin_LexicalResembler(t *testing.T) {
	tempRoot := filepath.Join(t.TempDir(), "tmp")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		t.Fatalf("mkdir temp root: %v", err)
	}
	stubTempRoots(t, tempRoot)

	// Same basename, different parent: /A/tmp is the root, /B/tmp/project is
	// the resembler. A comparison that dropped the parent — matching on the
	// tail, or on a suffix — would call this temporary.
	elsewhere := t.TempDir()
	resembler := filepath.Join(elsewhere, "tmp", "project")
	if err := os.MkdirAll(resembler, 0o755); err != nil {
		t.Fatalf("mkdir resembler: %v", err)
	}
	if reason, isTemp := TempOriginReason(resembler); isTemp {
		t.Errorf("resembler %q classified temporary against root %q (reason %q) — a real project's queue would be withdrawn",
			resembler, tempRoot, reason)
	}

	// Control: the instrument WOULD have found something. A path genuinely
	// beneath the injected root classifies temporary, so the negative above is
	// a measurement rather than a silent no-op.
	inside := filepath.Join(tempRoot, "project")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatalf("mkdir inside: %v", err)
	}
	if _, isTemp := TempOriginReason(inside); !isTemp {
		t.Fatalf("control failed: %q is beneath the injected root %q yet classified NOT temporary — the negative above proves nothing",
			inside, tempRoot)
	}
}
