//go:build windows

package cli

import (
	"os"
	"os/exec"
)

// codexDirectAnchorPID is the pid the worktree lock names on the direct
// path. Windows has no exec-in-place: the launcher starts the Codex child and
// waits for it, so the waiting launcher's pid lives exactly as long as the
// session does. If only the launcher is killed while the child survives, the
// lock reads as dead too early — the known residual of this platform.
func codexDirectAnchorPID() int { return os.Getpid() }

// defaultCodexDirectLaunch starts Codex and waits for it; the child's exit
// code reaches the caller through the returned error. It writes no factory
// state: the codex factory path is retired (SPEC-CODEX-FACTORY-RETIRE-001
// REQ-CFR-008).
func defaultCodexDirectLaunch(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
