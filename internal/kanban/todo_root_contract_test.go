package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAdoptionLandsWhereConsumersRead pins the path contract both resolvers and
// every caller share. Adoption used to write <root>/backlog.json while callers
// resolve the store through BacklogPathForRoot — <root>/.moai/state/todo/
// backlog.json — so an adopted queue was moved somewhere nothing reads and the
// operator's cards silently disappeared from `moai todo`.
func TestAdoptionLandsWhereConsumersRead(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return home, nil }
	defer func() { HomeDirFn = orig }()

	seedLocalQueue(t, proj, 1)

	root := ResolveTodoQueueRootAdopting(proj)
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

// TestBothEntryPointsServeTheCardsOnATemporaryBase covers what decision D-2
// (REQ-WTQ-005) exists for: the console must not report an empty queue while
// `moai todo` reports N.
//
// INTENTIONAL UPDATE (t621): the predecessor made adoptLocalTodoQueue's
// best-effort write FAIL and then asserted the two resolvers returned the same
// string. Both halves stopped carrying the claim. The function is gone, so its
// blocker was inert; and the two resolvers now share one body, so comparing
// their returns is a tautology — it would hold with every branch below it
// deleted. The claim is therefore asserted where it can still fail: on what
// each entry point READS. A temporary base keeps its queue project-local, so
// this is the project-local half; the home half is the _NonTemp copy.
func TestBothEntryPointsServeTheCardsOnATemporaryBase(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return home, nil }
	defer func() { HomeDirFn = orig }()

	seedLocalQueue(t, proj, 1)
	if reason, isTemp := TempOriginReason(proj); !isTemp {
		t.Fatalf("precondition: this half needs a temporary base; %q classified non-temporary (reason %q)", proj, reason)
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{"pure", BacklogPathForRoot(ResolveTodoQueueRoot(proj))},
		{"adopting", BacklogPathForRootAdopting(ResolveTodoQueueRootAdopting(proj))},
	} {
		rec, err := NewBacklogStore(tc.path).LoadPure()
		if err != nil || rec == nil {
			t.Errorf("%s entry point reads no queue at %s (err %v)", tc.name, tc.path, err)
			continue
		}
		if len(rec.Items) != 1 {
			t.Errorf("%s entry point reads %d cards at %s, want 1", tc.name, len(rec.Items), tc.path)
		}
	}
}
