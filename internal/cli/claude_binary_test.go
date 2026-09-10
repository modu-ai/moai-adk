package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// claude_binary_test.go covers the Claude Code binary pin (issue #1697):
// a configured pin is launched INSTEAD of the PATH lookup, an unset pin keeps
// the exec.LookPath("claude") fallback, and an invalid pin fails the launch
// with an actionable error instead of silently falling back.

// writeExecutable creates an executable file at path with a marker body and
// returns the path. Tests never execute the file directly (the launch seam is
// captured), so the body is inert; only existence and the mode bit matter.
func writeExecutable(t *testing.T, path string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
	return path
}

// fakePATHDir creates a directory holding a fake `claude` executable and
// returns the directory (to install as PATH) plus the binary path. Any
// resolution that reaches PATH under this fake resolves to this binary — a
// test asserting a different path therefore proves the pin won.
func fakePATHDir(t *testing.T) (dir, claudeBin string) {
	t.Helper()
	dir = t.TempDir()
	claudeBin = writeExecutable(t, filepath.Join(dir, "claude"))
	return dir, claudeBin
}

// fakeMoaiProject creates a minimal MoAI project root (a .moai directory, the
// findProjectRoot marker) under t.TempDir and chdirs into it. Launch-path
// writes (settings.local.json sync) land inside the temp project, not the
// repository.
func fakeMoaiProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("create .moai: %v", err)
	}
	t.Chdir(root)
	return root
}

func TestValidateClaudeBinaryPin_AcceptsExecutable(t *testing.T) {
	pin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-2.1.263"))
	got, err := validateClaudeBinaryPin(pin, "test-source")
	if err != nil {
		t.Fatalf("valid pin rejected: %v", err)
	}
	if got != pin {
		t.Errorf("resolved %q, want the pinned path %q", got, pin)
	}
}

func TestValidateClaudeBinaryPin_MissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-claude")
	_, err := validateClaudeBinaryPin(missing, "test-source")
	if err == nil {
		t.Fatal("missing pin accepted — pin validation must fail loud")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error should wrap the underlying stat error (fs.ErrNotExist), got: %v", err)
	}
	for _, want := range []string{missing, "test-source", "PATH lookup"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q (actionable: pin, source, fallback)", err, want)
		}
	}
}

func TestValidateClaudeBinaryPin_NotExecutable(t *testing.T) {
	plain := filepath.Join(t.TempDir(), "claude")
	if err := os.WriteFile(plain, []byte("not executable"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := validateClaudeBinaryPin(plain, "test-source")
	if err == nil {
		t.Fatal("non-executable pin accepted — pin validation must fail loud")
	}
	if !strings.Contains(err.Error(), "not executable") {
		t.Errorf("error %q should say the pin is not executable", err)
	}
	if !strings.Contains(err.Error(), "chmod +x") {
		t.Errorf("error %q should carry the chmod +x remedy", err)
	}
}

func TestValidateClaudeBinaryPin_Directory(t *testing.T) {
	dir := t.TempDir()
	_, err := validateClaudeBinaryPin(dir, "test-source")
	if err == nil {
		t.Fatal("directory pin accepted — pin validation must fail loud")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("error %q should say the pin is a directory", err)
	}
}

// TestResolveLaunchClaudeBinary_EnvPinWinsOverPATH is the pin-over-PATH proof
// at the resolution layer: with the env pin configured, the fake PATH copy of
// `claude` is NOT what resolves.
func TestResolveLaunchClaudeBinary_EnvPinWinsOverPATH(t *testing.T) {
	fakeMoaiProject(t)
	pathDir, pathClaude := fakePATHDir(t)
	t.Setenv("PATH", pathDir)
	pin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-pinned"))
	t.Setenv(config.EnvClaudeBin, pin)

	got, err := resolveLaunchClaudeBinary()
	if err != nil {
		t.Fatalf("resolve with env pin: %v", err)
	}
	if got != pin {
		t.Errorf("resolved %q, want the pinned %q (PATH copy %q must not win)", got, pin, pathClaude)
	}
}

// TestResolveLaunchClaudeBinary_ConfigPinWinsOverPATH proves the durable
// config key (llm.claude_bin in llm.yaml) is honored when the env var is
// unset.
func TestResolveLaunchClaudeBinary_ConfigPinWinsOverPATH(t *testing.T) {
	root := fakeMoaiProject(t)
	pathDir, pathClaude := fakePATHDir(t)
	t.Setenv("PATH", pathDir)
	t.Setenv(config.EnvClaudeBin, "")

	pin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-from-config"))
	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("create sections dir: %v", err)
	}
	llmYAML := "llm:\n  claude_bin: " + pin + "\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(llmYAML), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	got, err := resolveLaunchClaudeBinary()
	if err != nil {
		t.Fatalf("resolve with config pin: %v", err)
	}
	if got != pin {
		t.Errorf("resolved %q, want the config-pinned %q (PATH copy %q must not win)", got, pin, pathClaude)
	}
}

// TestResolveLaunchClaudeBinary_EnvOverridesConfig fixes the precedence
// between the two pin surfaces: env var wins (matches the project-wide
// env-overrides-file configuration priority).
func TestResolveLaunchClaudeBinary_EnvOverridesConfig(t *testing.T) {
	root := fakeMoaiProject(t)
	t.Setenv("PATH", t.TempDir()) // no claude here — resolution must not reach PATH
	envPin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-from-env"))
	t.Setenv(config.EnvClaudeBin, envPin)

	sectionsDir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sectionsDir, 0o755); err != nil {
		t.Fatalf("create sections dir: %v", err)
	}
	configPin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-from-config"))
	llmYAML := "llm:\n  claude_bin: " + configPin + "\n"
	if err := os.WriteFile(filepath.Join(sectionsDir, "llm.yaml"), []byte(llmYAML), 0o644); err != nil {
		t.Fatalf("write llm.yaml: %v", err)
	}

	got, err := resolveLaunchClaudeBinary()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != envPin {
		t.Errorf("resolved %q, want the env pin %q (env must override the config key %q)", got, envPin, configPin)
	}
}

// TestResolveLaunchClaudeBinary_UnsetFallsBackToPATH is the regression guard:
// with no pin on either surface, the PATH lookup behavior is unchanged.
func TestResolveLaunchClaudeBinary_UnsetFallsBackToPATH(t *testing.T) {
	fakeMoaiProject(t)
	pathDir, pathClaude := fakePATHDir(t)
	t.Setenv("PATH", pathDir)
	t.Setenv(config.EnvClaudeBin, "")

	got, err := resolveLaunchClaudeBinary()
	if err != nil {
		t.Fatalf("resolve without pin: %v", err)
	}
	if got != pathClaude {
		t.Errorf("resolved %q, want the PATH-resolved %q (fallback must be unchanged)", got, pathClaude)
	}
}

// TestResolveLaunchClaudeBinary_MissingPinFailsLoud proves an invalid pin is a
// launch error, NOT a silent PATH fallback: PATH here holds a working fake
// `claude`, and the resolution must still refuse.
func TestResolveLaunchClaudeBinary_MissingPinFailsLoud(t *testing.T) {
	fakeMoaiProject(t)
	pathDir, _ := fakePATHDir(t)
	t.Setenv("PATH", pathDir)
	missing := filepath.Join(t.TempDir(), "no-such-claude")
	t.Setenv(config.EnvClaudeBin, missing)

	got, err := resolveLaunchClaudeBinary()
	if err == nil {
		t.Fatalf("invalid pin resolved to %q — must fail loud, never fall back to PATH", got)
	}
	if !strings.Contains(err.Error(), "PATH lookup") {
		t.Errorf("error %q should tell the operator how to fall back", err)
	}
}

// TestLaunchClaudeDefault_LaunchesPinnedBinary is the end-to-end core proof:
// with a pin configured, launchClaudeDefault hands THE PINNED PATH to the
// exec seam — not a PATH lookup. PATH is set to a directory without any
// `claude`, so a fallback would fail the launch instead of launching the
// wrong binary.
func TestLaunchClaudeDefault_LaunchesPinnedBinary(t *testing.T) {
	fakeMoaiProject(t)
	t.Setenv("PATH", t.TempDir()) // no `claude` on PATH — fallback would fail
	pin := writeExecutable(t, filepath.Join(t.TempDir(), "claude-pinned"))
	t.Setenv(config.EnvClaudeBin, pin)

	origExec := execOrSpawnClaudeFunc
	defer func() { execOrSpawnClaudeFunc = origExec }()
	var capturedBin string
	var capturedArgs []string
	execOrSpawnClaudeFunc = func(bin string, args, env []string) error {
		capturedBin = bin
		capturedArgs = args
		return nil
	}

	if err := launchClaudeDefault("", nil); err != nil {
		t.Fatalf("launch with pin: %v", err)
	}
	if capturedBin != pin {
		t.Errorf("launched %q, want the pinned binary %q", capturedBin, pin)
	}
	if len(capturedArgs) == 0 || capturedArgs[0] != "claude" {
		t.Errorf("argv[0] = %v, want \"claude\" (the launcher's own argv convention)", capturedArgs)
	}
}

// TestLaunchClaudeDefault_PinErrorBlocksLaunch proves an invalid pin surfaces
// as the launch error and never reaches the exec seam.
func TestLaunchClaudeDefault_PinErrorBlocksLaunch(t *testing.T) {
	fakeMoaiProject(t)
	t.Setenv("PATH", t.TempDir())
	missing := filepath.Join(t.TempDir(), "no-such-claude")
	t.Setenv(config.EnvClaudeBin, missing)

	origExec := execOrSpawnClaudeFunc
	execCalled := false
	execOrSpawnClaudeFunc = func(bin string, args, env []string) error {
		execCalled = true
		return nil
	}
	defer func() { execOrSpawnClaudeFunc = origExec }()

	err := launchClaudeDefault("", nil)
	if err == nil {
		t.Fatal("launch with an invalid pin must fail")
	}
	if execCalled {
		t.Error("exec seam reached despite an invalid pin — the error must block the launch")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("error %q should name the invalid pin %q", err, missing)
	}
}

// TestLaunchClaudeDefault_UnsetPinFallsBackToPATH is the launch-path
// regression guard: with no pin anywhere, the binary handed to the exec seam
// is the PATH-resolved `claude`, byte-for-byte the pre-pin behavior.
func TestLaunchClaudeDefault_UnsetPinFallsBackToPATH(t *testing.T) {
	fakeMoaiProject(t)
	pathDir, pathClaude := fakePATHDir(t)
	t.Setenv("PATH", pathDir)
	t.Setenv(config.EnvClaudeBin, "")

	origExec := execOrSpawnClaudeFunc
	defer func() { execOrSpawnClaudeFunc = origExec }()
	var capturedBin string
	execOrSpawnClaudeFunc = func(bin string, args, env []string) error {
		capturedBin = bin
		return nil
	}

	if err := launchClaudeDefault("", nil); err != nil {
		t.Fatalf("launch without pin: %v", err)
	}
	if capturedBin != pathClaude {
		t.Errorf("launched %q, want the PATH-resolved %q (fallback behavior must be unchanged)", capturedBin, pathClaude)
	}
}
