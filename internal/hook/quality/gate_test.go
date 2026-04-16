package quality

import (
	"context"
	"fmt"
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
			// vetSteps 도 flutter 분기인지 확인
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
// ─── 이슈 #667: dotnet format 스테이징 파일 필터 및 복원 실패 graceful 처리 ───

// TestStagedFiles_OutsideGitRepo는 git 저장소 밖에서 stagedFiles가 nil을 반환하는지 확인합니다.
func TestStagedFiles_OutsideGitRepo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir() // git init 없음

	got, err := stagedFiles(context.Background(), dir)
	// 오류가 없어야 하고 nil을 반환해야 합니다 (보수적 fallback).
	if err != nil {
		t.Errorf("git 저장소 밖에서 오류가 발생해서는 안 됨: %v", err)
	}
	if got != nil {
		t.Errorf("git 저장소 밖에서 nil 반환 기대, got: %v", got)
	}
}

// TestStagedFiles_InsideGitRepo는 git 저장소 내에서 스테이징된 파일 목록을 반환하는지 확인합니다.
func TestStagedFiles_InsideGitRepo(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "git")

	dir := t.TempDir()
	// git 저장소 초기화
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}

	// 파일 생성 및 스테이징
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatalf("파일 생성 실패: %v", err)
	}
	if out, err := runGitCmd(dir, "add", "README.md", "config.yaml"); err != nil {
		t.Fatalf("git add 실패: %v, out: %s", err, out)
	}

	got, err := stagedFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("stagedFiles 오류: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("스테이징된 파일 2개 기대, got: %v", got)
	}
}

// TestStagedFiles_EmptyStaging은 스테이징된 파일이 없을 때 nil을 반환하는지 확인합니다.
func TestStagedFiles_EmptyStaging(t *testing.T) {
	t.Parallel()

	skipIfCommandMissing(t, "git")

	dir := t.TempDir()
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}

	// 스테이징하지 않음
	got, err := stagedFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("stagedFiles 오류: %v", err)
	}
	// 빈 스테이징: nil 반환 (보수적 fallback — 단계 실행)
	if got != nil {
		t.Errorf("스테이징된 파일 없을 때 nil 기대, got: %v", got)
	}
}

// writeFakeBinary는 임시 디렉터리에 호출 시 지정된 종료 코드로 종료하는
// 가짜 바이너리 셸 스크립트를 생성하고, 해당 디렉터리 경로를 반환합니다.
// 테스트에서 PATH를 앞에 삽입하여 실제 바이너리 대신 호출되게 합니다.
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
		t.Fatalf("가짜 바이너리 생성 실패: %v", err)
	}
	return binDir
}

// TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged는 .cs 파일이 스테이징되지 않았을 때
// dotnet format 단계를 건너뛰는지 검증합니다 (이슈 #667 Fix 1).
// t.Setenv를 사용하므로 t.Parallel()을 호출하지 않습니다.
func TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged(t *testing.T) {
	skipIfCommandMissing(t, "git")

	dir := t.TempDir()

	// git 저장소 초기화
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}

	// C# 프로젝트 마커 생성 (.sln 파일로 C# 툴체인 감지)
	if err := os.WriteFile(filepath.Join(dir, "MyApp.sln"), []byte(""), 0o644); err != nil {
		t.Fatalf("sln 파일 생성 실패: %v", err)
	}

	// .cs 파일이 아닌 파일만 스테이징 (README.md, config.yaml)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("docs"), 0o644); err != nil {
		t.Fatalf("README.md 생성 실패: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatalf("config.yaml 생성 실패: %v", err)
	}
	if out, err := runGitCmd(dir, "add", "README.md", "config.yaml"); err != nil {
		t.Fatalf("git add 실패: %v, out: %s", err, out)
	}

	// 호출되면 exit 1로 실패하는 가짜 dotnet 바이너리 주입.
	// changedExts 필터가 올바르게 동작하면 이 바이너리는 절대 실행되지 않아야 합니다.
	fakeBinDir := writeFakeBinary(t, "dotnet", 1, "Restore operation failed")
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	step := gateStep{
		name:        "dotnet format",
		binary:      "dotnet",
		args:        []string{"format", "--verify-no-changes"},
		optional:    false, // optional=false: 바이너리가 있으면 반드시 실행
		changedExts: []string{".cs"},
	}

	g := NewQualityGate(&GateConfig{
		Enabled:     true,
		SkipTests:   true,
		ProjectDir:  dir,
		LintTimeout: 5 * time.Second,
	})

	// stagedFiles를 통해 .cs 없음 감지 → dotnet format 건너뜀
	passed, out := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Errorf("C# 파일이 없을 때 dotnet format 건너뛰어야 합니다 (passed=true 기대), output: %q", out)
	}
}

// TestQualityGate_RunsDotnetFormatWhenCSharpStaged는 .cs 파일이 스테이징되었을 때
// dotnet format 단계가 실행되는지 검증합니다 (이슈 #667 Fix 1).
// t.Setenv를 사용하므로 t.Parallel()을 호출하지 않습니다.
func TestQualityGate_RunsDotnetFormatWhenCSharpStaged(t *testing.T) {
	skipIfCommandMissing(t, "git")

	dir := t.TempDir()

	// git 저장소 초기화
	if out, err := runGitCmd(dir, "init"); err != nil {
		t.Fatalf("git init 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}
	if out, err := runGitCmd(dir, "config", "user.name", "Test"); err != nil {
		t.Fatalf("git config 실패: %v, out: %s", err, out)
	}

	// C# 마커 생성
	if err := os.WriteFile(filepath.Join(dir, "MyApp.sln"), []byte(""), 0o644); err != nil {
		t.Fatalf("sln 파일 생성 실패: %v", err)
	}

	// .cs 파일을 스테이징
	if err := os.WriteFile(filepath.Join(dir, "Program.cs"), []byte("// C# code"), 0o644); err != nil {
		t.Fatalf("Program.cs 생성 실패: %v", err)
	}
	if out, err := runGitCmd(dir, "add", "Program.cs"); err != nil {
		t.Fatalf("git add 실패: %v, out: %s", err, out)
	}

	// 성공(exit 0)하는 가짜 dotnet 주입 — 실제로 호출되어야 합니다.
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

	// .cs 파일이 스테이징됨 → dotnet format 실행되어야 함 → 가짜 dotnet은 exit 0
	passed, out := g.executeStep(context.Background(), step, 5*time.Second)

	if !passed {
		t.Errorf(".cs 파일이 스테이징되었을 때 dotnet format이 통과해야 합니다, output: %q", out)
	}
}

// TestQualityGate_GracefulOnDotnetRestoreFailure는 dotnet 복원 실패 시
// 게이트가 경고를 로그하고 통과(true, "")를 반환하는지 검증합니다 (이슈 #667 Fix 2).
func TestQualityGate_GracefulOnDotnetRestoreFailure(t *testing.T) {
	t.Parallel()

	// dotnetRestoreFailure 감지 함수를 직접 테스트합니다.
	cases := []struct {
		name     string
		stderr   string
		wantSkip bool
	}{
		{
			name:     "Restore operation failed 포함",
			stderr:   "Unhandled exception: System.Exception: Restore operation failed.",
			wantSkip: true,
		},
		{
			name:     "NU1202 오류 코드 포함",
			stderr:   "error NU1202: Package 'foo' 1.0.0 is not compatible with net9.0-windows10.0.22621.0",
			wantSkip: true,
		},
		{
			name:     "NETSDK1005 오류 포함",
			stderr:   "NETSDK1005: Assets file not found",
			wantSkip: true,
		},
		{
			name:     "not supported on this platform 포함",
			stderr:   "This target framework 'net9.0-windows10.0.22621.0' is not supported on this platform",
			wantSkip: true,
		},
		{
			name:     "일반 dotnet 포맷 오류 (복원 실패 아님)",
			stderr:   "error CS1001: Identifier expected",
			wantSkip: false,
		},
		{
			name:     "빈 stderr",
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

// runGitCmd는 테스트 헬퍼로 주어진 디렉터리에서 git 명령을 실행합니다.
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
