//go:build !windows

package quality

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// affected_scope_gate_test.go — the Go test step's wiring of the affected scope
// (SPEC-PRECOMMIT-AFFECTED-SCOPE-001 M3: AC-PAS-006 gate half, AC-PAS-008 gate
// half, AC-PAS-010, AC-PAS-014). Fake binaries on PATH record each invocation's
// argv, so the exact argument list of every scenario is observed without
// running a real test suite. The affected-scope queries are injected.

// recordingBin installs recording fakes for the named binaries and puts them
// first on PATH. Each invocation appends "<name> <argv...>" to the returned log.
func recordingBin(t *testing.T, names ...string) string {
	t.Helper()
	binDir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "calls.log")
	for _, name := range names {
		script := "#!/bin/sh\necho \"" + name + " $*\" >> '" + logPath + "'\nexit 0\n"
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755); err != nil {
			t.Fatalf("write fake %s: %v", name, err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

// recordedCalls returns the non-empty lines of the recording log.
func recordedCalls(t *testing.T, logPath string) []string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read %s: %v", logPath, err)
	}
	var calls []string
	for _, l := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(l) != "" {
			calls = append(calls, l)
		}
	}
	return calls
}

// callsFor returns the recorded calls of one binary.
func callsFor(calls []string, bin string) []string {
	var out []string
	for _, c := range calls {
		if strings.HasPrefix(c, bin+" ") {
			out = append(out, c)
		}
	}
	return out
}

// newGoScopeProject creates a one-module Go project the toolchain detector
// recognises.
func newGoScopeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeScopeFiles(t, dir, map[string]string{
		"go.mod":  scopeGoMod,
		"main.go": "package main\n\nfunc main() {}\n",
	})
	return dir
}

// fixedQueries injects a changed list and a package graph.
func fixedQueries(files []string, prefix string, g moduleGraph) *scopeQueries {
	return &scopeQueries{
		changes: func(context.Context, string) (commitChanges, error) {
			return commitChanges{Files: files, ModulePrefix: prefix}, nil
		},
		graph: func(context.Context, string, string) (moduleGraph, error) { return g, nil },
	}
}

// preCommitGateConfig is the gate configuration of the pre-commit context with
// only the steps under test left on.
func preCommitGateConfig(dir string) *GateConfig {
	cfg := DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.TypecheckEnabled = false
	cfg.AstGrepGate = nil
	cfg.GraphFreshness = nil
	cfg.PreCommitContext = true
	cfg.PreCommitTestScope = "affected"
	return cfg
}

func runGateWith(t *testing.T, cfg *GateConfig, q *scopeQueries) (bool, string) {
	t.Helper()
	g := NewQualityGate(cfg)
	g.scopeQ = q
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return g.Run(ctx)
}

// AC-PAS-006 (gate half): an empty selection is an explicit skip of the Go test
// step, never a pass.
func TestQualityGate_AffectedScope_EmptySelectionSkips(t *testing.T) {
	calls := recordingBin(t, "go")
	dir := newGoScopeProject(t)

	passed, out := runGateWith(t, preCommitGateConfig(dir), fixedQueries([]string{"docs/note.md"}, "", baseScopeGraph()))
	if !passed {
		t.Fatalf("gate failed on a docs-only commit:\n%s", out)
	}
	if got := callsFor(recordedCalls(t, calls), "go"); slices.ContainsFunc(got, func(c string) bool { return strings.HasPrefix(c, "go test") }) {
		t.Errorf("go test ran for a docs-only commit: %v", got)
	}
	if !strings.Contains(out, "go test: skipped") {
		t.Errorf("the summary does not record the go test step as skipped:\n%s", out)
	}
	if !strings.Contains(out, "no package owns a changed file") {
		t.Errorf("the skip carries no stated reason:\n%s", out)
	}
	if strings.Contains(out, "go test: executed") {
		t.Errorf("the summary reports go test as executed:\n%s", out)
	}
}

// AC-PAS-008 (gate half): a failed changed-file or package-graph query falls
// back to the full module with a notice naming the failed query.
func TestQualityGate_AffectedScope_QueryFailureFallsBack(t *testing.T) {
	good := fixedQueries([]string{"a/x.go"}, "", baseScopeGraph())
	cases := []struct {
		name   string
		q      *scopeQueries
		notice string
	}{
		{
			name: "changed-file query fails",
			q: &scopeQueries{
				changes: func(context.Context, string) (commitChanges, error) {
					return commitChanges{}, errors.New("not a git repository")
				},
				graph: good.graph,
			},
			notice: "changed-file query",
		},
		{
			name: "package-graph query fails",
			q: &scopeQueries{
				changes: good.changes,
				graph: func(context.Context, string, string) (moduleGraph, error) {
					return moduleGraph{}, errors.New("package example.com/m/bad is incomplete")
				},
			},
			notice: "package-graph query",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := recordingBin(t, "go")
			dir := newGoScopeProject(t)
			passed, out := runGateWith(t, preCommitGateConfig(dir), tc.q)
			if !passed {
				t.Fatalf("gate failed:\n%s", out)
			}
			var tests []string
			for _, c := range callsFor(recordedCalls(t, calls), "go") {
				if strings.HasPrefix(c, "go test") {
					tests = append(tests, c)
				}
			}
			if want := []string{"go test ./..."}; !slices.Equal(tests, want) {
				t.Errorf("go test calls = %v, want %v (the full module)", tests, want)
			}
			if !strings.Contains(out, tc.notice) || !strings.Contains(out, "falling back to the full module") {
				t.Errorf("the summary does not name the failed query %q and the fallback:\n%s", tc.notice, out)
			}
		})
	}
}

// AC-PAS-001 / AC-PAS-007 wiring: the selected packages reach go test as
// module-root-relative directory patterns, and a widening trigger runs ./... .
func TestQualityGate_AffectedScope_ArgumentList(t *testing.T) {
	t.Run("selection narrows to module-root-relative patterns", func(t *testing.T) {
		calls := recordingBin(t, "go")
		dir := newGoScopeProject(t)
		passed, out := runGateWith(t, preCommitGateConfig(dir), fixedQueries([]string{"a/x.go"}, "", baseScopeGraph()))
		if !passed {
			t.Fatalf("gate failed:\n%s", out)
		}
		var tests []string
		for _, c := range callsFor(recordedCalls(t, calls), "go") {
			if strings.HasPrefix(c, "go test") {
				tests = append(tests, c)
			}
		}
		if want := []string{"go test ./a ./b"}; !slices.Equal(tests, want) {
			t.Errorf("go test calls = %v, want %v", tests, want)
		}
		if !strings.Contains(out, "selected 2, dropped 0") {
			t.Errorf("the summary lacks the selected/dropped counts:\n%s", out)
		}
	})
	t.Run("a widening trigger runs the full module and says why", func(t *testing.T) {
		calls := recordingBin(t, "go")
		dir := newGoScopeProject(t)
		passed, out := runGateWith(t, preCommitGateConfig(dir), fixedQueries([]string{"a/x.go", "go.mod"}, "", baseScopeGraph()))
		if !passed {
			t.Fatalf("gate failed:\n%s", out)
		}
		var tests []string
		for _, c := range callsFor(recordedCalls(t, calls), "go") {
			if strings.HasPrefix(c, "go test") {
				tests = append(tests, c)
			}
		}
		if want := []string{"go test ./..."}; !slices.Equal(tests, want) {
			t.Errorf("go test calls = %v, want %v", tests, want)
		}
		if !strings.Contains(out, "go.mod") || !strings.Contains(out, "widened") {
			t.Errorf("the summary does not state the widening and its trigger:\n%s", out)
		}
	})
	t.Run("build tags survive the replaced argument list", func(t *testing.T) {
		calls := recordingBin(t, "go")
		dir := newGoScopeProject(t)
		cfg := preCommitGateConfig(dir)
		cfg.GoBuildTags = "goolm"
		passed, out := runGateWith(t, cfg, fixedQueries([]string{"a/x.go"}, "", baseScopeGraph()))
		if !passed {
			t.Fatalf("gate failed:\n%s", out)
		}
		if got := callsFor(recordedCalls(t, calls), "go"); !slices.Contains(got, "go test -tags=goolm ./a ./b") {
			t.Errorf("go calls = %v, want one \"go test -tags=goolm ./a ./b\"", got)
		}
	})
}

// AC-PAS-010: outside the pre-commit context, in the full mode, and for the
// non-Go toolchains the test scope is exactly what it was before.
func TestQualityGate_AffectedScope_ContextIsolation(t *testing.T) {
	narrowing := fixedQueries([]string{"a/x.go"}, "", baseScopeGraph())

	goTestCalls := func(t *testing.T, cfg *GateConfig) []string {
		t.Helper()
		calls := recordingBin(t, "go")
		dir := newGoScopeProject(t)
		cfg.ProjectDir = dir
		if passed, out := runGateWith(t, cfg, narrowing); !passed {
			t.Fatalf("gate failed:\n%s", out)
		}
		return callsFor(recordedCalls(t, calls), "go")
	}

	t.Run("without the pre-commit flag the Go test step is test ./...", func(t *testing.T) {
		cfg := preCommitGateConfig("")
		cfg.PreCommitContext = false
		got := goTestCalls(t, cfg)
		if !slices.Contains(got, "go test ./...") || slices.Contains(got, "go test ./a ./b") {
			t.Errorf("go calls = %v, want go test ./... and no narrowed call", got)
		}
	})
	t.Run("mode full restores test ./...", func(t *testing.T) {
		cfg := preCommitGateConfig("")
		cfg.PreCommitTestScope = "full"
		got := goTestCalls(t, cfg)
		if !slices.Contains(got, "go test ./...") || slices.Contains(got, "go test ./a ./b") {
			t.Errorf("go calls = %v, want go test ./... and no narrowed call", got)
		}
	})
	t.Run("the vet step is identical with and without the flag", func(t *testing.T) {
		with := preCommitGateConfig("")
		without := preCommitGateConfig("")
		without.PreCommitContext = false
		vetOf := func(calls []string) []string {
			var vets []string
			for _, c := range calls {
				if strings.HasPrefix(c, "go vet") {
					vets = append(vets, c)
				}
			}
			return vets
		}
		a, b := vetOf(goTestCalls(t, with)), vetOf(goTestCalls(t, without))
		if len(a) == 0 || !slices.Equal(a, b) {
			t.Errorf("vet calls differ between the shapes: with=%v without=%v", a, b)
		}
	})

	nonGo := []struct {
		name   string
		marker map[string]string
		bin    string
	}{
		{"node", map[string]string{"package.json": "{\"name\":\"x\"}\n"}, "npm"},
		{"python", map[string]string{"pyproject.toml": "[project]\nname = \"x\"\n"}, "pytest"},
	}
	for _, tc := range nonGo {
		t.Run(tc.name+" test step is identical with the flag", func(t *testing.T) {
			run := func(flag bool) []string {
				calls := recordingBin(t, tc.bin)
				dir := t.TempDir()
				writeScopeFiles(t, dir, tc.marker)
				cfg := preCommitGateConfig(dir)
				cfg.PreCommitContext = flag
				if passed, out := runGateWith(t, cfg, narrowing); !passed {
					t.Fatalf("gate failed:\n%s", out)
				}
				return callsFor(recordedCalls(t, calls), tc.bin)
			}
			with, without := run(true), run(false)
			if len(without) == 0 || !slices.Equal(with, without) {
				t.Errorf("%s calls differ: with the flag %v, without %v", tc.bin, with, without)
			}
		})
	}
}

// AC-PAS-014: every step receives exactly the budget it received before —
// 30 s vet, 60 s lint, 120 s test, 300 s typecheck — at the definition site and
// at the call sites that hand it to the step runner, in every shape.
func TestQualityGate_StepBudgetsUnchanged(t *testing.T) {
	d := DefaultGateConfig()
	if d.VetTimeout != 30*time.Second || d.LintTimeout != 60*time.Second ||
		d.TestTimeout != 120*time.Second || d.TypecheckTimeout != 300*time.Second || d.LockWait != 30*time.Second {
		t.Fatalf("DefaultGateConfig budgets = vet %s lint %s test %s typecheck %s lock %s, want 30s/60s/120s/300s/30s",
			d.VetTimeout, d.LintTimeout, d.TestTimeout, d.TypecheckTimeout, d.LockWait)
	}

	want := map[string]time.Duration{
		"go vet":          30 * time.Second,
		"golangci-lint":   60 * time.Second,
		"go test":         120 * time.Second,
		typecheckStepName: 300 * time.Second,
	}
	for _, shape := range []struct {
		name     string
		preCommt bool
	}{{"standalone", false}, {"pre-commit affected scope", true}} {
		t.Run(shape.name, func(t *testing.T) {
			// golangci-lint and the typecheck command are optional steps the
			// gate skips when their binary is absent, so fakes must exist.
			fakes := writeFakeBinary(t, "golangci-lint", 0, "")
			fakeTC := writeFakeBinary(t, "typecheck-cmd", 0, "")
			t.Setenv("PATH", fakes+string(os.PathListSeparator)+fakeTC+string(os.PathListSeparator)+os.Getenv("PATH"))
			dir := newGoScopeProject(t)
			writeScopeFiles(t, dir, map[string]string{".golangci.yml": "version: \"2\"\n"})

			cfg := DefaultGateConfig()
			cfg.ProjectDir = dir
			cfg.TypecheckCommand = "typecheck-cmd"
			cfg.AstGrepGate = nil
			cfg.GraphFreshness = nil
			cfg.PreCommitContext = shape.preCommt
			cfg.PreCommitTestScope = "affected"

			got := map[string]time.Duration{}
			g := NewQualityGate(cfg)
			g.scopeQ = fixedQueries([]string{"a/x.go"}, "", baseScopeGraph())
			g.runner = func(_ context.Context, stepName, _ string, timeout time.Duration, _ string, _ ...string) (bool, string) {
				got[stepName] = timeout
				return true, ""
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			if passed, out := g.Run(ctx); !passed {
				t.Fatalf("gate failed:\n%s", out)
			}
			for step, budget := range want {
				if got[step] != budget {
					t.Errorf("step %q received budget %s, want %s", step, got[step], budget)
				}
			}
		})
	}
}
