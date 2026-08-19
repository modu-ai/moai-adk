// Package execerr keeps subprocess exit failures printable without exposing
// raw *exec.ExitError in %w wrap chains.
//
// *exec.ExitError happens to satisfy the CLI's ExitCoder interface
// (`ExitCode() int`), so a %w-wrapped subprocess failure that reaches
// cmd/moai would be mistaken for an intentional exit code: main adopts the
// subprocess's raw code and fang's error handler stays silent (card t130;
// the t41/t129 rc=128 silent-failure recurrences). internal/cli's
// ResolveExitCode refuses the raw type at the seam; this package is the
// producer-side half — wrap sites describe the failure through StatusDetail
// instead of chaining the raw error. The CommandError type in
// internal/core/git is the same discipline for that package's own execGit.
package execerr

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// StatusDetail describes a subprocess failure for an error message: the
// captured stderr when the process ran and printed one, otherwise
// "exited with status N". For errors that are not subprocess exits (binary
// missing, start failure) it returns err.Error() unchanged — those types do
// not satisfy ExitCoder and were always safe to wrap.
//
// Use it as the error text at exec wrap sites:
//
//	out, err := cmd.Output()
//	if err != nil {
//	    return "", fmt.Errorf("git log failed: %s", execerr.StatusDetail(err))
//	}
//
// os/exec populates ExitError.Stderr automatically for Output() and
// CombinedOutput() calls; sites that capture stderr into their own buffer
// append it themselves and only need the status line.
func StatusDetail(err error) string {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if s := strings.TrimSpace(string(ee.Stderr)); s != "" {
			return fmt.Sprintf("exited with status %d: %s", ee.ExitCode(), s)
		}
		return fmt.Sprintf("exited with status %d", ee.ExitCode())
	}
	return err.Error()
}
