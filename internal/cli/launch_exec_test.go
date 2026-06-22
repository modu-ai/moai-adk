package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestExecOrSpawnClaude_PosixBuildTagGate verifies REQ-CGH-001 / AC-CGH-001
// Scenario 1b: the raw syscall.Exec call is gated behind a POSIX-only build tag.
// launcher.go (the shared launch path) must NOT contain a raw syscall.Exec call;
// the POSIX exec lives only in launch_exec_posix.go (//go:build !windows) and the
// Windows spawn-and-exit lives in launch_exec_windows.go (//go:build windows).
//
// This is a source-structure regression guard: it asserts the launcher.go shared
// path delegates to execOrSpawnClaude rather than calling syscall.Exec inline, so
// that the Windows cross-compile path never hits the unguarded POSIX-only call.
func TestExecOrSpawnClaude_PosixBuildTagGate(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	cliDir := filepath.Dir(thisFile)

	// 1. launcher.go MUST NOT call syscall.Exec inline (only comments may mention it).
	launcherSrc, err := os.ReadFile(filepath.Join(cliDir, "launcher.go"))
	if err != nil {
		t.Fatalf("read launcher.go: %v", err)
	}
	if strings.Contains(string(launcherSrc), "syscall.Exec(") {
		t.Errorf("launcher.go must not call syscall.Exec inline; the POSIX-only call "+
			"belongs in launch_exec_posix.go behind //go:build !windows (REQ-CGH-001)")
	}
	if !strings.Contains(string(launcherSrc), "execOrSpawnClaude(") {
		t.Errorf("launcher.go must delegate the launch to execOrSpawnClaude (REQ-CGH-001)")
	}

	// 2. The POSIX variant exists with the !windows build tag and the syscall.Exec call.
	posixSrc, err := os.ReadFile(filepath.Join(cliDir, "launch_exec_posix.go"))
	if err != nil {
		t.Fatalf("read launch_exec_posix.go: %v", err)
	}
	if !strings.Contains(string(posixSrc), "//go:build !windows") {
		t.Errorf("launch_exec_posix.go must carry the //go:build !windows tag (REQ-CGH-001)")
	}
	if !strings.Contains(string(posixSrc), "syscall.Exec(") {
		t.Errorf("launch_exec_posix.go must contain the POSIX syscall.Exec call (REQ-CGH-001)")
	}

	// 3. The Windows variant exists with the windows build tag and spawns a child.
	winSrc, err := os.ReadFile(filepath.Join(cliDir, "launch_exec_windows.go"))
	if err != nil {
		t.Fatalf("read launch_exec_windows.go: %v", err)
	}
	if !strings.Contains(string(winSrc), "//go:build windows") {
		t.Errorf("launch_exec_windows.go must carry the //go:build windows tag (REQ-CGH-001)")
	}
	if strings.Contains(string(winSrc), "syscall.Exec(") {
		t.Errorf("launch_exec_windows.go must NOT call syscall.Exec (POSIX-only); it spawns "+
			"a child and exits instead (REQ-CGH-001)")
	}
	if !strings.Contains(string(winSrc), "exec.Command(") {
		t.Errorf("launch_exec_windows.go must spawn a child via exec.Command (REQ-CGH-001)")
	}
}
