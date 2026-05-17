package sandbox

import "errors"

// Sentinel errors for the sandbox execution layer.
// All sentinels support errors.Is via wrapping.
//
// @MX:NOTE: [AUTO] Sentinel errors are the primary error signaling mechanism for sandbox failures.
// @MX:SPEC: SPEC-V3R2-RT-003 REQ-012/040/041/042/043

// ErrSandboxBackendUnavailable is returned when the required sandbox backend
// binary (bwrap, sandbox-exec, docker) is not available on the current host.
// Agents MUST NOT fall back to unsandboxed execution on this error.
var ErrSandboxBackendUnavailable = errors.New("sandbox backend unavailable")

// ErrSandboxProfileInvalid is returned when profile generation produces an
// invalid SBPL string or bwrap argument sequence.
var ErrSandboxProfileInvalid = errors.New("sandbox profile invalid")

// ErrSandboxRequired is returned when security.yaml has sandbox.required: true
// and an agent declares sandbox: none without a justification.
var ErrSandboxRequired = errors.New("sandbox required by security policy")

// ErrSandboxOutputTruncated is returned when a sandboxed process produces output
// exceeding the configured MaxOutputBytes limit (default 16 MiB).
var ErrSandboxOutputTruncated = errors.New("sandbox output truncated to 16 MiB")

// ErrSandboxSetuidDenied is returned when a sandboxed process attempts to invoke
// a setuid binary (sudo, su) that the sandbox backend denies.
var ErrSandboxSetuidDenied = errors.New("sandbox denied setuid escalation")

// SandboxDeniedError is a structured error for path-level sandbox denial.
// It wraps the context (attempted path, reason) so callers can surface a
// clear SystemMessage naming the denied path.
type SandboxDeniedError struct {
	Path   string
	Reason string
}

// Error implements the error interface.
func (e *SandboxDeniedError) Error() string {
	if e.Path != "" {
		return "sandbox-denied: " + e.Path + " (" + e.Reason + ")"
	}
	return "sandbox-denied: " + e.Reason
}

// Is enables errors.Is matching against *SandboxDeniedError.
func (e *SandboxDeniedError) Is(target error) bool {
	_, ok := target.(*SandboxDeniedError)
	return ok
}
