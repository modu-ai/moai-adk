package sandbox

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// TestBubblewrap_ExecHello is a Linux smoke test for bubblewrap execution.
// Skipped on non-Linux systems.
func TestBubblewrap_ExecHello(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not available")
	}

	b := NewBubblewrapBackend()
	if !b.Available() {
		t.Skip("bubblewrap backend reports unavailable")
	}

	opts := SandboxOptions{
		WritableScope:  []string{t.TempDir()},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	out, err := b.Exec(opts, []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("bubblewrap exec: %v", err)
	}
	if !strings.Contains(string(out), "hello") {
		t.Errorf("bubblewrap exec: expected 'hello' in output, got %q", string(out))
	}
}

// TestBubblewrap_FileWriteScopeEPERM verifies that bwrap denies writes outside scope.
// Skipped on non-Linux systems.
func TestBubblewrap_FileWriteScopeEPERM(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not available")
	}

	b := NewBubblewrapBackend()
	if !b.Available() {
		t.Skip("bubblewrap backend reports unavailable")
	}

	scope := t.TempDir()
	opts := SandboxOptions{
		WritableScope:  []string{scope},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// scope 밖에 쓰기 시도 → 실패해야 함
	_, err := b.Exec(opts, []string{"sh", "-c", "touch /etc/passwd"})
	if err == nil {
		t.Error("bubblewrap should have denied write to /etc/passwd outside scope")
	}
}

// TestBubblewrap_NetworkBlocked verifies that bwrap blocks non-allowlisted network.
// Skipped on non-Linux. Integration-level — checks process exit failure.
func TestBubblewrap_NetworkBlocked(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not available")
	}

	b := NewBubblewrapBackend()
	if !b.Available() {
		t.Skip("bubblewrap backend reports unavailable")
	}

	// 네트워크 차단 상태에서 curl 시도 → curl은 exit code != 0
	opts := SandboxOptions{
		WritableScope:    []string{t.TempDir()},
		NetworkAllowlist: []string{}, // 빈 allowlist = 모두 차단
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	_, err := b.Exec(opts, []string{"sh", "-c", "curl -sS --max-time 2 https://evil.example.com || exit 1"})
	if err == nil {
		t.Error("bubblewrap should have blocked network access to evil.example.com")
	}
}

// TestBubblewrap_SetuidDenied verifies bwrap denies setuid escalation.
// Skipped on non-Linux.
func TestBubblewrap_SetuidDenied(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not available")
	}

	b := NewBubblewrapBackend()
	if !b.Available() {
		t.Skip("bubblewrap backend reports unavailable")
	}

	opts := SandboxOptions{
		WritableScope:  []string{t.TempDir()},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// sudo 실행 시도 → 실패해야 함 (user namespaces 내에서 setuid 불가)
	_, err := b.Exec(opts, []string{"sudo", "id"})
	if err == nil {
		t.Error("bubblewrap should have denied sudo (setuid escalation)")
	}
}

// TestBubblewrap_ArgsDeterministic verifies that bwrap args generation is stable.
// Skipped on non-Linux.
func TestBubblewrap_ArgsDeterministic(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree", "/moai/state"},
		ReadOnlyScope:    []string{"/usr", "/lib"},
		NetworkAllowlist: []string{"github.com", "pypi.org"},
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	first, err := GenerateBwrapArgs(opts)
	if err != nil {
		t.Fatalf("GenerateBwrapArgs: %v", err)
	}

	for i := range 10 {
		got, err := GenerateBwrapArgs(opts)
		if err != nil {
			t.Fatalf("run %d: GenerateBwrapArgs: %v", i, err)
		}
		if strings.Join(got, "|") != strings.Join(first, "|") {
			t.Fatalf("run %d: GenerateBwrapArgs is non-deterministic", i)
		}
	}

	_ = errors.New // ensure errors imported
}
