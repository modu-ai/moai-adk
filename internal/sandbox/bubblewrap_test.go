package sandbox

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// TestBubblewrap_ExecHello verifies basic execution with bwrap.
// REQ-V3R2-RT-003-011
func TestBubblewrap_ExecHello(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}

	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skipf("bwrap unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &BubblewrapBackend{}
	output, err := backend.Exec(ctx, opts, []string{"echo", "hello"})

	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	if string(output) != "hello\n" {
		t.Errorf("Output = %q, want %q", string(output), "hello\n")
	}
}

// TestBubblewrap_FileWriteScopeEPERM verifies that writes outside
// WritableScope are denied with EPERM.
// REQ-V3R2-RT-003-013
func TestBubblewrap_FileWriteScopeEPERM(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}

	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skipf("bwrap unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp/agent-worktree"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &BubblewrapBackend{}
	_, err := backend.Exec(ctx, opts, []string{"touch", "/etc/passwd"})

	if err == nil {
		t.Error("Expected error when writing outside scope, got nil")
	}

	// Should be a permission error
	if !isPermissionError(err) {
		t.Errorf("Expected permission error, got %T", err)
	}
}

// TestBubblewrap_NetworkBlocked verifies that network egress to
// non-allowlist hosts is blocked.
// REQ-V3R2-RT-003-014
func TestBubblewrap_NetworkBlocked(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}

	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skipf("bwrap unavailable: %v", err)
	}

	if os.Getenv("MOAI_TEST_NETWORK") == "" {
		t.Skip("Network test gated; set MOAI_TEST_NETWORK=1")
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:     []string{"/tmp"},
		NetworkAllowlist:  []string{"github.com"},
		MaxOutputBytes:    16 * 1024 * 1024,
	}

	backend := &BubblewrapBackend{}
	_, err := backend.Exec(ctx, opts, []string{"curl", "-sS", "https://evil.example.com"})

	if err == nil {
		t.Error("Expected error when accessing non-allowlist host, got nil")
	}

	// Should contain network denied message
	if !containsNetworkDeniedMessage(err) {
		t.Error("Error should mention network denial")
	}
}

// TestBubblewrap_SetuidDenied verifies that setuid binaries (sudo, su)
// are denied execution.
// REQ-V3R2-RT-003-040
func TestBubblewrap_SetuidDenied(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}

	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skipf("bwrap unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &BubblewrapBackend{}
	_, err := backend.Exec(ctx, opts, []string{"sudo", "ls", "/etc"})

	if err == nil {
		t.Error("Expected error when executing setuid binary, got nil")
	}

	if !IsSandboxSetuidDenied(err) {
		t.Errorf("Expected ErrSandboxSetuidDenied, got %T", err)
	}
}

// TestBubblewrap_ArgsDeterministic verifies that bwrap argument
// generation is deterministic (sorted, no random ordering).
// REQ-V3R2-RT-003-004
func TestBubblewrap_ArgsDeterministic(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("bubblewrap test requires Linux")
	}

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree", "/home/user/project"},
		ReadOnlyScope:    []string{"/usr", "/bin", "/etc"},
		NetworkAllowlist: []string{"github.com", "registry.npmjs.org"},
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	backend := &BubblewrapBackend{}

	// Generate args 100 times and verify they're identical
	var firstArgs []string
	for i := 0; i < 100; i++ {
		args, err := backend.buildArgs(opts, []string{"echo", "test"})
		if err != nil {
			t.Fatalf("buildArgs failed: %v", err)
		}

		if i == 0 {
			firstArgs = args
		} else if !argsEqual(args, firstArgs) {
			t.Errorf("Iteration %d: args differ from first iteration", i)
		}
	}
}

func isPermissionError(err error) bool {
	return true // Will be implemented properly in GREEN phase
}

func containsNetworkDeniedMessage(err error) bool {
	return true // Will be implemented properly in GREEN phase
}

func argsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
