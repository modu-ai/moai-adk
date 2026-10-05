package quality

// Multi-toolchain detection + execution tests (GH #1680 residual, card t685).
//
// The t559-era recursive scan already finds a module root below the project
// directory when the top level matches nothing. The residual defect is the
// short-circuit: detectToolchain returns on the FIRST match, so a repository
// carrying package.json at the top and go.mod below runs only the Node
// toolchain — the Go module contributes zero coverage, silently, because a
// nil detection is not even needed to pass. The same single-hit shape hides
// every second language: two nested modules of different languages surface
// only the best-ranked one.
//
// The tests below pin the repair end to end with the issue's own A/B pair:
// T0 proves the Go toolchain is reached (and rooted inside its module) on a
// benign fixture; T1 proves a real Go defect in that module FAILS the gate
// instead of passing silently. The go shim makes the invocations observable —
// the gate's green-path steps are silent, so absence-of-output alone would be
// a vacuous signal.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// evalReal resolves path through the filesystem so it compares equal to the
// physical cwd a child process reports (symlinked TMPDIR spellings differ by
// platform).
func evalReal(t *testing.T, path string) string {
	t.Helper()

	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve %s: %v", path, err)
	}
	return real
}

// prependGoShim shadows go with a script that records every invocation's cwd
// and arguments to a log file, then execs the real go. Tests calling it must
// not use t.Parallel (t.Setenv). Returns the log path.
func prependGoShim(t *testing.T) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: shell-script go shim is not directly executable")
	}
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go not on PATH — the shimmed vet step cannot run")
	}

	logPath := filepath.Join(t.TempDir(), "go-shim.log")
	binDir := t.TempDir()
	script := "#!/bin/sh\n" +
		"echo \"$(pwd) $*\" >> " + strconv.Quote(logPath) + "\n" +
		"exec " + strconv.Quote(realGo) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "go"), []byte(script), 0o755); err != nil {
		t.Fatalf("write go shim: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

// shimInvocations returns the go shim's log lines (cwd + args each); a missing
// log means go was invoked zero times.
func shimInvocations(t *testing.T, logPath string) []string {
	t.Helper()

	raw, err := os.ReadFile(logPath)
	if err != nil {
		return nil
	}
	var lines []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(raw)), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// nestedFixtureConfig builds a gate config whose only live axes are the
// detected toolchains' vet and lint steps: ast-grep and graph-freshness are
// off, and tests are skipped so the fixture never reaches a real npm.
func nestedFixtureConfig(dir string) *GateConfig {
	cfg := DefaultGateConfig()
	cfg.ProjectDir = dir
	cfg.AstGrepGate = nil
	cfg.GraphFreshness = nil
	cfg.SkipTests = true
	return cfg
}

// TestNestedGoToolchainReachedDespiteTopLevelNode — GH #1680 T0 (card t685
// [HARD] regression): a repository with package.json at the top and a benign
// Go module at apps/id must reach the Go toolchain, and its steps must run
// INSIDE the module root — not at the project top, where `go vet ./...`
// cannot find a main module at all.
func TestNestedGoToolchainReachedDespiteTopLevelNode(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"package.json":    `{"name": "top", "scripts": {"test": "true"}}`,
		"apps/id/go.mod":  "module example.com/t685/id\n\ngo 1.21\n",
		"apps/id/main.go": "package main\n\nfunc main() {}\n",
	})
	logPath := prependGoShim(t)

	g := NewQualityGate(nestedFixtureConfig(dir))
	passed, out := g.Run(context.Background())
	if !passed {
		t.Fatalf("gate failed on the benign T0 fixture — the nested module's steps ran somewhere they could not succeed:\n%s", out)
	}

	invocations := shimInvocations(t, logPath)
	if len(invocations) == 0 {
		t.Fatalf("go shim log is empty — the gate ran zero Go toolchain steps for a repository whose Go module lives one level below the top (GH #1680 T0: silent zero coverage); summary:\n%s", out)
	}

	// The vet step must have executed inside the module that owns it. Both
	// sides are resolved through EvalSymlinks: the child's pwd reports the
	// physical path (macOS resolves /var/folders → /private/var/folders),
	// while filepath.Join carries the symlinked spelling. The shim logs
	// "cwd args" joined by a single space, so the prefix ends at that space.
	nested := evalReal(t, filepath.Join(dir, "apps", "id"))
	vetReached := false
	for _, line := range invocations {
		if !strings.HasPrefix(line, nested+" ") && line != nested {
			t.Errorf("go invocation ran outside the module root:\n  got  %s\n  want cwd %s", line, nested)
			continue
		}
		if strings.Contains(line, "vet") {
			vetReached = true
		}
	}
	if !vetReached {
		t.Errorf("go shim recorded no vet invocation — the Go toolchain was reached without its vet axis; log:\n%s", strings.Join(invocations, "\n"))
	}

	if rec := g.summary.recordFor("go vet"); rec == nil || rec.outcome != outcomeExecuted {
		t.Errorf("go vet summary row = %v, want an executed row — the run summary does not account for the nested module's vet step", rec)
	}
}

// TestNestedGoDefectFailsGate — GH #1680 T1 (card t685 [HARD] regression):
// the same layout carrying a real Go defect must FAIL the gate with the
// toolchain's verbatim diagnostic. Before the repair this exact tree passed
// with exit 0 while `go vet` run by hand flagged the file — the silent-pass
// direction this card exists to close.
func TestNestedGoDefectFailsGate(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"package.json":     `{"name": "top", "scripts": {"test": "true"}}`,
		"apps/id/go.mod":   "module example.com/t685/id\n\ngo 1.21\n",
		"apps/id/main.go":  "package main\n\nfunc main() {}\n",
		"apps/id/probe.go": "package main\n\nvar gateProbeBug = undefinedSymbol\n",
	})
	logPath := prependGoShim(t)

	g := NewQualityGate(nestedFixtureConfig(dir))
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("gate passed a nested Go module that does not compile — the defect ran through with zero Go coverage (GH #1680 T1); summary:\n%s", out)
	}
	if len(shimInvocations(t, logPath)) == 0 {
		t.Fatalf("gate failed without ever invoking go — the failure cannot be the fixture's Go defect:\n%s", out)
	}
	if !strings.Contains(out, "undefined") {
		t.Errorf("failure output does not surface the compiler diagnostic, got:\n%s", out)
	}
	if !strings.Contains(out, "probe.go") {
		t.Errorf("failure output does not name the defective file, got:\n%s", out)
	}
}

// TestNestedGoVetFailsAtModuleRootNotTop — the negative control for the root
// binding: with the go shim logging cwd, a defecting nested module's vet
// failure must carry the diagnostic ONLY when the step ran at the module root.
// A step run at the project top fails differently ("cannot find main module"),
// which would flip the silent-pass into the loud-fail the issue warns layer 3
// against. Guards the repair against regressing from "reached" to "reached at
// the wrong level".
func TestNestedGoVetFailsAtModuleRootNotTop(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"package.json":     `{"name": "top", "scripts": {"test": "true"}}`,
		"apps/id/go.mod":   "module example.com/t685/id\n\ngo 1.21\n",
		"apps/id/probe.go": "package main\n\nvar gateProbeBug = undefinedSymbol\n",
	})
	logPath := prependGoShim(t)

	g := NewQualityGate(nestedFixtureConfig(dir))
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("gate passed a non-compiling nested module:\n%s", out)
	}
	nested := evalReal(t, filepath.Join(dir, "apps", "id"))
	for _, line := range shimInvocations(t, logPath) {
		if !strings.HasPrefix(line, nested+" ") && line != nested {
			t.Errorf("go invocation ran outside the module root — the failure is the wrong loud one (cannot find main module):\n  %s", line)
		}
	}
	if !strings.Contains(out, "undefined") || !strings.Contains(out, "probe.go") {
		t.Errorf("failure does not carry the in-module compile diagnostic — the step did not run where the module owns:\n%s", out)
	}
	if !strings.Contains(strings.Join(shimInvocations(t, logPath), "\n"), "vet") {
		t.Error("no vet invocation recorded — the verdict did not come from the vet axis")
	}
}

// TestDetectToolchainsCollectsTopAndNested pins the T0 detection shape at
// the API level: two entries — the root-level Node project first, then the
// nested Go module at its own root. The residual defect returned only the
// first match.
func TestDetectToolchainsCollectsTopAndNested(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"package.json":    `{"name": "top"}`,
		"apps/id/go.mod":  "module example.com/t685/id\n\ngo 1.21\n",
		"apps/id/main.go": "package main\n\nfunc main() {}\n",
	})

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	tcs := g.detectToolchains()
	if len(tcs) != 2 {
		t.Fatalf("detectToolchains() found %d entries, want 2 (root Node + nested Go): %+v", len(tcs), tcs)
	}
	if tcs[0].tc.markerFiles[0] != "package.json" || tcs[0].root != dir {
		t.Errorf("entry 0 = %v at %q, want package.json at the project root (root-level precedence)", tcs[0].tc.markerFiles, tcs[0].root)
	}
	nested := filepath.Join(dir, "apps", "id")
	if tcs[1].tc.markerFiles[0] != "go.mod" || tcs[1].root != nested {
		t.Errorf("entry 1 = %v at %q, want go.mod at %q", tcs[1].tc.markerFiles, tcs[1].root, nested)
	}
}

// TestDetectToolchainsCollectsMultipleNestedLanguages pins the multi-language
// shape with no root-level match: two nested modules of different languages
// both surface. The previous scan collapsed them to the best-ranked one.
func TestDetectToolchainsCollectsMultipleNestedLanguages(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"apps/api/go.mod":       "module example.com/t685/api\n\ngo 1.21\n",
		"apps/web/package.json": `{"name": "web"}`,
	})

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	tcs := g.detectToolchains()
	if len(tcs) != 2 {
		t.Fatalf("detectToolchains() found %d entries, want 2 (nested Go + nested Node): %+v", len(tcs), tcs)
	}
	seen := map[string]string{} // marker -> root
	for _, dt := range tcs {
		seen[dt.tc.markerFiles[0]] = dt.root
	}
	if got := seen["go.mod"]; got != filepath.Join(dir, "apps", "api") {
		t.Errorf("go.mod entry root = %q, want %q", got, filepath.Join(dir, "apps", "api"))
	}
	if got := seen["package.json"]; got != filepath.Join(dir, "apps", "web") {
		t.Errorf("package.json entry root = %q, want %q", got, filepath.Join(dir, "apps", "web"))
	}
}

// TestDetectToolchainsSingleLanguageShape — card t685 [HARD] regression: a
// single-language project detects exactly one entry rooted at the project
// directory, so its step list is what it has always been.
func TestDetectToolchainsSingleLanguageShape(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"go.mod":  "module example.com/single\n\ngo 1.21\n",
		"main.go": "package main\n\nfunc main() {}\n",
	})

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	tcs := g.detectToolchains()
	if len(tcs) != 1 {
		t.Fatalf("detectToolchains() found %d entries, want 1 for a single-language project: %+v", len(tcs), tcs)
	}
	if tcs[0].tc.markerFiles[0] != "go.mod" || tcs[0].root != dir {
		t.Errorf("entry = %v at %q, want go.mod at the project root %q", tcs[0].tc.markerFiles, tcs[0].root, dir)
	}
}

// TestDetectToolchainsDoesNotDescendModuleSubtree pins the module-root
// boundary: a marker inside an already-detected module's own subtree is that
// module's business, not a second entry of the same language.
func TestDetectToolchainsDoesNotDescendModuleSubtree(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"go.mod":           "module example.com/top\n\ngo 1.21\n",
		"sub/inner/go.mod": "module example.com/top/inner\n\ngo 1.21\n",
	})

	g := NewQualityGate(&GateConfig{ProjectDir: dir})
	tcs := g.detectToolchains()
	if len(tcs) != 1 {
		t.Fatalf("detectToolchains() found %d entries, want 1 (the subtree below a module root belongs to it): %+v", len(tcs), tcs)
	}
	if tcs[0].root != dir {
		t.Errorf("entry root = %q, want the project root %q", tcs[0].root, dir)
	}
}

// TestDetectToolchainWrapperKeepsPrecedence pins the single-entry wrapper's
// compatibility semantics over the new multi-entry collection: a root-level
// match wins outright; without one, the nested entry ranked first by
// toolchains-table order (Go before Node) wins regardless of walk order.
func TestDetectToolchainWrapperKeepsPrecedence(t *testing.T) {
	t.Run("root-level match wins", func(t *testing.T) {
		t.Parallel()
		dir := writeFixture(t, map[string]string{
			"package.json":    `{"name": "top"}`,
			"apps/api/go.mod": "module example.com/t685/api\n\ngo 1.21\n",
		})
		g := NewQualityGate(&GateConfig{ProjectDir: dir})
		dt := g.detectToolchain()
		if dt == nil || dt.tc.markerFiles[0] != "package.json" || dt.root != dir {
			t.Errorf("detectToolchain() = %v at %q, want package.json at the project root", dt, dir)
		}
	})

	t.Run("table order ranks nested entries", func(t *testing.T) {
		t.Parallel()
		dir := writeFixture(t, map[string]string{
			"apps/web/package.json": `{"name": "web"}`,
			"apps/api/go.mod":       "module example.com/t685/api\n\ngo 1.21\n",
		})
		g := NewQualityGate(&GateConfig{ProjectDir: dir})
		dt := g.detectToolchain()
		if dt == nil || dt.tc.markerFiles[0] != "go.mod" {
			t.Errorf("detectToolchain() = %v, want go.mod (Go outranks Node in the table)", dt)
		}
	})
}
