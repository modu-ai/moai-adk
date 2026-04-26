package quality

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// GateConfig holds configuration for the QualityGate.
type GateConfig struct {
	// Enabled controls whether the quality gate runs at all.
	Enabled bool
	// SkipTests skips the test step when true, useful for quick commits.
	SkipTests bool
	// VetTimeout is the maximum duration allowed for the vet/lint step.
	VetTimeout time.Duration
	// LintTimeout is the maximum duration allowed for the linter step.
	LintTimeout time.Duration
	// TestTimeout is the maximum duration allowed for the test step.
	TestTimeout time.Duration
	// ProjectDir is the project root directory used for language detection.
	// When empty, the current working directory is used.
	ProjectDir string
	// AstGrepGate configures the ast-grep domain rule scan step.
	// When nil, ast-grep scanning is skipped.
	AstGrepGate *AstGrepGateConfig
	// DisabledSteps disables specific steps by name (issue #667 Fix 3).
	// Keys are step names (e.g., "dotnet format"); a value of false skips that step.
	// Example: map[string]bool{"dotnet format": false}
	DisabledSteps map[string]bool
}

// DefaultGateConfig returns a GateConfig with production-safe defaults.
func DefaultGateConfig() *GateConfig {
	return &GateConfig{
		Enabled:     true,
		SkipTests:   false,
		VetTimeout:  30 * time.Second,
		LintTimeout: 60 * time.Second,
		TestTimeout: 120 * time.Second,
		AstGrepGate: DefaultAstGrepGateConfig(),
	}
}

// langToolchain defines the quality gate steps for a specific language.
type langToolchain struct {
	// markerFiles are files whose presence identifies this language (checked in order).
	markerFiles []string
	// vetSteps are vet/analyze commands run in order. Each is optional (skips if binary missing).
	vetSteps []gateStep
	// lintSteps are linter commands run in order. Each is optional.
	lintSteps []gateStep
	// testStep is the test command. Nil means no test step available.
	testStep *gateStep
}

// gateStep represents a single quality gate command.
type gateStep struct {
	name        string
	binary      string
	args        []string
	optional    bool     // If true, skip silently when binary is not found.
	configFiles []string // If non-empty, skip step when none of these files exist in project dir.
	// changedExts is the list of file extensions that trigger this step (issue #667 Fix 1).
	// When empty, the step always runs regardless of staged files.
	// When non-empty, the step is skipped if no staged file has one of these extensions.
	// Future heavy language-specific linters can opt-in using this pattern.
	changedExts []string
}

// toolchains defines quality gate steps per language.
// Order matters: first match by marker file wins.
var toolchains = []langToolchain{
	// Go: go.mod
	{
		markerFiles: []string{"go.mod"},
		vetSteps:    []gateStep{{name: "go vet", binary: "go", args: []string{"vet", "./..."}}},
		lintSteps: []gateStep{{
			name: "golangci-lint", binary: "golangci-lint", args: []string{"run"}, optional: true,
			configFiles: []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"},
		}},
		testStep: &gateStep{name: "go test", binary: "go", args: []string{"test", "./..."}},
	},
	// Node.js (TypeScript/JavaScript): package.json
	{
		markerFiles: []string{"package.json"},
		lintSteps: []gateStep{{
			name: "eslint", binary: "npx", args: []string{"eslint", "."}, optional: true,
			configFiles: []string{
				"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs",
				"eslint.config.ts", "eslint.config.mts", "eslint.config.cts",
				".eslintrc.js", ".eslintrc.cjs", ".eslintrc.yaml", ".eslintrc.yml", ".eslintrc.json", ".eslintrc",
			},
		}},
		testStep: &gateStep{name: "npm test", binary: "npm", args: []string{"test", "--", "--passWithNoTests"}},
	},
	// Python: pyproject.toml, setup.py, requirements.txt
	{
		markerFiles: []string{"pyproject.toml", "setup.py", "requirements.txt"},
		lintSteps: []gateStep{
			{name: "ruff", binary: "ruff", args: []string{"check", "."}, optional: true,
				configFiles: []string{"ruff.toml", ".ruff.toml", "pyproject.toml"},
			},
			{name: "mypy", binary: "mypy", args: []string{"."}, optional: true,
				configFiles: []string{"mypy.ini", ".mypy.ini", "pyproject.toml", "setup.cfg"},
			},
		},
		testStep: &gateStep{name: "pytest", binary: "pytest", args: nil, optional: true},
	},
	// Rust: Cargo.toml
	{
		markerFiles: []string{"Cargo.toml"},
		lintSteps:   []gateStep{{name: "cargo clippy", binary: "cargo", args: []string{"clippy", "--", "-D", "warnings"}, optional: true}},
		testStep:    &gateStep{name: "cargo test", binary: "cargo", args: []string{"test"}},
	},
	// Java: pom.xml (Maven) or build.gradle (Gradle)
	{
		markerFiles: []string{"pom.xml", "build.gradle", "build.gradle.kts"},
		lintSteps: []gateStep{{
			name: "checkstyle", binary: "checkstyle", args: []string{"-c", "/google_checks.xml", "src/"}, optional: true,
			configFiles: []string{"checkstyle.xml", "google_checks.xml", ".checkstyle"},
		}},
		testStep: &gateStep{name: "mvn test", binary: "mvn", args: []string{"test"}, optional: true},
	},
	// Kotlin: build.gradle.kts with .kt files
	{
		markerFiles: []string{"build.gradle.kts"},
		lintSteps: []gateStep{{
			name: "ktlint", binary: "ktlint", args: nil, optional: true,
			configFiles: []string{".editorconfig", ".ktlint"},
		}},
		testStep: &gateStep{name: "gradle test", binary: "gradle", args: []string{"test"}, optional: true},
	},
	// C#/.NET: *.csproj or *.sln
	// changedExts: skip dotnet format when no .cs file is staged.
	// Prevents NuGet restore failures on macOS for projects targeting Windows-only TFMs
	// (e.g., net9.0-windows10.0.22621.0) (issue #667).
	{
		markerFiles: []string{"*.csproj", "*.sln"},
		lintSteps:   []gateStep{{name: "dotnet format", binary: "dotnet", args: []string{"format", "--verify-no-changes"}, optional: true, changedExts: []string{".cs"}}},
		testStep:    &gateStep{name: "dotnet test", binary: "dotnet", args: []string{"test"}},
	},
	// Ruby: Gemfile
	{
		markerFiles: []string{"Gemfile"},
		lintSteps: []gateStep{{
			name: "rubocop", binary: "rubocop", args: nil, optional: true,
			configFiles: []string{".rubocop.yml", ".rubocop.yaml"},
		}},
		testStep: &gateStep{name: "rspec", binary: "rspec", args: nil, optional: true},
	},
	// PHP: composer.json
	{
		markerFiles: []string{"composer.json"},
		lintSteps: []gateStep{{
			name: "phpstan", binary: "phpstan", args: []string{"analyse"}, optional: true,
			configFiles: []string{"phpstan.neon", "phpstan.neon.dist", "phpstan.dist.neon"},
		}},
		testStep: &gateStep{name: "phpunit", binary: "phpunit", args: nil, optional: true},
	},
	// Swift: Package.swift
	{
		markerFiles: []string{"Package.swift"},
		lintSteps: []gateStep{{
			name: "swiftlint", binary: "swiftlint", args: nil, optional: true,
			configFiles: []string{".swiftlint.yml", ".swiftlint.yaml"},
		}},
		testStep: &gateStep{name: "swift test", binary: "swift", args: []string{"test"}},
	},
	// Dart/Flutter: pubspec.yaml
	// NOTE: Flutter projects are detected dynamically by inspecting pubspec.yaml
	// content in detectToolchain — Flutter's `package:test` dependency is provided
	// via `flutter_test` from the SDK, so `dart test` fails ("Could not find
	// package `test`"). We switch to `flutter test` / `flutter analyze` for
	// Flutter projects while keeping `dart` for pure Dart CLI projects.
	// See issue #652.
	{
		markerFiles: []string{"pubspec.yaml"},
		vetSteps:    []gateStep{{name: "dart analyze", binary: "dart", args: []string{"analyze"}}},
		testStep:    &gateStep{name: "dart test", binary: "dart", args: []string{"test"}, optional: true},
	},
	// Elixir: mix.exs
	{
		markerFiles: []string{"mix.exs"},
		lintSteps: []gateStep{{
			name: "credo", binary: "mix", args: []string{"credo"}, optional: true,
			configFiles: []string{".credo.exs"},
		}},
		testStep: &gateStep{name: "mix test", binary: "mix", args: []string{"test"}},
	},
	// Scala: build.sbt
	{
		markerFiles: []string{"build.sbt"},
		lintSteps: []gateStep{{
			name: "scalafix", binary: "scalafix", args: nil, optional: true,
			configFiles: []string{".scalafix.conf"},
		}},
		testStep: &gateStep{name: "sbt test", binary: "sbt", args: []string{"test"}, optional: true},
	},
	// Haskell: cabal project or stack
	{
		markerFiles: []string{"*.cabal", "stack.yaml"},
		testStep:    &gateStep{name: "cabal test", binary: "cabal", args: []string{"test"}, optional: true},
	},
	// Zig: build.zig
	{
		markerFiles: []string{"build.zig"},
		testStep: &gateStep{name: "zig test", binary: "zig", args: []string{"test"}, optional: true},
	},
}

// QualityGate runs deterministic quality checks before git commit.
// It detects the project language from marker files and runs the appropriate toolchain.
// If no language is detected, the gate passes silently.
type QualityGate struct {
	config *GateConfig

	// stagedCache caches the git diff --cached result exactly once per Run call.
	stagedCache        []string
	stagedCacheReady   bool // true when the query is complete (even if result is nil)
	stagedCacheNil     bool // true when nil was returned (conservative fallback)
}

// NewQualityGate creates a QualityGate with the given configuration.
// If cfg is nil, DefaultGateConfig is used.
func NewQualityGate(cfg *GateConfig) *QualityGate {
	if cfg == nil {
		cfg = DefaultGateConfig()
	}
	return &QualityGate{config: cfg}
}

// @MX:ANCHOR: [AUTO] Quality gate executor; primary entry point called by multiple hook handlers before git operations
// @MX:REASON: fan_in=35, invoked by PreCommit, SubagentStop, and TeammateIdle handlers; returns block/pass decision that controls git flow
// Run executes quality gate checks sequentially.
// Returns (passed bool, output string) where output contains error details on failure.
// When gate is disabled (config.Enabled == false), returns (true, "") immediately.
// The gate detects the project language and runs the corresponding toolchain.
func (g *QualityGate) Run(ctx context.Context) (bool, string) {
	if !g.config.Enabled {
		return true, ""
	}

	tc := g.detectToolchain()
	if tc == nil {
		// No recognized language — pass silently.
		return true, ""
	}

	// Step 1: vet steps
	for _, step := range tc.vetSteps {
		if ok, out := g.executeStep(ctx, step, g.config.VetTimeout); !ok {
			return false, out
		}
	}

	// Step 2: lint steps
	for _, step := range tc.lintSteps {
		if ok, out := g.executeStep(ctx, step, g.config.LintTimeout); !ok {
			return false, out
		}
	}

	// Step 2.5: ast-grep domain rules
	// ASTG-UPGRADE-001: switched to RunAstGrepGateV2 which uses the unified Scanner
	if g.config.AstGrepGate != nil && g.config.AstGrepGate.Enabled {
		projectDir := g.config.ProjectDir
		if projectDir == "" {
			projectDir, _ = os.Getwd()
		}
		if ok, out := RunAstGrepGateV2(ctx, projectDir, g.config.AstGrepGate); !ok {
			return false, out
		}
	}

	// Step 3: test step (skippable)
	if !g.config.SkipTests && tc.testStep != nil {
		if ok, out := g.executeStep(ctx, *tc.testStep, g.config.TestTimeout); !ok {
			return false, out
		}
	}

	return true, ""
}

// detectToolchain finds the matching toolchain by checking marker files in ProjectDir.
func (g *QualityGate) detectToolchain() *langToolchain {
	dir := g.config.ProjectDir
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if dir == "" {
		return nil
	}

	for i := range toolchains {
		for _, marker := range toolchains[i].markerFiles {
			if strings.Contains(marker, "*") {
				// Glob pattern (e.g., "*.csproj")
				matches, err := filepath.Glob(filepath.Join(dir, marker))
				if err == nil && len(matches) > 0 {
					return resolveDartFlutter(&toolchains[i], dir)
				}
			} else {
				if fileExists(filepath.Join(dir, marker)) {
					return resolveDartFlutter(&toolchains[i], dir)
				}
			}
		}
	}

	return nil
}

// resolveDartFlutter returns a Flutter-specific toolchain variant when the
// matched Dart toolchain's pubspec.yaml declares a Flutter SDK dependency,
// and the pure Dart variant otherwise. Flutter projects require
// `flutter test` / `flutter analyze` because `package:test` is provided
// transitively via `flutter_test` from the Flutter SDK (issue #652).
//
// Non-Dart toolchains are returned unchanged.
func resolveDartFlutter(tc *langToolchain, dir string) *langToolchain {
	// Only process toolchain entries whose first marker is pubspec.yaml.
	if len(tc.markerFiles) == 0 || tc.markerFiles[0] != "pubspec.yaml" {
		return tc
	}
	if !isFlutterProject(filepath.Join(dir, "pubspec.yaml")) {
		return tc
	}
	// Return a new langToolchain with flutter binary substitutions so we do
	// not mutate the package-level toolchains slice.
	return &langToolchain{
		markerFiles: tc.markerFiles,
		vetSteps:    []gateStep{{name: "flutter analyze", binary: "flutter", args: []string{"analyze"}}},
		testStep:    &gateStep{name: "flutter test", binary: "flutter", args: []string{"test"}, optional: true},
	}
}

// isFlutterProject reports whether the given pubspec.yaml declares the
// Flutter SDK as a dependency. Detection heuristic:
//   - "sdk: flutter" substring appears (Dart or Flutter dependency block)
//   - or "flutter:" top-level section appears (Flutter-specific config)
//
// Missing or unreadable files return false (safe fallback to `dart`).
func isFlutterProject(pubspecPath string) bool {
	data, err := os.ReadFile(pubspecPath)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "sdk: flutter") ||
		strings.Contains(content, "sdk:flutter") ||
		hasFlutterSection(content)
}

// hasFlutterSection reports whether pubspec content has a top-level
// `flutter:` section (not a dependency named "flutter").
func hasFlutterSection(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimRight(line, " \t")
		// Top-level section starts at column 0 with "flutter:"
		if trimmed == "flutter:" {
			return true
		}
	}
	return false
}

// executeStep runs a single gate step. Optional steps skip silently when the binary is missing.
// Steps with configFiles skip silently when none of the listed config files exist.
// Steps with changedExts skip silently when no staged file matches any of the listed extensions.
func (g *QualityGate) executeStep(ctx context.Context, step gateStep, timeout time.Duration) (bool, string) {
	// Fix 3: explicitly disable via DisabledSteps configuration
	if disabled, ok := g.config.DisabledSteps[step.name]; ok && !disabled {
		return true, ""
	}

	if step.optional {
		if _, err := exec.LookPath(step.binary); err != nil {
			return true, ""
		}
	}
	if len(step.configFiles) > 0 && !g.anyConfigFileExists(step.configFiles) {
		return true, ""
	}

	// Fix 1: skip when no staged file matches changedExts.
	// If stagedFiles lookup fails or we are outside a git repository, run the step conservatively.
	if len(step.changedExts) > 0 {
		dir := g.config.ProjectDir
		if dir == "" {
			dir, _ = os.Getwd()
		}
		staged := g.cachedStagedFiles(ctx, dir)
		// If staged is nil, cannot determine — run step conservatively
		if staged != nil && !hasStagedExt(staged, step.changedExts) {
			return true, ""
		}
	}

	return g.runStep(ctx, step.name, timeout, step.binary, step.args...)
}

// cachedStagedFiles queries stagedFiles exactly once per Run call and caches the result.
// Also caches when stagedFiles returns nil (conservative fallback).
func (g *QualityGate) cachedStagedFiles(ctx context.Context, dir string) []string {
	if !g.stagedCacheReady {
		files, _ := stagedFiles(ctx, dir)
		g.stagedCache = files
		g.stagedCacheReady = true
		g.stagedCacheNil = files == nil
	}
	if g.stagedCacheNil {
		return nil
	}
	return g.stagedCache
}

// hasStagedExt returns true if any file in the staged list has an extension in exts.
func hasStagedExt(staged []string, exts []string) bool {
	for _, f := range staged {
		ext := filepath.Ext(f)
		for _, want := range exts {
			if strings.EqualFold(ext, want) {
				return true
			}
		}
	}
	return false
}

// stagedFiles runs git diff --cached --name-only and returns the list of staged files.
// Returns nil when outside a git repository, the git binary is absent, or there are no staged files.
// Also returns nil on error (conservative fallback); callers must run the step when nil is returned.
func stagedFiles(ctx context.Context, dir string) ([]string, error) {
	// Check whether the git binary exists
	if _, err := exec.LookPath("git"); err != nil {
		return nil, nil //nolint:nilerr // conservative fallback
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// Outside a git repository or command failed — conservative fallback
		return nil, nil //nolint:nilerr
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		// No staged files — return nil (conservative: run the step)
		return nil, nil
	}

	lines := strings.Split(raw, "\n")
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

// anyConfigFileExists returns true if at least one of the given config files exists in ProjectDir.
func (g *QualityGate) anyConfigFileExists(configFiles []string) bool {
	dir := g.config.ProjectDir
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if dir == "" {
		return false
	}
	for _, cf := range configFiles {
		if fileExists(filepath.Join(dir, cf)) {
			return true
		}
	}
	return false
}

// runStep executes a single quality gate command with the given timeout.
// Returns (true, "") on success, (false, errorMessage) on failure or timeout.
func (g *QualityGate) runStep(ctx context.Context, stepName string, timeout time.Duration, name string, args ...string) (bool, string) {
	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(stepCtx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return true, ""
	}

	// Merge stdout and stderr; stderr typically has the diagnostics.
	combined := stderr.String() + stdout.String()
	output := strings.TrimSpace(combined)

	// Distinguish timeout from other failures (REQ-GATE-009).
	if stepCtx.Err() == context.DeadlineExceeded {
		msg := fmt.Sprintf("quality gate timed out: %s exceeded %s", stepName, timeout)
		return false, msg
	}

	// Fix 2: when a NuGet restore failure (cross-platform TFM mismatch) is detected in the dotnet step,
	// log a warning and pass. Other linters still propagate failures (issue #667).
	if strings.Contains(strings.ToLower(stepName), "dotnet") && isDotnetRestoreFailure(combined) {
		slog.Warn("dotnet restore failed: assumed cross-platform TFM mismatch, skipping",
			"step", stepName,
			"hint", "a Windows-only TFM may have failed to restore on macOS",
		)
		return true, ""
	}

	if output == "" {
		output = err.Error()
	}
	return false, fmt.Sprintf("quality gate failed: %s\n\n%s", stepName, output)
}

// isDotnetRestoreFailure checks whether the stderr/stdout output contains NuGet restore failure markers.
// Detects the error pattern that occurs when a Windows-only TFM (e.g., net9.0-windows10.0.22621.0)
// fails to restore on macOS (issue #667 Fix 2).
// This function applies only to the dotnet step; other linters propagate failures as-is.
func isDotnetRestoreFailure(output string) bool {
	markers := []string{
		"Restore operation failed",
		"NU1202",
		"NETSDK1005",
		"not supported on this platform",
	}
	for _, m := range markers {
		if strings.Contains(output, m) {
			return true
		}
	}
	return false
}

// isGitCommitRe matches git commit commands.
// Matches: git commit, git commit -m "...", git commit --amend, git commit --no-verify, etc.
// Does NOT match: git commit --help, echo "git commit".
var isGitCommitRe = regexp.MustCompile(`^\s*git\s+commit\b`)

// IsGitCommit reports whether command is a git commit invocation.
// --no-verify and --amend flags do not bypass the gate (REQ-GATE-011).
func IsGitCommit(command string) bool {
	return isGitCommitRe.MatchString(command)
}

// fileExists returns true if the path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
