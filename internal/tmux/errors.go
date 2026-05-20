// Package tmux provides tmux session management for parallel SPEC
// development. It supports creating sessions with multiple panes,
// automatic command execution, and configurable layouts.
package tmux

import "errors"

// Sentinel errors for tmux operations.
// All errors can be checked with errors.Is().
var (
	// ErrTmuxNotFound indicates the tmux binary is not in PATH.
	ErrTmuxNotFound = errors.New("tmux: tmux not found in PATH")

	// ErrSessionExists indicates a session with the same name already exists.
	ErrSessionExists = errors.New("tmux: session already exists")

	// ErrSessionFailed indicates a failure creating or managing a session.
	ErrSessionFailed = errors.New("tmux: failed to create session")

	// ErrNoPanes indicates the session configuration has no panes.
	ErrNoPanes = errors.New("tmux: no panes configured")

	// ErrTmuxSensitiveInjectFailed indicates that injection of a sensitive
	// environment variable failed BEFORE the value reached tmux. Callers MUST
	// NOT fall back to the argv-exposing `set-environment <k> <v>` path after
	// receiving this error — doing so would re-introduce the CWE-214 leak that
	// InjectSensitiveEnv was added to close.
	//
	// SPEC-V3R5-SECURITY-CRIT-001 P0-2 (REQ-SEC-002-007, AC-SEC-007).
	ErrTmuxSensitiveInjectFailed = errors.New("tmux: sensitive env inject failed")
)
