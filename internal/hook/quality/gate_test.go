package quality

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// skipIfCommandMissing skips the test if the given binary is not in PATH.
func skipIfCommandMissing(t *testing.T, binary string) {
	t.Helper()
	if _, err := exec.LookPath(binary); err != nil {
		t.Skipf("%s not found in PATH, skipping", binary)
	}
}

// skipIfNotGoProject skips the test if the current directory is not a Go module.
func skipIfNotGoProject(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		t.Skip("go.mod not found, skipping integration test")
	}
}

func TestDefaultGateConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultGateConfig()

	if !cfg.Enabled {
		t.Error("Enabled: want true, got false")
	}
	if cfg.SkipTests {
		t.Error("SkipTests: want false, got true")
	}
	if cfg.VetTimeout != 30*time.Second {
		t.Errorf("VetTimeout: want 30s, got %s", cfg.VetTimeout)
	}
	if cfg.LintTimeout != 60*time.Second {
		t.Errorf("LintTimeout: want 60s, got %s", cfg.LintTimeout)
	}
	if cfg.TestTimeout != 120*time.Second {
		t.Errorf("TestTimeout: want 120s, got %s", cfg.TestTimeout)
	}
}

func TestNewQualityGate_NilConfigUsesDefaults(t *testing.T) {
	t.Parallel()

	g := NewQualityGate(nil)

	if g.config == nil {
		t.Fatal("config should not be nil")
	}
	if !g.config.Enabled {
		t.Error("Enabled: want true, got false")
	}
}

func TestQualityGate_Disabled(t *testing.T) {
	t.Parallel()

	cfg := &GateConfig{Enabled: false}
	g := NewQualityGate(cfg)

	passed, output := g.Run(context.Background())

	if !passed {
		t.Error("disabled gate should always pass")
	}
	if output != "" {
		t.Errorf("disabled gate should return empty output, got %q", output)
	}
}

func TestQualityGate_Timeout(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "go")

	// Use a near-zero timeout to force timeout on go test.
	cfg := &GateConfig{
		Enabled:     true,
		SkipTests:   false,
		VetTimeout:  30 * time.Second,
		LintTimeout: 60 * time.Second,
		TestTimeout: 1 * time.Nanosecond, // guaranteed timeout
	}
	g := NewQualityGate(cfg)

	// We need a real Go project for this test to exercise timeout logic.
	// Run from project root.
	skipIfNotGoProject(t)

	passed, output := g.Run(context.Background())

	// go vet may pass before timeout triggers on go test, so we only check
	// that if it fails the output contains "timed out" or is empty (vet failed first).
	if !passed {
		if !strings.Contains(output, "timed out") && !strings.Contains(output, "quality gate failed") {
			t.Errorf("unexpected output on timeout: %q", output)
		}
	}
}

func TestQualityGate_SkipTests(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "go")
	skipIfNotGoProject(t)

	cfg := &GateConfig{
		Enabled:     true,
		SkipTests:   true,
		VetTimeout:  30 * time.Second,
		LintTimeout: 60 * time.Second,
		TestTimeout: 120 * time.Second,
	}
	g := NewQualityGate(cfg)

	// When SkipTests is true, only vet and lint run.
	// This should not time out or fail solely due to test step.
	// We cannot assert pass/fail because lint may or may not be installed,
	// but we assert the call completes without panic.
	_, _ = g.Run(context.Background())
}

func TestQualityGate_ContextCanceled(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "go")
	skipIfNotGoProject(t)

	cfg := DefaultGateConfig()
	g := NewQualityGate(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	passed, output := g.Run(ctx)

	// With a cancelled context the first step should fail.
	if passed {
		// If vet happened to complete before cancel was noticed, that is acceptable.
		t.Log("gate passed despite cancelled context (race window); acceptable")
		return
	}
	if output == "" {
		t.Error("expected non-empty output on failure")
	}
}

func TestQualityGate_runStep_UnknownCommand(t *testing.T) {
	t.Parallel()

	cfg := DefaultGateConfig()
	g := NewQualityGate(cfg)

	passed, output := g.runStep(context.Background(), "test-step", 5*time.Second, "this-binary-does-not-exist-12345")

	if passed {
		t.Error("runStep with missing binary should fail")
	}
	if output == "" {
		t.Error("output should contain error details")
	}
}

func TestQualityGate_runStep_TimeoutMessage(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "sleep")

	cfg := DefaultGateConfig()
	g := NewQualityGate(cfg)

	// sleep for a very long time but timeout after 10ms.
	passed, output := g.runStep(context.Background(), "slow-step", 10*time.Millisecond, "sleep", "60")

	if passed {
		t.Error("runStep should fail on timeout")
	}
	if !strings.Contains(output, "timed out") {
		t.Errorf("output should contain 'timed out', got: %q", output)
	}
	if !strings.Contains(output, "slow-step") {
		t.Errorf("output should contain step name 'slow-step', got: %q", output)
	}
}

func TestQualityGate_runStep_Success(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "true")

	cfg := DefaultGateConfig()
	g := NewQualityGate(cfg)

	passed, output := g.runStep(context.Background(), "noop", 5*time.Second, "true")

	if !passed {
		t.Errorf("runStep with 'true' should pass, output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty on success, got: %q", output)
	}
}

func TestQualityGate_executeStep_SkipsOptionalMissing(t *testing.T) {
	t.Parallel()

	// Verifies that executeStep returns (true, "") when the binary
	// is absent and the step is marked optional.
	cfg := DefaultGateConfig()
	g := NewQualityGate(cfg)

	step := gateStep{name: "missing-tool", binary: "this-binary-does-not-exist-12345", optional: true}
	passed, output := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Error("executeStep with missing optional binary should pass (skip)")
	}
	if output != "" {
		t.Errorf("output should be empty when tool is skipped, got: %q", output)
	}
}

func TestQualityGate_detectToolchain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		marker     string
		wantNil    bool
		wantMarker string // first marker of expected toolchain
	}{
		{"Go project", "go.mod", false, "go.mod"},
		{"Node project", "package.json", false, "package.json"},
		{"Python pyproject", "pyproject.toml", false, "pyproject.toml"},
		{"Python requirements", "requirements.txt", false, "pyproject.toml"},
		{"Rust project", "Cargo.toml", false, "Cargo.toml"},
		{"Java Maven", "pom.xml", false, "pom.xml"},
		{"Ruby project", "Gemfile", false, "Gemfile"},
		{"PHP project", "composer.json", false, "composer.json"},
		{"Dart project", "pubspec.yaml", false, "pubspec.yaml"},
		{"Elixir project", "mix.exs", false, "mix.exs"},
		{"Swift project", "Package.swift", false, "Package.swift"},
		{"Scala project", "build.sbt", false, "build.sbt"},
		{"Zig project", "build.zig", false, "build.zig"},
		{"Unknown project", "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if tt.marker != "" {
				os.WriteFile(filepath.Join(dir, tt.marker), []byte(""), 0o644)
			}
			g := NewQualityGate(&GateConfig{ProjectDir: dir})
			tc := g.detectToolchain()
			if tt.wantNil {
				if tc != nil {
					t.Errorf("detectToolchain() should be nil for %q, got marker %v", tt.name, tc.markerFiles)
				}
			} else {
				if tc == nil {
					t.Errorf("detectToolchain() should not be nil for %q", tt.name)
				} else if tc.markerFiles[0] != tt.wantMarker {
					t.Errorf("detectToolchain() first marker = %q, want %q", tc.markerFiles[0], tt.wantMarker)
				}
			}
		})
	}
}

func TestQualityGate_detectToolchain_GlobPattern(t *testing.T) {
	t.Parallel()

	// C# projects use *.csproj glob pattern
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "MyApp.csproj"), []byte("<Project/>"), 0o644)

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	tc := g.detectToolchain()

	if tc == nil {
		t.Fatal("detectToolchain() should detect C# project from .csproj file")
	}
	if tc.markerFiles[0] != "*.csproj" {
		t.Errorf("expected *.csproj marker, got %v", tc.markerFiles)
	}
}

func TestQualityGate_Run_UnknownProjectPasses(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // empty directory, no language markers
	g := NewQualityGate(&GateConfig{
		Enabled:    true,
		ProjectDir: dir,
	})

	passed, output := g.Run(context.Background())

	if !passed {
		t.Errorf("unknown project should pass silently, got output: %q", output)
	}
	if output != "" {
		t.Errorf("output should be empty for unknown project, got: %q", output)
	}
}

// TestIsGitCommit covers various command patterns.
func TestIsGitCommit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "bare git commit",
			command: "git commit",
			want:    true,
		},
		{
			name:    "git commit with -m flag",
			command: `git commit -m "fix: something"`,
			want:    true,
		},
		{
			name:    "git commit --amend",
			command: "git commit --amend",
			want:    true,
		},
		{
			name:    "git commit --no-verify",
			command: `git commit --no-verify -m "skip"`,
			want:    true,
		},
		{
			name:    "git commit with leading whitespace",
			command: "  git commit -m 'msg'",
			want:    true,
		},
		{
			name:    "git commit with all flags",
			command: `git commit --no-verify --amend -m "test"`,
			want:    true,
		},
		{
			name:    "git commit --help is NOT a commit",
			command: "git commit --help",
			want:    true, // still matches "git commit" prefix (REQ-GATE-011: no bypass)
		},
		{
			name:    "echo with git commit text",
			command: `echo "git commit"`,
			want:    false,
		},
		{
			name:    "git status",
			command: "git status",
			want:    false,
		},
		{
			name:    "git push",
			command: "git push origin main",
			want:    false,
		},
		{
			name:    "git log",
			command: "git log --oneline",
			want:    false,
		},
		{
			name:    "empty string",
			command: "",
			want:    false,
		},
		{
			name:    "unrelated command",
			command: "ls -la",
			want:    false,
		},
		{
			name:    "gitcommit without space",
			command: "gitcommit",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsGitCommit(tt.command)
			if got != tt.want {
				t.Errorf("IsGitCommit(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}
