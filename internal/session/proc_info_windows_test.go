//go:build windows

package session

// proc_info_windows_test.go — behavioral tests for the Toolhelp32 ancestry
// reader. This file compiles only on Windows: the repository's windows CI leg
// runs it, while local development on any other platform never compiles it
// (cross-build does not compile tests; `GOOS=windows go vet` is the local
// compile proof). The platform-neutral walk logic above this seam is covered
// by the injected procInfoFunc tests in session_pid_test.go.

import (
	"os"
	"os/exec"
	"testing"
)

// TestPlatformProcInfoSelf asserts the snapshot finds the calling process and
// reports a parent. Windows keeps the parent PID in the entry even when the
// parent has already exited, so ppid > 0 holds regardless of who spawned the
// test binary.
func TestPlatformProcInfoSelf(t *testing.T) {
	ppid, comm, ok := platformProcInfo(os.Getpid())
	if !ok {
		t.Fatal("platformProcInfo(self) reported unsupported")
	}
	if comm == "" {
		t.Error("platformProcInfo(self) returned an empty comm")
	}
	if ppid <= 0 {
		t.Errorf("platformProcInfo(self) ppid = %d, want > 0", ppid)
	}
}

// TestPlatformProcInfoDetectsDeadPid is the headline behavior: a PID absent
// from the snapshot reads as ok=false, which is how the ancestry walk detects
// a dead ancestor on Windows — isProcessAlive is conservative (always true)
// here, so snapshot absence is the only dead signal the walk has. The child is
// `cmd /c exit`; cmd is guaranteed present on every Windows this file
// compiles on. Between Wait() returning and the snapshot being taken the OS
// could in theory reuse the PID; the window is a few milliseconds on an idle
// runner, so that race is theoretical and accepted.
func TestPlatformProcInfoDetectsDeadPid(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	deadPID := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait child: %v", err)
	}

	if _, _, ok := platformProcInfo(deadPID); ok {
		t.Errorf("platformProcInfo(%d) reported an exited process as present", deadPID)
	}
}

// TestPlatformProcInfoInvalidPid covers the guard clause: non-positive PIDs
// are rejected without touching the process table.
func TestPlatformProcInfoInvalidPid(t *testing.T) {
	for _, pid := range []int{0, -1} {
		if _, _, ok := platformProcInfo(pid); ok {
			t.Errorf("platformProcInfo(%d) = ok, want rejected", pid)
		}
	}
}

// TestNormalizeWindowsComm pins the wrapper-matching contract: the shared
// wrapperProcessNames map is keyed on lowercase bare names, so the OS-reported
// executable name must lose its case and its ".exe" suffix before lookup.
func TestNormalizeWindowsComm(t *testing.T) {
	cases := map[string]string{
		"Moai.EXE":   "moai",
		"cmd.exe":    "cmd",
		"pwsh":       "pwsh",
		"PowerShell": "powershell",
	}
	for in, want := range cases {
		if got := normalizeWindowsComm(in); got != want {
			t.Errorf("normalizeWindowsComm(%q) = %q, want %q", in, got, want)
		}
	}
}
