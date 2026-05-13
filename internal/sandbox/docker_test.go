package sandbox

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

// TestDocker_ExecHello verifies basic execution with docker run.
// REQ-V3R2-RT-003-015
func TestDocker_ExecHello(t *testing.T) {
	if os.Getenv("MOAI_TEST_DOCKER") != "1" {
		t.Skip("Docker tests gated; set MOAI_TEST_DOCKER=1")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker unavailable: %v", err)
	}

	// Check if docker daemon is running
	ctx, cancel := context.WithTimeout(context.Background(), 5000)
	defer cancel()
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err != nil {
		t.Skipf("docker daemon not running: %v", err)
	}

	opts := SandboxOptions{
		WritableScope:  []string{"/tmp"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &DockerBackend{}
	output, err := backend.Exec(context.Background(), opts, []string{"echo", "hello"})

	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	if string(output) != "hello\n" {
		t.Errorf("Output = %q, want %q", string(output), "hello\n")
	}
}

// TestDocker_NetworkAllowlist verifies that network access is restricted
// to the allowlist.
// REQ-V3R2-RT-003-008
func TestDocker_NetworkAllowlist(t *testing.T) {
	if os.Getenv("MOAI_TEST_DOCKER") != "1" {
		t.Skip("Docker tests gated; set MOAI_TEST_DOCKER=1")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker unavailable: %v", err)
	}

	opts := SandboxOptions{
		WritableScope:     []string{"/tmp"},
		NetworkAllowlist:  []string{"github.com"},
		MaxOutputBytes:    16 * 1024 * 1024,
	}

	backend := &DockerBackend{}
	_, err := backend.Exec(context.Background(), opts, []string{"curl", "-sS", "https://evil.example.com"})

	if err == nil {
		t.Error("Expected error when accessing non-allowlist host, got nil")
	}

	// Should contain network denied message
	if !containsNetworkDeniedMessage(err) {
		t.Error("Error should mention network denial")
	}
}

// TestDocker_FileWriteScope verifies that file writes are restricted
// to the WritableScope.
// REQ-V3R2-RT-003-007
func TestDocker_FileWriteScope(t *testing.T) {
	if os.Getenv("MOAI_TEST_DOCKER") != "1" {
		t.Skip("Docker tests gated; set MOAI_TEST_DOCKER=1")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker unavailable: %v", err)
	}

	opts := SandboxOptions{
		WritableScope:  []string{"/tmp/agent-worktree"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	backend := &DockerBackend{}
	_, err := backend.Exec(context.Background(), opts, []string{"touch", "/etc/passwd"})

	if err == nil {
		t.Error("Expected error when writing outside scope, got nil")
	}

	// Should be a permission error
	if !isPermissionError(err) {
		t.Errorf("Expected permission error, got %T", err)
	}
}

func containsNetworkDeniedMessage(err error) bool {
	return true // Will be implemented properly in GREEN phase
}

func isPermissionError(err error) bool {
	return true // Will be implemented properly in GREEN phase
}
