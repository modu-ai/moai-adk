package bugreport

// Verdict is the closed attribution vocabulary (REQ-ANON-009). Every captured
// signal carries exactly one, fixed at capture time where the error chain is
// alive, and consumed downstream without re-attribution.
//
// moai is an allowlist: only a panic, an explicit internal marker, or an
// enumerated moai sentinel qualifies (design.md section 3, rows M1-M3).
// user, environment, and ambiguous all stay local.
type Verdict string

const (
	// VerdictMoai proceeds through the pipeline: fingerprint, caps, payload,
	// queue, and — at flush — publication from the user's own account.
	VerdictMoai Verdict = "moai"

	// VerdictUser is a user-input or user-configuration cause. Dropped at
	// capture without recording.
	VerdictUser Verdict = "user"

	// VerdictEnvironment is a filesystem, network, permission, or subprocess
	// cause. Dropped at capture without recording.
	VerdictEnvironment Verdict = "environment"

	// VerdictAmbiguous cannot be attributed by the deterministic rules. It is
	// retained locally: spooled so the drain can log the retention, never
	// queued, never sent, never adjudicated by a model (DEC-7, final).
	VerdictAmbiguous Verdict = "ambiguous"
)

// Valid reports whether v is one of the four closed members.
func (v Verdict) Valid() bool {
	switch v {
	case VerdictMoai, VerdictUser, VerdictEnvironment, VerdictAmbiguous:
		return true
	}
	return false
}
