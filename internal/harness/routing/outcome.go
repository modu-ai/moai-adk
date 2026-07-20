package routing

// DeriveOutcome derives a terminal routing outcome deterministically from
// machine evidence, per the fixed precedence of REQ-HEV-013:
//
//	(1) explicit abort marker — any abort-kind evidence entry -> abort
//	(2) any gate_exit evidence with non-zero exit          -> fail
//	(3) at least one terminal passing machine signal        -> success
//	    (declared-final gate_exit "0", audit-score, or verify-path marked terminal)
//	(4) otherwise                                           -> non-terminal (stay pending)
//
// The second return value reports whether a terminal outcome was derived: false
// means the row is not yet closeable (multi-turn pipeline still in flight).
//
// DeriveOutcome accepts NO free-text outcome override — outcome is a pure
// function of the machine evidence (REQ-HEV-006, un-fakeable contract). The two
// writer-internal finalizations, reroute (REQ-HEV-010) and the lazy staleness
// sweep (REQ-HEV-014), assign their outcome directly and DO NOT call this
// function.
//
// @MX:NOTE: [AUTO] Fixed precedence abort > non-zero-gate > terminal-pass > pending;
// no code path accepts a caller-supplied outcome (AP-1).
func DeriveOutcome(refs []EvidenceRef) (Outcome, bool) {
	// Precedence 1: an explicit abort marker is inherently terminal.
	for _, r := range refs {
		if r.Kind == KindAbort {
			return OutcomeAbort, true
		}
	}
	// Precedence 2: any gate_exit with a non-zero exit code closes the row as fail.
	for _, r := range refs {
		if r.Kind == KindGateExit && r.Value != "" && r.Value != "0" {
			return OutcomeFail, true
		}
	}
	// Precedence 3: at least one terminal passing machine signal closes as success.
	for _, r := range refs {
		if !r.Terminal {
			continue
		}
		switch r.Kind {
		case KindGateExit:
			if r.Value == "0" {
				return OutcomeSuccess, true
			}
		case KindAuditScore, KindVerifyPath:
			return OutcomeSuccess, true
		}
	}
	// Precedence 4: no terminal signal — leave pending.
	return "", false
}
