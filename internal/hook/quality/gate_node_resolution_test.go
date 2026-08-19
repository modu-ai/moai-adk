package quality

// gate_node_resolution_test.go — Node toolchain test-step run-form resolution
// (SPEC-HARNESS-GATE-TEST-001).
//
// The gate's Node test step was hardcoded to `npm test -- --passWithNoTests`.
// For a package whose `test` script is watch-prone (bare `vitest`, or a
// `--watch`/`--watchAll` token), that command never self-terminates and the
// step dies at TestTimeout with the suite at 0% — the gate could not finish on
// such projects at all (three consecutive consumer cards bypassed the gate via
// SKIP_MOAI_PRECOMMIT for exactly this reason). resolveNodeTestStep rewrites
// the step to a self-terminating (run-form) command using the 3-tier
// resolution, and this file pins that decision table:
//
//	(i)   scripts.test:run present → `npm run test:run`
//	(ii)  scripts.test watch-prone → runner's non-watch flag appended
//	      (vitest `--run`, jest `--ci`)
//	(iii) anything else → `npm test` unchanged
//
// `--passWithNoTests` must survive on tiers (ii)/(iii) (an empty test suite
// must not regress the gate); tier (i) appends no flags because the test:run
// script is author-owned run-form and a turbo-delegating script hard-errors
// on unknown arguments (measured 2026-08-19). Every package.json parse
// failure must fall back to tier (iii) rather than failing the commit.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// nodeToolchain returns the Node entry from the package-level table.
func nodeToolchain(t *testing.T) *langToolchain {
	t.Helper()
	for i := range toolchains {
		if len(toolchains[i].markerFiles) > 0 && toolchains[i].markerFiles[0] == "package.json" {
			return &toolchains[i]
		}
	}
	t.Fatal("Node toolchain not found in the toolchains table")
	return nil
}

// nodeBaseStep returns a copy of the table's Node test step, the exact shape
// resolveNodeTestStep receives from Run.
func nodeBaseStep(t *testing.T) gateStep {
	t.Helper()
	tc := nodeToolchain(t)
	if tc.testStep == nil {
		t.Fatal("Node toolchain has no test step")
	}
	return *tc.testStep
}

// writePackageJSON writes the given package.json content in a fresh temp dir
// and returns the dir.
func writePackageJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "package.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestResolveNodeTestStep_DecisionTable pins the 3-tier resolution (tiers (i),
// (ii), (iii) — SPEC REQ-HGT-001/002/003). Each subtest names the tier it
// covers.
func TestResolveNodeTestStep_DecisionTable(t *testing.T) {
	base := nodeBaseStep(t)

	t.Run("tier i: test:run present invokes that script", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "vitest", "test:run": "vitest run"}}`)
		got := resolveNodeTestStep(base, dir)
		if got.name != "npm run test:run" {
			t.Errorf("name = %q, want %q", got.name, "npm run test:run")
		}
		if got.binary != "npm" {
			t.Errorf("binary = %q, want %q", got.binary, "npm")
		}
		// No flags appended: the test:run script is author-owned run-form.
		// Measurement (2026-08-19): appending `--passWithNoTests` via npm's
		// `--` makes a turbo-delegating script (`turbo run test:run
		// --passWithNoTests`) hard-error on the unknown argument, breaking
		// the gate on monorepo roots entirely.
		wantArgs := []string{"run", "test:run"}
		if !reflect.DeepEqual(got.args, wantArgs) {
			t.Errorf("args = %v, want %v", got.args, wantArgs)
		}
	})

	t.Run("tier i: test:run present wins even when test is self-terminating", func(t *testing.T) {
		// Priority is (i) > (ii) > (iii): a dedicated run-form script is the
		// strongest signal regardless of what `test` looks like.
		dir := writePackageJSON(t, `{"scripts": {"test": "node --test", "test:run": "vitest run"}}`)
		if got := resolveNodeTestStep(base, dir); got.name != "npm run test:run" {
			t.Errorf("name = %q, want %q", got.name, "npm run test:run")
		}
	})

	t.Run("tier i: test:run fires without a test script at all", func(t *testing.T) {
		// `npm test` on a package without scripts.test errors before any flag
		// applies; a present test:run is the usable command (REQ-HGT-001 keys
		// on test:run existence, not on test's presence).
		dir := writePackageJSON(t, `{"scripts": {"test:run": "vitest run"}}`)
		if got := resolveNodeTestStep(base, dir); got.name != "npm run test:run" {
			t.Errorf("name = %q, want %q", got.name, "npm run test:run")
		}
	})

	t.Run("tier ii: bare vitest gets --run appended", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "vitest"}}`)
		got := resolveNodeTestStep(base, dir)
		if got.name != "npm test --run" {
			t.Errorf("name = %q, want %q", got.name, "npm test --run")
		}
		wantArgs := []string{"test", "--", "--passWithNoTests", "--run"}
		if !reflect.DeepEqual(got.args, wantArgs) {
			t.Errorf("args = %v, want %v", got.args, wantArgs)
		}
	})

	t.Run("tier ii: vitest --watch token gets --run appended", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "vitest --watch"}}`)
		got := resolveNodeTestStep(base, dir)
		if got.name != "npm test --run" {
			t.Errorf("name = %q, want %q", got.name, "npm test --run")
		}
		wantArgs := []string{"test", "--", "--passWithNoTests", "--run"}
		if !reflect.DeepEqual(got.args, wantArgs) {
			t.Errorf("args = %v, want %v", got.args, wantArgs)
		}
	})

	t.Run("tier ii: vitest --watchAll token gets --run appended", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "vitest --watchAll"}}`)
		if got := resolveNodeTestStep(base, dir); got.name != "npm test --run" {
			t.Errorf("name = %q, want %q", got.name, "npm test --run")
		}
	})

	t.Run("tier ii: jest --watchAll token gets --ci appended", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "jest --watchAll"}}`)
		got := resolveNodeTestStep(base, dir)
		if got.name != "npm test --ci" {
			t.Errorf("name = %q, want %q", got.name, "npm test --ci")
		}
		wantArgs := []string{"test", "--", "--passWithNoTests", "--ci"}
		if !reflect.DeepEqual(got.args, wantArgs) {
			t.Errorf("args = %v, want %v", got.args, wantArgs)
		}
	})

	t.Run("tier ii: jest --watch token gets --ci appended", func(t *testing.T) {
		// --ci is chosen over --watchAll=false because it uniformly negates
		// both jest watch modes (appending --watchAll=false would leave a
		// `jest --watch` script still watching).
		dir := writePackageJSON(t, `{"scripts": {"test": "jest --watch"}}`)
		if got := resolveNodeTestStep(base, dir); got.name != "npm test --ci" {
			t.Errorf("name = %q, want %q", got.name, "npm test --ci")
		}
	})

	t.Run("tier iii: plain self-terminating script is preserved verbatim", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "node --test"}}`)
		got := resolveNodeTestStep(base, dir)
		if !reflect.DeepEqual(got, base) {
			t.Errorf("step rewritten for a plain script: %+v, want unchanged %+v", got, base)
		}
	})

	t.Run("tier iii: bare jest is self-terminating and preserved", func(t *testing.T) {
		// REQ-HGT-002 defines watch-prone as bare vitest or --watch-family
		// tokens; bare jest runs once by default and stays on tier (iii).
		dir := writePackageJSON(t, `{"scripts": {"test": "jest"}}`)
		if got := resolveNodeTestStep(base, dir); !reflect.DeepEqual(got, base) {
			t.Errorf("step rewritten for bare jest: %+v, want unchanged", got)
		}
	})

	t.Run("tier iii: vitest run subcommand is already run-form and preserved", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"test": "vitest run"}}`)
		if got := resolveNodeTestStep(base, dir); !reflect.DeepEqual(got, base) {
			t.Errorf("step rewritten for `vitest run`: %+v, want unchanged", got)
		}
	})

	t.Run("tier iii: watch-prone script of an unknown runner is preserved", func(t *testing.T) {
		// The gate knows the non-watch flag for vitest and jest only. For any
		// other runner, guessing a flag risks breaking the test command
		// outright; tier (iii) keeps current behavior (REQ-HGT-003).
		dir := writePackageJSON(t, `{"scripts": {"test": "mocha --watch"}}`)
		if got := resolveNodeTestStep(base, dir); !reflect.DeepEqual(got, base) {
			t.Errorf("step rewritten for unknown runner: %+v, want unchanged", got)
		}
	})

	t.Run("tier iii: no test script at all keeps current behavior", func(t *testing.T) {
		dir := writePackageJSON(t, `{"scripts": {"build": "vite build"}}`)
		if got := resolveNodeTestStep(base, dir); !reflect.DeepEqual(got, base) {
			t.Errorf("step rewritten without a test script: %+v, want unchanged", got)
		}
	})
}

// TestResolveNodeTestStep_ParseFailureFallback pins tier (iii) on every
// package.json parse failure (REQ-HGT-003): the gate must fall back to the
// current `npm test` command, never fail because the manifest was unreadable.
func TestResolveNodeTestStep_ParseFailureFallback(t *testing.T) {
	base := nodeBaseStep(t)

	cases := map[string]func(t *testing.T) string{
		"missing package.json": func(t *testing.T) string {
			return t.TempDir() // no package.json written
		},
		"invalid JSON": func(t *testing.T) string {
			return writePackageJSON(t, `{not json`)
		},
		"no scripts key": func(t *testing.T) string {
			return writePackageJSON(t, `{"name": "x", "version": "1.0.0"}`)
		},
		"scripts not an object": func(t *testing.T) string {
			return writePackageJSON(t, `{"scripts": []}`)
		},
		"empty dir string": func(t *testing.T) string {
			return ""
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			dir := setup(t)
			if got := resolveNodeTestStep(base, dir); !reflect.DeepEqual(got, base) {
				t.Errorf("parse failure rewrote the step: %+v, want unchanged %+v", got, base)
			}
		})
	}
}

// TestResolveNodeTestStep_RepoSurfaceFixtures pins the resolution against the
// real script shapes of the consuming monorepo (mo.ai.kr): root (turbo with
// test:run), apps/www and apps/bo (bare vitest with test:run). All three
// surfaces carry a test:run script, so all three must resolve to tier (i) and
// self-terminate under the gate.
func TestResolveNodeTestStep_RepoSurfaceFixtures(t *testing.T) {
	base := nodeBaseStep(t)

	fixtures := map[string]string{
		"root (turbo + test:run)": `{"scripts": {` +
			`"test": "PATH=$HOME/.bun/bin:$PATH turbo run test --", ` +
			`"test:run": "turbo run test:run"}}`,
		"apps/www (vitest + test:run)": `{"scripts": {` +
			`"test": "vitest", ` +
			`"test:run": "vitest run"}}`,
		"apps/bo (vitest + test:run)": `{"scripts": {` +
			`"test": "vitest", ` +
			`"test:run": "vitest run"}}`,
	}
	for name, pkg := range fixtures {
		t.Run(name, func(t *testing.T) {
			dir := writePackageJSON(t, pkg)
			got := resolveNodeTestStep(base, dir)
			if got.name != "npm run test:run" {
				t.Errorf("name = %q, want %q — this surface must self-terminate via tier (i)",
					got.name, "npm run test:run")
			}
			wantArgs := []string{"run", "test:run"}
			if !reflect.DeepEqual(got.args, wantArgs) {
				t.Errorf("args = %v, want %v", got.args, wantArgs)
			}
		})
	}
}

// TestResolveNodeTestStep_NonNodeStepUnchanged guards the transform's scope:
// only the Node test step is rewritten; every other toolchain's test step
// passes through untouched (REQ-HGT-004).
func TestResolveNodeTestStep_NonNodeStepUnchanged(t *testing.T) {
	dir := writePackageJSON(t, `{"scripts": {"test": "vitest", "test:run": "vitest run"}}`)
	steps := []gateStep{
		{name: "go test", binary: "go", args: []string{"test", "./..."}},
		{name: "cargo test", binary: "cargo", args: []string{"test"}},
	}
	for _, step := range steps {
		if got := resolveNodeTestStep(step, dir); !reflect.DeepEqual(got, step) {
			t.Errorf("non-Node step was rewritten: %+v, want unchanged %+v", got, step)
		}
	}
}

// TestResolveNodeTestStep_DoesNotMutatePackageTable asserts the resolution
// never writes back into the package-level toolchains slice — mirroring the
// resolvePytestRunner mutation guard.
func TestResolveNodeTestStep_DoesNotMutatePackageTable(t *testing.T) {
	before := nodeBaseStep(t)
	dir := writePackageJSON(t, `{"scripts": {"test": "vitest", "test:run": "vitest run"}}`)

	_ = resolveNodeTestStep(before, dir)

	after := nodeBaseStep(t)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("resolution mutated the package-level toolchains table: %+v -> %+v", before, after)
	}
}
