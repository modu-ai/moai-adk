package quality

// gate_summary_test.go — SPEC-GATE-THREE-AXES-001 M1 (card t235).
//
// A passing gate run used to emit zero bytes: runStep collapsed its success
// path to (true, ""), Run dropped the test step's pass-side value outright,
// and the CLI printed nothing when the output was empty. A run that replayed a
// build cache and a run that genuinely executed the suite were byte-identical
// silence. These tests bind the execution summary that closes that hole.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// --- helpers --------------------------------------------------------------

// outcomeTokens is the closed set AC-GTA-001 draws a completed run's outcome
// from. outcomeNotReached is deliberately absent: a run that reaches a verdict
// reaches every configured step.
var outcomeTokens = []string{"executed", "skipped", "disabled"}

// parseSummaryOutcomes maps each summary row's label to its outcome field (the
// text between the label and the detail separator). Rows are recognised by the
// "- " prefix the renderer writes.
//
// The first ": " is the label separator: a label may itself carry a colon
// ("npm run test:run"), but never a colon followed by a space.
func parseSummaryOutcomes(t *testing.T, out string) map[string]string {
	t.Helper()

	got := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		row := strings.TrimPrefix(trimmed, "- ")
		idx := strings.Index(row, ": ")
		if idx < 0 {
			t.Fatalf("summary row %q carries no label separator", row)
		}
		label := row[:idx]
		rest := row[idx+len(": "):]
		if sep := strings.Index(rest, summaryDetailSep); sep >= 0 {
			rest = rest[:sep]
		}
		got[label] = strings.TrimSpace(rest)
	}
	return got
}

// assertExactlyOneOutcomeToken enforces AC-GTA-001's "exactly one outcome
// token" on an outcome field. A field carrying two tokens, or none, fails.
func assertExactlyOneOutcomeToken(t *testing.T, label, field string) {
	t.Helper()

	hits := 0
	for _, tok := range outcomeTokens {
		if strings.Contains(field, tok) {
			hits++
		}
	}
	if hits != 1 {
		t.Errorf("step %q outcome field %q carries %d outcome tokens, want exactly 1", label, field, hits)
	}
}

// goFixture lays out a Go module whose steps the gate will find.
func goFixture(t *testing.T, withSource bool) string {
	t.Helper()

	// No `go` directive: naming a version below the installed toolchain sends
	// the vet and test steps looking for another toolchain to download, which
	// costs the fixture tens of seconds and reaches the network.
	files := map[string]string{"go.mod": "module t235fixture\n"}
	if withSource {
		files["main.go"] = "package main\n\nfunc main() {}\n"
	}
	return writeFixture(t, files)
}

// --- AC-GTA-002 sleeper ---------------------------------------------------

// helperSummarySleepEnv carries the sleep duration, in milliseconds, to the
// re-executed helper below. Absent, the helper is a no-op.
const helperSummarySleepEnv = "MOAI_TEST_SUMMARY_SLEEP_MS"

// TestHelperSummarySleep is not a test. Re-executed as a child by
// summarySleeperCommand, it sleeps for a controlled duration so AC-GTA-002 has
// a step whose wall-clock cost is known in advance. The child bounds itself:
// it exits on its own, and -test.timeout caps it from outside, so nothing is
// left running if the parent returns early.
func TestHelperSummarySleep(t *testing.T) {
	raw := os.Getenv(helperSummarySleepEnv)
	if raw == "" {
		t.Skip("helper process only")
	}
	ms, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("%s = %q: %v", helperSummarySleepEnv, raw, err)
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// summarySleeperCommand returns a whitespace-separated command line — the form
// gate.typecheck.command is parsed from — that sleeps for the configured span.
func summarySleeperCommand() string {
	return strings.Join([]string{
		os.Args[0],
		"-test.run=^TestHelperSummarySleep$",
		"-test.timeout=60s",
	}, " ")
}

// --- AC-GTA-001 -----------------------------------------------------------

// TestSummaryIsCompleteAndVariesWithWhatExecuted — AC-GTA-001.
//
// Two copies of one fixture differing only in gate.disabled_steps. Both
// summaries must name every configured step with exactly one outcome token,
// and the two must differ in the test step's outcome and nowhere else.
//
// disabled_steps is read with an inverted convention: a FALSE value skips the
// step (gate.go executeStep, first guard). A fixture written with the
// intuitive {"go test": true} leaves the step running in both copies and the
// required difference disappears.
func TestSummaryIsCompleteAndVariesWithWhatExecuted(t *testing.T) {
	dir := goFixture(t, true)

	newCfg := func() *GateConfig {
		cfg := DefaultGateConfig()
		cfg.ProjectDir = dir
		cfg.AstGrepGate = nil
		cfg.VetTimeout = 60 * time.Second
		cfg.TestTimeout = 60 * time.Second
		return cfg
	}

	offCfg := newCfg()
	offCfg.DisabledSteps = map[string]bool{"go test": false}
	offPassed, offOut := NewQualityGate(offCfg).Run(context.Background())
	if !offPassed {
		t.Fatalf("fixture with the test step turned off did not reach a passing verdict: %s", offOut)
	}

	onPassed, onOut := NewQualityGate(newCfg()).Run(context.Background())
	if !onPassed {
		t.Fatalf("fixture with the test step enabled did not reach a passing verdict: %s", onOut)
	}

	off := parseSummaryOutcomes(t, offOut)
	on := parseSummaryOutcomes(t, onOut)

	// Every step the Go toolchain configures must be named in both summaries.
	for _, label := range []string{"go vet", typecheckStepName, "golangci-lint", "go test"} {
		if field, ok := off[label]; ok {
			assertExactlyOneOutcomeToken(t, label, field)
		} else {
			t.Fatalf("disabled-test summary omits the configured step %q; summary was:\n%s", label, offOut)
		}
		if field, ok := on[label]; ok {
			assertExactlyOneOutcomeToken(t, label, field)
		} else {
			t.Fatalf("enabled-test summary omits the configured step %q; summary was:\n%s", label, onOut)
		}
	}

	if !strings.Contains(off["go test"], "disabled") {
		t.Errorf("test step outcome with disabled_steps set = %q, want it reported as disabled", off["go test"])
	}
	if !strings.Contains(on["go test"], "executed") {
		t.Errorf("test step outcome without disabled_steps = %q, want it reported as executed", on["go test"])
	}

	for label, offField := range off {
		if label == "go test" {
			continue
		}
		onField, ok := on[label]
		if !ok {
			t.Errorf("step %q present in one summary and absent from the other", label)
			continue
		}
		if outcomeOf(offField) != outcomeOf(onField) {
			t.Errorf("step %q outcome differs between runs that executed the same work: %q vs %q", label, offField, onField)
		}
	}
}

// outcomeOf reduces an outcome field to its token, dropping the measured
// duration that legitimately differs between two runs of the same step.
func outcomeOf(field string) string {
	for _, tok := range outcomeTokens {
		if strings.Contains(field, tok) {
			return tok
		}
	}
	return field
}

// --- AC-GTA-002 -----------------------------------------------------------

// TestSummaryDurationIsMeasuredNotConstant — AC-GTA-002.
//
// One step sleeps a controlled D and must be reported at >= D; a step that
// returns immediately must be reported at < D. No constant — and in
// particular not the configured timeout budget, which both fixtures share —
// satisfies both bounds.
func TestSummaryDurationIsMeasuredNotConstant(t *testing.T) {
	const d = 1200 * time.Millisecond

	t.Setenv(helperSummarySleepEnv, strconv.Itoa(int(d/time.Millisecond)))

	slowGate := typecheckOnlyGate(t, summarySleeperCommand())
	if _, out := slowGate.Run(context.Background()); out == "" {
		t.Fatalf("slow run emitted no summary at all")
	}
	slow := slowGate.summary.recordFor(typecheckStepName)
	if slow == nil {
		t.Fatalf("slow run recorded no %s row", typecheckStepName)
	}

	fastGate := typecheckOnlyGate(t, cmdExitZero)
	fastGate.Run(context.Background())
	fast := fastGate.summary.recordFor(typecheckStepName)
	if fast == nil {
		t.Fatalf("fast run recorded no %s row", typecheckStepName)
	}

	if slow.duration < d {
		t.Errorf("step that slept %s reported %s; the reported duration is not measured", d, slow.duration)
	}
	if fast.duration >= d {
		t.Errorf("step that returned immediately reported %s, at or above the %s the other step slept", fast.duration, d)
	}
}

// typecheckOnlyGate builds a gate over a source-free Go module, so the vet,
// lint and test steps all skip and the typecheck override is the one step that
// executes.
func typecheckOnlyGate(t *testing.T, command string) *QualityGate {
	t.Helper()

	cfg := DefaultGateConfig()
	cfg.ProjectDir = goFixture(t, false)
	cfg.AstGrepGate = nil
	cfg.SkipTests = true
	cfg.TypecheckEnabled = true
	cfg.TypecheckCommand = command
	return NewQualityGate(cfg)
}

// --- AC-GTA-003 -----------------------------------------------------------

// TestSummaryDistinguishesAllFiveSkipPaths — AC-GTA-003.
//
// executeStep can decline to run a step by five ordered guards, first match
// wins. Each fixture below is built to clear every preceding guard so it
// reaches the path it targets; the five reasons must be mutually distinct and
// each must name its own observation.
func TestSummaryDistinguishesAllFiveSkipPaths(t *testing.T) {
	type fixture struct {
		path       string
		build      func(t *testing.T) (*QualityGate, gateStep)
		mustSay    []string
		mustNotSay []string
	}

	fixtures := []fixture{
		{
			path: "a-config-off",
			build: func(t *testing.T) (*QualityGate, gateStep) {
				step := gateStep{name: "alpha", binary: "go", args: []string{"version"}}
				g := gateFor(t, t.TempDir())
				// FALSE turns the step off — the inverted convention.
				g.config.DisabledSteps = map[string]bool{"alpha": false}
				return g, step
			},
			mustSay:    []string{"disabled_steps"},
			mustNotSay: []string{"PATH", "config file"},
		},
		{
			path: "b-optional-binary-absent",
			build: func(t *testing.T) (*QualityGate, gateStep) {
				step := gateStep{name: "bravo", binary: "t235-no-such-binary", optional: true}
				return gateFor(t, t.TempDir()), step
			},
			mustSay:    []string{"PATH", "t235-no-such-binary"},
			mustNotSay: []string{"disabled_steps", "config file"},
		},
		{
			path: "c-config-files-absent",
			build: func(t *testing.T) (*QualityGate, gateStep) {
				// Not optional, so guard (b) cannot fire first.
				step := gateStep{
					name: "charlie", binary: "go", args: []string{"version"},
					configFiles: []string{".t235-absent.yml"},
				}
				return gateFor(t, t.TempDir()), step
			},
			mustSay:    []string{"config file", ".t235-absent.yml"},
			mustNotSay: []string{"PATH", "staged"},
		},
		{
			path: "d-no-staged-match",
			build: func(t *testing.T) (*QualityGate, gateStep) {
				// The staged lookup must SUCCEED for this guard to skip: with a
				// nil result the step runs conservatively, so the fixture is an
				// initialised repository with a staged non-matching file.
				dir := stagedRepo(t, "notes.txt")
				step := gateStep{
					name: "delta", binary: "go", args: []string{"version"},
					changedExts: []string{".py"},
				}
				return gateFor(t, dir), step
			},
			mustSay:    []string{"staged"},
			mustNotSay: []string{"PATH", "disabled_steps"},
		},
		{
			path: "e-no-project-source",
			build: func(t *testing.T) (*QualityGate, gateStep) {
				step := gateStep{
					name: "echo", binary: "go", args: []string{"version"},
					sourceExts: []string{".py"},
				}
				return gateFor(t, writeFixture(t, map[string]string{"pyproject.toml": "[project]\n"})), step
			},
			mustSay:    []string{"source file"},
			mustNotSay: []string{"staged", "PATH"},
		},
	}

	reasons := make(map[string]string, len(fixtures))
	for _, f := range fixtures {
		t.Run(f.path, func(t *testing.T) {
			g, step := f.build(t)
			g.summary = newRunSummary()
			g.summary.seed(step.name)

			ok, _ := g.executeStep(context.Background(), step, 30*time.Second)
			if !ok {
				t.Fatalf("skip path %s failed the gate instead of skipping", f.path)
			}

			rec := g.summary.recordFor(step.name)
			if rec == nil {
				t.Fatalf("skip path %s recorded no summary row", f.path)
			}
			if rec.reason == "" {
				t.Fatalf("skip path %s recorded no reason", f.path)
			}
			for _, want := range f.mustSay {
				if !strings.Contains(rec.reason, want) {
					t.Errorf("reason %q does not name its own observation (missing %q)", rec.reason, want)
				}
			}
			for _, unwanted := range f.mustNotSay {
				if strings.Contains(rec.reason, unwanted) {
					t.Errorf("reason %q claims an observation from a different skip path (%q)", rec.reason, unwanted)
				}
			}
			reasons[f.path] = rec.reason
		})
	}

	if len(reasons) != len(fixtures) {
		t.Fatalf("only %d of %d skip paths produced a reason", len(reasons), len(fixtures))
	}
	seen := make(map[string]string, len(reasons))
	for path, reason := range reasons {
		if other, dup := seen[reason]; dup {
			t.Errorf("skip paths %s and %s share one reason text %q; the five paths must be mutually distinct", other, path, reason)
		}
		seen[reason] = path
	}
}

// gateFor builds a bare gate rooted at dir.
func gateFor(t *testing.T, dir string) *QualityGate {
	t.Helper()
	return &QualityGate{config: &GateConfig{Enabled: true, ProjectDir: dir}}
}

// stagedRepo initialises a git repository holding one staged file, so the
// staged-file lookup succeeds and returns a non-empty list.
func stagedRepo(t *testing.T, name string) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git is not on PATH; the staged-extension guard cannot be reached: %v", err)
	}
	dir := writeFixture(t, map[string]string{name: "staged content\n"})
	for _, args := range [][]string{{"init"}, {"add", name}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v in fixture: %v\n%s", args, err, out)
		}
	}
	return dir
}

// --- AC-GTA-004 -----------------------------------------------------------

// TestSummaryReportsExecutedCommandLineNotLabel — AC-GTA-004.
//
// The assertion binds to argv — step.binary plus step.args as handed to the
// process launcher — and explicitly not to gateStep.name. The two coincide on
// only one of resolveNodeTestStep's three tiers. Tiers (ii) and (iii) carry
// `-- --passWithNoTests` on the command line and not in the label, and that
// flag decides whether an empty suite counts as a pass.
func TestSummaryReportsExecutedCommandLineNotLabel(t *testing.T) {
	cases := []struct {
		tier    string
		scripts string
		wantCmd string
	}{
		{"i-test-run-script", `{"scripts":{"test:run":"vitest run"}}`, "npm run test:run"},
		{"ii-watch-prone", `{"scripts":{"test":"vitest"}}`, "npm test -- --passWithNoTests --run"},
		{"iii-unchanged", `{"scripts":{"test":"echo ok"}}`, "npm test -- --passWithNoTests"},
	}

	for _, tc := range cases {
		t.Run(tc.tier, func(t *testing.T) {
			cfg := DefaultGateConfig()
			cfg.ProjectDir = writeFixture(t, map[string]string{"package.json": tc.scripts})
			cfg.AstGrepGate = nil
			cfg.TypecheckEnabled = false
			cfg.LintTimeout = 20 * time.Second
			cfg.TestTimeout = 20 * time.Second

			// The verdict itself is not the subject: npm may be absent, and the
			// fixture's runner certainly is. What must hold either way is that
			// the summary names the command line the gate handed to the launcher.
			_, out := NewQualityGate(cfg).Run(context.Background())
			if !strings.Contains(out, tc.wantCmd) {
				t.Errorf("summary does not report the executed command line %q; got:\n%s", tc.wantCmd, out)
			}
		})
	}
}

// --- AC-GTA-006 -----------------------------------------------------------

// TestSummaryDoesNotDropAnExistingNotice — AC-GTA-006.
func TestSummaryDoesNotDropAnExistingNotice(t *testing.T) {
	run := func(t *testing.T, ag *AstGrepGateConfig) string {
		t.Helper()
		cfg := DefaultGateConfig()
		cfg.ProjectDir = goFixture(t, false)
		cfg.SkipTests = true
		cfg.AstGrepGate = ag
		passed, out := NewQualityGate(cfg).Run(context.Background())
		if !passed {
			t.Fatalf("fixture did not reach a passing verdict: %s", out)
		}
		return out
	}

	t.Run("rules-dir-unconfigured", func(t *testing.T) {
		out := run(t, &AstGrepGateConfig{Enabled: true, RulesDir: ""})
		if !strings.Contains(out, astGrepReasonRulesDirUnconfigured) {
			t.Errorf("the ast-grep notice was dropped when the summary was emitted; got:\n%s", out)
		}
		if !strings.Contains(out, summaryHeaderPrefix) {
			t.Errorf("output carries the notice but no execution summary; got:\n%s", out)
		}
	})

	// The scanner-absent notice is the AC's named fixture. It is only reachable
	// on a machine without sg — with the scanner present the step really scans
	// and reports something else.
	t.Run("scanner-absent", func(t *testing.T) {
		if _, err := exec.LookPath("sg"); err == nil {
			t.Skip("sg is installed; the scanner-absent notice is unreachable here")
		}
		dir := t.TempDir()
		rules := filepath.Join(dir, "rules")
		if err := os.MkdirAll(rules, 0o755); err != nil {
			t.Fatalf("mkdir rules: %v", err)
		}
		out := run(t, &AstGrepGateConfig{Enabled: true, RulesDir: rules})
		if !strings.Contains(out, astGrepReasonScannerUnavailable) {
			t.Errorf("the scanner-absent notice was dropped; got:\n%s", out)
		}
		if !strings.Contains(out, summaryHeaderPrefix) {
			t.Errorf("output carries the notice but no execution summary; got:\n%s", out)
		}
	})
}

// --- AC-GTA-007 -----------------------------------------------------------

// TestSummaryReportsUnreachedStepsAsNotReached — AC-GTA-007.
//
// A summary pre-populated from configuration and overwritten as steps complete
// would leave an `executed` on a step the run never reached. The aborting
// fixture exposes exactly that residue.
func TestSummaryReportsUnreachedStepsAsNotReached(t *testing.T) {
	cfg := DefaultGateConfig()
	cfg.ProjectDir = goFixture(t, true)
	cfg.AstGrepGate = nil
	cfg.VetTimeout = 60 * time.Second
	// The typecheck axis is the second step; make it fail so the lint and test
	// steps after it are never reached.
	cfg.TypecheckCommand = cmdExitNonZero

	g := NewQualityGate(cfg)
	passed, out := g.Run(context.Background())
	if passed {
		t.Fatalf("fixture with a failing second step passed the gate: %s", out)
	}

	for _, label := range []string{"golangci-lint", "go test"} {
		rec := g.summary.recordFor(label)
		if rec == nil {
			t.Fatalf("aborting run omits the configured step %q from its summary", label)
		}
		if rec.outcome != outcomeNotReached {
			t.Errorf("step %q was never reached but is reported as %q", label, rec.outcome)
		}
		if rec.duration != 0 {
			t.Errorf("step %q was never reached but carries duration %s", label, rec.duration)
		}
	}
	if !strings.Contains(out, string(outcomeNotReached)) {
		t.Errorf("failure output does not report the unreached steps; got:\n%s", out)
	}
}
