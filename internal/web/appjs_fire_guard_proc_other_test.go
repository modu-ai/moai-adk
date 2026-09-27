//go:build !unix

package web

import (
	"os/exec"
	"testing"
)

// fireGuardSetProcessGroup is a no-op off unix. The gated test-browser CI job
// that runs the fire guard is Linux-only; process-group teardown lives in the
// unix variant (card t1098).
func fireGuardSetProcessGroup(cmd *exec.Cmd) {}

// fireGuardKillAndReap keeps the parent-only teardown off unix.
func fireGuardKillAndReap(t *testing.T, cmd *exec.Cmd, done <-chan struct{}) {
	t.Helper()
	_ = cmd.Process.Kill()
	<-done
}
