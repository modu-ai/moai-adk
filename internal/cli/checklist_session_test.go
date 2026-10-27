package cli

// checklist_session_test.go — SPEC-TUI-CHECKLIST-001 M3: the init plan table,
// the activation predicate, the capture rules and the session end. Every test
// that changes a package seam runs without t.Parallel and restores it through
// t.Cleanup (spec.md §4 C-5).

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// forceChecklistOn turns the activation predicate on as a whole, the seam the
// rich-golden and wiring tests use (design.md §3.1).
func forceChecklistOn(t *testing.T) {
	t.Helper()
	orig := checklistForceOnFn
	checklistForceOnFn = func() bool { return true }
	t.Cleanup(func() { checklistForceOnFn = orig })
}

// checklistRig is a command with buffer writers, a started session and the
// collector the session hands out — the pieces runInit holds in R1.
type checklistRig struct {
	cmd     *cobra.Command
	out     *bytes.Buffer
	errBuf  *bytes.Buffer
	sess    *checklistSession
	collect *warnCollector
}

func newChecklistRig(t *testing.T) *checklistRig {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	// The session reads the profile locale for the catalogue language; keep a
	// developer's real profile out of the labels.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	forceChecklistOn(t)
	cmd := &cobra.Command{Use: "init"}
	cmd.Flags().Bool("verbose", false, "")
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	sess := beginChecklistSession(cmd, nil)
	collect := sess.collector(printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr())))
	return &checklistRig{cmd: cmd, out: out, errBuf: errBuf, sess: sess, collect: collect}
}

// all returns both real writers' bytes joined, the "both output writers" view.
func (r *checklistRig) all() string { return r.out.String() + r.errBuf.String() }

func TestChecklistPlanGroupCap(t *testing.T) {
	if len(initChecklistPlan) == 0 {
		t.Fatal("init plan table is empty: an empty swept set asserts nothing")
	}
	for _, g := range initChecklistPlan {
		if n := len(g.Items); n < 1 || n > 8 {
			t.Errorf("group %q has %d items, want 1..8", g.Key, n)
		}
		gk := "checklist.group." + g.Key
		if got := tui.Translate(gk, "en"); got == gk || got == "" {
			t.Errorf("group key %q does not resolve in the en catalogue (got %q)", gk, got)
		}
		for _, it := range g.Items {
			ik := "checklist.item." + g.Key + "." + it
			if got := tui.Translate(ik, "en"); got == ik || got == "" {
				t.Errorf("item key %q does not resolve in the en catalogue (got %q)", ik, got)
			}
		}
	}
}

func TestChecklistActivationPredicate(t *testing.T) {
	allTrue := checklistInputs{
		StdoutTTY: true, StderrTTY: true, ProcessStreams: true,
		CIUnset: true, TermNotDumb: true, VerboseOff: true,
	}
	cases := []struct {
		name   string
		mutate func(*checklistInputs)
		want   bool
	}{
		{"all_true", func(*checklistInputs) {}, true},
		{"stdout_not_tty", func(in *checklistInputs) { in.StdoutTTY = false }, false},
		{"stderr_not_tty", func(in *checklistInputs) { in.StderrTTY = false }, false},
		{"writer_replaced", func(in *checklistInputs) { in.ProcessStreams = false }, false},
		{"ci_set", func(in *checklistInputs) { in.CIUnset = false }, false},
		{"term_dumb", func(in *checklistInputs) { in.TermNotDumb = false }, false},
		{"verbose_on", func(in *checklistInputs) { in.VerboseOff = false }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := allTrue
			tc.mutate(&in)
			if got := checklistActive(in); got != tc.want {
				t.Errorf("checklistActive(%+v) = %v, want %v", in, got, tc.want)
			}
		})
	}
}

// TestChecklistCapturePassThrough: bytes written while no item is open reach
// the real writer at once and unchanged; bytes written while an item is open
// stay off the screen on success and are expanded when the item fails.
func TestChecklistCapturePassThrough(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := newChecklistRig(t)
		r.sess.openItem("templates", "detect")
		before := r.errBuf.Len()
		_, _ = r.cmd.ErrOrStderr().Write([]byte("inside\n"))
		if r.errBuf.Len() != before {
			t.Fatalf("a byte written inside an open item reached the real writer: %q", r.errBuf.String())
		}
		r.sess.closeOpenItem()

		before = r.errBuf.Len()
		_, _ = r.cmd.ErrOrStderr().Write([]byte("outside\n"))
		if got := r.errBuf.String()[before:]; got != "outside\n" {
			t.Fatalf("outside bytes must pass through immediately and unchanged, got %q", got)
		}
		r.sess.end(nil)
		if strings.Contains(r.all(), "inside") {
			t.Errorf("inside must stay collapsed after a successful item, output:\n%s", r.errBuf.String())
		}
	})
	t.Run("fail_expands", func(t *testing.T) {
		r := newChecklistRig(t)
		r.sess.openItem("templates", "detect")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("inside\n"))
		r.sess.failOpenItem(errors.New("boom"))
		if strings.Count(r.all(), "inside") != 1 {
			t.Errorf("inside must be expanded once under a failed item, output:\n%s", r.errBuf.String())
		}
		r.sess.end(nil)
	})
	t.Run("partial_line_is_flushed_as_detail", func(t *testing.T) {
		r := newChecklistRig(t)
		r.sess.openItem("templates", "detect")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("no-newline"))
		r.sess.failOpenItem(errors.New("boom"))
		if strings.Count(r.all(), "no-newline") != 1 {
			t.Errorf("a partial line recorded in an item must surface on failure, output:\n%s", r.errBuf.String())
		}
		r.sess.end(nil)
	})
}

// TestChecklistSessionRestoresWriters: a finished session hands the original
// writers back, whatever way the run ended.
func TestChecklistSessionRestoresWriters(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{{"nil", nil}, {"error", errors.New("boom")}} {
		t.Run(tc.name, func(t *testing.T) {
			r := newChecklistRig(t)
			if r.cmd.ErrOrStderr() == r.errBuf || r.cmd.OutOrStdout() == r.out {
				t.Fatal("an active session must replace both command writers")
			}
			r.sess.end(tc.err)
			if r.cmd.ErrOrStderr() != r.errBuf || r.cmd.OutOrStdout() != r.out {
				t.Error("end() must restore the original command writers")
			}
		})
	}
}

// TestChecklistInactiveSessionIsNoOp: with any activation condition false the
// session replaces nothing and draws nothing.
func TestChecklistInactiveSessionIsNoOp(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cmd := &cobra.Command{Use: "init"}
	cmd.Flags().Bool("verbose", false, "")
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	sess := beginChecklistSession(cmd, nil)
	if cmd.ErrOrStderr() != errBuf || cmd.OutOrStdout() != out {
		t.Fatal("an inactive session must leave the command writers alone")
	}
	sess.openItem("templates", "detect")
	sess.skipItem("hooks", "prepush")
	sess.end(errors.New("boom"))
	if out.Len() != 0 || errBuf.Len() != 0 {
		t.Errorf("an inactive session drew output: stdout=%q stderr=%q", out.String(), errBuf.String())
	}
}

// TestChecklistAbortFlushes: a run that ends without a summary still shows its
// warnings exactly once and closes its items truthfully. The subtests here
// drive the session directly; the runInit-level subtests join in M4.
func TestChecklistAbortFlushes(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		r := newChecklistRig(t)
		r.collect.Warn("w-err") // outside any item: an immediate line goes out
		r.sess.openItem("templates", "deploy")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("detail-x\n"))
		r.sess.end(errors.New("deploy blew up"))

		got := r.all()
		for _, want := range []string{"✗", "detail-x", "deploy blew up"} {
			if !strings.Contains(got, want) {
				t.Errorf("output lacks %q:\n%s", want, got)
			}
		}
		if n := strings.Count(got, "w-err"); n != 1 {
			t.Errorf("w-err appears %d times, want exactly 1 (the immediate line; the session end must not rewrite it)", n)
		}
		for _, banned := range []string{"warning(s) during init:", "Done"} {
			if strings.Contains(got, banned) {
				t.Errorf("output must not contain %q on an aborted run:\n%s", banned, got)
			}
		}
	})
	t.Run("error_collect_after_fail", func(t *testing.T) {
		r := newChecklistRig(t)
		rep := newChecklistReporter(r.sess)
		rep.StepStart("Initialization", "Creating project structure")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("detail-x\n"))
		rep.StepError(errors.New("template deployment failed"))
		r.collect.Collect("c-err") // no item is open now: Collect writes nothing by itself
		r.sess.end(errors.New("template deployment failed"))

		got := r.all()
		if n := strings.Count(got, "c-err"); n != 1 {
			t.Errorf("c-err appears %d times, want exactly 1 (written by the session end):\n%s", n, got)
		}
		if !strings.Contains(got, "✗") || !strings.Contains(got, "detail-x") {
			t.Errorf("failed item and its detail missing:\n%s", got)
		}
		if strings.Contains(got, "warning(s) during init:") {
			t.Errorf("no panel on an aborted run:\n%s", got)
		}
	})
	t.Run("nil_return_open_item", func(t *testing.T) {
		r := newChecklistRig(t)
		r.collect.Warn("w-nil")
		r.collect.Collect("c-nil")
		r.sess.openItem("templates", "deploy")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("detail-e\n"))
		end := beginPreserved(r.cmd.ErrOrStderr(), preservedNotice)
		_, _ = r.cmd.ErrOrStderr().Write([]byte("keep-e\n"))
		end()
		r.sess.end(nil)

		got := r.all()
		if !strings.Contains(got, "✓") {
			t.Errorf("an item closed by a nil return must end with the done glyph:\n%s", got)
		}
		for _, banned := range []string{"✗", "detail-e", "Done"} {
			if strings.Contains(got, banned) {
				t.Errorf("output must not contain %q:\n%s", banned, got)
			}
		}
		for _, once := range []string{"keep-e", "w-nil", "c-nil"} {
			if n := strings.Count(got, once); n != 1 {
				t.Errorf("%q appears %d times, want exactly 1:\n%s", once, n, got)
			}
		}
	})
	t.Run("error_after_finished_items", func(t *testing.T) {
		r := newChecklistRig(t)
		r.sess.openItem("templates", "detect")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("fold-a\n"))
		r.sess.closeOpenItem() // explicit Done
		r.sess.openItem("templates", "validate")
		_, _ = r.cmd.ErrOrStderr().Write([]byte("fold-b\n"))
		r.sess.openItem("templates", "deploy") // implicit Done of validate
		_, _ = r.cmd.ErrOrStderr().Write([]byte("fold-c\n"))
		r.sess.end(errors.New("boom"))

		got := r.all()
		for _, once := range []string{"fold-a", "fold-b", "fold-c"} {
			if n := strings.Count(got, once); n != 1 {
				t.Errorf("%q appears %d times, want exactly 1:\n%s", once, n, got)
			}
		}
		if a, b := strings.Index(got, "fold-a"), strings.Index(got, "fold-b"); a > b {
			t.Errorf("finished items' detail out of item order:\n%s", got)
		}
		for _, label := range []string{checklistLabel("templates.detect"), checklistLabel("templates.validate")} {
			if n := strings.Count(checklistCollapse(got), label); n != 1 {
				t.Errorf("finished line %q drawn %d times, want 1 (no redraw):\n%s", label, n, got)
			}
		}
	})
	// The two runInit-level subtests (SPEC-TUI-CHECKLIST-001 M4) live with the
	// init wiring tests.
	t.Run("cancel", checklistAbortCancel)
	t.Run("resumed_after_validate_fail", checklistAbortResumed)
}

// TestChecklistSummaryListsOutsideWarnings: a warning raised outside any item
// is listed once by the summary block, and the session end does not write it a
// second time; a Collect-origin warning (never shown) is listed too.
func TestChecklistSummaryListsOutsideWarnings(t *testing.T) {
	r := newChecklistRig(t)
	r.sess.openItem("templates", "deploy")
	r.sess.closeOpenItem()
	r.collect.Warn("w-out")
	r.collect.Collect("c-out")
	two := 2
	r.sess.summary(tui.ChecklistSummary{CreatedFiles: &two, CreatedDirs: &two, Warnings: r.collect.Count()})
	r.sess.end(nil)

	got := r.all()
	if n := strings.Count(got, "w-out"); n != 2 {
		t.Errorf("w-out appears %d times, want 2 (immediate line + summary list):\n%s", n, got)
	}
	if n := strings.Count(got, "c-out"); n != 1 {
		t.Errorf("c-out appears %d times, want 1 (summary list only):\n%s", n, got)
	}
	if !strings.Contains(got, "2 warnings") {
		t.Errorf("summary headline lacks the warning count:\n%s", got)
	}
	if strings.Contains(got, "warning(s) during init:") {
		t.Errorf("no panel while the checklist is active:\n%s", got)
	}
}

// TestChecklistReporterMapping: the executor's step names map onto the plan
// table's templates items, and a repeated Validation start keeps one line.
func TestChecklistReporterMapping(t *testing.T) {
	r := newChecklistRig(t)
	rep := newChecklistReporter(r.sess)
	rep.StepStart("Detection", "Analyzing project structure")
	rep.StepComplete("done")
	rep.StepStart("Validation", "Using mode: tdd")
	rep.StepComplete("")
	rep.StepStart("Validation", "Validating project structure")
	rep.StepStart("Initialization", "Creating project structure")
	rep.StepUpdate("writing")
	rep.StepComplete("Created 2 files")
	r.sess.end(nil)

	got := r.errBuf.String()
	labels := []string{
		tui.Translate("checklist.item.templates.detect", "en"),
		tui.Translate("checklist.item.templates.validate", "en"),
		tui.Translate("checklist.item.templates.deploy", "en"),
	}
	for _, l := range labels {
		if n := strings.Count(got, "✓ "+l); n != 1 {
			t.Errorf("finished line for %q appears %d times, want 1:\n%s", l, n, got)
		}
	}
	if strings.Contains(got, "Using mode") || strings.Contains(got, "writing") {
		t.Errorf("step messages are details and must stay collapsed on success:\n%s", got)
	}
}
