package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

func gtdTouch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
}

// F4 (turn-end gate relay #3, card t1538): a queue in the MoAI home layout,
// <home>/.moai/db/<key>/todo, is walked up to the home directory — which
// itself carries a .moai — and the old walk took the home directory for the
// PROJECT root, picking the home project's factory database instead of the
// queue's own sibling <home>/.moai/db/<key>/factory/factory.db.
func TestGTDFactoryDBForHomeLayoutQueueIsItsOwnSibling(t *testing.T) {
	home := t.TempDir()
	moaiHome := filepath.Join(home, ".moai")
	t.Setenv(paths.EnvHome, moaiHome)
	queueDir := filepath.Join(moaiHome, "db", "projkey", "todo")
	if err := os.MkdirAll(queueDir, 0o700); err != nil {
		t.Fatal(err)
	}
	store := NewBacklogStore(filepath.Join(queueDir, "backlog.json"))

	// The home directory resolves as a project too, and has its own factory
	// database — the wrong answer for this queue.
	homeFactory, err := homestate.FactoryDBPath(home)
	if err != nil {
		t.Fatal(err)
	}
	gtdTouch(t, homeFactory)

	// No sibling yet: the queue's project has no factory database, and the
	// home project's is not a substitute.
	if got := gtdFactoryDBForStore(store); got != "" {
		t.Fatalf("factory db = %q for a home-layout queue with no sibling, want \"\" — the home directory is not this queue's project", got)
	}

	sibling := filepath.Join(moaiHome, "db", "projkey", "factory", "factory.db")
	gtdTouch(t, sibling)
	if got := gtdFactoryDBForStore(store); got != sibling {
		t.Fatalf("factory db = %q, want the queue's own sibling %q", got, sibling)
	}
}

// The project-local layout keeps resolving through the project root: a queue
// at <root>/.moai/state/todo belongs to <root>'s factory database.
func TestGTDFactoryDBForProjectLocalQueueUsesProjectRoot(t *testing.T) {
	root, store, _ := runtimeFixture(t)
	want, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	gtdTouch(t, want)
	if got := gtdFactoryDBForStore(store); got != want {
		t.Fatalf("factory db = %q, want the project root's %q", got, want)
	}
}
