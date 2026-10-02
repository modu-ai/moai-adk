//go:build windows

package kanban

import (
	"os/exec"
	"testing"
)

// deadPIDWin moved verbatim from the Windows clear suite deleted at M6
// (SPEC-LAUNCHER-ENTRY-FLAGS-001); integration_lock_mutation_windows_test.go
// still calls it.
//
// deadPIDWin returns the PID of a process that has positively terminated, by
// spawning and reaping a Windows child (cmd.exe exits immediately). This
// file compiles only under GOOS=windows, so the helper is a Windows
// implementation, not a skip: GOOS=windows go vet verifies it compiles here.
func deadPIDWin(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("cmd", "/c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("spawn sacrificial process: %v", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait sacrificial process: %v", err)
	}
	if !defaultProcessAlive(-1) {
		_ = pid // sanity shape only; the real probe check follows
	}
	if processAlive(pid) {
		t.Fatalf("sacrificial pid %d still observed live; cannot construct a dead owner", pid)
	}
	return pid
}
