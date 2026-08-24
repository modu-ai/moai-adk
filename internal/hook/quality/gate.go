package quality

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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
	// GoBuildTags is a space- or comma-separated list of Go build tags (e.g.
	// "goolm") applied to the Go toolchain vet/test/lint steps when non-empty.
	// Sourced from the project .moai/config/build-tags file so a project that
	// requires non-default build tags (cgo-alternative pure-Go libraries such as
	// goolm, sqlite-nocgo) is vetted under the same tags it compiles with;
	// otherwise go vet ./... fails on the cgo variant (libolm fatal) and the
	// gate false-negatives. Ignored for non-Go toolchains.
	GoBuildTags string
	// AstGrepGate configures the ast-grep domain rule scan step.
	// When nil, ast-grep scanning is skipped.
	AstGrepGate *AstGrepGateConfig
	// DisabledSteps disables specific steps by name (issue #667 Fix 3).
	// Keys are step names (e.g., "dotnet format"); a value of false skips that step.
	// Example: map[string]bool{"dotnet format": false}
	DisabledSteps map[string]bool
	// TypecheckEnabled controls the typecheck axis. Default true.
	TypecheckEnabled bool
	// TypecheckCommand overrides the language default and enables the axis for
	// any language. Split on whitespace (strings.Fields) and executed directly,
	// so shell metacharacters are not interpreted — a pipeline or a redirect
	// belongs in a script the command invokes.
	TypecheckCommand string
	// TypecheckTimeout is the maximum duration allowed for the typecheck step.
	TypecheckTimeout time.Duration
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
		// The axis ships on: a project with no type-check surface reports the
		// skip and passes, so enabling it by default costs nothing while
		// closing the hole for every project that does have one.
		TypecheckEnabled: true,
		TypecheckTimeout: 300 * time.Second,
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
	// typecheckStep is the type-check command, resolved against the project at
	// run time by resolveTypecheckStep. Nil means the language has no default
	// type-check axis — the gate then reports the absence rather than passing
	// quietly, and gate.typecheck.command supplies one for any language.
	typecheckStep *gateStep
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
	// sourceExts lists the extensions that give this step something to check.
	// When the project tree holds none of them, the step is skipped with a
	// hint instead of run.
	//
	// This exists for tools that treat an empty source set as a usage error
	// rather than a vacuous pass. mypy is the case: `mypy .` on a directory
	// with no .py files prints "There are no .py[i] files in directory" and
	// exits 2, so a freshly scaffolded project — a pyproject.toml and nothing
	// else yet — failed its first gate for having no code. pytest already
	// treats its own empty-suite exit as a pass (see runStep); this is the
	// same judgement moved one step earlier, to the point where the tool has
	// nothing to look at.
	//
	// Skipping BEFORE running, rather than forgiving an exit code after, is
	// deliberate: mypy's exit 2 also covers a genuinely broken config, and
	// forgiving the code would hide that. Absence of sources is checked
	// directly instead.
	//
	// Distinct from changedExts, which asks what is staged for THIS commit;
	// this asks whether the project has any such source at all.
	sourceExts []string
}

// toolchains defines quality gate steps per language.
// Order matters: first match by marker file wins.
var toolchains = []langToolchain{
	// Go: go.mod
	{
		markerFiles: []string{"go.mod"},
		// sourceExts: a module with a go.mod but not one .go file yet is a
		// scaffold, and `go vet ./...` / `go test ./...` exit non-zero there
		// ("matched no packages"). Skipping keeps a first commit from being
		// blocked for having no code — the same policy the Python steps carry.
		vetSteps: []gateStep{{name: "go vet", binary: "go", args: []string{"vet", "./..."}, sourceExts: []string{".go"}}},
		lintSteps: []gateStep{{
			name: "golangci-lint", binary: "golangci-lint", args: []string{"run"}, optional: true,
			configFiles: []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"},
			sourceExts:  []string{".go"},
		}},
		testStep: &gateStep{name: "go test", binary: "go", args: []string{"test", "./..."}, sourceExts: []string{".go"}},
	},
	// Node.js (TypeScript/JavaScript): package.json
	//
	// This entry has no vetSteps, which is what made Node the one type-unsafe
	// blind spot: mypy sits on the lint axis, Dart analyze on the vet axis, go
	// vet type-checks as it runs, and a compiled language's test step compiles
	// first. Node had only lint and test, and a project whose linter config is
	// absent (biome instead of eslint, say) skipped lint too — leaving a
	// type-broken build to pass the gate. typecheckStep closes that hole.
	{
		markerFiles:   []string{"package.json"},
		typecheckStep: nodeTypecheckStep(),
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
		// changedExts: skip the Python linters when no .py file is staged,
		// mirroring the .cs scoping on the C#/.NET entry below (issue #1265).
		lintSteps: []gateStep{
			{name: "ruff", binary: "ruff", args: []string{"check", "."}, optional: true,
				configFiles: []string{"ruff.toml", ".ruff.toml", "pyproject.toml"},
				changedExts: []string{".py"},
			},
			{name: "mypy", binary: "mypy", args: []string{"."}, optional: true,
				configFiles: []string{"mypy.ini", ".mypy.ini", "pyproject.toml", "setup.cfg"},
				changedExts: []string{".py"},
				// ruff needs no counterpart: `ruff check .` on a source-free
				// tree reports "All checks passed!" and exits 0. mypy is the
				// one Python step that treats no sources as an error.
				sourceExts: []string{".py", ".pyi"},
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
		testStep:    &gateStep{name: "dotnet test", binary: "dotnet", args: []string{"test"}, optional: true},
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
		testStep:    &gateStep{name: "zig test", binary: "zig", args: []string{"test"}, optional: true},
	},
}

// QualityGate runs deterministic quality checks before git commit.
// It detects the project language from marker files and runs the appropriate toolchain.
// If no language is detected, the gate passes silently.
type QualityGate struct {
	config *GateConfig

	// stagedCache caches the git diff --cached result exactly once per Run call.
	stagedCache      []string
	stagedCacheReady bool // true when the query is complete (even if result is nil)
	stagedCacheNil   bool // true when nil was returned (conservative fallback)
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
// @MX:REASON: fan_in=35, invoked by SubagentStop and TeammateIdle handlers; returns block/pass decision that controls git flow
// Run executes quality gate checks sequentially.
// Returns (passed bool, output string). On failure output carries the error
// details. On success output is usually empty, but carries a non-blocking
// notice when a step passed with something to report — an ast-grep step that
// skipped because sg is absent, for instance. Callers must not read an empty
// success output as the only possible success shape.
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
	var vetReason string
	for _, step := range tc.vetSteps {
		ok, out := g.executeStep(ctx, step, g.config.VetTimeout)
		if !ok {
			return false, out
		}
		vetReason = appendReason(vetReason, out)
	}

	// passReason carries a passing step's notice out to Run's caller. Dropping
	// it here is what left an absent ast-grep scanner indistinguishable from a
	// clean scan: the step reported the skip and this frame threw it away.
	passReason := vetReason

	// Step 1.5: typecheck axis.
	//
	// It runs after vet and before lint so a type error surfaces before the
	// slower style pass, and so a project whose linter is absent still gets a
	// correctness gate. Every outcome is reported: a skip is not a failure,
	// but it is never silent — that silence is what let a broken build through.
	if g.config.TypecheckEnabled {
		step, reason, ok := resolveTypecheckStep(tc.typecheckStep, resolveQualityProjectDir(*g.config, "QualityGate.Run.typecheck"), g.config.TypecheckCommand)
		switch {
		case ok:
			passed, out := g.executeStep(ctx, step, g.config.TypecheckTimeout)
			if !passed {
				return false, out
			}
			passReason = appendReason(passReason, out)
		default:
			passReason = appendReason(passReason, reason)
		}
	}

	// Step 2: lint steps
	for _, step := range tc.lintSteps {
		ok, out := g.executeStep(ctx, step, g.config.LintTimeout)
		if !ok {
			return false, out
		}
		passReason = appendReason(passReason, out)
	}

	// Step 2.5: ast-grep domain rules
	// ASTG-UPGRADE-001: switched to RunAstGrepGateV2 which uses the unified Scanner
	if g.config.AstGrepGate != nil && g.config.AstGrepGate.Enabled {
		// REQ-HCWA-007: route cwd resolution through resolveQualityProjectDir.
		projectDir := resolveQualityProjectDir(*g.config, "QualityGate.Run.astgrep")
		ok, out := RunAstGrepGateV2(ctx, projectDir, g.config.AstGrepGate)
		if !ok {
			return false, out
		}
		passReason = appendReason(passReason, out)
	}

	// Step 3: test step (skippable)
	if !g.config.SkipTests && tc.testStep != nil {
		// The Node test step is resolved to a self-terminating (run-form)
		// command right before execution, reading the project's package.json
		// scripts from the resolved project dir.
		// Every other toolchain's step passes through unchanged.
		// REQ-HCWA-007: route cwd resolution through resolveQualityProjectDir.
		testStep := resolveNodeTestStep(*tc.testStep, resolveQualityProjectDir(*g.config, "QualityGate.Run.nodeTestStep"))
		if ok, out := g.executeStep(ctx, testStep, g.config.TestTimeout); !ok {
			return false, out
		}
	}

	return true, passReason
}

// detectToolchain finds the matching toolchain by checking marker files in ProjectDir.
func (g *QualityGate) detectToolchain() *langToolchain {
	// REQ-HCWA-007: route cwd resolution through resolveQualityProjectDir.
	dir := resolveQualityProjectDir(*g.config, "QualityGate.detectToolchain")
	if dir == "" {
		return nil
	}

	for i := range toolchains {
		for _, marker := range toolchains[i].markerFiles {
			if strings.Contains(marker, "*") {
				// Glob pattern (e.g., "*.csproj")
				matches, err := filepath.Glob(filepath.Join(dir, marker))
				if err == nil && len(matches) > 0 {
					return resolvePythonRunner(resolveGoBuildTags(resolveDartFlutter(&toolchains[i], dir), g.config.GoBuildTags), dir)
				}
			} else {
				if fileExists(filepath.Join(dir, marker)) {
					return resolvePythonRunner(resolveGoBuildTags(resolveDartFlutter(&toolchains[i], dir), g.config.GoBuildTags), dir)
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

// resolveGoBuildTags returns a Go-toolchain variant whose vet/test/lint
// steps carry the Go build-tags flag when tags is non-empty, so the gate
// vets and tests the project under the same build tags it compiles with.
// A non-Go toolchain and an empty tags value are returned unchanged. The
// returned toolchain is a clone (the package-level toolchains slice is never
// mutated), mirroring resolveDartFlutter.
func resolveGoBuildTags(tc *langToolchain, tags string) *langToolchain {
	if tags == "" {
		return tc
	}
	// Only the Go toolchain (marker go.mod) carries go vet/test/golangci-lint.
	if len(tc.markerFiles) == 0 || tc.markerFiles[0] != "go.mod" {
		return tc
	}
	clone := &langToolchain{markerFiles: tc.markerFiles}
	clone.vetSteps = make([]gateStep, len(tc.vetSteps))
	for i := range tc.vetSteps {
		clone.vetSteps[i] = applyGoBuildTags(tc.vetSteps[i], tags)
	}
	clone.lintSteps = make([]gateStep, len(tc.lintSteps))
	for i := range tc.lintSteps {
		clone.lintSteps[i] = applyGoBuildTags(tc.lintSteps[i], tags)
	}
	if tc.testStep != nil {
		ts := applyGoBuildTags(*tc.testStep, tags)
		clone.testStep = &ts
	}
	return clone
}

// resolvePythonRunner returns a Python-toolchain variant whose pytest step is
// resolved against project-local runners. Non-Python toolchains and toolchains
// without a test step are returned unchanged. The returned toolchain is a clone
// (the package-level toolchains slice is never mutated), mirroring
// resolveDartFlutter and resolveGoBuildTags.
func resolvePythonRunner(tc *langToolchain, dir string) *langToolchain {
	if tc == nil || tc.testStep == nil || tc.testStep.name != pytestStepName || dir == "" {
		return tc
	}
	resolved := resolvePytestRunner(*tc.testStep, dir)
	if resolved.binary == tc.testStep.binary {
		return tc // no project-local runner found; leave the table entry alone
	}
	clone := &langToolchain{
		markerFiles: tc.markerFiles,
		vetSteps:    tc.vetSteps,
		lintSteps:   tc.lintSteps,
	}
	clone.testStep = &resolved
	return clone
}

// resolvePytestRunner rewrites the pytest step to invoke the runner the project
// actually owns, in order: an in-venv pytest, `uv run pytest`, `poetry run
// pytest`, then the bare `pytest` fallback.
//
// A bare `pytest` off $PATH is wrong for most modern Python projects: a project
// using .venv, uv, or poetry has no global pytest, so the optional step silently
// no-opped — or worse, resolved an interpreter belonging to a different project
// (issue #1265).
//
// The in-venv form is an absolute path we have already proved exists, so the
// step drops `optional` — the $PATH LookPath skip in executeStep would otherwise
// discard a runner we know is there. The uv/poetry and bare forms stay optional:
// those binaries live on $PATH and a missing one must skip, not block a commit.
// The step name is preserved throughout so runStep's pytest exit-5
// no-tests-collected handling continues to apply.
func resolvePytestRunner(step gateStep, dir string) gateStep {
	if step.name != pytestStepName || dir == "" {
		return step
	}

	if venv := venvPytestPath(dir); venv != "" {
		step.binary = venv
		step.args = nil
		step.optional = false
		return step
	}
	if hasUVProject(dir) {
		step.binary = "uv"
		step.args = []string{"run", "pytest"}
		return step
	}
	if fileExists(filepath.Join(dir, "poetry.lock")) {
		step.binary = "poetry"
		step.args = []string{"run", "pytest"}
		return step
	}
	return step
}

// venvPytestPath returns the path to an in-venv pytest executable, or "" when
// none exists. Both the POSIX (.venv/bin) and Windows (.venv/Scripts) layouts
// are probed so the resolution is not host-specific.
func venvPytestPath(dir string) string {
	candidates := []string{
		filepath.Join(dir, ".venv", "bin", "pytest"),
		filepath.Join(dir, ".venv", "Scripts", "pytest.exe"),
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

// hasUVProject reports whether the project is uv-managed: a uv.lock file, or a
// [tool.uv] section in pyproject.toml.
func hasUVProject(dir string) bool {
	if fileExists(filepath.Join(dir, "uv.lock")) {
		return true
	}
	data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "[tool.uv") {
			return true
		}
	}
	return false
}

// applyGoBuildTags returns a copy of step with the Go build-tags flag injected.
// go-subcommand steps (binary "go") get "-tags=<tags>" right after the
// subcommand; golangci-lint gets "--build-tags=<tags>" appended (golangci-lint
// uses a different flag name than go). Other steps are returned unchanged.
// An empty tags value returns the step unchanged.
func applyGoBuildTags(step gateStep, tags string) gateStep {
	if tags == "" {
		return step
	}
	switch step.binary {
	case "go":
		// Insert -tags=<tags> immediately after the subcommand (args[0]).
		args := make([]string, 0, len(step.args)+1)
		if len(step.args) > 0 {
			args = append(args, step.args[0])
		}
		args = append(args, "-tags="+tags)
		if len(step.args) > 1 {
			args = append(args, step.args[1:]...)
		}
		out := step
		out.args = args
		return out
	case "golangci-lint":
		args := append([]string{}, step.args...)
		args = append(args, "--build-tags="+tags)
		out := step
		out.args = args
		return out
	default:
		return step
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

// nodeTestStepName identifies the Node toolchain's test step. The resolution
// keys off this name, mirroring how resolvePythonRunner keys off
// pytestStepName.
const nodeTestStepName = "npm test"

// nodeTestRunScript is the npm script name that denotes a self-terminating
// test command (the strongest run-form signal, tier (i)).
const nodeTestRunScript = "test:run"

// resolveNodeTestStep rewrites the Node toolchain's test step into a
// self-terminating (run-form) command when the project's package.json makes
// that possible. The hardcoded `npm test --` was watch-prone for packages
// whose `test` script never exits (bare `vitest`, or a `--watch`/`--watchAll`
// token), so the step always died at TestTimeout with the suite at 0%.
// Resolution tiers, in priority order:
//
//	(i)   scripts.test:run present → `npm run test:run`
//	(ii)  scripts.test watch-prone → the runner's non-watch flag appended to
//	      `npm test` (vitest `--run`; jest `--ci`)
//	(iii) anything else → `npm test` unchanged
//
// Tier (i) appends no flags: the test:run script is an author-owned run-form
// command, and appending `--passWithNoTests` via npm's `--` makes a
// turbo-delegating script (`turbo run test:run --passWithNoTests`) hard-error
// on the unknown argument, breaking the gate on monorepo roots (measured
// 2026-08-19). `--passWithNoTests` stays appended on tiers (ii)/(iii): a
// package with an empty test suite must keep passing the gate rather than
// regress. Parse failures — missing package.json, invalid JSON, absent
// scripts map — fall back to tier (iii) so the gate never fails because it
// could not read the manifest. Non-Node steps are returned unchanged
// (REQ-HGT-004).
func resolveNodeTestStep(step gateStep, dir string) gateStep {
	if step.name != nodeTestStepName || dir == "" {
		return step
	}
	scripts, ok := readPackageJSONScripts(filepath.Join(dir, "package.json"))
	if !ok {
		return step
	}
	if strings.TrimSpace(scripts[nodeTestRunScript]) != "" {
		return gateStep{
			name:   "npm run test:run",
			binary: "npm",
			args:   []string{"run", nodeTestRunScript},
		}
	}
	if flag := nodeNonWatchFlag(scripts["test"]); flag != "" {
		return gateStep{
			name:   nodeTestStepName + " " + flag,
			binary: "npm",
			args:   append([]string{"test", "--", "--passWithNoTests"}, flag),
		}
	}
	return step
}

// readPackageJSONScripts parses the scripts map out of the package.json at
// path. ok is false when the file is missing, unreadable, not valid JSON, the
// root is not an object, or scripts is absent/not a string map — every
// failure routes the Node test-step resolution to the tier-(iii) fallback.
func readPackageJSONScripts(path string) (map[string]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, false
	}
	if pkg.Scripts == nil {
		return nil, false
	}
	return pkg.Scripts, true
}

// nodeNonWatchFlag returns the runner's non-watch flag for a watch-prone test
// script: `--run` for vitest, `--ci` for jest. jest takes `--ci` rather than
// `--watchAll=false` because `--ci` uniformly negates both `--watch` and
// `--watchAll` (appending `--watchAll=false` would leave a `jest --watch`
// script still watching). Empty when the script is not watch-prone or the
// runner is unknown — for an unknown runner, guessing a flag risks breaking
// the test command outright, so tier (iii) preserves current behavior.
func nodeNonWatchFlag(script string) string {
	if !nodeScriptWatchProne(script) {
		return ""
	}
	for _, tok := range strings.Fields(script) {
		if strings.Contains(tok, "vitest") {
			return "--run"
		}
	}
	for _, tok := range strings.Fields(script) {
		if strings.Contains(tok, "jest") {
			return "--ci"
		}
	}
	return ""
}

// nodeScriptWatchProne reports whether a test script will not self-terminate:
// (a) it carries a `--watch` or `--watchAll` token (exact match, so an already
// negated `--watch=false` stays non-watch-prone), or (b) it is a bare vitest
// invocation — first token `vitest` without a `run` subcommand, vitest's
// default being watch mode. Bare jest runs once by default and is NOT
// watch-prone (REQ-HGT-002 defines the predicate).
func nodeScriptWatchProne(script string) bool {
	fields := strings.Fields(script)
	if len(fields) == 0 {
		return false
	}
	for _, tok := range fields {
		if tok == "--watch" || tok == "--watchAll" {
			return true
		}
	}
	if fields[0] != "vitest" {
		return false
	}
	for _, tok := range fields[1:] {
		if tok == "run" {
			return false
		}
	}
	return true
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
		// REQ-HCWA-007: route cwd resolution through resolveQualityProjectDir.
		dir := resolveQualityProjectDir(*g.config, "QualityGate.executeStep.extfilter")
		staged := g.cachedStagedFiles(ctx, dir)
		// If staged is nil, cannot determine — run step conservatively
		if staged != nil && !hasStagedExt(staged, step.changedExts) {
			return true, ""
		}
	}

	// Skip when the project holds no source this step could check. A scaffold
	// that has declared its language but not written code yet must not fail
	// its first gate for having no code.
	if len(step.sourceExts) > 0 {
		dir := resolveQualityProjectDir(*g.config, "QualityGate.executeStep.srcfilter")
		if found, determined := projectHasSourceFile(dir, step.sourceExts); determined && !found {
			slog.Warn("no sources for step: treating as skip",
				"step", step.name,
				"extensions", strings.Join(step.sourceExts, ","),
				"hint", "add sources, or set gate.disabled_steps in .moai/config/sections/gate.yaml",
			)
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

// sourceScanEntryCap bounds projectHasSourceFile's walk. A source-free tree is
// established by exhausting the walk, so on a very large repository that answer
// could cost a full traversal. Past the cap the scan gives up and reports
// undetermined, which runs the step — the conservative direction, and the same
// one taken when the directory cannot be read at all.
const sourceScanEntryCap = 20000

// sourceScanSkipDirs are directory names never worth descending into when
// asking whether a project has sources of its own: dependency trees, build
// output, and caches. A .py under node_modules or .venv is somebody else's.
var sourceScanSkipDirs = map[string]struct{}{
	".git": {}, ".hg": {}, ".svn": {},
	"node_modules": {}, "vendor": {},
	".venv": {}, "venv": {}, "site-packages": {}, "__pycache__": {},
	".tox": {}, ".nox": {}, ".mypy_cache": {}, ".ruff_cache": {}, ".pytest_cache": {},
	"dist": {}, "build": {}, "target": {}, ".next": {}, ".output": {},
}

// projectHasSourceFile reports whether dir's tree holds a file with one of the
// given extensions, stopping at the first match.
//
// The second return says whether the question was answered at all. An empty
// dir, an unreadable tree, or a walk that outgrows sourceScanEntryCap answers
// false — the caller then runs the step rather than skipping it, so an
// uncertain scan never suppresses a check.
func projectHasSourceFile(dir string, exts []string) (found bool, determined bool) {
	if dir == "" || len(exts) == 0 {
		return false, false
	}

	visited := 0
	errStop := errors.New("stop")
	errBudget := errors.New("budget")

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// The root failing is the whole question failing — a missing or
			// unreadable project directory must answer undetermined, so the
			// caller runs the step rather than skipping it.
			if path == dir {
				return err
			}
			// A permission hole somewhere below is skipped instead: it must
			// not decide the whole answer.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		visited++
		if visited > sourceScanEntryCap {
			return errBudget
		}
		if d.IsDir() {
			if path == dir {
				return nil
			}
			if _, skip := sourceScanSkipDirs[d.Name()]; skip {
				return fs.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(d.Name())
		for _, want := range exts {
			if strings.EqualFold(ext, want) {
				found = true
				return errStop
			}
		}
		return nil
	})

	switch {
	case errors.Is(walkErr, errStop):
		return true, true
	case walkErr == nil:
		return false, true
	default:
		// errBudget, or the root itself being unreadable.
		return false, false
	}
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
	// REQ-HCWA-007: route cwd resolution through resolveQualityProjectDir.
	dir := resolveQualityProjectDir(*g.config, "QualityGate.anyConfigFileExists")
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
	// Which budget can kill this step is decided BEFORE it starts, by comparing
	// the caller's remaining time (the hook dispatcher's, when the gate runs
	// from a hook) with the step's own. Deciding afterwards from ctx.Err() would
	// misattribute the narrow case where the step deadline fires first and the
	// parent deadline passes while the process is still shutting down — both
	// contexts then read DeadlineExceeded.
	parentBudget, parentBinds := time.Duration(0), false
	if deadline, ok := ctx.Deadline(); ok {
		parentBudget = time.Until(deadline)
		parentBinds = parentBudget <= timeout
	}

	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(stepCtx, name, args...)
	// Every step's arguments are cwd-relative ("./...", ".", a bare "test").
	// Without an explicit Dir the child inherits the calling process's cwd,
	// which is the project root only by coincidence — so a gate configured for
	// one directory would grade another. Under `go test` that coincidence
	// breaks: the cwd is the package under test, so a "go test ./..." step
	// re-executes the suite that invoked it.
	if dir := resolveQualityProjectDir(*g.config, "QualityGate.runStep"); dir != "" {
		cmd.Dir = dir
	}
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

	// pytest exits 5 for "no tests collected" (EXIT_NOTESTSCOLLECTED). Nothing is
	// wrong with the code — the suite is simply empty — so a project that has not
	// written its first test must not be blocked from committing (issue #1265).
	// Scoped to the pytest step: exit 5 carries no such meaning for other tools.
	if stepName == pytestStepName && exitCodeOf(err) == pytestNoTestsCollected {
		slog.Warn("pytest collected no tests: treating as pass",
			"step", stepName,
			"hint", "add tests, or set gate.skip_tests / gate.disabled_steps in .moai/config/sections/gate.yaml",
		)
		return true, ""
	}

	// Distinguish timeout from other failures (REQ-GATE-009).
	if stepCtx.Err() == context.DeadlineExceeded {
		// A parent cancellation propagates to stepCtx, so DeadlineExceeded here
		// says nothing about WHICH budget ran out. Blaming the step's own budget
		// unconditionally produced impossible reasons — a 30s dispatcher budget
		// expiring mid-`go test` reported "go test exceeded 2m0s" (card t218).
		if parentBinds {
			return false, fmt.Sprintf(
				"quality gate timed out: the overall gate budget ran out while running %s (%s of it remained when the step started; the step's own %s budget was not exceeded)",
				stepName, parentBudget.Round(time.Millisecond), timeout)
		}
		return false, fmt.Sprintf("quality gate timed out: %s exceeded %s", stepName, timeout)
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

// pytestStepName is the gate step name whose exit codes carry pytest semantics.
const pytestStepName = "pytest"

// pytestNoTestsCollected is pytest's documented EXIT_NOTESTSCOLLECTED status.
// It means the run found no tests, not that anything failed.
const pytestNoTestsCollected = 5

// exitCodeOf extracts a process exit status from a command error.
// Returns -1 when err is not an exit error (timeout, binary not found, ...).
func exitCodeOf(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
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
