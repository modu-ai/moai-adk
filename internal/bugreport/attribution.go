package bugreport

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/config"
	"gopkg.in/yaml.v3"
)

// Reason is the call-site-asserted attribution token (design.md section 3,
// mode A): a token an emit site supplies because the sentinel lives in a
// package bugreport cannot import (internal/template's errors, os/exec's
// types) or because a subprocess exit carries no importable type. The token
// is closed; an unknown reason never creates a moai verdict.
type Reason string

const (
	// ReasonExec is asserted by an exec-typed call site: its site maps
	// *exec.Error, exec.ErrNotFound, or *exec.ExitError to this token, because
	// bugreport must not import os/exec (REQ-ANON-025). A subprocess exit —
	// for example git in a user's repository state — is not moai's defect.
	ReasonExec Reason = "exec"

	ReasonPathTraversal     Reason = "path_traversal"
	ReasonNotFound          Reason = "not_found"
	ReasonPreserveIntegrity Reason = "preserve_integrity"
	ReasonMissingKey        Reason = "missing_key"
	ReasonUnexpandedToken   Reason = "unexpanded_token"
	ReasonInvalidJSON       Reason = "invalid_json"
)

// Valid reports membership in the closed reason set.
func (r Reason) Valid() bool {
	switch r {
	case ReasonExec, ReasonPathTraversal, ReasonNotFound, ReasonPreserveIntegrity,
		ReasonMissingKey, ReasonUnexpandedToken, ReasonInvalidJSON:
		return true
	}
	return false
}

// Attribute evaluates the ordered, first-match rule table (design.md
// section 3) over a captured signal. Error chains are inspected with
// errors.Is and errors.As ONLY — error text is never read anywhere in this
// package. moai is an ALLOWLIST: only a panic (M1), the internal marker
// (M2), or an enumerated moai token (M3) produces VerdictMoai; every
// unrecognised error falls to row F and reads ambiguous, which stays local.
//
// The environment rows (A) and user rows (U) run before the moai rows (M),
// so an internal marker wrapping an *fs.PathError resolves to environment —
// the cause is the environment even though moai noticed it.
//
// The returned reason echoes the call-site token when one decided the row,
// so the spool line carries why.
func Attribute(kind Kind, err error, reason Reason) (Verdict, Reason) {
	// The hook_timeout register verdict is unconditional-ambiguous (design
	// section 2: "load and a stuck handler look alike without more data"),
	// and it routes BEFORE the generic error-chain rows: the timeout
	// sentinels the registry captures under this kind — context.
	// DeadlineExceeded first among them — implement net.Error, so the A2
	// network row classified environment and Capture discarded the signal
	// the SPEC retains locally (review-gate P2; DEC-7).
	if kind == KindHookTimeout {
		return VerdictAmbiguous, ""
	}

	// A1: filesystem / syscall.
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return VerdictEnvironment, ""
	}
	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		return VerdictEnvironment, ""
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return VerdictEnvironment, ""
	}

	// A2: network.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return VerdictEnvironment, ""
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return VerdictEnvironment, ""
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return VerdictEnvironment, ""
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return VerdictEnvironment, ""
	}

	// A3: permission / cancellation.
	if errors.Is(err, os.ErrPermission) || errors.Is(err, context.Canceled) {
		return VerdictEnvironment, ""
	}

	// A3a: the call site asserted the exec token.
	if reason == ReasonExec {
		return VerdictEnvironment, ReasonExec
	}

	// U1: config sentinels (internal/config is importable, row D).
	if errors.Is(err, config.ErrConfigNotFound) || errors.Is(err, config.ErrInvalidConfig) ||
		errors.Is(err, config.ErrInvalidYAML) || errors.Is(err, config.ErrSectionTypeMismatch) ||
		errors.Is(err, config.ErrInvalidDevelopmentMode) {
		return VerdictUser, ""
	}

	// U2: JSON decode errors.
	var jsonSyntax *json.SyntaxError
	if errors.As(err, &jsonSyntax) {
		return VerdictUser, ""
	}
	var jsonType *json.UnmarshalTypeError
	if errors.As(err, &jsonType) {
		return VerdictUser, ""
	}

	// U3: yaml type errors. Measured 2026-10-07 (TestYAMLSyntaxErrorReaches-
	// Fallback): a yaml SYNTAX error is untyped in v3 and reaches row F,
	// which is safe because F is not moai.
	var yamlType *yaml.TypeError
	if errors.As(err, &yamlType) {
		return VerdictUser, ""
	}

	// M1: a panic (the recover site asserted the kind).
	if kind == KindPanic {
		return VerdictMoai, ""
	}

	// M2: the internal marker.
	if IsInternalMarker(err) {
		return VerdictMoai, ""
	}

	// M3: a moai token.
	switch reason {
	case ReasonPathTraversal, ReasonNotFound, ReasonPreserveIntegrity, ReasonMissingKey:
		return VerdictMoai, reason
	}

	// F: the fallback — every unrecognised error, an unmarked hook handler
	// error, the hook timeout, and the tokens unexpanded_token/invalid_json.
	// Ambiguous is retained locally: recorded so the drain logs the retention,
	// never queued, never sent, never adjudicated (DEC-7, final).
	return VerdictAmbiguous, reason
}
