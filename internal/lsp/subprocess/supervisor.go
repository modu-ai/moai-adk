package subprocess

import (
	"context"
	"os"
	"os/exec"
	"syscall"
)

// ExitEvent is delivered on the channel returned by Supervisor.Watch when
// the monitored subprocess terminates for any reason.
type ExitEvent struct {
	// ExitCode is the exit code reported by the OS.
	// For signal-killed processes on Unix this is typically -1.
	ExitCode int

	// Err holds the raw error from cmd.Wait (may be *exec.ExitError).
	Err error

	// Signaled is true when the process was terminated by a signal.
	Signaled bool
}

// Supervisor monitors a running language server subprocess.
//
// It exposes Watch for crash detection, Signal for graceful shutdown, and Kill
// for forced termination (REQ-LC-005, REQ-LC-031).
//
// Supervisor owns the single call to cmd.Wait(); callers MUST NOT call
// cmd.Wait() independently after creating a Supervisor.
//
// Typical usage:
//
//	ch := sv.Watch(ctx)
//	select {
//	case ev := <-ch:
//	    // Handle abnormal server termination (degraded state)
//	}
//
// @MX:ANCHOR: [AUTO] Supervisor — subprocess lifecycle monitor used by core.Client and degraded-state handler
// @MX:REASON: fan_in >= 3 — core.Client.Start, crash recovery handler, and integration tests all use Supervisor
type Supervisor struct {
	cmd    *exec.Cmd
	doneCh chan struct{} // closed when cmd.Wait completes
	exitEv ExitEvent    // stored result from cmd.Wait; safe after doneCh is closed
}

// NewSupervisor creates a Supervisor for the already-started subprocess described
// by result. The LaunchResult.Cmd must have been started before calling this.
//
// NewSupervisor immediately starts a background goroutine that calls cmd.Wait
// exactly once. Callers MUST NOT call result.Cmd.Wait after this point.
//
// @MX:WARN: [AUTO] NewSupervisor starts a goroutine that owns cmd.Wait for the subprocess lifetime
// @MX:REASON: Wait() blocks forever if stdio pipes are not closed — callers must close stdin/stdout/stderr first
func NewSupervisor(result *LaunchResult) *Supervisor {
	s := &Supervisor{
		cmd:    result.Cmd,
		doneCh: make(chan struct{}),
	}
	// Wait goroutine: run exactly once to prevent race conditions
	go func() {
		err := result.Cmd.Wait()
		s.exitEv = buildExitEvent(err)
		close(s.doneCh)
	}()
	return s
}

// Watch returns a channel that receives a single ExitEvent when the subprocess
// terminates, then closes. If ctx is cancelled before the process exits, the
// channel is closed without an ExitEvent.
//
// Multiple Watch calls are safe; each gets its own independent channel.
func (s *Supervisor) Watch(ctx context.Context) <-chan ExitEvent {
	ch := make(chan ExitEvent, 1)

	go func() {
		defer close(ch)
		select {
		case <-s.doneCh:
			// doneCh is already closed or closes when Wait completes
			// exitEv is read only after doneCh is closed, so this is race-free
			ch <- s.exitEv
		case <-ctx.Done():
			// Context cancelled: notify caller by closing the empty channel
		}
	}()

	return ch
}

// Signal sends sig to the subprocess.
// Returns an error if the process has already exited or the signal is invalid.
func (s *Supervisor) Signal(sig os.Signal) error {
	return s.cmd.Process.Signal(sig)
}

// Kill sends SIGKILL to the subprocess, forcing immediate termination.
// Returns an error if the process has already exited.
func (s *Supervisor) Kill() error {
	return s.cmd.Process.Signal(syscall.SIGKILL)
}

// buildExitEvent converts the error from cmd.Wait into an ExitEvent.
func buildExitEvent(err error) ExitEvent {
	if err == nil {
		return ExitEvent{ExitCode: 0}
	}

	exitErr, ok := err.(*exec.ExitError) //nolint:errorlint // exec.ExitError is not wrapped
	if !ok {
		return ExitEvent{Err: err}
	}

	code := exitErr.ExitCode()
	signaled := false

	// Unix: ProcessState.Sys() returns syscall.WaitStatus
	if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
		signaled = ws.Signaled()
	}

	return ExitEvent{
		ExitCode: code,
		Err:      err,
		Signaled: signaled,
	}
}
