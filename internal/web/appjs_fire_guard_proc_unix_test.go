//go:build unix

package web

import (
	"errors"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// fireGuardSetProcessGroup puts Chrome in its own process group so teardown
// can signal the browser and every child it forks (zygote, gpu-process,
// utility services, renderers) in one call (card t1098).
func fireGuardSetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// fireGuardKillAndReap SIGKILLs Chrome's whole process group, waits for the
// parent to be reaped, then polls until no group member remains. Waiting on
// the parent alone is not enough on Linux: the NetworkService utility was
// observed alive after the parent exited, still writing under the profile,
// so the profile TempDir's RemoveAll failed with "directory not empty".
func fireGuardKillAndReap(t *testing.T, cmd *exec.Cmd, done <-chan struct{}) {
	t.Helper()
	pgid := cmd.Process.Pid
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	<-done
	deadline := time.Now().Add(10 * time.Second)
	for {
		err := syscall.Kill(-pgid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		if err != nil {
			t.Errorf("probe Chrome process group %d: %v — cannot confirm the group is gone, so the profile-dir cleanup may race survivors", pgid, err)
			return
		}
		if time.Now().After(deadline) {
			t.Errorf("Chrome process group %d still has members 10s after SIGKILL — the profile-dir cleanup would race the survivors", pgid)
			return
		}
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		time.Sleep(50 * time.Millisecond)
	}
}
