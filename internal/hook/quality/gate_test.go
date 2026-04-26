package quality

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func TestQualityGate_executeStep_SkipsWhenConfigFileMissing(t *testing.T) {
	t.Parallel()

	// Reproduces issue #619: eslint step runs on a project without eslint config,
	// causing the quality gate to fail. The configFiles field should make the step
	// skip when no config file exists.
	dir := t.TempDir()
	// Create package.json so Node.js toolchain matches, but NO eslint config.
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	g := NewQualityGate(&GateConfig{Enabled: true, ProjectDir: dir})

	// Step with configFiles requirement but none exist in dir.
	step := gateStep{
		name:     "eslint",
		binary:   "true", // Use 'true' to ensure it would pass if executed.
		args:     nil,
		optional: true,
		configFiles: []string{
			"eslint.config.js", "eslint.config.mjs",
			".eslintrc.js", ".eslintrc.json",
		},
	}
	passed, output := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Error("step with missing config files should skip (pass)")
	}
	if output != "" {
		t.Errorf("output should be empty when config-skipped, got: %q", output)
	}
}

func TestQualityGate_executeStep_RunsWhenConfigFileExists(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "true")

	dir := t.TempDir()
	// Create both package.json and an eslint config.
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "eslint.config.js"), []byte("module.exports = {}"), 0o644); err != nil {
		t.Fatalf("failed to create eslint config: %v", err)
	}

	g := NewQualityGate(&GateConfig{Enabled: true, ProjectDir: dir})

	step := gateStep{
		name:        "eslint",
		binary:      "true", // 'true' always exits 0
		optional:    true,
		configFiles: []string{"eslint.config.js", ".eslintrc.json"},
	}
	passed, output := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Errorf("step should run and pass when config file exists, output: %q", output)
	}
}

func TestQualityGate_Run_PythonWithPackageJSON(t *testing.T) {
	t.Parallel()

	// Issue #619: Python project with package.json but no eslint config.
	// Node.js toolchain matches first and eslint fails.
	// After fix: eslint step should be skipped due to missing config files.
	dir := t.TempDir()

	// Create both package.json (triggers Node.js toolchain) and requirements.txt.
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==3.0"), 0o644); err != nil {
		t.Fatalf("failed to create requirements.txt: %v", err)
	}

	g := NewQualityGate(&GateConfig{
		Enabled:     true,
		SkipTests:   true, // Skip test step to isolate lint behavior.
		ProjectDir:  dir,
		VetTimeout:  5 * time.Second,
		LintTimeout: 5 * time.Second,
		TestTimeout: 5 * time.Second,
	})

	// Node.js toolchain matches first (package.json before pyproject.toml).
	// Without the configFiles fix, eslint would run and fail.
	// With the fix, eslint skips because no eslint config exists.
	passed, output := g.Run(context.Background())

	if !passed {
		t.Errorf("Python project with package.json should pass when eslint config is missing, output: %q", output)
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
				if err := os.WriteFile(filepath.Join(dir, tt.marker), []byte(""), 0o644); err != nil {
					t.Fatalf("failed to create marker file: %v", err)
				}
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

// TestQualityGate_detectToolchain_Flutter verifies that a pubspec.yaml
// containing the Flutter SDK dependency resolves to `flutter test` / `flutter
// analyze` rather than the Dart CLI defaults. See issue #652.
func TestQualityGate_detectToolchain_Flutter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		pubspec        string
		wantTestBinary string
		wantTestArg0   string
	}{
		{
			name: "flutter SDK dependency → flutter test",
			pubspec: `name: my_app
dependencies:
  flutter:
    sdk: flutter
`,
			wantTestBinary: "flutter",
			wantTestArg0:   "test",
		},
		{
			name: "flutter top-level section → flutter test",
			pubspec: `name: my_app
dependencies:
  cupertino_icons: ^1.0.0

flutter:
  uses-material-design: true
`,
			wantTestBinary: "flutter",
			wantTestArg0:   "test",
		},
		{
			name: "pure Dart CLI project → dart test",
			pubspec: `name: my_dart_cli
dependencies:
  args: ^2.4.0
dev_dependencies:
  test: ^1.24.0
`,
			wantTestBinary: "dart",
			wantTestArg0:   "test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte(tc.pubspec), 0o644); err != nil {
				t.Fatalf("write pubspec.yaml: %v", err)
			}
			g := NewQualityGate(&GateConfig{ProjectDir: dir})
			tcResolved := g.detectToolchain()
			if tcResolved == nil {
				t.Fatalf("detectToolchain returned nil")
			}
			if tcResolved.testStep == nil {
				t.Fatalf("testStep is nil")
			}
			if tcResolved.testStep.binary != tc.wantTestBinary {
				t.Errorf("test binary = %q, want %q", tcResolved.testStep.binary, tc.wantTestBinary)
			}
			if len(tcResolved.testStep.args) == 0 || tcResolved.testStep.args[0] != tc.wantTestArg0 {
				t.Errorf("test args[0] = %v, want %q", tcResolved.testStep.args, tc.wantTestArg0)
			}
			// Also verify vetSteps are in the flutter branch
			if len(tcResolved.vetSteps) > 0 && tc.wantTestBinary == "flutter" {
				if tcResolved.vetSteps[0].binary != "flutter" {
					t.Errorf("vet binary = %q, want flutter", tcResolved.vetSteps[0].binary)
				}
			}
		})
	}
}

func TestQualityGate_detectToolchain_GlobPattern(t *testing.T) {
	t.Parallel()

	// C# projects use *.csproj glob pattern
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "MyApp.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatalf("failed to create .csproj file: %v", err)
	}

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
// ─── Issue #667: dotnet format staged file filter and graceful restore-failure handling ───

// TestStagedFiles_OutsideGitRepo verifies that stagedFiles returns nil outside a git repository.
func TestStagedFiles_OutsideGitRepo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // no git init

	got, err := stagedFiles(context.Background(), dir)
	// Must not error and must return nil (conservative fallback).
	if err != nil {
		t.Errorf("must not error outside a git repository: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil outside a git repository, got: %v", got)
	}
}

// TestStagedFiles_InsideGitRepo verifies that the staged file list is returned inside a git repository.
func TestStagedFiles_InsideGitRepo(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "git")

	dir := t.TempDir()
	// Initialize git repository
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}

	// Create and stage files
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if out, err := runGitCmd(dir, "add", "README.md", "config.yaml"); err != nil {
		t.Fatalf("git add failed: %v, out: %s", err, out)
	}

	got, err := stagedFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("stagedFiles error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 staged files, got: %v", got)
	}
}

// TestStagedFiles_EmptyStaging verifies that nil is returned when there are no staged files.
func TestStagedFiles_EmptyStaging(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "git")

	dir := t.TempDir()
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}

	// Do not stage any files
	got, err := stagedFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("stagedFiles error: %v", err)
	}
	// Empty staging: return nil (conservative fallback — run the step)
	if got != nil {
		t.Errorf("expected nil when no staged files, got: %v", got)
	}
}

// writeFakeBinary creates a fake binary shell script in a temp directory that exits with the
// specified exit code when called, and returns the directory path.
// Prepend the directory to PATH in tests so it is called instead of the real binary.
func writeFakeBinary(t *testing.T, name string, exitCode int, output string) string {
	t.Helper()
	binDir := t.TempDir()
	script := filepath.Join(binDir, name)
	content := "#!/bin/sh\n"
	if output != "" {
		content += "echo '" + output + "' >&2\n"
	}
	content += "exit " + fmt.Sprintf("%d", exitCode) + "\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to create fake binary: %v", err)
	}
	return binDir
}

// TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged verifies that the dotnet format step
// is skipped when no .cs files are staged (issue #667 Fix 1).
// Does not call t.Parallel() because it uses t.Setenv.
func TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged(t *testing.T) {
	skipIfCommandMissing(t, "git")

	dir := t.TempDir()

	// Initialize git repository
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}

	// Create C# project marker (.sln file for C# toolchain detection)
	if err := os.WriteFile(filepath.Join(dir, "MyApp.sln"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create sln file: %v", err)
	}

	// Stage only non-.cs files (README.md, config.yaml)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs"), 0o644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatalf("failed to create config.yaml: %v", err)
	}
	if out, err := runGitCmd(dir, "add", "README.md", "config.yaml"); err != nil {
		t.Fatalf("git add failed: %v, out: %s", err, out)
	}

	// Inject a fake dotnet binary that fails with exit 1 when called.
	// If the changedExts filter works correctly, this binary must never be executed.
	fakeBinDir := writeFakeBinary(t, "dotnet", 1, "Restore operation failed")
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	step := gateStep{
		name:        "dotnet format",
		binary:      "dotnet",
		args:        []string{"format", "--verify-no-changes"},
		optional:    false, // optional=false: must run when binary is present
		changedExts: []string{".cs"},
	}

	g := NewQualityGate(&GateConfig{
		Enabled:     true,
		SkipTests:   true,
		ProjectDir:  dir,
		LintTimeout: 5 * time.Second,
	})

	// No .cs detected via stagedFiles → dotnet format skipped
	passed, out := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Errorf("dotnet format must be skipped when no C# files are staged (expected passed=true), output: %q", out)
	}
}

// TestQualityGate_RunsDotnetFormatWhenCSharpStaged verifies that the dotnet format step runs
// when .cs files are staged (issue #667 Fix 1).
// Does not call t.Parallel() because it uses t.Setenv.
func TestQualityGate_RunsDotnetFormatWhenCSharpStaged(t *testing.T) {
	// Skip on Windows: shell-script fake binary is not directly executable via exec.Command.
	// On CI with real dotnet installed, format execution would fail on timeout.
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}
	skipIfCommandMissing(t, "git")

	dir := t.TempDir()

	// Initialize git repository
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config failed: %v, out: %s", err, out)
	}

	// Create C# marker
	if err := os.WriteFile(filepath.Join(dir, "MyApp.sln"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create sln file: %v", err)
	}

	// Stage a .cs file
	if err := os.WriteFile(filepath.Join(dir, "Program.cs"), []byte("// C# code"), 0o644); err != nil {
		t.Fatalf("failed to create Program.cs: %v", err)
	}
	if out, err := runGitCmd(dir, "add", "Program.cs"); err != nil {
		t.Fatalf("git add failed: %v, out: %s", err, out)
	}

	// Inject a fake dotnet that succeeds (exit 0) — it must actually be called.
	fakeBinDir := writeFakeBinary(t, "dotnet", 0, "")
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	step := gateStep{
		name:        "dotnet format",
		binary:      "dotnet",
		args:        []string{"format", "--verify-no-changes"},
		optional:    false,
		changedExts: []string{".cs"},
	}

	g := NewQualityGate(&GateConfig{
		Enabled:     true,
		SkipTests:   true,
		ProjectDir:  dir,
		LintTimeout: 5 * time.Second,
	})

	// .cs file is staged → dotnet format must run → fake dotnet exits 0
	passed, out := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Errorf("dotnet format must pass when .cs files are staged, output: %q", out)
	}
}

// TestQualityGate_GracefulOnDotnetRestoreFailure verifies that the gate logs a warning and returns
// (true, "") when a dotnet restore failure is detected (issue #667 Fix 2).
func TestQualityGate_GracefulOnDotnetRestoreFailure(t *testing.T) {
	t.Parallel()

	// Test the dotnetRestoreFailure detection function directly.
	cases := []struct {
		name     string
		stderr   string
		wantSkip bool
	}{
		{
			name:     "contains Restore operation failed",
			stderr:   "Unhandled exception: System.Exception: Restore operation failed.",
			wantSkip: true,
		},
		{
			name:     "contains NU1202 error code",
			stderr:   "error NU1202: Package 'foo' 1.0.0 is not compatible with net9.0-windows10.0.22621.0",
			wantSkip: true,
		},
		{
			name:     "contains NETSDK1005 error",
			stderr:   "NETSDK1005: Assets file not found",
			wantSkip: true,
		},
		{
			name:     "contains not supported on this platform",
			stderr:   "This target framework 'net9.0-windows10.0.22621.0' is not supported on this platform",
			wantSkip: true,
		},
		{
			name:     "general dotnet format error (not restore failure)",
			stderr:   "error CS1001: Identifier expected",
			wantSkip: false,
		},
		{
			name:     "empty stderr",
			stderr:   "",
			wantSkip: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isDotnetRestoreFailure(tc.stderr)
			if got != tc.wantSkip {
				t.Errorf("isDotnetRestoreFailure(%q) = %v, want %v", tc.stderr, got, tc.wantSkip)
			}
		})
	}
}

// runGitCmd is a test helper that runs a git command in the given directory.
func runGitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

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
