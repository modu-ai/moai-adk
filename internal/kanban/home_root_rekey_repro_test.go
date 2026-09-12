package kanban

// home_root_rekey_repro_test.go — t621: the SECOND layer of the queue-root
// resolution t549 opened. t549 repaired homeTodoQueueRoot's NO-home branch
// (it returned a state directory where callers expect a root). Two residues
// were left on the branch where home DOES resolve:
//
//	A. Re-keying. homeTodoQueueRoot returns ~/.moai/todo/<key(base)>, and every
//	   consumer extends that with BacklogPathForRoot, which re-derives the home
//	   location from whatever root it is handed — producing
//	   ~/.moai/db/<key(~/.moai/todo/<key(base)>)>/todo. The key is computed from
//	   a home path instead of the project, so the adopting resolver MOVES the
//	   queue to a doubly-keyed directory the pure resolver never names.
//
//	B. Layout blindness. fallbackTodoQueueRoot decides read-through with
//	   os.Stat on the `backlog.json` name, but the storage engine's steady state
//	   is a sibling `backlog.db` with no json beside it. A real queue in the
//	   normal layout is therefore invisible to the predicate.
//
// Every test here overrides process-global seams (HomeDirFn, TempRootsFn,
// MOAI_HOME), so none of them run in parallel.

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// reachHomeBranch installs the seams that put base on the home-resolvable
// fallback branch of both resolvers, and fails the test if any of the four
// earlier branches would intercept it first. Each precondition is a way a
// test below could pass without ever reaching the code under test.
func reachHomeBranch(t *testing.T, base string) (home string) {
	t.Helper()
	t.Setenv(paths.EnvHome, "")
	home = t.TempDir()
	stubHome(t, home)
	declareNonTemporary(t)

	if _, ok := primaryCheckoutRoot(base); ok {
		t.Fatalf("precondition: base %q must not resolve as a git checkout", base)
	}
	if explicitMoaiHome() {
		t.Fatalf("precondition: MOAI_HOME must not be an absolute override")
	}
	if reason, isTemp := TempOriginReason(base); isTemp {
		t.Fatalf("precondition: base %q still classifies temporary (reason %q)", base, reason)
	}
	if !pathInsideTempDir(base) {
		t.Fatalf("precondition: base %q is not inside os.TempDir(), fallback branch unreachable", base)
	}
	if _, ok := homeTodoQueueRoot(base); !ok {
		t.Fatalf("precondition: home must resolve, so the home branch is the one under test")
	}
	return home
}

// TestT621_HomeBranchRootIsNotReKeyed — residue A.
//
// The invariant: whatever root a resolver returns, extending it with
// BacklogPathForRoot must name the project's ONE queue file. A root that is
// itself re-keyed by the extension names a second location, and the two
// resolvers then disagree about where the operator's cards live — the exact
// fork this resolution exists to prevent.
func TestT621_HomeBranchRootIsNotReKeyed(t *testing.T) {
	proj := t.TempDir()
	reachHomeBranch(t, proj)

	canonical := BacklogPathForRoot(proj)
	seedLocalQueue(t, proj, 3)
	t.Logf("canonical queue path = %s", canonical)

	// Control: the fixture is readable where the project's own root names it.
	// Without it a zero below could be an unreadable fixture, not a wrong path.
	if n, err := pureItemCount(canonical); n != 3 {
		t.Fatalf("control: canonical read holds %d items (err %v), want 3", n, err)
	}

	homeRoot, _ := homeTodoQueueRoot(proj)
	t.Logf("homeTodoQueueRoot        = %s", homeRoot)
	t.Logf("BacklogPathForRoot(home) = %s", BacklogPathForRoot(homeRoot))

	// The adopting resolver runs FIRST: it is the one with the side effect, and
	// the pure resolver must agree with the tree it leaves behind.
	adopting := ResolveTodoQueueRootAdopting(proj)
	pure := ResolveTodoQueueRoot(proj)
	adoptingPath, purePath := BacklogPathForRoot(adopting), BacklogPathForRoot(pure)
	t.Logf("adopting root=%s path=%s", adopting, adoptingPath)
	t.Logf("pure     root=%s path=%s", pure, purePath)

	if adoptingPath != purePath {
		t.Errorf("the resolvers name two different queue files:\n  adopting %s\n  pure     %s",
			adoptingPath, purePath)
	}
	for _, tc := range []struct {
		name string
		path string
	}{{"adopting", adoptingPath}, {"pure", purePath}} {
		// The identity, not merely a readable queue: agreeing on one WRONG
		// location satisfies "both read 3 items" and "the two agree", which is
		// how a re-keyed root passes both without naming the project's queue.
		if tc.path != canonical {
			t.Errorf("%s resolver's queue = %q, want the project's own %q",
				tc.name, tc.path, canonical)
		}
		if n, err := pureItemCount(tc.path); n != 3 {
			t.Errorf("%s resolver reads %d items through its root (err %v), want 3", tc.name, n, err)
		}
	}
}

// TestT621_StatuslineAnchorAndCommandPathReadOneQueue — residue A, stated as
// the damage rather than as a path shape.
//
// Not every surface goes through this file's resolvers. The statusline reads
// the queue under the STATE ANCHOR — the project root itself
// (internal/statusline/backlog.go resolveBoardRoot → BacklogCountsForRoot), and
// the console's file watch registers kanban.StateDirForRoot(projectRoot)
// (internal/web/events.go). A resolver root that carries a different project
// key therefore does not merely name an odd directory: it forks the queue
// between the command path and every anchor-based surface, which is the failure
// this resolution was written to prevent.
func TestT621_StatuslineAnchorAndCommandPathReadOneQueue(t *testing.T) {
	proj := t.TempDir()
	reachHomeBranch(t, proj)

	// Write the way `moai todo` does: resolve the root, then open the store
	// under it.
	root := ResolveTodoQueueRootAdopting(proj)
	commandPath := BacklogPathForRoot(root)
	if _, _, err := NewBacklogStore(commandPath).Add("a card added through the command path"); err != nil {
		t.Fatalf("add through the command path: %v", err)
	}
	t.Logf("command path wrote %s", commandPath)

	// Control: the card is really there. A zero below must be a wrong path,
	// not a failed write.
	if n, err := pureItemCount(commandPath); n != 1 {
		t.Fatalf("control: the command path's own read holds %d items (err %v), want 1", n, err)
	}

	// Read the way the statusline does: from the project anchor, never through
	// the resolver.
	counts := BacklogCountsForRoot(proj)
	t.Logf("anchor read: available=%v queued=%d picked=%d (anchor path %s)",
		counts.Available, counts.Queued, counts.Picked, BacklogPathForRoot(proj))
	if counts.Queued != 1 {
		t.Errorf("the anchor surface counts %d queued cards (available=%v) while the command path holds 1 — the queue is forked",
			counts.Queued, counts.Available)
	}
}

// TestT621_FallbackReadsThroughToADbOnlyQueue — residue B.
//
// The engine's steady state is {backlog.db, no backlog.json}. A queue in that
// layout must be as visible to the read-through predicate as a legacy json one,
// or the pure resolver reports an empty fallback while the cards sit unread.
func TestT621_FallbackReadsThroughToADbOnlyQueue(t *testing.T) {
	proj := t.TempDir()
	reachHomeBranch(t, proj)

	canonical := BacklogPathForRoot(proj)
	if _, _, err := NewBacklogStore(canonical).Add("a card in the engine's steady-state layout"); err != nil {
		t.Fatalf("seed db queue: %v", err)
	}

	// Precondition: the layout really is db-only. If a json were present the
	// unrepaired predicate would find it and the test would pass vacuously.
	layout := inspectBacklogLayout(canonical)
	t.Logf("layout at %s: db=%v json=%v", canonical, layout.dbExists, layout.jsonExists)
	if !layout.dbExists || layout.jsonExists {
		t.Fatalf("precondition: want the db-only layout, got db=%v json=%v", layout.dbExists, layout.jsonExists)
	}
	if n, err := pureItemCount(canonical); n != 1 {
		t.Fatalf("control: canonical read holds %d items (err %v), want 1", n, err)
	}

	root := ResolveTodoQueueRoot(proj)
	path := BacklogPathForRoot(root)
	t.Logf("pure root=%s path=%s", root, path)
	if n, err := pureItemCount(path); n != 1 {
		t.Errorf("the pure resolver reads %d items (err %v), want the 1 queued card", n, err)
	}
}

// TestT621_AdoptionSeesADbOnlyLocalQueue — residue B on the adopting side.
//
// adoptLocalTodoQueue decides "is there anything to adopt" with the same
// json-name stat, and its move carries the json file alone. A db-only local
// queue is therefore adopted as nothing; worse, a {db, json} pair would be
// split across two roots. The property asserted is the one the operator sees:
// after adoption the resolver's own root names a readable queue holding the
// cards.
func TestT621_AdoptionSeesADbOnlyLocalQueue(t *testing.T) {
	proj := t.TempDir()
	reachHomeBranch(t, proj)

	canonical := BacklogPathForRoot(proj)
	if _, _, err := NewBacklogStore(canonical).Add("a card in the engine's steady-state layout"); err != nil {
		t.Fatalf("seed db queue: %v", err)
	}
	if layout := inspectBacklogLayout(canonical); !layout.dbExists || layout.jsonExists {
		t.Fatalf("precondition: want the db-only layout, got db=%v json=%v", layout.dbExists, layout.jsonExists)
	}

	root := ResolveTodoQueueRootAdopting(proj)
	path := BacklogPathForRoot(root)
	t.Logf("adopting root=%s path=%s", root, path)
	if n, err := pureItemCount(path); n != 1 {
		t.Errorf("the adopting resolver reads %d items (err %v), want the 1 queued card", n, err)
	}
	// The queue must not have been split in two by a move that carried one
	// artifact of a layout.
	if root != proj {
		if from, to := inspectBacklogLayout(canonical), inspectBacklogLayout(path); from.dbExists && to.dbExists {
			t.Errorf("the queue exists at BOTH %s and %s after adoption", canonical, path)
		}
	}
	_ = filepath.Dir(path)
}
