package quality

// gate_oxlint_lint_test.go — Node toolchain oxlint lint coverage
// (issue #1631, card t550, SPEC-GATE-OXLINT-DETECT-001).
//
// The Node entry's lintSteps carried two elements (eslint, biome), each gated
// on its own config files. A project whose linter is oxlint matches neither,
// so its lint axis ran zero steps while the gate reported a pass. Measured on
// tree d060e0d13: all four of oxlint's own config filenames produced 0 lint
// steps (.moai/reports/t550/baseline.md).
//
// These tests pin the fix: an oxlint project gets an oxlint step that
// executes, a red oxlint run fails the gate, each of the four documented
// config filenames selects the step on its own, the entry's declared shape
// matches oxlint's discovery list exactly, and the eslint / biome / no-linter
// controls stay unchanged.

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

// oxlintConfigFilenames is oxlint's own documented config discovery set, per
// https://oxc.rs/docs/guide/usage/linter/config.html (accessed 2026-09-10).
// The set is CLOSED: REQ-002 states "No other filename is recognized", so
// AC-004 asserts set equality against it rather than containment.
var oxlintConfigFilenames = []string{
	".oxlintrc.json",
	".oxlintrc.jsonc",
	"oxlint.config.ts",
	"oxlint.config.mts",
}

// oxlintGate builds a gate whose only live axis is lint, mirroring biomeGate:
// vet is absent from the Node toolchain, typecheck and ast-grep are turned
// off, and tests are skipped. The fake npx prepended by the caller keeps the
// oxlint step from ever reaching a real npx (which would fetch oxlint over the
// network).
func oxlintGate(t *testing.T, dir string) *QualityGate {
	t.Helper()

	cfg := DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.AstGrepGate = nil
	cfg.GraphFreshness = nil
	cfg.SkipTests = true
	cfg.TypecheckEnabled = false
	return NewQualityGate(cfg)
}

// containsString reports whether haystack holds needle.
func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// AC-001 — TestNodeLintRunsOxlintStepOnOxlintProject: a project whose linter
// is oxlint (.oxlintrc.json present, no eslint or biome config) must get an
// oxlint lint step that actually executes, with `npx oxlint` recorded in the
// run summary.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintRunsOxlintStepOnOxlintProject(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json":   `{"name": "oxlint-project", "scripts": {"test:run": "vitest run"}}`,
		".oxlintrc.json": `{}`,
	})
	prependFakeNPX(t, 0, "")

	g := oxlintGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a clean oxlint project: %s", out)
	}

	rec := g.summary.recordFor("oxlint")
	if rec == nil {
		t.Fatalf("no oxlint row in the run summary — the Node toolchain has no oxlint lint step; summary:\n%s", out)
	}
	if rec.outcome != outcomeExecuted {
		t.Errorf("oxlint row outcome = %q (reason %q), want %q", rec.outcome, rec.reason, outcomeExecuted)
	}
	if rec.command != "npx oxlint" {
		t.Errorf("oxlint row command = %q, want %q", rec.command, "npx oxlint")
	}
}

// AC-002 — TestNodeLintOxlintViolationFailsGate: issue #1631's reported
// symptom. An oxlint project with a live lint error must fail the gate, and
// the failure output must name the oxlint step.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintOxlintViolationFailsGate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json":   `{"name": "oxlint-project", "scripts": {"test:run": "vitest run"}}`,
		".oxlintrc.json": `{}`,
	})
	prependFakeNPX(t, 1, "src/probe.ts:1:1 eslint(no-unused-vars): probe violation")

	g := oxlintGate(t, dir)
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("gate passed a red oxlint run — the lint axis never looked at oxlint; summary:\n%s", out)
	}
	if !strings.Contains(out, "quality gate failed: oxlint") {
		t.Errorf("failure output does not name the oxlint step: %q", out)
	}
}

// AC-003 — TestNodeLintOxlintConfigFilenames: each of the four documented
// config filenames selects the oxlint step, measured one fixture per name.
// A single fixture would measure one name and leave three asserted but
// unmeasured, so a typo in any of the three would pass unseen behind the
// fourth. The subtest names carry the filename so the run reports which case
// was exercised — four cases, not three.
//
// The oxlint.config.ts / oxlint.config.mts cases double as a mis-claim
// control: the eslint entry's own list carries eslint.config.ts /
// eslint.config.mts, which differ only in prefix, so eslint must still skip.
// Does not call t.Parallel() because subtests use t.Setenv.
func TestNodeLintOxlintConfigFilenames(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	for _, filename := range oxlintConfigFilenames {
		t.Run("oxlint_config/"+filename, func(t *testing.T) {
			dir := writeFixture(t, map[string]string{
				"package.json": `{"name": "oxlint-project"}`,
				filename:       `{}`,
			})
			prependFakeNPX(t, 0, "")

			g := oxlintGate(t, dir)
			passed, out := g.Run(context.Background())
			if !passed {
				t.Fatalf("gate failed on a clean oxlint project carrying %s: %s", filename, out)
			}

			rec := g.summary.recordFor("oxlint")
			if rec == nil {
				t.Fatalf("no oxlint row for config file %s; summary:\n%s", filename, out)
			}
			if rec.outcome != outcomeExecuted {
				t.Errorf("config file %s: oxlint outcome = %q (reason %q), want %q",
					filename, rec.outcome, rec.reason, outcomeExecuted)
			}

			// Mis-claim control: eslint's list carries eslint.config.ts /
			// eslint.config.mts, one prefix away from two of these names.
			eslint := g.summary.recordFor("eslint")
			if eslint == nil {
				t.Fatalf("config file %s: eslint row disappeared from the run summary", filename)
			}
			if eslint.outcome != outcomeSkipped {
				t.Errorf("config file %s: eslint outcome = %q, want %q (no eslint config in this fixture)",
					filename, eslint.outcome, outcomeSkipped)
			}
		})
	}
}

// AC-004 — TestNodeLintOxlintStepShape pins the oxlint entry against the
// package-level toolchains table. configFiles is asserted set-EQUAL to
// oxlintConfigFilenames, in both directions with distinguishable messages:
// REQ-002 states a CLOSED norm ("No other filename is recognized"), and
// containment would constrain only the shrinking direction — a fifth name
// added to the list would leave every behavioural AC green while violating
// that norm.
func TestNodeLintOxlintStepShape(t *testing.T) {
	tc := nodeToolchain(t)

	var oxlint *gateStep
	for i := range tc.lintSteps {
		if tc.lintSteps[i].name == "oxlint" {
			oxlint = &tc.lintSteps[i]
			break
		}
	}
	if oxlint == nil {
		t.Fatalf("Node toolchain lintSteps has no oxlint entry: %+v", tc.lintSteps)
	}
	if oxlint.binary != "npx" {
		t.Errorf("oxlint binary = %q, want %q", oxlint.binary, "npx")
	}
	if got := strings.Join(oxlint.args, " "); got != "oxlint" {
		t.Errorf("oxlint args = %q, want %q", got, "oxlint")
	}
	if !oxlint.optional {
		t.Errorf("oxlint optional = false, want true (an absent npx must not fail the gate)")
	}

	if len(oxlint.configFiles) != len(oxlintConfigFilenames) {
		t.Errorf("oxlint configFiles has %d entries, want exactly %d: %v",
			len(oxlint.configFiles), len(oxlintConfigFilenames), oxlint.configFiles)
	}
	for _, want := range oxlintConfigFilenames {
		if !containsString(oxlint.configFiles, want) {
			t.Errorf("missing config file name: %s (declared: %v)", want, oxlint.configFiles)
		}
	}
	for _, got := range oxlint.configFiles {
		if !containsString(oxlintConfigFilenames, got) {
			t.Errorf("unexpected config file name: %s (oxlint does not read it; declared: %v)",
				got, oxlint.configFiles)
		}
	}
}

// AC-005 — CONTROL: an eslint project is unchanged. eslint executes; biome
// AND oxlint both skip on absent config. If oxlint executes here, the entry
// is not config-gated and the change has become "run every linter everywhere".
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintEslintProjectUnaffectedByOxlint(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json":     `{"name": "eslint-project"}`,
		"eslint.config.js": `export default [];`,
	})
	prependFakeNPX(t, 0, "")

	g := oxlintGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a clean eslint project: %s", out)
	}

	assertLintOutcome(t, g, "eslint", outcomeExecuted, out)
	assertLintCommand(t, g, "eslint", "npx eslint .")
	assertConfigAbsentSkip(t, g, "biome", out)
	assertConfigAbsentSkip(t, g, "oxlint", out)
	if n := executedLintCount(t, g); n != 1 {
		t.Errorf("executed lint steps = %d, want 1; summary:\n%s", n, out)
	}
}

// AC-006 — CONTROL: a biome project is unchanged. biome executes; eslint AND
// oxlint both skip on absent config.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintBiomeProjectUnaffectedByOxlint(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "biome-project"}`,
		"biome.json":   `{}`,
	})
	prependFakeNPX(t, 0, "")

	g := oxlintGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a clean biome project: %s", out)
	}

	assertLintOutcome(t, g, "biome", outcomeExecuted, out)
	assertLintCommand(t, g, "biome", "npx biome check .")
	assertConfigAbsentSkip(t, g, "eslint", out)
	assertConfigAbsentSkip(t, g, "oxlint", out)
	if n := executedLintCount(t, g); n != 1 {
		t.Errorf("executed lint steps = %d, want 1; summary:\n%s", n, out)
	}
}

// AC-007 — CONTROL: a linter-free Node scaffold still passes, runs zero lint
// steps, and emits a visible config-absent notice for each of the three lint
// entries — three notices, up from the two on the baseline tree. A failure
// here means the change started blocking first commits and scaffolds.
// Does not call t.Parallel() because it uses t.Setenv.
func TestNodeLintLinterFreeScaffoldStillPasses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script fake binary is not directly executable")
	}

	dir := writeFixture(t, map[string]string{
		"package.json": `{"name": "bare-project"}`,
	})
	// npx present (so the optional-binary guard passes) but never executed:
	// every lint step skips on config before any command runs.
	prependFakeNPX(t, 0, "")

	g := oxlintGate(t, dir)
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on a config-free project: %s", out)
	}

	if n := executedLintCount(t, g); n != 0 {
		t.Errorf("executed lint steps = %d on a linter-free scaffold, want 0; summary:\n%s", n, out)
	}

	notices := 0
	for _, label := range []string{"eslint", "biome", "oxlint"} {
		assertConfigAbsentSkip(t, g, label, out)
		if !strings.Contains(out, label) {
			t.Errorf("pass output lacks a visible notice naming %q:\n%s", label, out)
			continue
		}
		notices++
	}
	if notices != 3 {
		t.Errorf("visible config-absent notices = %d, want 3 (eslint, biome, oxlint)", notices)
	}
}

// assertLintOutcome checks one lint row's outcome.
func assertLintOutcome(t *testing.T, g *QualityGate, label string, want stepOutcome, out string) {
	t.Helper()
	rec := g.summary.recordFor(label)
	if rec == nil {
		t.Fatalf("no %s row in the run summary; summary:\n%s", label, out)
	}
	if rec.outcome != want {
		t.Errorf("%s outcome = %q (reason %q), want %q", label, rec.outcome, rec.reason, want)
	}
}

// assertLintCommand checks one executed lint row's command line.
func assertLintCommand(t *testing.T, g *QualityGate, label, want string) {
	t.Helper()
	rec := g.summary.recordFor(label)
	if rec == nil {
		t.Fatalf("no %s row in the run summary", label)
	}
	if rec.command != want {
		t.Errorf("%s command = %q, want %q", label, rec.command, want)
	}
}

// assertConfigAbsentSkip checks that a lint row skipped for the config-absent
// reason specifically — a skip for any other reason is a different observation.
func assertConfigAbsentSkip(t *testing.T, g *QualityGate, label, out string) {
	t.Helper()
	rec := g.summary.recordFor(label)
	if rec == nil {
		t.Fatalf("no %s row in the run summary; summary:\n%s", label, out)
	}
	if rec.outcome != outcomeSkipped {
		t.Errorf("%s outcome = %q, want %q (no %s config in this fixture)", label, rec.outcome, outcomeSkipped, label)
		return
	}
	if !strings.Contains(rec.reason, "config file") {
		t.Errorf("%s skipped for %q, want the config-files-absent reason", label, rec.reason)
	}
}

// executedLintCount counts the Node lint rows that actually executed.
func executedLintCount(t *testing.T, g *QualityGate) int {
	t.Helper()
	tc := nodeToolchain(t)
	n := 0
	for i := range tc.lintSteps {
		if rec := g.summary.recordFor(tc.lintSteps[i].name); rec != nil && rec.outcome == outcomeExecuted {
			n++
		}
	}
	return n
}
