package closure

import (
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/gitenv"
)

// TestMain drops the variables that name a repository (GIT_DIR, GIT_WORK_TREE,
// …) before any test runs. This package's tests execute the closuretest
// fixture builder, which shells out to git, so the same GH #1691 hazard as
// internal/closure/gitio applies here even though the structural check in
// internal/gitenv only recognizes direct git-start shapes in _test.go files.
func TestMain(m *testing.M) {
	if err := gitenv.ScrubProcess(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
