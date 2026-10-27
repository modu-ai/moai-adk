package cli

// checklist_wiring_warn_test.go — SPEC-TUI-CHECKLIST-001 M4, second part: the
// warning placement, the abort and resume paths, the channel split, the prompt
// state and the counts of the `moai init` wiring (AC-TCL-016..020). Every test
// runs the real runInit through buffer writers; the injection points are the two
// test seams of the card, never a change to init.go.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// warnInjector records the command and collector the session hands out and
// writes the outside-of-item warnings at that moment (before any item).
type warnInjector struct {
	cmd *cobra.Command
	c   *warnCollector
}

// installWarnInjection sets checklistSessionStartFn to record the pair and call
// outside, and wraps the real executor so inside runs right after the
// "Initialization" step starts (templates.deploy open). Both seams restore on
// cleanup.
func installWarnInjection(t *testing.T, outside func(c *warnCollector), inside func(cmd *cobra.Command, c *warnCollector)) *warnInjector {
	t.Helper()
	inj := &warnInjector{}
	orig := checklistSessionStartFn
	checklistSessionStartFn = func(cmd *cobra.Command, c *warnCollector) {
		inj.cmd, inj.c = cmd, c
		if outside != nil {
			outside(c)
		}
	}
	t.Cleanup(func() { checklistSessionStartFn = orig })
	if inside != nil {
		injectIntoStep(t, "Initialization", func() { inside(inj.cmd, inj.c) })
	}
	return inj
}

// legacyInitRun runs the fixed init with the activation predicate off (the
// buffers are not terminals) and returns the root and the two raw streams.
func legacyInitRun(t *testing.T) (root, stdout, stderr string, err error) {
	t.Helper()
	root, _ = checklistFallbackEnv(t)
	cmd, out, errBuf := newChecklistInitCmd(t, nil)
	err = runInit(cmd, nil)
	return root, out.String(), errBuf.String(), err
}

// TestChecklistWarningsExactlyOnce is AC-TCL-016.
func TestChecklistWarningsExactlyOnce(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		installWarnInjection(t,
			func(c *warnCollector) { c.Warn("zz-w-out"); c.Collect("zz-c-out") },
			func(cmd *cobra.Command, c *warnCollector) {
				c.Warn("zz-w-in")
				c.Collect("zz-c-in")
				emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "zz-s-in")
			})
		_, stdout, stderr, err := checklistRichRun(t, nil)
		if err != nil {
			t.Fatalf("init: %v\n%s", err, stderr)
		}
		all := checklistCollapse(stdout) + checklistCollapse(stderr)
		for token, want := range map[string]int{"zz-w-in": 1, "zz-c-in": 1, "zz-s-in": 1, "zz-w-out": 2, "zz-c-out": 1} {
			if n := strings.Count(all, token); n != want {
				t.Errorf("%q appears %d times, want %d:\n%s", token, n, want, checklistCollapse(stderr))
			}
		}
		if strings.Contains(all, "warning(s) during init:") {
			t.Errorf("the legacy warning panel header appeared on an active run")
		}
		if !strings.Contains(checklistCollapse(stderr), "· 4 warnings ·") {
			t.Errorf("summary does not report 4 warnings (collector count, s-in not counted):\n%s", checklistCollapse(stderr))
		}
	})

	t.Run("init_inactive_control", func(t *testing.T) {
		installWarnInjection(t,
			func(c *warnCollector) { c.Warn("zz-w-out"); c.Collect("zz-c-out") },
			func(cmd *cobra.Command, c *warnCollector) {
				c.Warn("zz-w-in")
				c.Collect("zz-c-in")
				emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "zz-s-in")
			})
		_, stdout, stderr, err := legacyInitRun(t)
		if err != nil {
			t.Fatalf("init: %v\n%s", err, stderr)
		}
		all := stdout + stderr
		if n := strings.Count(all, "warning(s) during init:"); n != 1 {
			t.Errorf("legacy panel header appears %d times, want 1:\n%s", n, stderr)
		}
		if !strings.Contains(all, "4 warning(s) collected") {
			t.Errorf("legacy card does not carry the collector count 4:\n%s", all)
		}
	})

	t.Run("init_stdout_severity", func(t *testing.T) {
		installWarnInjection(t, nil, func(cmd *cobra.Command, _ *warnCollector) {
			emitSeverityLine(cmd.OutOrStdout(), sevErr, resolveTheme(), "zz-s-out")
		})
		_, stdout, stderr, err := checklistRichRun(t, nil)
		if err != nil {
			t.Fatalf("init: %v\n%s", err, stderr)
		}
		cOut, cErr := checklistCollapse(stdout), checklistCollapse(stderr)
		if n := strings.Count(cOut+cErr, "zz-s-out"); n != 1 {
			t.Fatalf("s-out appears %d times across both writers, want 1:\n%s", n, cErr)
		}
		lines := checklistLines(cErr)
		label := checklistLabel("templates.deploy")
		found := false
		for i, l := range lines {
			if !strings.Contains(l, label) {
				continue
			}
			found = true
			if !strings.Contains(l, "!") || strings.Contains(l, "✓") {
				t.Errorf("deploy item line %q must end with the warning glyph and not the done glyph", l)
			}
			if i+1 >= len(lines) || !strings.Contains(lines[i+1], "zz-s-out") {
				t.Errorf("s-out is not directly under the deploy item line:\n%s", cErr)
			}
		}
		if !found {
			t.Fatalf("deploy item line not found:\n%s", cErr)
		}
	})
}

// checklistAbortCancel is TestChecklistAbortFlushes/cancel (AC-TCL-017 b).
func checklistAbortCancel(t *testing.T) {
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
	runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) { return nil, wizard.ErrCancelled }
	t.Cleanup(func() { runWizardFn = origWizard })

	cmd, out, errBuf := newChecklistInitCmd(t, map[string]string{"non-interactive": "false"})
	origOut, origErr := cmd.OutOrStdout(), cmd.ErrOrStderr()
	if err := runInit(cmd, nil); err != nil {
		t.Fatalf("a cancelled wizard must return nil, got %v", err)
	}
	all := out.String() + errBuf.String()
	if n := strings.Count(all, "Initialization cancelled."); n != 1 {
		t.Errorf("cancel message appears %d times, want 1:\n%s", n, all)
	}
	if n := strings.Count(all, "MoAI-ADK init"); n != 0 {
		t.Errorf("a run with no opened item drew the checklist header %d times, want 0", n)
	}
	if cmd.OutOrStdout() != origOut || cmd.ErrOrStderr() != origErr {
		t.Errorf("the command writers were not handed back after the cancelled run")
	}
}

// checklistAbortResumed is TestChecklistAbortFlushes/resumed_after_validate_fail
// (AC-TCL-017 c): the real event order of a resume — the validator reports its
// error, then the resume runs and the command returns nil.
func checklistAbortResumed(t *testing.T) {
	resumeHome(t)
	projectDir := filepath.Join(t.TempDir(), "resume-proj")
	resumeSeedRun(t, projectDir, "claude")

	forceChecklistOn(t)
	checklistRichClock(t)
	t.Setenv("NO_COLOR", "1")
	installWarnInjection(t, func(c *warnCollector) { c.Warn("zz-w-resume"); c.Collect("zz-c-resume") }, nil)

	cmd := newInitTestCmd()
	cmd.Flags().Bool("verbose", false, "")
	if err := cmd.Flags().Set("llm", "claude"); err != nil {
		t.Fatal(err)
	}
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	origOut, origErr := cmd.OutOrStdout(), cmd.ErrOrStderr()
	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("the resume must return nil, got %v\n%s", err, errBuf.String())
	}
	rawErr := errBuf.String()
	cErr := checklistCollapse(rawErr)
	all := checklistCollapse(out.String()) + cErr

	lines := checklistLines(cErr)
	detect, validate := -1, -1
	for i, l := range lines {
		switch {
		case strings.Contains(l, checklistLabel("templates.detect")):
			detect = i
		case strings.Contains(l, checklistLabel("templates.validate")):
			validate = i
		}
	}
	if detect < 0 || validate < 0 {
		t.Fatalf("detect/validate item lines missing:\n%s", cErr)
	}
	if !strings.Contains(lines[detect], "✓") {
		t.Errorf("detect line %q must end with the done glyph", lines[detect])
	}
	if !strings.Contains(lines[validate], "✗") {
		t.Errorf("validate line %q must end with the failed glyph", lines[validate])
	}
	// The failure expands under its item: the indented lines that follow carry
	// the error text (a detail line such as "Mode validated" may come first).
	indentedErr := false
	for _, l := range lines[validate+1:] {
		if !strings.HasPrefix(l, "    ") {
			break
		}
		if strings.Contains(l, "project already initialized") {
			indentedErr = true
		}
	}
	if !indentedErr {
		t.Errorf("the validate failure is not followed by the indented error text:\n%s", cErr)
	}
	for token, want := range map[string]int{"✗": 1, "Initialized MoAI project (resumed": 1, "zz-w-resume": 1, "zz-c-resume": 1, "warning(s) during init:": 0, "Done ·": 0} {
		if n := strings.Count(all, token); n != want {
			t.Errorf("%q appears %d times, want %d:\n%s", token, n, want, cErr)
		}
	}
	if i := strings.Index(rawErr, "✗"); i >= 0 && strings.Contains(rawErr[i:], "\r\x1b[2K") {
		t.Errorf("the failed line was redrawn after it was written")
	}
	if cmd.OutOrStdout() != origOut || cmd.ErrOrStderr() != origErr {
		t.Errorf("the command writers were not handed back after the resumed run")
	}
}

// TestChecklistChannelPartition is AC-TCL-018.
func TestChecklistChannelPartition(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		_, stdout, stderr, err := checklistRichRun(t, nil)
		if err != nil {
			t.Fatalf("init: %v\n%s", err, stderr)
		}
		cOut, cErr := checklistCollapse(stdout), checklistCollapse(stderr)
		if n := strings.Count(cOut, "MoAI-ADK init"); n != 0 {
			t.Errorf("stdout carries the checklist header %d times, want 0", n)
		}
		if n := strings.Count(cErr, "MoAI-ADK init"); n != 1 {
			t.Errorf("stderr carries the checklist header %d times, want 1", n)
		}
		for _, key := range []string{"templates.detect", "templates.validate", "templates.deploy", "config.autonomy", "verify.manifest"} {
			label := checklistLabel(key)
			if !strings.Contains(cErr, label) {
				t.Errorf("item %q is missing from stderr", label)
			}
			if strings.Contains(cOut, label) {
				t.Errorf("item %q leaked onto stdout", label)
			}
		}
	})
}

var checklistOpenSpinnerRe = regexp.MustCompile(`[⠀-⣿]`)

// TestChecklistNoOpenLineDuringPrompt is AC-TCL-019 (init): when the wizard seam
// is called no item is open and no unfinished spinner line is on the screen.
func TestChecklistNoOpenLineDuringPrompt(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		checklistFallbackEnv(t)
		forceChecklistOn(t)
		checklistRichClock(t)
		origInteractive := isInteractiveStdin
		isInteractiveStdin = func() bool { return true }
		t.Cleanup(func() { isInteractiveStdin = origInteractive })
		origDeps := deps
		deps = nil
		t.Cleanup(func() { deps = origDeps })

		var inj warnInjector
		origStart := checklistSessionStartFn
		checklistSessionStartFn = func(cmd *cobra.Command, c *warnCollector) { inj.cmd, inj.c = cmd, c }
		t.Cleanup(func() { checklistSessionStartFn = origStart })

		cmd, _, errBuf := newChecklistInitCmd(t, map[string]string{"non-interactive": "false"})
		var openKey string
		var seenKey, seenSpinner bool
		var rawAtPrompt string
		origWizard := runWizardFn
		t.Cleanup(func() { runWizardFn = origWizard })
		// followWizardLang wraps runWizardFn at session start, so the probe is the
		// inner function the wrapper calls.
		runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
			sess := inj.c.checklistSession()
			sess.mu.Lock()
			openKey = sess.curKey
			seenKey = true
			sess.mu.Unlock()
			rawAtPrompt = errBuf.String()
			seenSpinner = checklistOpenSpinnerRe.MatchString(rawAtPrompt) || strings.Contains(rawAtPrompt, "\r\x1b[2K")
			return &wizard.WizardResult{ConversationLang: "en", DevelopmentMode: "tdd"}, nil
		}
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		if !seenKey {
			t.Fatal("the wizard seam was never called: the probe asserted nothing")
		}
		if openKey != "" {
			t.Errorf("an item (%q) was open when the wizard seam was called", openKey)
		}
		if seenSpinner || rawAtPrompt != "" && strings.Contains(rawAtPrompt, "MoAI-ADK init") {
			t.Errorf("the screen carried a checklist line when the wizard seam was called:\n%q", rawAtPrompt)
		}
	})
}

// checklistCountsOf extracts the numbers the completion output carries.
type checklistCounts struct {
	dirs, files, warnings string
	disk                  []string
	err                   error
}

var (
	legacyDirsRe  = regexp.MustCompile(`(\d+) dirs`)
	legacyFilesRe = regexp.MustCompile(`(\d+) files`)
	legacyWarnRe  = regexp.MustCompile(`(\d+) warning\(s\) collected`)
	activeFilesRe = regexp.MustCompile(`(\d+) files created`)
	activeDirsRe  = regexp.MustCompile(`(\d+) directories`)
	activeWarnRe  = regexp.MustCompile(`(\d+) warnings`)
)

func firstGroup(t *testing.T, re *regexp.Regexp, s string) string {
	t.Helper()
	m := re.FindStringSubmatch(s)
	if m == nil {
		t.Fatalf("no match for %s in:\n%s", re, s)
	}
	return m[1]
}

func diskListing(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.Walk(dir, func(p string, info os.FileInfo, werr error) error {
		if werr != nil {
			return werr
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		if !info.IsDir() {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	sort.Strings(files)
	return files
}

// TestChecklistSummaryMatchesLegacyCounts is AC-TCL-020.
func TestChecklistSummaryMatchesLegacyCounts(t *testing.T) {
	inject := func(t *testing.T) {
		installWarnInjection(t,
			func(c *warnCollector) { c.Warn("zz-w-count"); c.Collect("zz-c-count") },
			func(cmd *cobra.Command, _ *warnCollector) {
				emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "zz-s-count")
			})
	}
	var legacy, active checklistCounts

	t.Run("init_legacy", func(t *testing.T) {
		inject(t)
		root, stdout, stderr, err := legacyInitRun(t)
		legacy.err = err
		all := stdout + stderr
		legacy.dirs = firstGroup(t, legacyDirsRe, all)
		legacy.files = firstGroup(t, legacyFilesRe, all)
		legacy.warnings = firstGroup(t, legacyWarnRe, all)
		legacy.disk = diskListing(t, filepath.Join(root, "proj"))
	})
	t.Run("init_active", func(t *testing.T) {
		inject(t)
		root, stdout, stderr, err := checklistRichRun(t, nil)
		active.err = err
		all := checklistCollapse(stdout) + checklistCollapse(stderr)
		active.dirs = firstGroup(t, activeDirsRe, all)
		active.files = firstGroup(t, activeFilesRe, all)
		active.warnings = firstGroup(t, activeWarnRe, all)
		active.disk = diskListing(t, filepath.Join(root, "proj"))
	})

	t.Run("init", func(t *testing.T) {
		if legacy.dirs == "" || active.dirs == "" {
			t.Fatal("a run did not report its counts: the comparison asserted nothing")
		}
		if legacy.err != nil || active.err != nil {
			t.Errorf("return values differ or failed: legacy %v, active %v", legacy.err, active.err)
		}
		if legacy.dirs != active.dirs || legacy.files != active.files {
			t.Errorf("created counts differ: legacy %s dirs %s files, active %s dirs %s files", legacy.dirs, legacy.files, active.dirs, active.files)
		}
		if legacy.warnings != "2" || active.warnings != "2" {
			t.Errorf("warning counts: legacy %s, active %s, want 2 and 2 (the severity warning is not counted)", legacy.warnings, active.warnings)
		}
		if strings.Join(legacy.disk, "\n") != strings.Join(active.disk, "\n") {
			t.Errorf("deployed file lists differ (legacy %d files, active %d files)", len(legacy.disk), len(active.disk))
		}
		if len(legacy.disk) == 0 {
			t.Error("no deployed files were listed: the comparison asserted nothing")
		}
	})
}
