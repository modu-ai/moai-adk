package cli

// checklist_wiring_test.go — SPEC-TUI-CHECKLIST-001 M4: the `moai init` wiring.
// Every test here runs the real runInit through buffer writers with the
// activation predicate forced on (or, for the gate tests, driven through the
// two input seams). Seam-changing tests never run in parallel and restore the
// seam through t.Cleanup (spec.md §4 C-5).

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/core/project"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// checklistRichFixedElapsed is the clock the rich runs report, so the summary
// headline is the same on every run.
const checklistRichFixedElapsed = 1400 * time.Millisecond

// checklistRichGoldenPath resolves a rich golden from this source file.
func checklistRichGoldenPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "checklist", "rich", name+".golden")
}

// checklistRichClock freezes the checklist clock: the first reading is the
// start, every later one is start+checklistRichFixedElapsed.
func checklistRichClock(t *testing.T) {
	t.Helper()
	orig := checklistClockFn
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	calls := 0
	checklistClockFn = func() time.Time {
		calls++
		if calls == 1 {
			return base
		}
		return base.Add(checklistRichFixedElapsed)
	}
	t.Cleanup(func() { checklistClockFn = orig })
}

var (
	checklistAnsiRe  = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	checklistEraseRe = regexp.MustCompile(`[^\n]*\r\x1b\[2K`)
)

// checklistCollapse applies the spec §3.3 line unit: lines overwritten in place
// (`\r\x1b[2K`) are folded to their final text and ANSI sequences are stripped.
func checklistCollapse(raw string) string {
	return checklistAnsiRe.ReplaceAllString(checklistEraseRe.ReplaceAllString(raw, ""), "")
}

// checklistLines splits collapsed text into the §3.3 line unit.
func checklistLines(collapsed string) []string {
	collapsed = strings.TrimSuffix(collapsed, "\n")
	if collapsed == "" {
		return nil
	}
	return strings.Split(collapsed, "\n")
}

// checklistRichNormalize is the fallback normalization plus the platform string
// of the environment line.
func checklistRichNormalize(raw, root string) string {
	s := checklistFallbackNormalize(raw, root)
	return strings.ReplaceAll(s, runtime.GOOS+"/"+runtime.GOARCH, "<PLATFORM>")
}

// checklistCombined renders both collapsed streams as the rich golden layout.
func checklistCombined(stdout, stderr string) string {
	return "== stdout ==\n" + checklistCollapse(stdout) + "== stderr ==\n" + checklistCollapse(stderr)
}

// checklistRichCompare compares (or, under UPDATE_GOLDEN=1, rewrites) one rich
// golden and returns the normalized text.
func checklistRichCompare(t *testing.T, name, root, raw string) string {
	t.Helper()
	got := checklistRichNormalize(raw, root)
	path := checklistRichGoldenPath(name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		return got
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("rich golden %s is missing: %v", path, err)
	}
	if got != string(want) {
		t.Fatalf("%s differs from the rich golden %s\n%s", name, path, checklistFallbackFirstDiff(string(want), got))
	}
	return got
}

// newChecklistInitCmd builds the fixed non-interactive init command of the M0
// fixture plus the --verbose flag, with fresh buffer writers. overrides replace
// the fixed flag values.
func newChecklistInitCmd(t *testing.T, overrides map[string]string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	cmd := initModeTestCmd()
	cmd.Flags().Bool("verbose", false, "")
	flags := map[string]string{
		"non-interactive": "true",
		"language":        "go",
		"mode":            "tdd",
		"no-hooks":        "false",
	}
	for k, v := range overrides {
		flags[k] = v
	}
	for name, val := range flags {
		if err := cmd.Flags().Set(name, val); err != nil {
			t.Fatalf("set --%s=%s: %v", name, val, err)
		}
	}
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	return cmd, out, errBuf
}

// checklistRichRun runs the fixed init with the checklist forced on and returns
// the temp root, the two raw streams and the run error.
func checklistRichRun(t *testing.T, overrides map[string]string) (root, stdout, stderr string, err error) {
	t.Helper()
	root, _ = checklistFallbackEnv(t)
	forceChecklistOn(t)
	checklistRichClock(t)
	cmd, out, errBuf := newChecklistInitCmd(t, overrides)
	err = runInit(cmd, nil)
	return root, out.String(), errBuf.String(), err
}

// checklistItemLine returns the collapsed stderr lines that carry label.
func checklistItemLines(collapsedStderr, label string) []string {
	var hits []string
	for _, l := range checklistLines(collapsedStderr) {
		if strings.Contains(l, label) {
			hits = append(hits, l)
		}
	}
	return hits
}

// requireItemEnds asserts that exactly one item line carries label and that it
// ends with the wanted glyph and not the other one.
func requireItemEnds(t *testing.T, collapsedStderr, label, want, notWant string) {
	t.Helper()
	hits := checklistItemLines(collapsedStderr, label)
	if len(hits) != 1 {
		t.Errorf("item %q appears on %d lines, want exactly 1:\n%s", label, len(hits), collapsedStderr)
		return
	}
	if !strings.Contains(hits[0], want) {
		t.Errorf("item line %q lacks the %s glyph", hits[0], want)
	}
	if strings.Contains(hits[0], notWant) {
		t.Errorf("item line %q must not carry the %s glyph", hits[0], notWant)
	}
}

func checklistLabel(key string) string { return tui.Translate("checklist.item."+key, "en") }

// TestChecklistInitConditionalSkip is AC-TCL-001 (command level): a step init
// did not take is drawn with the skipped glyph, never the done glyph.
func TestChecklistInitConditionalSkip(t *testing.T) {
	t.Run("no_hooks", func(t *testing.T) {
		_, stdout, stderr, err := checklistRichRun(t, map[string]string{"no-hooks": "true"})
		if err != nil {
			t.Fatalf("init: %v\n%s", err, stderr)
		}
		se := checklistCollapse(stderr)
		requireItemEnds(t, se, checklistLabel("hooks.prepush"), "○", "✓")
		requireItemEnds(t, se, checklistLabel("hooks.precommit"), "○", "✓")
		all := checklistCollapse(stdout) + se
		for _, banned := range []string{"Pre-push hook installed", "Pre-commit hook installed"} {
			if n := strings.Count(all, banned); n != 0 {
				t.Errorf("%q appears %d times, want 0", banned, n)
			}
		}
		if n := strings.Count(all, "this checkout is shared across concurrent sessions;"); n != 1 {
			t.Errorf("P5 advisory appears %d times, want exactly 1", n)
		}
	})
	t.Run("mcp_declined", func(t *testing.T) {
		_, stdout, stderr, err := checklistRichRun(t, nil)
		if err != nil {
			t.Fatalf("init: %v\n%s", err, stderr)
		}
		se := checklistCollapse(stderr)
		requireItemEnds(t, se, checklistLabel("mcp.entry"), "○", "✓")
		if n := strings.Count(checklistCollapse(stdout)+se, "Provisioned the moai MCP server entry"); n != 0 {
			t.Errorf("MCP provision announcement appears %d times, want 0", n)
		}
	})
	t.Run("control_not_skipped", func(t *testing.T) {
		root, _ := checklistFallbackEnv(t)
		_ = root
		forceChecklistOn(t)
		checklistRichClock(t)
		origInteractive := isInteractiveStdin
		isInteractiveStdin = func() bool { return true }
		t.Cleanup(func() { isInteractiveStdin = origInteractive })
		origDeps := deps
		deps = nil
		t.Cleanup(func() { deps = origDeps })
		origWizard := runWizardFn
		runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
			return &wizard.WizardResult{ConversationLang: "en", DevelopmentMode: "tdd"}, nil
		}
		t.Cleanup(func() { runWizardFn = origWizard })

		cmd, _, errBuf := newChecklistInitCmd(t, map[string]string{"non-interactive": "false"})
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		se := checklistCollapse(errBuf.String())
		for _, k := range []string{"hooks.prepush", "hooks.precommit", "mcp.entry"} {
			requireItemEnds(t, se, checklistLabel(k), "✓", "○")
		}
	})
}

// checklistDirectFragments are the five direct-print line fragments spec §3.3
// counts separately from the checklist's own lines.
var checklistDirectFragments = []string{
	"Deploying core templates only (slim mode)",
	"Common skills and agents install into your user folders",
	"Initializing MoAI project...",
	"user-asset ensure: ",
	"this checkout is shared across concurrent sessions;",
}

// checklistFragmentLines returns, across both collapsed streams, every line that
// carries frag with its leading spaces trimmed.
func checklistFragmentLines(frag string, collapsed ...string) []string {
	var hits []string
	for _, c := range collapsed {
		for _, l := range checklistLines(c) {
			if strings.Contains(l, frag) {
				hits = append(hits, strings.TrimLeft(l, " "))
			}
		}
	}
	return hits
}

// checklistPassThroughFragment marks the lines writeInstallSummary prints for a
// converted asset the Codex face could not map. The ensure writes them straight
// to the command writer before any checklist item opens, so the checklist does
// not draw them.
const checklistPassThroughFragment = "unconverted reference (no Codex-side mapping"

// fixedUserAssetEnsure returns an ensure seam that prints a fixed header (the
// "user-asset ensure: " direct fragment) and n pass-through lines, whatever the
// embedded catalog holds.
func fixedUserAssetEnsure(n int) func(string, []string, io.Writer) error {
	return func(_ string, _ []string, out io.Writer) error {
		_, _ = io.WriteString(out, "user-asset ensure: 12 installed, 0 refreshed, 0 removed, 0 collision-skipped, 0 divergence-preserved\n")
		for i := 0; i < n; i++ {
			_, _ = io.WriteString(out, "  unconverted reference (no Codex-side mapping — report to the harness maintainers): fixture/asset-"+strings.Repeat("x", i%3+1)+"\n")
		}
		return nil
	}
}

// checklistBudgetCounts splits the collapsed lines into the checklist-drawn
// bucket C, the five direct-print fragments, and the user-asset pass-through
// lines (the ensure's own output, not drawn by the checklist).
func checklistBudgetCounts(lines []string) (c, direct, passthrough int) {
	for _, l := range lines {
		isDirect := false
		for _, frag := range checklistDirectFragments {
			if strings.Contains(l, frag) {
				isDirect = true
				break
			}
		}
		switch {
		case isDirect:
			direct++
		case strings.Contains(l, checklistPassThroughFragment):
			passthrough++
		}
	}
	c = len(lines) - direct - passthrough
	return c, direct, passthrough
}

// TestChecklistScreenBudget is AC-TCL-011: the lines the checklist draws stay
// within the budget, the five direct lines neither grow nor shrink against the
// M0 baseline, and the collapsed output equals the rich golden.
func TestChecklistScreenBudget(t *testing.T) {
	root, stdout, stderr, err := checklistRichRun(t, nil)
	if err != nil {
		t.Fatalf("init: %v\n%s", err, stderr)
	}
	cOut, cErr := checklistCollapse(stdout), checklistCollapse(stderr)

	// The M0 baseline of the same fixture, normalized the way its goldens were.
	readBaseline := func(name string) string {
		b, rerr := os.ReadFile(filepath.Join(checklistFallbackGoldenDir(), name+".golden"))
		if rerr != nil {
			t.Fatalf("baseline golden %s: %v", name, rerr)
		}
		return string(b)
	}
	baseOut, baseErr := readBaseline("init-clean-stdout"), readBaseline("init-clean-stderr")
	richOutN := checklistFallbackNormalize(cOut, root)
	richErrN := checklistFallbackNormalize(cErr, root)

	// The "init" budget runs on a fixed user-asset fixture, never on the embedded
	// catalog, so the count cannot drift with the catalog (394 installed / 30
	// unconverted today). The ensure seam prints the header (one of the five
	// direct fragments) and n pass-through lines, exactly as writeInstallSummary
	// does for a real run. n=30 mirrors the unconverted count the real catalog
	// produces at this commit and is large enough that counting those lines in C
	// would break the budget (23 + 30 > 24): a calculation that mixed them in
	// cannot pass.
	for _, tc := range []struct {
		name string
		n    int
	}{
		{name: "init_no_unconverted", n: 0},
		{name: "init", n: 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stubEnsure(t, fixedUserAssetEnsure(tc.n))
			_, so, se, rerr := checklistRichRun(t, nil)
			if rerr != nil {
				t.Fatalf("init: %v\n%s", rerr, se)
			}
			all := append(checklistLines(checklistCollapse(so)), checklistLines(checklistCollapse(se))...)
			c, direct, passthrough := checklistBudgetCounts(all)
			if c > 24 {
				t.Errorf("checklist-drawn lines C = %d (total %d, direct %d, pass-through %d), want <= 24", c, len(all), direct, passthrough)
			}
			if direct != 5 {
				t.Errorf("direct lines = %d, want 5", direct)
			}
			if passthrough != tc.n {
				t.Errorf("pass-through (user-asset ensure) lines = %d, want %d", passthrough, tc.n)
			}
			if tc.n > 0 {
				// Guard on the fixture: the mutant calculation (every line outside
				// the five direct fragments, pass-through included) must break the
				// budget here, or the fixture proves nothing.
				if mc := len(all) - direct; mc <= 24 {
					t.Errorf("fixture too small to discriminate: mutant C = %d <= 24", mc)
				}
			}
		})
	}

	t.Run("direct_lines_unchanged", func(t *testing.T) {
		for _, frag := range checklistDirectFragments {
			rich := checklistFragmentLines(frag, richOutN, richErrN)
			base := checklistFragmentLines(frag, baseOut, baseErr)
			if len(rich) != 1 || len(base) != 1 {
				t.Errorf("fragment %q: rich %d times, baseline %d times, want 1 and 1", frag, len(rich), len(base))
				continue
			}
			if rich[0] != base[0] {
				t.Errorf("fragment %q changed text\n  baseline: %q\n  rich:     %q", frag, base[0], rich[0])
			}
		}
	})

	t.Run("rich_golden", func(t *testing.T) {
		checklistRichCompare(t, "init-clean", root, checklistCombined(stdout, stderr))
	})

	t.Run("no_raw_keys", func(t *testing.T) {
		if strings.Contains(cOut+cErr, "checklist.") {
			t.Errorf("a raw catalogue key reached the output:\n%s", cErr)
		}
		entries, derr := os.ReadDir(filepath.Dir(checklistRichGoldenPath("x")))
		if derr != nil {
			t.Fatalf("rich golden dir: %v", derr)
		}
		if len(entries) == 0 {
			t.Fatal("no rich goldens found: an empty swept set asserts nothing")
		}
		for _, e := range entries {
			b, rerr := os.ReadFile(filepath.Join(filepath.Dir(checklistRichGoldenPath("x")), e.Name()))
			if rerr != nil {
				t.Fatalf("read %s: %v", e.Name(), rerr)
			}
			if strings.Contains(string(b), "checklist.") {
				t.Errorf("rich golden %s carries a raw catalogue key", e.Name())
			}
		}
	})
}

// checklistGateRun runs the fixed init with the activation predicate driven by
// the two input seams (never the force-on seam) and returns the raw streams.
func checklistGateRun(t *testing.T, overrides map[string]string, setup func(*testing.T), tty func(out, errBuf io.Writer) func(io.Writer) bool, replaced bool) (root, stdout, stderr string) {
	t.Helper()
	root, _ = checklistFallbackEnv(t)
	t.Setenv("TERM", "xterm-256color")
	if setup != nil {
		setup(t)
	}
	cmd, out, errBuf := newChecklistInitCmd(t, overrides)

	origTTY, origStreams := checklistIsTTYFn, checklistProcessStreamsFn
	t.Cleanup(func() { checklistIsTTYFn, checklistProcessStreamsFn = origTTY, origStreams })
	checklistIsTTYFn = tty(out, errBuf)
	if replaced {
		// The command writers are buffers that are not the process streams.
		xo, xe := &bytes.Buffer{}, &bytes.Buffer{}
		checklistProcessStreamsFn = func() (io.Writer, io.Writer) { return xo, xe }
	} else {
		checklistProcessStreamsFn = func() (io.Writer, io.Writer) { return out, errBuf }
	}
	if err := runInit(cmd, nil); err != nil {
		t.Fatalf("init: %v\n%s", err, errBuf.String())
	}
	return root, out.String(), errBuf.String()
}

func allTTY(_, _ io.Writer) func(io.Writer) bool { return func(io.Writer) bool { return true } }

// requireFallbackOutput asserts the run equals the M0 clean baseline and drew
// no escape sequence.
func requireFallbackOutput(t *testing.T, root, stdout, stderr string) {
	t.Helper()
	checklistFallbackCompare(t, "init-clean-stdout", root, stdout)
	checklistFallbackCompare(t, "init-clean-stderr", root, stderr)
	for name, s := range map[string]string{"stdout": stdout, "stderr": stderr} {
		if strings.Contains(s, "\x1b[") {
			t.Errorf("%s of a fallback run carries an escape sequence", name)
		}
	}
}

// TestChecklistGateCommandLevel is AC-TCL-012: every activation condition, taken
// away one at a time, keeps the output byte-for-byte on the baseline path.
func TestChecklistGateCommandLevel(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(t *testing.T)
		tty      func(out, errBuf io.Writer) func(io.Writer) bool
		replaced bool
	}{
		{name: "init_stdout_not_tty", tty: func(out, _ io.Writer) func(io.Writer) bool { return func(w io.Writer) bool { return w != out } }},
		{name: "init_stderr_not_tty", tty: func(_, errBuf io.Writer) func(io.Writer) bool { return func(w io.Writer) bool { return w != errBuf } }},
		{name: "init_ci_set", setup: func(t *testing.T) { t.Setenv("CI", "true") }, tty: allTTY},
		{name: "init_term_dumb", setup: func(t *testing.T) { t.Setenv("TERM", "dumb") }, tty: allTTY},
		{name: "init_writer_replaced", tty: allTTY, replaced: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, so, se := checklistGateRun(t, nil, tc.setup, tc.tty, tc.replaced)
			requireFallbackOutput(t, r, so, se)
		})
	}
	t.Run("control_all_true", func(t *testing.T) {
		_, so, se := checklistGateRun(t, nil, nil, allTTY, false)
		if !strings.Contains(se, "MoAI-ADK init") {
			t.Errorf("control run drew no checklist header:\n%s", se)
		}
		if !strings.Contains(so+se, "\x1b[") {
			t.Errorf("control run drew no escape sequence")
		}
	})
}

// TestChecklistVerboseForcesFallback is AC-TCL-014 (init): --verbose keeps the
// baseline output even when every other condition holds; without it the
// checklist draws.
func TestChecklistVerboseForcesFallback(t *testing.T) {
	t.Run("flag_registered", func(t *testing.T) {
		if initCmd.Flags().Lookup("verbose") == nil {
			t.Fatal("init does not register --verbose")
		}
		if !strings.Contains(initCmd.Flags().FlagUsages(), "--verbose") {
			t.Errorf("init usage does not list --verbose:\n%s", initCmd.Flags().FlagUsages())
		}
	})
	t.Run("init", func(t *testing.T) {
		root, so, se := checklistGateRun(t, map[string]string{"verbose": "true"}, nil, allTTY, false)
		requireFallbackOutput(t, root, so, se)
	})
	t.Run("init_control_without_verbose", func(t *testing.T) {
		_, _, se := checklistGateRun(t, nil, nil, allTTY, false)
		if !strings.Contains(se, "MoAI-ADK init") {
			t.Errorf("control run without --verbose drew no checklist:\n%s", se)
		}
	})
}

// injectingExecutor wraps the real init executor and runs inject right after
// the named step starts, so a test can write while the matching checklist item
// is open.
type injectingExecutor struct {
	inner initPhaseExecutor
	step  string
	fn    func()
}

type injectingReporter struct {
	project.ProgressReporter
	step string
	fn   func()
}

func (r *injectingReporter) StepStart(name, message string) {
	r.ProgressReporter.StepStart(name, message)
	if name == r.step && r.fn != nil {
		r.fn()
	}
}

func (e *injectingExecutor) SetReporter(r project.ProgressReporter) {
	e.inner.SetReporter(&injectingReporter{ProgressReporter: r, step: e.step, fn: e.fn})
}

func (e *injectingExecutor) Execute(ctx context.Context, opts project.InitOptions) (*project.InitResult, error) {
	return e.inner.Execute(ctx, opts)
}

// injectIntoStep makes the real executor call fn right after step starts.
func injectIntoStep(t *testing.T, step string, fn func()) {
	t.Helper()
	orig := newInitPhaseExecutorFn
	newInitPhaseExecutorFn = func(d project.Detector, m project.MethodologyDetector, v project.ProjectValidator, i project.Initializer) initPhaseExecutor {
		return &injectingExecutor{inner: orig(d, m, v, i), step: step, fn: fn}
	}
	t.Cleanup(func() { newInitPhaseExecutorFn = orig })
}
