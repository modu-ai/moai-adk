package cli

// SPEC-CLI-TUX-V3-002 M2d — warning collector + completion card contracts
// (REQ-TUX2-013/014/016, AC-TUX2-012/013/014).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/config"
)

// TestWarningCollector_SummaryPanelOnce asserts the consolidated stderr
// summary: N collected warnings re-emitted exactly once in a single panel
// carrying the count and every message (AC-TUX2-012).
func TestWarningCollector_SummaryPanelOnce(t *testing.T) {
	var out, errBuf bytes.Buffer
	base := printer.New(printer.WithWriters(&out, &errBuf), printer.WithMode(printer.ModePlain))
	wc := newWarnCollector(base)

	wc.Warn("first warning: %s", "alpha")
	wc.Warn("second warning: %s", "beta")
	wc.Collect("third warning: gamma") // collect-only (no individual re-emission)

	if got := wc.Count(); got != 3 {
		t.Fatalf("collector must hold 3 warnings, got %d", got)
	}

	var panel bytes.Buffer
	wc.emitSummary(&panel)
	summary := panel.String()

	for _, want := range []string{"first warning: alpha", "second warning: beta", "third warning: gamma"} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary must carry %q, got:\n%s", want, summary)
		}
	}
	if !strings.Contains(summary, "3") {
		t.Errorf("summary must carry the warning count 3, got:\n%s", summary)
	}

	// Exactly once: a second emission attempt must be a no-op.
	before := panel.Len()
	wc.emitSummary(&panel)
	if panel.Len() != before {
		t.Errorf("summary must be emitted exactly once; second call added output")
	}

	// Individual warnings still stream to stderr as they occur (Warn path),
	// and the collect-only path stays silent until the summary.
	inline := errBuf.String()
	if !strings.Contains(inline, "first warning: alpha") {
		t.Errorf("Warn must still emit the individual stderr line, got %q", inline)
	}
	if strings.Contains(inline, "third warning: gamma") {
		t.Errorf("Collect must not emit an individual line, got %q", inline)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay clean, got %q", out.String())
	}
}

// TestWarningCollector_ZeroWarningsNoPanel asserts the empty-collector edge:
// no warnings -> no panel at all (acceptance.md §C).
func TestWarningCollector_ZeroWarningsNoPanel(t *testing.T) {
	var out, errBuf bytes.Buffer
	base := printer.New(printer.WithWriters(&out, &errBuf), printer.WithMode(printer.ModePlain))
	wc := newWarnCollector(base)

	var panel bytes.Buffer
	wc.emitSummary(&panel)
	if panel.Len() != 0 {
		t.Errorf("zero warnings must emit no panel, got %q", panel.String())
	}
}

// TestInitStdoutCleanWarningChannel asserts the channel discipline
// (REQ-TUX2-014, AC-TUX2-013): warnings never reach stdout — neither the
// individual lines nor the summary panel.
func TestInitStdoutCleanWarningChannel(t *testing.T) {
	var out, errBuf bytes.Buffer
	base := printer.New(printer.WithWriters(&out, &errBuf), printer.WithMode(printer.ModePlain))
	wc := newWarnCollector(base)

	wc.Warn("disk almost full")
	wc.Collect("template skipped")
	wc.emitSummary(&errBuf)

	if strings.Contains(out.String(), "Warning") || out.Len() != 0 {
		t.Errorf("stdout must never carry warning text, got %q", out.String())
	}
	if !strings.Contains(errBuf.String(), "disk almost full") {
		t.Errorf("stderr must carry the warnings, got %q", errBuf.String())
	}
}

// TestCompletionCardNextActions asserts the success card carries the
// next-action sequence and the warning-summary pointer (REQ-TUX2-016,
// AC-TUX2-014).
func TestCompletionCardNextActions(t *testing.T) {
	card := buildInitSuccessCard("myproj", 12, 96, 0)
	for _, want := range []string{"cd myproj", "moai cc", "/moai plan", "12", "96"} {
		if !strings.Contains(card, want) {
			t.Errorf("completion card must carry %q, got:\n%s", want, card)
		}
	}
	if strings.Contains(card, "warning") || strings.Contains(card, "Warning") {
		t.Errorf("zero-warning card must not mention warnings, got:\n%s", card)
	}

	withWarn := buildInitSuccessCard("myproj", 12, 96, 3)
	if !strings.Contains(withWarn, "3 warning") {
		t.Errorf("card must point to the stderr warning summary when warnings exist, got:\n%s", withWarn)
	}
}

// TestInitSuccessCard_TuiBoxPillLanguage asserts the init completion card uses
// the shared tui.Box + tui.Pill visual language mirroring the update
// classification summary (SPEC-CLI-TUX-INIT-UPDATE-001 REQ-TUXIU-031,
// AC-TUXIU-011). Under NO_COLOR the tui.Pill primitive degrades to bracketed
// "[label]" text (its monochrome contract), which is the greppable signal that
// the card routes counts through tui.Pill rather than the legacy
// RenderKeyValueLines "Directories  N created" form.
func TestInitSuccessCard_TuiBoxPillLanguage(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	card := buildInitSuccessCard("proj", 3, 7, 0)

	// Count summary rendered via tui.Pill (NO_COLOR bracketed form).
	for _, want := range []string{"[3 dirs]", "[7 files]"} {
		if !strings.Contains(card, want) {
			t.Errorf("init success card must render counts via tui.Pill (want %q), got:\n%s", want, card)
		}
	}
	// Rendered inside a tui.Box rounded border (structural border runes present
	// regardless of colour).
	if !strings.ContainsAny(card, "╭╮╰╯─│") {
		t.Errorf("init success card must render inside a tui.Box border, got:\n%s", card)
	}
	// Next-action sequence preserved (data unchanged).
	for _, want := range []string{"cd proj", "moai cc", "/moai plan"} {
		if !strings.Contains(card, want) {
			t.Errorf("init success card must retain next-action %q, got:\n%s", want, card)
		}
	}
}

// TestExistingProjectUpdateRedirectHint asserts REQ-TUX2-015 (AC-TUX2-015):
// re-running moai init on an already-initialized project without --force
// surfaces a redirect hint naming `moai update` as the likely intent.
func TestExistingProjectUpdateRedirectHint(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv(config.EnvSkipBinaryUpdate, "1")

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	projectDir := filepath.Join(t.TempDir(), "existing-proj")
	if err := os.MkdirAll(filepath.Join(projectDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	err := runInit(cmd, []string{projectDir})
	if err == nil {
		t.Fatal("expected an already-initialized error without --force")
	}
	combined := err.Error() + errBuf.String()
	if !strings.Contains(combined, "moai update") {
		t.Errorf("error surface must include the 'moai update' redirect hint, got:\nerr=%v\nstderr=%s", err, errBuf.String())
	}
	if !strings.Contains(combined, "--force") {
		t.Errorf("error surface must still mention --force, got:\nerr=%v\nstderr=%s", err, errBuf.String())
	}
}
