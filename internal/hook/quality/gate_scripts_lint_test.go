package quality

// Node scripts.lint coverage (issue #1631 residual, card t687).
//
// biome (t233) and oxlint (SPEC-GATE-OXLINT-DETECT-001) closed the "gate
// runs no linter at all" half of #1631. This card closes the reporter's
// first preference: a project that declares scripts.lint in its package.json
// has told the gate exactly which lint command to run, and the gate never
// looked at it. The config-gated eslint/biome/oxlint entries stay for
// projects without the script, unchanged.
//
// Precedence (the design judgment the card asks to pin): when scripts.lint
// is present it REPLACES the config-gated entries for that toolchain — the
// project's own command outranks the gate's guesses, and running both would
// lint the tree twice. The replaced entries stay visible in the run summary
// as skips, so nothing trades silence for the substitution. A watch-prone
// lint script never self-terminates; it is reported, not run — a hang to the
// lint timeout would be a false red on every commit (the same visible-skip
// judgment the source-free and config-free skips carry).

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// prependFakes puts fake binary directories ahead of the real PATH. Tests
// calling it must not use t.Parallel (t.Setenv).
func prependFakes(t *testing.T, dirs ...string) {
	t.Helper()

	joined := strings.Join(dirs, string(os.PathListSeparator)) + string(os.PathListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", joined)
}

// TestScriptsLintViolationFailsGate — the issue's reported symptom, inverted
// into a regression pin: a project whose own lint command fails must fail the
// gate with the step and the diagnostic named. Before this card the gate ran
// no lint at all for such a project and passed.
func TestScriptsLintViolationFailsGate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "scripts-lint", "scripts": {"lint": "exit 1"}}`,
	})
	npmDir := writeFakeBinary(t, "npm", 1, "scripts.lint probe violation")
	prependFakes(t, npmDir)

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("gate passed a project whose own scripts.lint fails — the declared lint command never ran; summary:\n%s", out)
	}
	if !strings.Contains(out, "quality gate failed: npm run lint") {
		t.Errorf("failure output does not name the npm run lint step:\n%s", out)
	}
	if !strings.Contains(out, "scripts.lint probe violation") {
		t.Errorf("failure output does not carry the lint command's diagnostic:\n%s", out)
	}
	rec := g.summary.recordFor("npm run lint")
	if rec == nil || rec.outcome != outcomeExecuted {
		t.Fatalf("npm run lint row = %v, want an executed row", rec)
	}
	if rec.command != "npm run lint" {
		t.Errorf("npm run lint command = %q, want %q", rec.command, "npm run lint")
	}
}

// TestScriptsLintCleanProjectPasses — the green half of the same pair: a
// clean scripts.lint run passes, with the step visible as executed (not one
// of the config-gated entries quietly skipping).
func TestScriptsLintCleanProjectPasses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "scripts-lint", "scripts": {"lint": "exit 0"}}`,
	})
	npmDir := writeFakeBinary(t, "npm", 0, "")
	prependFakes(t, npmDir)

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed a clean scripts.lint project:\n%s", out)
	}
	rec := g.summary.recordFor("npm run lint")
	if rec == nil || rec.outcome != outcomeExecuted {
		t.Fatalf("npm run lint row = %v, want an executed row; summary:\n%s", rec, out)
	}
}

// TestScriptsLintSupersedesConfigEntries — the precedence pin: with
// scripts.lint present, the config-gated entries (biome here) do not also
// run. The fake npx exits 1, so a gate that still ran biome would fail.
func TestScriptsLintSupersedesConfigEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "both", "scripts": {"lint": "exit 0"}}`,
		"biome.json":   `{}`,
	})
	npmDir := writeFakeBinary(t, "npm", 0, "")
	npxDir := writeFakeBinary(t, "npx", 1, "biome must not run beside scripts.lint")
	prependFakes(t, npmDir, npxDir)

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed — biome ran beside the project's own scripts.lint:\n%s", out)
	}
	if rec := g.summary.recordFor("biome"); rec != nil {
		t.Errorf("biome row = %v, want no row (superseded entries are not seeded)", rec)
	}
	if rec := g.summary.recordFor("npm run lint"); rec == nil || rec.outcome != outcomeExecuted {
		t.Errorf("npm run lint row = %v, want an executed row; summary:\n%s", rec, out)
	}
}

// TestWatchProneLintScriptSkippedVisibly — the watch defense: a lint script
// carrying --watch never self-terminates, so it is reported as a skip, not
// run (the fake npm exits 1, so a gate that ran it would fail). The replaced
// entries carry the watch-prone reason, keeping the substitution visible.
func TestWatchProneLintScriptSkippedVisibly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "watcher", "scripts": {"lint": "eslint . --watch"}}`,
	})
	npmDir := writeFakeBinary(t, "npm", 1, "npm must not run a watch-prone lint script")
	prependFakes(t, npmDir)

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed — the watch-prone lint script ran instead of being reported:\n%s", out)
	}
	if rec := g.summary.recordFor("npm run lint"); rec != nil {
		t.Errorf("npm run lint row = %v, want no row (watch-prone scripts are not run)", rec)
	}
	eslint := g.summary.recordFor("eslint")
	if eslint == nil {
		t.Fatalf("eslint row disappeared — the superseded entries must stay visible")
	}
	if !strings.Contains(eslint.reason, "watch-prone") {
		t.Errorf("eslint skip reason = %q, want the watch-prone notice", eslint.reason)
	}
}

// TestNoScriptsLintLeavesLintAxisUnchanged — card regression: projects
// without scripts.lint keep the exact config-gated behavior (biome runs on
// its config; no npm run lint row appears).
func TestNoScriptsLintLeavesLintAxisUnchanged(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "biome-project", "scripts": {"test:run": "vitest run"}}`,
		"biome.json":   `{}`,
	})
	npxDir := writeFakeBinary(t, "npx", 0, "")
	prependFakes(t, npxDir)

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a plain biome project:\n%s", out)
	}
	rec := g.summary.recordFor("biome")
	if rec == nil || rec.outcome != outcomeExecuted {
		t.Fatalf("biome row = %v, want an executed row — the config-gated axis must be untouched; summary:\n%s", rec, out)
	}
	if rec := g.summary.recordFor("npm run lint"); rec != nil {
		t.Errorf("npm run lint row = %v, want no row (no scripts.lint declared)", rec)
	}
}

// TestRealNpmScriptsLintViolationFailsGate — end-to-end with the real npm:
// npm itself executes the declared script and its failure verdict reaches
// the gate. Proves the fake-binary tests stand for the real npm run path.
func TestRealNpmScriptsLintViolationFailsGate(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not on PATH — the real-npm path cannot run")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "real-npm-lint", "scripts": {"lint": "exit 1"}}`,
	})

	g := biomeGate(t, dir)
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("gate passed a project whose real npm run lint exits 1; summary:\n%s", out)
	}
	if !strings.Contains(out, "npm run lint") {
		t.Errorf("failure output does not name the npm run lint step:\n%s", out)
	}
}
