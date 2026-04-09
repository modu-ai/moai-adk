package quality

import (
	"bytes"
	"context"
	"fmt"
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
}

// DefaultGateConfig returns a GateConfig with production-safe defaults.
func DefaultGateConfig() *GateConfig {
	return &GateConfig{
		Enabled:     true,
		SkipTests:   false,
		VetTimeout:  30 * time.Second,
		LintTimeout: 60 * time.Second,
		TestTimeout: 120 * time.Second,
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
	name     string
	binary   string
	args     []string
	optional bool // If true, skip silently when binary is not found.
}

// toolchains defines quality gate steps per language.
// Order matters: first match by marker file wins.
var toolchains = []langToolchain{
	// Go: go.mod
	{
		markerFiles: []string{"go.mod"},
		vetSteps:    []gateStep{{name: "go vet", binary: "go", args: []string{"vet", "./..."}}},
		lintSteps:   []gateStep{{name: "golangci-lint", binary: "golangci-lint", args: []string{"run"}, optional: true}},
		testStep:    &gateStep{name: "go test", binary: "go", args: []string{"test", "./..."}},
	},
	// Node.js (TypeScript/JavaScript): package.json
	{
		markerFiles: []string{"package.json"},
		lintSteps:   []gateStep{{name: "eslint", binary: "npx", args: []string{"eslint", "."}, optional: true}},
		testStep:    &gateStep{name: "npm test", binary: "npm", args: []string{"test", "--", "--passWithNoTests"}},
	},
	// Python: pyproject.toml, setup.py, requirements.txt
	{
		markerFiles: []string{"pyproject.toml", "setup.py", "requirements.txt"},
		lintSteps: []gateStep{
			{name: "ruff", binary: "ruff", args: []string{"check", "."}, optional: true},
			{name: "mypy", binary: "mypy", args: []string{"."}, optional: true},
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
		lintSteps:   []gateStep{{name: "checkstyle", binary: "checkstyle", args: []string{"-c", "/google_checks.xml", "src/"}, optional: true}},
		testStep:    &gateStep{name: "mvn test", binary: "mvn", args: []string{"test"}, optional: true},
	},
	// Kotlin: build.gradle.kts with .kt files
	{
		markerFiles: []string{"build.gradle.kts"},
		lintSteps:   []gateStep{{name: "ktlint", binary: "ktlint", args: nil, optional: true}},
		testStep:    &gateStep{name: "gradle test", binary: "gradle", args: []string{"test"}, optional: true},
	},
	// C#/.NET: *.csproj or *.sln
	{
		markerFiles: []string{"*.csproj", "*.sln"},
		lintSteps:   []gateStep{{name: "dotnet format", binary: "dotnet", args: []string{"format", "--verify-no-changes"}, optional: true}},
		testStep:    &gateStep{name: "dotnet test", binary: "dotnet", args: []string{"test"}},
	},
	// Ruby: Gemfile
	{
		markerFiles: []string{"Gemfile"},
		lintSteps:   []gateStep{{name: "rubocop", binary: "rubocop", args: nil, optional: true}},
		testStep:    &gateStep{name: "rspec", binary: "rspec", args: nil, optional: true},
	},
	// PHP: composer.json
	{
		markerFiles: []string{"composer.json"},
		lintSteps:   []gateStep{{name: "phpstan", binary: "phpstan", args: []string{"analyse"}, optional: true}},
		testStep:    &gateStep{name: "phpunit", binary: "phpunit", args: nil, optional: true},
	},
	// Swift: Package.swift
	{
		markerFiles: []string{"Package.swift"},
		lintSteps:   []gateStep{{name: "swiftlint", binary: "swiftlint", args: nil, optional: true}},
		testStep:    &gateStep{name: "swift test", binary: "swift", args: []string{"test"}},
	},
	// Dart/Flutter: pubspec.yaml
	{
		markerFiles: []string{"pubspec.yaml"},
		vetSteps:    []gateStep{{name: "dart analyze", binary: "dart", args: []string{"analyze"}}},
		testStep:    &gateStep{name: "dart test", binary: "dart", args: []string{"test"}, optional: true},
	},
	// Elixir: mix.exs
	{
		markerFiles: []string{"mix.exs"},
		lintSteps:   []gateStep{{name: "credo", binary: "mix", args: []string{"credo"}, optional: true}},
		testStep:    &gateStep{name: "mix test", binary: "mix", args: []string{"test"}},
	},
	// Scala: build.sbt
	{
		markerFiles: []string{"build.sbt"},
		lintSteps:   []gateStep{{name: "scalafix", binary: "scalafix", args: nil, optional: true}},
		testStep:    &gateStep{name: "sbt test", binary: "sbt", args: []string{"test"}, optional: true},
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
}

// NewQualityGate creates a QualityGate with the given configuration.
// If cfg is nil, DefaultGateConfig is used.
func NewQualityGate(cfg *GateConfig) *QualityGate {
	if cfg == nil {
		cfg = DefaultGateConfig()
	}
	return &QualityGate{config: cfg}
}

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
					return &toolchains[i]
				}
			} else {
				if fileExists(filepath.Join(dir, marker)) {
					return &toolchains[i]
				}
			}
		}
	}

	return nil
}

// executeStep runs a single gate step. Optional steps skip silently when the binary is missing.
func (g *QualityGate) executeStep(ctx context.Context, step gateStep, timeout time.Duration) (bool, string) {
	if step.optional {
		if _, err := exec.LookPath(step.binary); err != nil {
			return true, ""
		}
	}
	return g.runStep(ctx, step.name, timeout, step.binary, step.args...)
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

	if output == "" {
		output = err.Error()
	}
	return false, fmt.Sprintf("quality gate failed: %s\n\n%s", stepName, output)
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
