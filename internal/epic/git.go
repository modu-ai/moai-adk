package epic

import (
	"os/exec"
)

// newGitRevParseCommand builds the `git rev-parse HEAD` invocation for the
// given working directory. Kept in its own file so cross-platform build (B1)
// can override the command construction if a platform needs a different path
// lookup; the default os/exec implementation resolves `git` via PATH on every
// GOOS/GOARCH combination, so no build-tag split is required today.
func newGitRevParseCommand(dir string) *exec.Cmd {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	return cmd
}
