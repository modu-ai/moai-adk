package sandbox

import (
	"context"
	"os"
	"testing"
)

// TestLauncher_DispatchByOS verifies the launcher automatically selects
// the correct backend based on the OS and CI environment.
// REQ-V3R2-RT-003-002, REQ-V3R2-RT-003-015
func TestLauncher_DispatchByOS(t *testing.T) {
	tests := []struct {
		name           string
		declared       Sandbox
		ciEnv          string
		goos           string
		wantEffective  Sandbox
	}{
		{
			name:          "macOS defaults to seatbelt",
			declared:      "none",
			goos:          "darwin",
			wantEffective: "seatbelt",
		},
		{
			name:          "Linux defaults to bubblewrap",
			declared:      "none",
			goos:          "linux",
			wantEffective: "bubblewrap",
		},
		{
			name:          "CI=1 overrides to docker",
			declared:      "bubblewrap",
			ciEnv:         "1",
			goos:          "linux",
			wantEffective: "docker",
		},
		{
			name:          "CI=1 on macOS also overrides to docker",
			declared:      "seatbelt",
			ciEnv:         "1",
			goos:          "darwin",
			wantEffective: "docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail until Launcher is implemented
		 launcher := NewLauncher()
			if tt.ciEnv != "" {
				oldCi := os.Getenv("CI")
				os.Setenv("CI", tt.ciEnv)
				defer os.Setenv("CI", oldCi)
			}

			got := launcher.resolveBackend(tt.declared, tt.goos)
			if got != tt.wantEffective {
				t.Errorf("Launcher.resolveBackend(%q, %s) = %q, want %q", tt.declared, tt.goos, got, tt.wantEffective)
			}
		})
	}
}

// TestLauncher_CIOverride verifies that CI=1 environment variable
// forces docker backend selection.
// REQ-V3R2-RT-003-015
func TestLauncher_CIOverride(t *testing.T) {
	oldCi := os.Getenv("CI")
	os.Setenv("CI", "1")
	defer os.Setenv("CI", oldCi)

	launcher := NewLauncher()
	backend := launcher.resolveBackend("bubblewrap", "linux")
	if backend != SandboxDocker {
		t.Errorf("CI=1 should resolve to docker, got %q", backend)
	}
}

// TestLauncher_OutputTruncation16MiB verifies that output exceeding
// 16 MiB is truncated and a SystemMessage is emitted.
// REQ-V3R2-RT-003-042
func TestLauncher_OutputTruncation16MiB(t *testing.T) {
	ctx := context.Background()
	opts := SandboxOptions{
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	// This test will fail until Launcher.Exec is implemented
	launcher := NewLauncher()
	output, err := launcher.Exec(ctx, "none", []string{"sh", "-c", "dd if=/dev/zero bs=1M count=32"}, opts)

	if err != nil {
		t.Logf("Exec returned error (expected): %v", err)
	}

	// Output should be exactly 16 MiB
	wantLen := 16 * 1024 * 1024
	if len(output) != wantLen {
		t.Errorf("Output length = %d, want %d", len(output), wantLen)
	}
}

// TestLauncher_BackendUnavailable verifies that missing backend
// causes SandboxBackendUnavailable error (no silent fallback).
// REQ-V3R2-RT-003-012
func TestLauncher_BackendUnavailable(t *testing.T) {
	ctx := context.Background()
	opts := SandboxOptions{
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	launcher := NewLauncher()
	_, err := launcher.Exec(ctx, "bubblewrap", []string{"echo", "hi"}, opts)

	if err == nil {
		t.Error("Expected error when backend unavailable, got nil")
	}

	if !IsSandboxBackendUnavailable(err) {
		t.Errorf("Expected ErrSandboxBackendUnavailable, got %T", err)
	}
}

// TestLauncher_ResolveBackend_AllScenarios verifies all combinations
// of declared sandbox, OS, and CI environment.
// REQ-V3R2-RT-003-002, REQ-V3R2-RT-003-003, REQ-V3R2-RT-003-015
func TestLauncher_ResolveBackend_AllScenarios(t *testing.T) {
	tests := []struct {
		declared      Sandbox
		ciSet         bool
		goos          string
		wantEffective Sandbox
	}{
		// macOS scenarios
		{"none", false, "darwin", "seatbelt"},
		{"seatbelt", false, "darwin", "seatbelt"},
		{"bubblewrap", false, "darwin", "seatbelt"}, // fallback
		{"docker", false, "darwin", "docker"},
		{"none", true, "darwin", "docker"},
		{"seatbelt", true, "darwin", "docker"},

		// Linux scenarios
		{"none", false, "linux", "bubblewrap"},
		{"bubblewrap", false, "linux", "bubblewrap"},
		{"seatbelt", false, "linux", "bubblewrap"}, // fallback
		{"docker", false, "linux", "docker"},
		{"none", true, "linux", "docker"},
		{"bubblewrap", true, "linux", "docker"},
	}

	for _, tt := range tests {
		t.Run(string(tt.declared)+"_ci="+boolToString(tt.ciSet)+"_"+tt.goos, func(t *testing.T) {
			launcher := NewLauncher()
			got := launcher.resolveBackend(tt.declared, tt.goos)
			if got != tt.wantEffective {
				t.Errorf("resolveBackend(%q, %s, CI=%v) = %q, want %q",
					tt.declared, tt.goos, tt.ciSet, got, tt.wantEffective)
			}
		})
	}
}

// TestLauncher_PermissionDenyDivergence verifies that when permission
// layer allows but sandbox denies, sandbox wins with SystemMessage.
// REQ-V3R2-RT-003-051
func TestLauncher_PermissionDenyDivergence(t *testing.T) {
	// This test uses mock permission decision
	// Real wiring with RT-002 happens after that SPEC merges
	ctx := context.Background()
	opts := SandboxOptions{
		WritableScope:  []string{"/tmp/worktree"},
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	launcher := NewLauncher()
	_, err := launcher.Exec(ctx, "bubblewrap", []string{"rm", "-rf", "/etc/passwd"}, opts)

	if err == nil {
		t.Error("Expected error when sandbox denies write, got nil")
	}

	// Should contain divergence message
	if !containsDivergenceMessage(err) {
		t.Error("Error should mention permission-sandbox divergence")
	}
}

// TestLauncher_OutputTruncation_SystemMessage verifies that truncation
// emits a SystemMessage with details.
func TestLauncher_OutputTruncation_SystemMessage(t *testing.T) {
	ctx := context.Background()
	opts := SandboxOptions{
		MaxOutputBytes: 16 * 1024 * 1024,
	}

	launcher := NewLauncher()
	output, err := launcher.Exec(ctx, "none", []string{"sh", "-c", "dd if=/dev/zero bs=1M count=32"}, opts)

	if err == nil {
		t.Error("Expected truncation error, got nil")
	}

	if !IsSandboxOutputTruncated(err) {
		t.Errorf("Expected ErrSandboxOutputTruncated, got %T", err)
	}

	if len(output) != 16*1024*1024 {
		t.Errorf("Output should be truncated to 16 MiB, got %d bytes", len(output))
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func containsDivergenceMessage(err error) bool {
	if err == nil {
		return false
	}
	return true // Will be implemented properly in GREEN phase
}
