package kanban

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// TestMain drops the variables that name a repository (GIT_DIR, GIT_WORK_TREE,
// …) before any test runs, so the git fixtures in this package act on their
// own temp directories even when the test binary inherits a git hook's or a
// lane session's git context (GH #1691). The structural check in
// internal/gitenv fails if this call is removed.
func TestMain(m *testing.M) {
	if err := gitenv.ScrubProcess(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
