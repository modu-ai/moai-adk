package sandbox

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// TestSeatbelt_ExecHello is a macOS smoke test for seatbelt (sandbox-exec) execution.
// Skipped on non-macOS systems.
func TestSeatbelt_ExecHello(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("seatbelt test requires macOS")
	}
	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skip("sandbox-exec not available")
	}

	s := NewSeatbeltBackend()
	if !s.Available() {
		t.Skip("seatbelt backend reports unavailable")
	}

	opts := SandboxOptions{
		WritableScope:  []string{t.TempDir()},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	out, err := s.Exec(opts, []string{"echo", "hello"})
	if err != nil {
		t.Fatalf("seatbelt exec: %v", err)
	}
	if !strings.Contains(string(out), "hello") {
		t.Errorf("seatbelt exec: expected 'hello' in output, got %q", string(out))
	}
}

// TestSeatbelt_FileWriteScopeEPERM verifies that sandbox-exec denies writes outside scope.
// Skipped on non-macOS systems.
func TestSeatbelt_FileWriteScopeEPERM(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("seatbelt test requires macOS")
	}
	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skip("sandbox-exec not available")
	}

	s := NewSeatbeltBackend()
	if !s.Available() {
		t.Skip("seatbelt backend reports unavailable")
	}

	scope := t.TempDir()
	opts := SandboxOptions{
		WritableScope:  []string{scope},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// Attempt to write /etc/passwd outside scope -> must fail with EPERM.
	_, err := s.Exec(opts, []string{"sh", "-c", "touch /etc/passwd"})
	if err == nil {
		t.Error("seatbelt should have denied write to /etc/passwd outside writable scope")
	}
}

// TestSeatbelt_NetworkBlocked verifies that seatbelt blocks non-allowlisted network.
// Skipped on non-macOS.
func TestSeatbelt_NetworkBlocked(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("seatbelt test requires macOS")
	}
	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skip("sandbox-exec not available")
	}

	s := NewSeatbeltBackend()
	if !s.Available() {
		t.Skip("seatbelt backend reports unavailable")
	}

	opts := SandboxOptions{
		WritableScope:    []string{t.TempDir()},
		NetworkAllowlist: []string{}, // empty = deny all egress
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	// curl attempt → should fail due to network denial
	_, err := s.Exec(opts, []string{"sh", "-c", "curl -sS --max-time 2 https://evil.example.com || exit 1"})
	if err == nil {
		t.Error("seatbelt should have blocked network access to evil.example.com")
	}
}

// TestSeatbelt_SetuidDenied verifies sandbox-exec denies setuid escalation.
// Skipped on non-macOS.
func TestSeatbelt_SetuidDenied(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("seatbelt test requires macOS")
	}
	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		t.Skip("sandbox-exec not available")
	}

	s := NewSeatbeltBackend()
	if !s.Available() {
		t.Skip("seatbelt backend reports unavailable")
	}

	opts := SandboxOptions{
		WritableScope:  []string{t.TempDir()},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// Attempt sudo -> must fail.
	_, err := s.Exec(opts, []string{"sudo", "id"})
	if err == nil {
		t.Error("seatbelt should have denied sudo (setuid escalation)")
	}
}

// TestSeatbelt_SBPLDeterministic verifies that SBPL profile generation is stable.
// Skipped on non-macOS.
func TestSeatbelt_SBPLDeterministic(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("seatbelt test requires macOS")
	}

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree", "/moai/state"},
		ReadOnlyScope:    []string{"/usr", "/lib"},
		NetworkAllowlist: []string{"github.com", "pypi.org"},
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	first, err := GenerateSBPL(opts)
	if err != nil {
		t.Fatalf("GenerateSBPL: %v", err)
	}

	for i := range 10 {
		got, err := GenerateSBPL(opts)
		if err != nil {
			t.Fatalf("run %d: GenerateSBPL: %v", i, err)
		}
		if got != first {
			t.Fatalf("run %d: GenerateSBPL is non-deterministic", i)
		}
	}
}
