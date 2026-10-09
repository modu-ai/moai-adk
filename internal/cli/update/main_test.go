package update

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// TestMain drops the variables that name a repository (GIT_DIR, GIT_WORK_TREE,
// …) before any test runs: this package's fixtures drive the update flow,
// which shells out to git, and an inherited git context from a hook or a lane
// session would point those fixtures at the caller's checkout. The structural
// check in internal/gitenv fails if this call is removed.
func TestMain(m *testing.M) {
	if err := gitenv.ScrubProcess(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
