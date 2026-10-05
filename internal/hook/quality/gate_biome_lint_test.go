package quality

// gate_biome_lint_test.go — Node toolchain biome lint coverage (issue #1631, card t233).
//
// The Node entry's lintSteps carried exactly one element (eslint) gated on an
// eslint config file. A project using biome has no such file, so the lint axis
// skipped and the gate reported a pass on a tree with live lint errors —
// `npm run lint` exited 1 while `moai gate` exited 0. The skip itself is no
// longer silent on this tree (the run summary's config-files-absent reason,
// pinned by TestSummaryDistinguishesAllFiveSkipPaths); what was still missing
// is a biome entry, so the lint axis actually runs on a biome project.
//
// These tests pin that: a project carrying biome.json gets a biome lint step
// that executes, that a red biome run fails the gate, and that a project with
// no linter config at all still reports the eslint skip as a visible notice
// rather than silence.

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
)

// biomeGate builds a gate whose only live axis is lint: vet is absent from the
// Node toolchain, typecheck and ast-grep are turned off, and tests are skipped.
// The fake npx prepended by the caller keeps the biome step from ever reaching
// a real npx (which would try to download biome from the network).
func biomeGate(t *testing.T, dir string) *QualityGate {
	t.Helper()

	cfg := DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.AstGrepGate = nil
	cfg.GraphFreshness = nil
	cfg.SkipTests = true
	cfg.TypecheckEnabled = false
	return NewQualityGate(cfg)
}

// prependFakeNPX shadows npx with a fake binary that exits exitCode after
// printing output on stderr, and returns nothing. Tests calling it must not
// use t.Parallel (t.Setenv).
func prependFakeNPX(t *testing.T, exitCode int, output string) {
	t.Helper()

	fakeBinDir := writeFakeBinary(t, "npx", exitCode, output)
	t.Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// TestNodeLintRunsBiomeStepOnBiomeProject — issue #1631: a project whose
// linter is biome (biome.json present, no eslint config anywhere) must get a
// biome lint step that actually executes, with the biome command line
// recorded in the run summary.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintRunsBiomeStepOnBiomeProject(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "biome-project", "scripts": {"test:run": "vitest run"}}`,
		"biome.json":   `{}`,
	})
	prependFakeNPX(t, 0, "")

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a clean biome project: %s", out)
	}

	rec := g.summary.recordFor("biome")
	if rec == nil {
		t.Fatalf("no biome row in the run summary — the Node toolchain has no biome lint step; summary:\n%s", out)
	}
	if rec.outcome != outcomeExecuted {
		t.Errorf("biome row outcome = %q (reason %q), want %q", rec.outcome, rec.reason, outcomeExecuted)
	}
	if rec.command != "npx biome check ." {
		t.Errorf("biome row command = %q, want %q", rec.command, "npx biome check .")
	}

	// The eslint step must still skip with its config-absent notice — biome
	// coverage is additive, not a replacement for the eslint entry.
	eslint := g.summary.recordFor("eslint")
	if eslint == nil {
		t.Fatalf("eslint row disappeared from the run summary")
	}
	if eslint.outcome != outcomeSkipped {
		t.Errorf("eslint outcome = %q, want %q (no eslint config in this fixture)", eslint.outcome, outcomeSkipped)
	}
}

// TestNodeLintBiomeViolationFailsGate — issue #1631's reported symptom: a
// biome project with a live lint error (`npm run lint` exit 1) must fail the
// gate, not pass it. The fake npx exits 1 to stand in for the red lint run.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintBiomeViolationFailsGate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "biome-project", "scripts": {"test:run": "vitest run"}}`,
		"biome.json":   `{}`,
	})
	prependFakeNPX(t, 1, "src/probe.ts:1:1 lint/suspicious -- probe violation")

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("gate passed a red biome run — the lint axis never looked at biome; summary:\n%s", out)
	}
	if !strings.Contains(out, "quality gate failed: biome") {
		t.Errorf("failure output does not name the biome step: %q", out)
	}
}

// TestNodeLintBiomeStepShape pins the biome entry against the package-level
// toolchains table: the config gate must recognize both biome.json and
// biome.jsonc, so a project using the commented-JSON form is covered too.
func TestNodeLintBiomeStepShape(t *testing.T) {
	tc := nodeToolchain(t)

	var biome *gateStep
	for i := range tc.lintSteps {
		if tc.lintSteps[i].name == "biome" {
			biome = &tc.lintSteps[i]
			break
		}
	}
	if biome == nil {
		t.Fatalf("Node toolchain lintSteps has no biome entry: %+v", tc.lintSteps)
	}
	if biome.binary != "npx" {
		t.Errorf("biome binary = %q, want %q", biome.binary, "npx")
	}
	if got := strings.Join(biome.args, " "); got != "biome check ." {
		t.Errorf("biome args = %q, want %q", got, "biome check .")
	}
	for _, want := range []string{"biome.json", "biome.jsonc"} {
		found := false
		for _, cf := range biome.configFiles {
			if cf == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("biome configFiles %v lacks %q", biome.configFiles, want)
		}
	}
}

// TestNodeLintNoLinterConfigSkipIsVisible pins the notice half of issue #1631:
// a Node project with no linter config at all must not pass silently — the
// eslint skip reaches Run's caller through the run summary, keeping a skipped
// lint axis distinguishable from a clean one. Expected to hold on this tree
// already (the summary mechanism predates this card); this pins it against
// regression.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintNoLinterConfigSkipIsVisible(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "bare-project"}`,
	})
	// npx present (so the optional-binary guard passes) but never executed:
	// both lint steps skip on config before any command runs.
	prependFakeNPX(t, 0, "")

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a config-free project: %s", out)
	}
	for _, want := range []string{"eslint", "skipped", "config file"} {
		if !strings.Contains(out, want) {
			t.Errorf("pass output lacks %q — a skipped lint axis is silent again:\n%s", want, out)
		}
	}
}
