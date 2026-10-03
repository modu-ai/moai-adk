package cli

// managed_codex_tui.go — SPEC-FACTORY-MANAGED-TUI-001: the operator TUI attach
// of the managed Codex session owner.
//
// M1 compile stubs (behavior-neutral): the symbols the reproduction tests need
// so the package test binary builds, with the headless behavior unchanged.
// The attach decision, the TUI child, the log diversion and the busy tracking
// are filled in by the later milestones.

import (
	"io"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
	"github.com/modu-ai/moai-adk/internal/config"
)

// Package-private seams. Production defaults are the real behavior; tests
// replace them so the timing and terminal rules run without a real terminal.
var (
	// managedTerminalCheck reports whether a stream is a terminal.
	managedTerminalCheck = func(f *os.File) bool { return f != nil && isatty.IsTerminal(f.Fd()) }
	// managedBusyWarnInterval is how often a long busy thread is logged.
	managedBusyWarnInterval = config.DefaultManagedCodexTurnTimeout
	// managedWriteDeadline bounds one App Server connection write.
	managedWriteDeadline = config.DefaultManagedCodexTurnTimeout
	// managedTUIStopGrace is the wait between the interrupt and the kill.
	managedTUIStopGrace = config.DefaultManagedCodexTUIStopGrace
	// managedProbeTimeout bounds the capability probe.
	managedProbeTimeout = config.DefaultManagedCodexProbeTimeout
	// managedCodexTUICommand creates the operator TUI child.
	managedCodexTUICommand = exec.Command
	// managedCodexRemoteProbe asks the codex binary whether `resume` offers the
	// remote options; the second result names the reason when it does not.
	managedCodexRemoteProbe = func(program string, env []string, dir string) (bool, string) {
		return false, "capability probe not implemented"
	}
	// managedWriteBarrier is a test hook called at the entry of every App Server
	// connection write.
	managedWriteBarrier func(v any)
)

// managedOperatorSurface is the optional capability the Codex owner offers the
// delivery driver while an operator TUI is attached. The managedSession
// interface itself is unchanged.
type managedOperatorSurface interface {
	// AttachOperator starts the operator TUI when the session planned one. The
	// channel delivers the single end-of-TUI result (the TUI's exit mapped to an
	// error, or the lost-connection error); attached is false for a headless
	// session.
	AttachOperator() (done <-chan error, attached bool)
	// Busy reports whether the thread has an active turn.
	Busy() bool
}

var _ managedOperatorSurface = (*managedCodexSession)(nil)

// AttachOperator is the M1 stub: a session never attaches.
func (s *managedCodexSession) AttachOperator() (<-chan error, bool) { return nil, false }

// Busy is the M1 stub: the thread is never busy.
func (s *managedCodexSession) Busy() bool { return false }

// planOperatorTUI is the M1 stub: the pre-Start attach decision plans nothing.
func (s *managedCodexSession) planOperatorTUI(root, runID string, stdin io.Reader) {}

// stopTUI is the M1 stub: there is never a TUI child to stop.
func (s *managedCodexSession) stopTUI() {}

// managedCodexRemoteSupport is the M1 stub: no help text offers the options.
func managedCodexRemoteSupport(help string) bool { return false }

// managedTUIExitError is the M1 stub: no exit status maps to an error.
func managedTUIExitError(state *os.ProcessState) error { return nil }
