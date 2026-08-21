package quality

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeFixture lays out a project directory and returns its path.
func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// Cross-platform stand-ins for the `true` / `false` utilities.
//
// Windows ships neither, and the typecheck step is optional, so a LookPath miss
// SKIPS the step rather than failing it — a Windows run would have seen the gate
// pass and failed the blocking assertion for the wrong reason. `go` is present
// wherever these tests run, so its own exit codes carry the signal instead.
const (
	cmdExitZero    = "go version"
	cmdExitNonZero = "go t172-not-a-subcommand"
)

func typecheckConfig(dir string) *GateConfig {
	cfg := DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.SkipTests = true
	cfg.AstGrepGate = nil
	return cfg
}

// --- Tier selection -------------------------------------------------------

// TestTypecheckTierScriptWins — tier (b): package.json scripts.typecheck.
func TestTypecheckTierScriptWins(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json":  `{"scripts":{"typecheck":"turbo run typecheck"}}`,
		"tsconfig.json": `{"compilerOptions":{"strict":true}}`,
	})

	step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if !ok {
		t.Fatalf("resolution skipped (%s); want the script tier to fire", reason)
	}
	if step.binary != "npm" {
		t.Errorf("binary = %q, want npm", step.binary)
	}
	if got := strings.Join(step.args, " "); got != "run typecheck" {
		t.Errorf("args = %q, want %q", got, "run typecheck")
	}
}

// TestTypecheckTierTsconfigFallback — tier (c): tsconfig.json, no script.
func TestTypecheckTierTsconfigFallback(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json":  `{"scripts":{"build":"tsc"}}`,
		"tsconfig.json": `{"compilerOptions":{"strict":true}}`,
	})

	step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if !ok {
		t.Fatalf("resolution skipped (%s); want the tsconfig tier to fire", reason)
	}
	if step.binary != "npx" {
		t.Errorf("binary = %q, want npx", step.binary)
	}
	if got := strings.Join(step.args, " "); got != "--no-install tsc --noEmit" {
		t.Errorf("args = %q, want %q", got, "--no-install tsc --noEmit")
	}
}

// TestTypecheckSkipsWithoutTsconfigOrScript asserts the skip carries a reason.
//
// A silent skip here is exactly the shape of the incident this axis exists to
// repair: a missing config made the gate pass while the build was broken.
func TestTypecheckSkipsWithoutTsconfigOrScript(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json": `{"scripts":{"build":"vite build"}}`,
	})

	_, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if ok {
		t.Fatal("resolution fired with neither a typecheck script nor a tsconfig")
	}
	if strings.TrimSpace(reason) == "" {
		t.Fatal("skip carried no reason")
	}
	if !strings.Contains(reason, "tsconfig") {
		t.Errorf("reason %q does not say what was missing", reason)
	}
}

// TestSolutionStyleTsconfigIsRefused — a solution-style tsconfig (files: [] plus
// references only) type-checks nothing and exits 0, so accepting it would
// reinstate the vacuous pass this card repairs.
func TestSolutionStyleTsconfigIsRefused(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json":  `{"scripts":{"build":"tsc -b"}}`,
		"tsconfig.json": `{"files":[],"references":[{"path":"./packages/a"}]}`,
	})

	_, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if ok {
		t.Fatal("solution-style tsconfig accepted; it passes vacuously")
	}
	if !strings.Contains(reason, "solution-style") {
		t.Errorf("reason %q does not name the solution-style case", reason)
	}
}

// TestSolutionStyleWithScriptStillRuns — the script tier outranks the tsconfig
// shape, so a monorepo root delegating to turbo is not penalised.
func TestSolutionStyleWithScriptStillRuns(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json":  `{"scripts":{"typecheck":"turbo run typecheck"}}`,
		"tsconfig.json": `{"files":[],"references":[{"path":"./packages/a"}]}`,
	})

	step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
	if !ok {
		t.Fatalf("skipped (%s); the script tier must outrank the tsconfig shape", reason)
	}
	if got := strings.Join(step.args, " "); got != "run typecheck" {
		t.Errorf("args = %q, want the script tier", got)
	}
}

// TestTypecheckOverrideForcesAnyLanguage — tier (a).
func TestTypecheckOverrideForcesAnyLanguage(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})

	step, reason, ok := resolveTypecheckStep(nil, dir, "mypy --strict .")
	if !ok {
		t.Fatalf("override skipped (%s); it must fire for any language", reason)
	}
	if step.binary != "mypy" {
		t.Errorf("binary = %q, want mypy", step.binary)
	}
	if got := strings.Join(step.args, " "); got != "--strict ." {
		t.Errorf("args = %q, want %q", got, "--strict .")
	}
}

// TestTypecheckAbsentForLanguageReportsReason — a language with no default
// says so rather than passing quietly.
func TestTypecheckAbsentForLanguageReportsReason(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})

	_, reason, ok := resolveTypecheckStep(nil, dir, "")
	if ok {
		t.Fatal("a language without a typecheck default produced a step")
	}
	if !strings.Contains(reason, "gate.typecheck.command") {
		t.Errorf("reason %q does not tell the operator how to supply one", reason)
	}
}

// --- Gate integration -----------------------------------------------------

// TestGateBlocksOnTypecheckFailure is the headline RED: a failing typecheck
// must stop the gate, which is precisely what did not happen in the incident.
func TestGateBlocksOnTypecheckFailure(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})
	cfg := typecheckConfig(dir)
	cfg.TypecheckCommand = cmdExitNonZero
	cfg.SkipTests = true

	passed, out := NewQualityGate(cfg).Run(context.Background())
	if passed {
		t.Fatalf("gate passed with a failing typecheck; output = %q", out)
	}
	if !strings.Contains(strings.ToLower(out), "typecheck") {
		t.Errorf("failure output %q does not name the typecheck step", out)
	}
}

// TestGatePassesWhenTypecheckSucceeds is the GREEN counterpart — without it,
// the block above could be satisfied by a gate that always fails.
func TestGatePassesWhenTypecheckSucceeds(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})
	cfg := typecheckConfig(dir)
	cfg.TypecheckCommand = cmdExitZero

	passed, out := NewQualityGate(cfg).Run(context.Background())
	if !passed {
		t.Fatalf("gate failed with a passing typecheck; output = %q", out)
	}
}

// TestTypecheckDisabledByConfig — typecheck.enabled false.
func TestTypecheckDisabledByConfig(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})
	cfg := typecheckConfig(dir)
	cfg.TypecheckEnabled = false
	cfg.TypecheckCommand = cmdExitNonZero

	if passed, out := NewQualityGate(cfg).Run(context.Background()); !passed {
		t.Fatalf("disabled typecheck still blocked the gate: %q", out)
	}
}

// TestTypecheckDisabledByDisabledSteps — the step-name key already in use for
// every other step works for this axis too.
func TestTypecheckDisabledByDisabledSteps(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})
	cfg := typecheckConfig(dir)
	cfg.TypecheckCommand = cmdExitNonZero
	cfg.DisabledSteps = map[string]bool{typecheckStepName: false}

	if passed, out := NewQualityGate(cfg).Run(context.Background()); !passed {
		t.Fatalf("disabled_steps did not suppress the typecheck step: %q", out)
	}
}

// TestTypecheckSkipIsReportedNotSilent — the axis's whole point.
func TestTypecheckSkipIsReportedNotSilent(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{"go.mod": "module x\n"})
	cfg := typecheckConfig(dir)

	passed, out := NewQualityGate(cfg).Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a skipped typecheck; skipping is not failing: %q", out)
	}
	if !strings.Contains(out, typecheckStepName) {
		t.Fatalf("output %q does not report the skip — a silent skip is the defect being repaired", out)
	}
}

// TestTypecheckDefaultsAreProductionSafe pins the shipped defaults.
func TestTypecheckDefaultsAreProductionSafe(t *testing.T) {
	t.Parallel()

	cfg := DefaultGateConfig()
	if !cfg.TypecheckEnabled {
		t.Error("TypecheckEnabled default = false, want true")
	}
	if cfg.TypecheckTimeout != 300*time.Second {
		t.Errorf("TypecheckTimeout default = %s, want 300s", cfg.TypecheckTimeout)
	}
	if cfg.TypecheckCommand != "" {
		t.Errorf("TypecheckCommand default = %q, want empty", cfg.TypecheckCommand)
	}
}

// TestNodeToolchainCarriesTypecheckStep asserts the wiring exists where the
// incident happened: the Node entry previously had no vet axis at all.
func TestNodeToolchainCarriesTypecheckStep(t *testing.T) {
	t.Parallel()

	for _, tc := range toolchains {
		for _, marker := range tc.markerFiles {
			if marker != "package.json" {
				continue
			}
			if tc.typecheckStep == nil {
				t.Fatal("the Node toolchain has no typecheckStep; Node was the untyped hole")
			}
			return
		}
	}
	t.Fatal("no Node toolchain entry found")
}
