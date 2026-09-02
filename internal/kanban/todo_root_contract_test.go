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

// TestAdoptingAndPureResolversAgreeWhenAdoptionFails covers the branch decision
// D-2 (REQ-WTQ-005) exists for, on the one path no earlier criterion exercised.
// adoptLocalTodoQueue is best-effort; when it cannot write, the adopting path
// must still resolve to the queue that holds the cards rather than report an
// empty fallback while the console reports N.
func TestAdoptingAndPureResolversAgreeWhenAdoptionFails(t *testing.T) {
	home, proj := t.TempDir(), t.TempDir()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return home, nil }
	defer func() { HomeDirFn = orig }()

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
