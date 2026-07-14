package worktree

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// TestCreateTmuxSession_ResultBlock_OnStderrViaPrinter is the M4
// characterization test (SPEC-CLI-TUX-V3-005). The five tmux-session-creation
// status lines were migrated from fmt.Printf (stdout) to p.Info (stderr).
//
// This test pins the five lines BYTE-FOR-BYTE on stderr in ModePlain (fixed
// "·" info glyph, no ANSI) and confirms stdout carries none of them. The "· "
// prefix is emitted exclusively by the Printer status path, so observing it on
// stderr is also the reachability proof that p.Info was actually called
// (AC-TUX3-021) and that exactly five info lines were produced.
func TestCreateTmuxSession_ResultBlock_OnStderrViaPrinter(t *testing.T) {
	// Inject a buffer-backed Printer forced to ModePlain so the rendered bytes
	// are deterministic (escape-free, fixed info glyph).
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	captured := printer.New(
		printer.WithWriters(stdoutBuf, stderrBuf),
		printer.WithMode(printer.ModePlain),
	)
	orig := tmuxSessionPrinterFactory
	t.Cleanup(func() { tmuxSessionPrinterFactory = orig })
	tmuxSessionPrinterFactory = func() printer.Printer { return captured }

	const (
		projectName = "test-proj"
		specID      = "SPEC-M4-001"
		wtPath      = "/tmp/m4-wt"
	)
	cfg := &TmuxSessionConfig{
		ProjectName:  projectName,
		SpecID:       specID,
		WorktreePath: wtPath,
		ActiveMode:   "cc", // skip the GLM/CG env-injection path
	}

	// recordingSessionManager (defined in tmux_integration_sec_harden_test.go)
	// returns SessionName = cfg.Name, PaneCount = len(cfg.Panes). CreateTmuxSession
	// builds the session config with a single pane, so PaneCount = 1; Attached is
	// the zero value (false).
	if err := CreateTmuxSession(context.Background(), cfg, &recordingSessionManager{}); err != nil {
		t.Fatalf("CreateTmuxSession() unexpected error: %v", err)
	}

	// Build the exact expected stderr block. sessionName is derived the same way
	// CreateTmuxSession derives it (and the same value recordingSessionManager
	// echoes back as SessionName).
	sessionName := GenerateTmuxSessionName(projectName, specID)
	glyph := tui.StatusIcon("info")
	var want strings.Builder
	fmt.Fprintf(&want, "%s Tmux session created: %s\n", glyph, sessionName)
	fmt.Fprintf(&want, "%s Panes created: %d\n", glyph, 1)
	fmt.Fprintf(&want, "%s Attached: %v\n", glyph, false)
	fmt.Fprintf(&want, "%s Worktree path: %s\n", glyph, wtPath)
	fmt.Fprintf(&want, "%s To attach: tmux attach-session -t %s\n", glyph, sessionName)

	if got := stderrBuf.String(); got != want.String() {
		t.Errorf("stderr mismatch (M4 migration regressed the status block):\ngot:\n%s\nwant:\n%s", got, want.String())
	}
	if got := stdoutBuf.String(); got != "" {
		t.Errorf("stdout must be empty after M4 migration (status lines moved stdout->stderr), got:\n%s", got)
	}

	// Reachability + count guard (AC-TUX3-021): exactly five Printer-emitted
	// info lines on stderr, each prefixed with the info glyph.
	lines := strings.Split(strings.TrimRight(stderrBuf.String(), "\n"), "\n")
	const expectedInfoLines = 5
	if len(lines) != expectedInfoLines {
		t.Fatalf("expected %d info lines on stderr, got %d: %q", expectedInfoLines, len(lines), stderrBuf.String())
	}
	for i, ln := range lines {
		if prefix := glyph + " "; !strings.HasPrefix(ln, prefix) {
			t.Errorf("line %d is not a Printer info line (missing %q prefix): %q", i, prefix, ln)
		}
	}
}
