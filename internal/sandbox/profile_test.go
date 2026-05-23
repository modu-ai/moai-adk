package sandbox

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

// TestProfile_GenerateSBPL verifies macOS SBPL profile generation.
// RED: fails until profile.go::generateSBPL is created.
func TestProfile_GenerateSBPL(t *testing.T) {
	t.Parallel()

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree"},
		ReadOnlyScope:    []string{"/usr", "/lib"},
		NetworkAllowlist: []string{"github.com", "pypi.org"},
	}

	profile, err := GenerateSBPL(opts)
	if err != nil {
		t.Fatalf("GenerateSBPL: unexpected error: %v", err)
	}
	if profile == "" {
		t.Fatal("GenerateSBPL: returned empty profile")
	}

	// SBPL profiles must start with (version 1)
	if !strings.HasPrefix(profile, "(version 1)") {
		t.Errorf("SBPL profile must start with (version 1), got: %q", profile[:min(50, len(profile))])
	}

	// (deny default) is mandatory.
	if !strings.Contains(profile, "(deny default)") {
		t.Error("SBPL profile must contain (deny default)")
	}

	// writable scope appears
	if !strings.Contains(profile, "/tmp/worktree") {
		t.Error("SBPL profile must contain writable scope path")
	}

	// SBPL does not support host-specific TCP blocking (sandbox-exec constraint).
	// Instead, a non-empty allowlist must include `(allow network-outbound (remote tcp))`.
	if !strings.Contains(profile, "(allow network-outbound (remote tcp))") {
		t.Error("SBPL profile with non-empty allowlist must contain (allow network-outbound (remote tcp))")
	}

	// LSP carve-out included (REQ-021).
	if !strings.Contains(profile, ".cache") {
		t.Error("SBPL profile must contain LSP carve-out for ~/.cache")
	}
}

// TestProfile_GenerateBwrapArgs verifies Linux bwrap argument generation.
// RED: fails until profile.go::generateBwrapArgs is created.
func TestProfile_GenerateBwrapArgs(t *testing.T) {
	t.Parallel()

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree"},
		ReadOnlyScope:    []string{"/usr", "/lib"},
		NetworkAllowlist: []string{"github.com"},
	}

	args, err := GenerateBwrapArgs(opts)
	if err != nil {
		t.Fatalf("GenerateBwrapArgs: unexpected error: %v", err)
	}
	if len(args) == 0 {
		t.Fatal("GenerateBwrapArgs: returned empty args")
	}

	joined := strings.Join(args, " ")

	// Required bwrap flags.
	if !strings.Contains(joined, "--unshare-all") {
		t.Error("bwrap args must contain --unshare-all")
	}
	if !strings.Contains(joined, "--die-with-parent") {
		t.Error("bwrap args must contain --die-with-parent")
	}

	// writable scope mounts via --bind.
	if !strings.Contains(joined, "/tmp/worktree") {
		t.Error("bwrap args must contain writable scope path")
	}
}

// TestProfile_GenerateDockerSnippet verifies Docker snippet generation.
// RED: fails until profile.go::generateDockerSnippet is created.
func TestProfile_GenerateDockerSnippet(t *testing.T) {
	t.Parallel()

	opts := SandboxOptions{
		WritableScope:  []string{"/workspace"},
		DockerImage:    "alpine:latest",
	}

	snippet, err := GenerateDockerSnippet(opts)
	if err != nil {
		t.Fatalf("GenerateDockerSnippet: unexpected error: %v", err)
	}
	if snippet == "" {
		t.Fatal("GenerateDockerSnippet: returned empty snippet")
	}

	// must include docker run.
	if !strings.Contains(snippet, "docker run") {
		t.Error("Docker snippet must contain 'docker run'")
	}

	// --rm is a required flag for ephemeral execution.
	if !strings.Contains(snippet, "--rm") {
		t.Error("Docker snippet must contain --rm flag")
	}
}

// TestProfile_DeterministicChecksum_100Runs verifies profile generation is deterministic.
// RED: fails until profile.go produces stable output across runs.
func TestProfile_DeterministicChecksum_100Runs(t *testing.T) {
	t.Parallel()

	opts := SandboxOptions{
		WritableScope:    []string{"/tmp/worktree", "/moai/state"},
		ReadOnlyScope:    []string{"/usr", "/lib", "/etc"},
		NetworkAllowlist: []string{"github.com", "pypi.org", "proxy.golang.org"},
		EnvPassthrough:   []string{"GH_TOKEN"},
		MaxOutputBytes:   16 * 1024 * 1024,
	}

	// SBPL determinism.
	first, err := GenerateSBPL(opts)
	if err != nil {
		t.Fatalf("GenerateSBPL: %v", err)
	}
	firstHash := checksum(first)

	for i := range 100 {
		got, err := GenerateSBPL(opts)
		if err != nil {
			t.Fatalf("run %d: GenerateSBPL: %v", i, err)
		}
		if checksum(got) != firstHash {
			t.Fatalf("run %d: GenerateSBPL is non-deterministic", i)
		}
	}

	// bwrap determinism.
	firstArgs, err := GenerateBwrapArgs(opts)
	if err != nil {
		t.Fatalf("GenerateBwrapArgs: %v", err)
	}
	firstArgsHash := checksum(strings.Join(firstArgs, "|"))

	for i := range 100 {
		got, err := GenerateBwrapArgs(opts)
		if err != nil {
			t.Fatalf("run %d: GenerateBwrapArgs: %v", i, err)
		}
		if checksum(strings.Join(got, "|")) != firstArgsHash {
			t.Fatalf("run %d: GenerateBwrapArgs is non-deterministic", i)
		}
	}
}

func checksum(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
