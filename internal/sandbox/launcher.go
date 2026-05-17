package sandbox

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

// Timeout constants for sandbox operations.
const (
	// bwrapProbeTimeout is the timeout for the bwrap --version probe.
	bwrapProbeTimeout = 3 * time.Second
	// dockerProbeTimeout is the timeout for the docker info probe.
	dockerProbeTimeout = 5 * time.Second
	// execTimeout is the maximum time allowed for a sandboxed command.
	execTimeout = 5 * time.Minute
)

// @MX:NOTE: [AUTO] resolveBackend contains the OS × CI × role dispatch table
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-002/015

// writeRoles is the set of agent roles that require an active sandbox by default.
// Per REQ-V3R2-RT-003-003: implementer, tester, designer.
var writeRoles = map[string]bool{
	"implementer": true,
	"tester":      true,
	"designer":    true,
}

// isReadOnlyRole returns true for roles that default to sandbox: none.
// Per REQ-V3R2-RT-003-003: researcher, analyst, reviewer, architect.
func isReadOnlyRole(role string) bool {
	switch role {
	case "researcher", "analyst", "reviewer", "architect":
		return true
	}
	return false
}

// Launcher is the top-level sandbox facade. It dispatches sandbox execution
// to the appropriate per-OS backend and applies cross-cutting concerns:
// output truncation, env scrubbing, backend availability checking,
// and permission-divergence handling.
//
// @MX:ANCHOR: [AUTO] Launcher is the primary entry point for all sandbox execution
// @MX:REASON: Fan_in >= 3: doctor_sandbox.go, agent_lint.go, future agent_dispatch.go,
//             all sandbox tests — any interface change breaks all callers
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-002/012/015/050
type Launcher struct {
	// backends maps sandbox type to its implementation.
	// Initialized with OS-appropriate defaults; overrideable via SetBackend (for testing).
	backends map[Sandbox]SandboxBackend

	// sandboxRequired mirrors security.yaml sandbox.required field.
	sandboxRequired bool
}

// NewLauncher creates a Launcher with the default OS-appropriate backends.
func NewLauncher() *Launcher {
	l := &Launcher{
		backends: map[Sandbox]SandboxBackend{
			SandboxBubblewrap: NewBubblewrapBackend(),
			SandboxSeatbelt:   NewSeatbeltBackend(),
			SandboxDocker:     NewDockerBackend(),
		},
	}
	return l
}

// SetBackend replaces the backend for the given sandbox type.
// Primarily used in tests to inject mock backends.
func (l *Launcher) SetBackend(s Sandbox, b SandboxBackend) {
	if l.backends == nil {
		l.backends = make(map[Sandbox]SandboxBackend)
	}
	l.backends[s] = b
}

// SetSandboxRequired controls whether sandbox: none is rejected
// for implementer roles without justification.
func (l *Launcher) SetSandboxRequired(required bool) {
	l.sandboxRequired = required
}

// ResolveBackend returns the sandbox that would actually be used for the given
// declared value. For SandboxNone, it returns SandboxNone unchanged.
// CI=1 does not override a declared non-none sandbox.
func (l *Launcher) ResolveBackend(declared Sandbox) Sandbox {
	if declared != SandboxNone {
		return declared
	}
	return SandboxNone
}

// ResolveForRole returns the sandbox backend appropriate for the given agent role,
// taking into account the current OS and CI=1 environment variable.
//
// Rules (priority order):
//  1. CI=1 env var + write role → SandboxDocker
//  2. write role on macOS → SandboxSeatbelt
//  3. write role on Linux → SandboxBubblewrap
//  4. read-only role → SandboxNone
//  5. unknown role → SandboxNone (conservative default)
func (l *Launcher) ResolveForRole(role string) Sandbox {
	if !writeRoles[role] || isReadOnlyRole(role) {
		return SandboxNone
	}

	// CI=1 override: implementer/tester/designer → docker (REQ-V3R2-RT-003-015)
	if os.Getenv("CI") == "1" {
		return SandboxDocker
	}

	// OS-based default
	switch runtime.GOOS {
	case "darwin":
		return SandboxSeatbelt
	case "linux":
		return SandboxBubblewrap
	default:
		// Windows, etc. — Docker fallback
		return SandboxDocker
	}
}

// Exec runs cmd in the sandbox identified by s with the given options.
// Returns ErrSandboxBackendUnavailable when the backend is not present.
// Returns ErrSandboxOutputTruncated (alongside the truncated output) when
// output exceeds opts.MaxOutputBytes.
//
// Sandbox verdict takes priority over any external permission decision
// (REQ-V3R2-RT-003-051 / AC-V3R2-RT-003-16).
func (l *Launcher) Exec(s Sandbox, opts SandboxOptions, cmd []string) ([]byte, error) {
	// SandboxNone: direct exec without sandboxing
	if s == SandboxNone {
		if l.sandboxRequired {
			return nil, ErrSandboxRequired
		}
		// Unsandboxed exec — pass through (security.yaml sandbox.required=false)
		return unsandboxedExec(opts, cmd)
	}

	backend, ok := l.backends[s]
	if !ok {
		return nil, fmt.Errorf("%w: no backend registered for %q", ErrSandboxBackendUnavailable, s)
	}

	if !backend.Available() {
		return nil, ErrSandboxBackendUnavailable
	}

	out, err := backend.Exec(opts, cmd)
	// Sandbox verdict wins: if backend returns a sandbox denial error,
	// bubble it up with full context (AC-V3R2-RT-003-16).
	return out, err
}

// TruncateOutput truncates output to limit bytes and returns ErrSandboxOutputTruncated.
// If len(output) <= limit, returns the original output and nil error.
func TruncateOutput(output []byte, limit int64) ([]byte, error) {
	if int64(len(output)) <= limit {
		return output, nil
	}
	return output[:limit], fmt.Errorf("%w: output was %d bytes, truncated to %d",
		ErrSandboxOutputTruncated, len(output), limit)
}

// limitedWriter wraps a bytes.Buffer and caps writes at limit bytes.
type limitedWriter struct {
	buf     *bytes.Buffer
	limit   int64
	written int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.written
	if remaining <= 0 {
		return len(p), nil // silently drop
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	n, err := w.buf.Write(p)
	w.written += int64(n)
	return len(p), err // return original len to satisfy io.Writer contract
}

// unsandboxedExec runs cmd without any sandbox (fallback for sandbox: none).
func unsandboxedExec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("exec: empty command")
	}

	maxBytes := opts.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputBytes
	}

	// Import os/exec inline to avoid circular import with test package
	// We use a subprocess approach to avoid importing exec at package level
	// when the test package substitutes backends
	_ = maxBytes // unsandboxed exec just returns an error for now
	return nil, fmt.Errorf("unsandboxed exec not yet implemented — use sandbox: seatbelt or bubblewrap")
}
