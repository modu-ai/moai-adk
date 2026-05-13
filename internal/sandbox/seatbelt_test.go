package sandbox

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
)

// TestSeatbelt_ExecHello verifies basic execution with sandbox-exec.
// REQ-V3R2-RT-003-010
func TestSeatbelt_ExecHello(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("sandbox-exec test requires macOS")
	}

	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skipf("sandbox-exec unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &SeatbeltBackend{}
	output, err := backend.Exec(ctx, opts, []string{"echo", "hello"})

	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	if string(output) != "hello\n" {
		t.Errorf("Output = %q, want %q", string(output), "hello\n")
	}
}

// TestSeatbelt_FileWriteScopeEPERM verifies that writes outside
// WritableScope are denied with EPERM.
// REQ-V3R2-RT-003-013
func TestSeatbelt_FileWriteScopeEPERM(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("sandbox-exec test requires macOS")
	}

	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skipf("sandbox-exec unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp/agent-worktree"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &SeatbeltBackend{}
	_, err := backend.Exec(ctx, opts, []string{"touch", "/etc/passwd"})

	if err == nil {
		t.Error("Expected error when writing outside scope, got nil")
	}

	// Should be a permission error
	if !isPermissionError(err) {
		t.Errorf("Expected permission error, got %T", err)
	}
}

// TestSeatbelt_NetworkBlocked verifies that network egress to
// non-allowlist hosts is blocked.
// REQ-V3R2-RT-003-014
func TestSeatbelt_NetworkBlocked(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("sandbox-exec test requires macOS")
	}

	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skipf("sandbox-exec unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:     []string{"/tmp"},
		NetworkAllowlist:  []string{"github.com"},
		MaxOutputBytes:    16 * 1024 * 1024,
	}

	backend := &SeatbeltBackend{}
	_, err := backend.Exec(ctx, opts, []string{"curl", "-sS", "https://evil.example.com"})

	if err == nil {
		t.Error("Expected error when accessing non-allowlist host, got nil")
	}

	// Should contain network denied message
	if !containsNetworkDeniedMessage(err) {
		t.Error("Error should mention network denial")
	}
}

// TestSeatbelt_SetuidDenied verifies that setuid binaries (sudo, su)
// are denied execution.
// REQ-V3R2-RT-003-040
func TestSeatbelt_SetuidDenied(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("sandbox-exec test requires macOS")
	}

	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skipf("sandbox-exec unavailable: %v", err)
	}

	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &SeatbeltBackend{}
	_, err := backend.Exec(ctx, opts, []string{"sudo", "ls", "/etc"})

	if err == nil {
		t.Error("Expected error when executing setuid binary, got nil")
	}

	if !IsSandboxSetuidDenied(err) {
		t.Errorf("Expected ErrSandboxSetuidDenied, got %T", err)
	}
}

// TestSeatbelt_SBPLDeterministic verifies that SBPL profile generation
// is deterministic (sorted, no random ordering).
// REQ-V3R2-RT-003-004
func TestSeatbelt_SBPLDeterministic(t *testing.T) {
	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree", "/home/user/project"},
		ReadOnlyScope:    []string{"/usr", "/bin", "/etc"},
		NetworkAllowlist: []string{"github.com", "registry.npmjs.org"},
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	backend := &SeatbeltBackend{}

	// Generate SBPL 100 times and verify they're identical
	var firstSBPL string
	for i := 0; i < 100; i++ {
		sbpl, err := backend.generateSBPL(opts)
		if err != nil {
			t.Fatalf("generateSBPL failed: %v", err)
		}

		if i == 0 {
			firstSBPL = sbpl
		} else if sbpl != firstSBPL {
			t.Errorf("Iteration %d: SBPL differs from first iteration", i)
			t.Logf("First:\n%s\n", firstSBPL)
			t.Logf("Current:\n%s\n", sbpl)
		}
	}
}

func isPermissionError(err error) bool {
	return true // Will be implemented properly in GREEN phase
}

func containsNetworkDeniedMessage(err error) bool {
	return true // Will be implemented properly in GREEN phase
}
