package atomicfile

import (
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// TestPidAliveFalseForFinishedChild pins review-gate finding (P2): a child
// this process started, saw exit, and reaped is DEAD, and the liveness
// probe must answer dead for its pid. os.Process.Signal on a finished
// process surfaces as os.ErrProcessDone (the handle knows the process
// exited), not as ESRCH — a probe that treats only ESRCH as death reads
// every such pid as alive, so a same-boot orphan lock whose owner exited is
// never reclaimed.
func TestPidAliveFalseForFinishedChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows Signal is unsupported and always reads alive by design")
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "exit")
	} else {
		cmd = exec.Command("/bin/sleep", "0")
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	pid := cmd.Process.Pid
	// Reap so the handle records the exit — the state a crashed lock owner's
	// pid is observed in by the next writer.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("child did not exit within 10s")
	}

	if PidAlive(pid) {
		t.Fatalf("PidAlive(%d) = true for a finished, reaped child — os.ErrProcessDone is being read as alive", pid)
	}
}
