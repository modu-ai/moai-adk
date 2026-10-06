package quality

// gate_node_pm_test.go — Node package-manager-aware runner selection.
//
// The gate's Node axes (test / lint / typecheck) were hardcoded to npm/npx:
// a project that DECLARES its package manager (package.json "packageManager":
// "bun@1.4.2" — the corepack field) still received `npm test` / `npx eslint`
// invocations, which install nothing the project uses and can fail outright
// (operator report from a bun-based consumer project). These tests pin the
// declaration-driven selection:
//
//   - a VALID `bun@<version>` declaration selects the bun path on every axis;
//   - every other shape — other PMs, absent field, malformed value, parse
//     failure — keeps the existing npm/npx vectors byte-identical;
//   - the bun path never invokes npm or npx, never uses the builtin
//     `bun test` (explicit `bun run test` script form), never guesses bun
//     from a value that does not declare it;
//   - config-gated linters and tsc resolve ONLY locally-installed binaries
//     (nearest node_modules/.bin up to the repository boundary; no PATH or
//     global fallback), and their absence FAILS the axis rather than skipping.
//
// Real-bun execution cells (bun >= 1.4 verified semantics: `bun run test --
// --flag` forwards the flag after stripping `--`, exactly like npm) skip when
// the bun binary is absent so the suite stays portable.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// writeBunPackageJSON writes a bun-declared package.json (valid corepack
// declaration) with the given scripts object text into a fresh temp dir.
func writeBunPackageJSON(t *testing.T, scripts string) string {
	t.Helper()
	return writePackageJSON(t, `{"packageManager": "bun@1.4.2", "scripts": `+scripts+`}`)
}

// wantNoNPMinStep fails when a resolved step's binary is npm/npx — the
// zero-npm/npx invariant every bun-declared path must hold.
func wantNoNPMinStep(t *testing.T, step gateStep, axis string) {
	t.Helper()
	if step.binary == "npm" || step.binary == "npx" {
		t.Errorf("%s: bun-declared project resolved to %s step %+v", axis, step.binary, step)
	}
}

// ─────────────────────────────────────────────────────────────
// Test axis: the bun test step.
// ─────────────────────────────────────────────────────────────

// TestResolveNodeTestStep_BunVectors pins the three tiers under a bun
// declaration. Tier shape mirrors the npm table exactly, with `bun run`
// as the explicit script form (the builtin `bun test` is never used — a
// project whose runner is vitest/jest must not be handed to bun's own test
// runner) and mandatory execution (optional=false: a missing bun binary
// fails the step rather than skipping it).
func TestResolveNodeTestStep_BunVectors(t *testing.T) {
	base := nodeBaseStep(t)

	t.Run("tier i: test:run present runs bun run test:run with no flags", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"test": "vitest", "test:run": "vitest run"}`)
		got := resolveNodeTestStep(base, dir)
		want := gateStep{name: "bun run test:run", binary: "bun", args: []string{"run", "test:run"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("step = %+v, want %+v", got, want)
		}
		wantNoNPMinStep(t, got, "test tier i")
	})

	t.Run("tier ii: bare vitest script gets --run appended", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"test": "vitest"}`)
		got := resolveNodeTestStep(base, dir)
		want := gateStep{
			name:   "bun run test --run",
			binary: "bun",
			args:   []string{"run", "test", "--", "--passWithNoTests", "--run"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("step = %+v, want %+v", got, want)
		}
	})

	t.Run("tier ii: jest --watchAll gets --ci appended", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"test": "jest --watchAll"}`)
		got := resolveNodeTestStep(base, dir)
		want := gateStep{
			name:   "bun run test --ci",
			binary: "bun",
			args:   []string{"run", "test", "--", "--passWithNoTests", "--ci"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("step = %+v, want %+v", got, want)
		}
	})

	t.Run("tier iii: plain script keeps passWithNoTests policy", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"test": "node --test"}`)
		got := resolveNodeTestStep(base, dir)
		want := gateStep{
			name:   "bun run test",
			binary: "bun",
			args:   []string{"run", "test", "--", "--passWithNoTests"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("step = %+v, want %+v", got, want)
		}
	})

	t.Run("code-less manifest: no scripts key at all still selects bun tier iii", func(t *testing.T) {
		dir := writePackageJSON(t, `{"packageManager": "bun@1.4.2", "name": "scaffold"}`)
		got := resolveNodeTestStep(base, dir)
		if got.binary != "bun" || got.name != "bun run test" {
			t.Errorf("code-less bun manifest resolved to %+v, want bun run test", got)
		}
	})

	t.Run("builtin collision: args address the script, never `bun test`", func(t *testing.T) {
		// The vector must be `bun run test ...` — args[0]=="run". A bare
		// `bun test` would hand the project to bun's own test runner, which
		// knows nothing about a vitest/jest project.
		dir := writeBunPackageJSON(t, `{"test": "node --test"}`)
		got := resolveNodeTestStep(base, dir)
		if len(got.args) < 2 || got.args[0] != "run" || got.args[1] != "test" {
			t.Errorf("args = %v, want [run test ...] (explicit script form)", got.args)
		}
	})

	t.Run("mandatory: bun steps are never optional", func(t *testing.T) {
		// optional=true is the only skip-on-missing-binary path executeStep
		// carries; the bun steps must not take it. A missing bun binary has
		// to FAIL the commit, not pass it silently.
		for _, scripts := range []string{
			`{"test": "vitest", "test:run": "vitest run"}`,
			`{"test": "vitest"}`,
			`{"test": "node --test"}`,
		} {
			dir := writeBunPackageJSON(t, scripts)
			if got := resolveNodeTestStep(base, dir); got.optional {
				t.Errorf("bun test step is optional for scripts %s: %+v", scripts, got)
			}
		}
	})
}

// TestResolveNodeTestStep_NonBunDeclarationsPreserveNpm pins the
// byte-identical npm vectors for every declaration that is not a valid bun
// one — other PMs, no declaration, malformed bun values. Parse failure and
// malformed declarations must never be guessed into the bun path.
func TestResolveNodeTestStep_NonBunDeclarationsPreserveNpm(t *testing.T) {
	base := nodeBaseStep(t)

	cases := map[string]string{
		"npm@10 declared":          `{"packageManager": "npm@10.0.1", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"yarn@4 declared":          `{"packageManager": "yarn@4.1.0", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"pnpm@9 declared":          `{"packageManager": "pnpm@9.1.0", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"deno@2 declared":          `{"packageManager": "deno@2.0.0", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"no declaration":           `{"scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"bun without version":      `{"packageManager": "bun", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"bun empty version":        `{"packageManager": "bun@", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"bun wrong case":           `{"packageManager": "Bun@1.0.0", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"declaration not a PM":     `{"packageManager": "hello", "scripts": {"test": "vitest", "test:run": "vitest run"}}`,
		"invalid JSON manifest":    `{"packageManager": "bun@1.4.2", "scripts": {"test":` + "\n",
		"manifest missing on disk": "",
	}
	for name, pkg := range cases {
		t.Run(name, func(t *testing.T) {
			var dir string
			if pkg == "" {
				dir = t.TempDir()
			} else {
				dir = writePackageJSON(t, pkg)
			}
			got := resolveNodeTestStep(base, dir)
			// All these manifests carry scripts.test:run, so tier (i) fires
			// with the npm vector (except the invalid-JSON/missing cases,
			// which fall to tier (iii), also npm). Either way the binary must
			// be npm and the name an npm one.
			if got.binary != "npm" {
				t.Errorf("non-bun declaration selected binary %q (%+v)", got.binary, got)
			}
			if got.binary == "npm" && got.args[0] != "run" && got.args[0] != "test" {
				t.Errorf("unexpected npm args %v", got.args)
			}
			if strings.HasPrefix(got.name, "bun") {
				t.Errorf("non-bun declaration selected bun step %q", got.name)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Test axis: the bun lint step.
// ─────────────────────────────────────────────────────────────

// TestResolveNodeLintSteps_BunScriptLint pins the scripts.lint tier under
// bun: `bun run lint`, mandatory. The npm path keeps optional=true (skip when
// npm is absent); the bun path flips to mandatory per the declared-toolchain
// contract.
func TestResolveNodeLintSteps_BunScriptLint(t *testing.T) {
	steps := nodeToolchain(t).lintSteps

	t.Run("scripts.lint runs via bun, mandatory", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"lint": "eslint ."}`)
		resolved, script, watchProne := resolveNodeLintSteps(steps, dir)
		if watchProne {
			t.Fatal("plain eslint lint script reported watch-prone")
		}
		if script != "eslint ." {
			t.Errorf("script = %q, want %q", script, "eslint .")
		}
		want := []gateStep{{name: "bun run lint", binary: "bun", args: []string{"run", "lint"}}}
		if !reflect.DeepEqual(resolved, want) {
			t.Errorf("resolved = %+v, want %+v", resolved, want)
		}
	})

	t.Run("watch-prone lint script is skipped, never run", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"lint": "eslint . --watch"}`)
		resolved, _, watchProne := resolveNodeLintSteps(steps, dir)
		if !watchProne {
			t.Fatal("watch-prone lint script not reported as such")
		}
		if !reflect.DeepEqual(resolved, steps) {
			t.Errorf("watch-prone path must pass table steps through, got %+v", resolved)
		}
	})
}

// makeBinTool creates an executable tool file under dir's node_modules/.bin
// and returns its path.
func makeBinTool(t *testing.T, dir, tool string) string {
	t.Helper()
	binDir := filepath.Join(dir, "node_modules", ".bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(binDir, tool)
	if err := os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestResolveNodeLintSteps_BunConfigLintLocalBin pins the config-gated
// linter tier under bun: binaries resolve from the nearest
// node_modules/.bin up to (and including) the repository root, never past
// it, and never via PATH.
func TestResolveNodeLintSteps_BunConfigLintLocalBin(t *testing.T) {
	steps := nodeToolchain(t).lintSteps

	t.Run("local eslint replaces npx, args drop the tool token, mandatory", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"build": "vite build"}`)
		eslintBin := makeBinTool(t, dir, "eslint")

		resolved, _, _ := resolveNodeLintSteps(steps, dir)
		if len(resolved) != len(steps) {
			t.Fatalf("resolved %d steps, want %d", len(resolved), len(steps))
		}
		es := resolved[0]
		if es.name != "eslint" {
			t.Fatalf("first resolved step = %q, want eslint", es.name)
		}
		if es.binary != eslintBin {
			t.Errorf("eslint binary = %q, want local %q", es.binary, eslintBin)
		}
		if want := []string{"."}; !reflect.DeepEqual(es.args, want) {
			t.Errorf("eslint args = %v, want %v", es.args, want)
		}
		if es.optional {
			t.Error("bun config-lint step must be mandatory (optional=false)")
		}
		if len(es.configFiles) == 0 {
			t.Error("config guard must survive the bun rewrite")
		}
		// biome/oxlint entries are rewritten the same way; with no biome/oxlint
		// binaries installed locally they carry the expected-path form (the
		// deterministic-failure vector), which the missing-tool test below
		// exercises end to end.
	})

	t.Run("nested package resolves the nested node_modules first", func(t *testing.T) {
		repo := t.TempDir()
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		nested := filepath.Join(repo, "apps", "web")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, nested, "package.json", `{"packageManager": "bun@1.4.2", "scripts": {"build": "vite build"}}`)
		nestedBin := makeBinTool(t, nested, "eslint")
		repoBin := makeBinTool(t, repo, "eslint")

		resolved, _, _ := resolveNodeLintSteps(steps, nested)
		if resolved[0].binary != nestedBin {
			t.Errorf("nested package resolved %q, want the nearest %q (repo bin was %q)",
				resolved[0].binary, nestedBin, repoBin)
		}
	})

	t.Run("hoisted bin at the repository root resolves for a nested package", func(t *testing.T) {
		repo := t.TempDir()
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		nested := filepath.Join(repo, "apps", "web")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, nested, "package.json", `{"packageManager": "bun@1.4.2", "scripts": {"build": "vite build"}}`)
		repoBin := makeBinTool(t, repo, "eslint")

		resolved, _, _ := resolveNodeLintSteps(steps, nested)
		if resolved[0].binary != repoBin {
			t.Errorf("nested package resolved %q, want hoisted repo bin %q", resolved[0].binary, repoBin)
		}
	})

	t.Run("walk never crosses the repository boundary", func(t *testing.T) {
		outer := t.TempDir()
		repo := filepath.Join(outer, "repo")
		if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		pkg := filepath.Join(repo, "pkg")
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, pkg, "package.json", `{"packageManager": "bun@1.4.2", "scripts": {"build": "vite build"}}`)
		// The only eslint lives ABOVE the repository root.
		aboveRepo := makeBinTool(t, outer, "eslint")

		resolved, _, _ := resolveNodeLintSteps(steps, pkg)
		if resolved[0].binary == aboveRepo {
			t.Fatal("resolution crossed the repository boundary")
		}
		if resolved[0].binary != filepath.Join(pkg, "node_modules", ".bin", "eslint") {
			t.Errorf("unresolved tool must carry the expected nearest path, got %q", resolved[0].binary)
		}
	})

	t.Run("npm-declared project keeps the npx vector byte-identical", func(t *testing.T) {
		dir := writePackageJSON(t, `{"packageManager": "npm@10.0.1", "scripts": {"build": "vite build"}}`)
		resolved, _, _ := resolveNodeLintSteps(steps, dir)
		if !reflect.DeepEqual(resolved, steps) {
			t.Errorf("npm-declared lint steps rewritten: %+v, want unchanged %+v", resolved, steps)
		}
	})
}

// ─────────────────────────────────────────────────────────────
// Test axis: the bun typecheck step.
// ─────────────────────────────────────────────────────────────

// TestResolveTypecheckStep_BunTiers pins the typecheck tiers under bun.
// Tier (a) (gate.typecheck.command) is honoured VERBATIM — whether to
// translate an explicit npm/npx override for a bun project is a separate
// contract decision this change deliberately does not make.
func TestResolveTypecheckStep_BunTiers(t *testing.T) {
	t.Run("scripts.typecheck runs via bun, mandatory", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"typecheck": "tsc --noEmit"}`)
		step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
		if !ok {
			t.Fatalf("typecheck skipped: %s", reason)
		}
		want := gateStep{name: typecheckStepName, binary: "bun", args: []string{"run", "typecheck"}}
		if !reflect.DeepEqual(step, want) {
			t.Errorf("step = %+v, want %+v", step, want)
		}
	})

	t.Run("tsconfig resolves the local tsc binary, mandatory", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"build": "vite build"}`)
		writeFile(t, dir, "tsconfig.json", `{"include": ["src"]}`)
		tscBin := makeBinTool(t, dir, "tsc")

		step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
		if !ok {
			t.Fatalf("typecheck skipped: %s", reason)
		}
		if step.binary != tscBin {
			t.Errorf("tsc binary = %q, want local %q", step.binary, tscBin)
		}
		if want := []string{"--noEmit"}; !reflect.DeepEqual(step.args, want) {
			t.Errorf("tsc args = %v, want %v", step.args, want)
		}
		if step.optional {
			t.Error("bun tsc step must be mandatory")
		}
	})

	t.Run("missing local tsc carries the expected path, never PATH fallback", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"build": "vite build"}`)
		writeFile(t, dir, "tsconfig.json", `{"include": ["src"]}`)

		step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
		if !ok {
			t.Fatalf("typecheck skipped: %s", reason)
		}
		want := filepath.Join(dir, "node_modules", ".bin", "tsc")
		if step.binary != want {
			t.Errorf("unresolved tsc binary = %q, want expected-path form %q", step.binary, want)
		}
		if step.optional {
			t.Error("unresolved tsc must stay mandatory so execution fails, not skips")
		}
	})

	t.Run("solution-style tsconfig refusal survives the bun path", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"build": "vite build"}`)
		writeFile(t, dir, "tsconfig.json", `{"files": [], "references": [{"path": "./tsconfig.app.json"}]}`)

		_, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
		if ok {
			t.Fatal("solution-style tsconfig must stay refused under bun")
		}
		if !strings.Contains(reason, "solution-style") {
			t.Errorf("refusal reason lost: %q", reason)
		}
	})

	t.Run("no script and no tsconfig keeps the explicit skip reason", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"build": "vite build"}`)
		_, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
		if ok {
			t.Fatal("typecheck must stay skipped without script or tsconfig")
		}
		if !strings.Contains(reason, "no scripts.typecheck") {
			t.Errorf("skip reason lost: %q", reason)
		}
	})

	t.Run("explicit override wins verbatim on a bun project", func(t *testing.T) {
		dir := writeBunPackageJSON(t, `{"typecheck": "tsc --noEmit"}`)
		step, _, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "bun run typecheck")
		if !ok {
			t.Fatal("override must stay honoured")
		}
		if step.binary != "bun" || !reflect.DeepEqual(step.args, []string{"run", "typecheck"}) {
			t.Errorf("override mangled: %+v", step)
		}
	})

	t.Run("npm-declared project keeps the npx tsc vector", func(t *testing.T) {
		dir := writePackageJSON(t, `{"packageManager": "npm@10.0.1", "scripts": {"build": "vite build"}}`)
		writeFile(t, dir, "tsconfig.json", `{"include": ["src"]}`)
		step, reason, ok := resolveTypecheckStep(nodeTypecheckStep(), dir, "")
		if !ok {
			t.Fatalf("typecheck skipped: %s", reason)
		}
		if step.binary != "npx" {
			t.Errorf("npm-declared tsc binary = %q, want npx", step.binary)
		}
	})
}

// ─────────────────────────────────────────────────────────────
// Gate-level cells.
// ─────────────────────────────────────────────────────────────

// bunAvailable reports whether a real bun binary is on PATH.
func bunAvailable() bool {
	_, err := exec.LookPath("bun")
	return err == nil
}

// ─────────────────────────────────────────────────────────────
// Unit cells: the declaration parser and the local-bin walker.
// ─────────────────────────────────────────────────────────────

// TestNodePMFromDeclaration pins the corepack declaration parser: only a
// valid bun@version selects bun; the hash suffix is tolerated; every other
// shape — including near-misses — answers npm.
func TestNodePMFromDeclaration(t *testing.T) {
	cases := map[string]struct {
		decl string
		want nodePM
	}{
		"plain":                    {decl: "bun@1.4.2", want: nodePMBun},
		"with corepack hash":       {decl: "bun@1.0.0+sha512-deadbeef", want: nodePMBun},
		"whitespace padded":        {decl: "  bun@1.4.2  ", want: nodePMBun},
		"no version":               {decl: "bun", want: nodePMNpm},
		"empty version":            {decl: "bun@", want: nodePMNpm},
		"hash only after at":       {decl: "bun@+sha512-x", want: nodePMNpm},
		"wrong case":               {decl: "Bun@1.0.0", want: nodePMNpm},
		"npm declared":             {decl: "npm@10.0.1", want: nodePMNpm},
		"yarn declared":            {decl: "yarn@4.1.0", want: nodePMNpm},
		"pnpm declared":            {decl: "pnpm@9.1.0", want: nodePMNpm},
		"deno declared":            {decl: "deno@2.0.0", want: nodePMNpm},
		"garbage":                  {decl: "hello", want: nodePMNpm},
		"empty":                    {decl: "", want: nodePMNpm},
		"scoped-package lookalike": {decl: "@bun/thing@1.0.0", want: nodePMNpm},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := nodePMFromDeclaration(tc.decl); got != tc.want {
				t.Errorf("nodePMFromDeclaration(%q) = %v, want %v", tc.decl, got, tc.want)
			}
		})
	}
}

// TestReadNodeManifest_ScriptlessManifestIsReadable pins the semantic change
// that matters for bun: a manifest with a packageManager field but NO
// scripts key still reads ok (nil scripts behave as empty), so a scaffold
// bun project selects the bun arm. Invalid JSON stays !ok.
func TestReadNodeManifest_ScriptlessManifestIsReadable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"packageManager": "bun@1.4.2"}`)
	m, ok := readNodeManifest(filepath.Join(dir, "package.json"))
	if !ok {
		t.Fatal("scriptless manifest read failed")
	}
	if m.pm != nodePMBun {
		t.Errorf("pm = %v, want bun", m.pm)
	}
	if m.scripts != nil {
		t.Errorf("scripts = %v, want nil (reads as empty)", m.scripts)
	}

	writeFile(t, dir, "package.json", `{not json`)
	if _, ok := readNodeManifest(filepath.Join(dir, "package.json")); ok {
		t.Error("invalid JSON read as ok")
	}
}

// TestResolveNodeLocalBin_EdgeCells pins the walker's guard rails beyond the
// lint-axis cells above: executability is required (unix), a directory
// shaped like a tool never resolves, and a worktree's .git FILE witnesses
// the repository boundary the same way a .git directory does.
func TestResolveNodeLocalBin_EdgeCells(t *testing.T) {
	t.Run("non-executable file is not a tool", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix exec-bit check")
		}
		dir := t.TempDir()
		binDir := filepath.Join(dir, "node_modules", ".bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binDir, "eslint"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, ok := resolveNodeLocalBin(dir, "eslint"); ok {
			t.Error("non-executable file resolved as a tool")
		}
	})

	t.Run("directory shaped like a tool is not a tool", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "node_modules", ".bin", "eslint"), 0o755); err != nil {
			t.Fatal(err)
		}
		if _, ok := resolveNodeLocalBin(dir, "eslint"); ok {
			t.Error("directory resolved as a tool")
		}
	})

	t.Run("worktree .git file witnesses the boundary", func(t *testing.T) {
		repo := t.TempDir()
		// A linked worktree's .git is a file pointing at the real gitdir.
		writeFile(t, repo, ".git", "gitdir: /elsewhere/main/.git/worktrees/x\n")
		pkg := filepath.Join(repo, "pkg")
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			t.Fatal(err)
		}
		outer := filepath.Dir(repo)
		above := makeBinTool(t, outer, "eslint")

		if _, ok := resolveNodeLocalBin(pkg, "eslint"); ok {
			t.Fatal("walk crossed a file-shaped .git boundary")
		}
		_ = above
	})
}

// TestGateRun_BunDeclaredScriptAxesPassThroughRealBun runs the whole gate on
// a bun-declared fixture whose three script axes are echo-based, against the
// REAL bun binary: every axis must execute through `bun run` and pass. Skips
// when bun is absent (selection vectors are covered above regardless).
func TestGateRun_BunDeclaredScriptAxesPassThroughRealBun(t *testing.T) {
	if !bunAvailable() {
		t.Skip("bun binary absent; selection vectors covered by the resolver tests")
	}
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"packageManager": "bun@1.4.2",
		"scripts": {
			"test": "echo test-script-ran",
			"test:run": "echo test-run-script-ran",
			"lint": "echo lint-script-ran",
			"typecheck": "echo typecheck-script-ran"
		}
	}`)

	g := &QualityGate{config: &GateConfig{
		Enabled:          true,
		ProjectDir:       dir,
		TypecheckEnabled: true,
		VetTimeout:       30 * time.Second,
		LintTimeout:      30 * time.Second,
		TestTimeout:      30 * time.Second,
		TypecheckTimeout: 30 * time.Second,
	}}
	ok, msg := g.Run(context.Background())
	if !ok {
		t.Fatalf("gate failed on a green bun-declared project: %s", msg)
	}
	for _, want := range []string{"bun run test:run", "bun run lint", "bun run typecheck"} {
		if !strings.Contains(msg, want) {
			t.Errorf("run summary missing %q (msg: %s)", want, msg)
		}
	}
	if strings.Contains(msg, "npm test") || strings.Contains(msg, "npm run") {
		t.Errorf("npm invocation surfaced on a bun-declared project: %s", msg)
	}
}

// TestGateRun_BunMissingLocalToolFails pins the fail-not-skip contract: a
// bun-declared project with a linter config whose local binary is absent
// must FAIL the gate naming the expected location — never skip to green,
// never fall back to a PATH/global binary. This cell needs no bun install:
// the doomed absolute path fails at exec regardless.
func TestGateRun_BunMissingLocalToolFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"packageManager": "bun@1.4.2",
		"scripts": {"test": "echo test-ok"}
	}`)
	writeFile(t, dir, "eslint.config.js", "export default {};\n")

	g := &QualityGate{config: &GateConfig{
		Enabled:     true,
		ProjectDir:  dir,
		VetTimeout:  30 * time.Second,
		LintTimeout: 30 * time.Second,
		TestTimeout: 30 * time.Second,
	}}
	ok, msg := g.Run(context.Background())
	if ok {
		t.Fatal("gate passed although the required local eslint binary is absent")
	}
	want := filepath.Join(dir, "node_modules", ".bin", "eslint")
	if !strings.Contains(msg, want) {
		t.Errorf("failure does not name the expected binary location %q: %s", want, msg)
	}
}

// TestBunRunForwardingSemantics locks the measured bun 1.4 semantics the
// tier-(ii)/(iii) arg vectors rely on: `bun run test -- --flag` strips the
// `--` and forwards the flag to the script (identical to npm), and
// `bun run test` addresses the package script, never the builtin bun test
// runner. Skips when bun is absent.
func TestBunRunForwardingSemantics(t *testing.T) {
	if !bunAvailable() {
		t.Skip("bun binary absent")
	}
	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"scripts": {"test": "echo"}}`)

	fwd := exec.Command("bun", "run", "test", "--", "--passWithNoTests")
	fwd.Dir = dir
	out, err := fwd.Output()
	if err != nil {
		t.Fatalf("bun run test -- --passWithNoTests: %v", err)
	}
	if !strings.Contains(string(out), "--passWithNoTests") {
		t.Errorf("-- separator not stripped/forwarded: %q", string(out))
	}

	writeFile(t, dir, "package.json", `{"scripts": {"test": "echo SCRIPT-NOT-BUILTIN"}}`)
	scriptForm := exec.Command("bun", "run", "test")
	scriptForm.Dir = dir
	out, err = scriptForm.Output()
	if err != nil {
		t.Fatalf("bun run test: %v", err)
	}
	if !strings.Contains(string(out), "SCRIPT-NOT-BUILTIN") {
		t.Errorf("bun run test did not address the package script: %q", string(out))
	}
}
