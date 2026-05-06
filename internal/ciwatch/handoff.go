package ciwatch

// CheckResult holds the result of a single GitHub Actions check run.
// Fields map directly to `gh pr checks --json` output fields.
type CheckResult struct {
	// Name is the check context name (e.g. "Lint", "Test (ubuntu-latest)").
	Name string `json:"name"`
	// Status is the current status: "queued", "in_progress", or "completed".
	Status string `json:"status"`
	// Conclusion is the terminal result: "success", "failure", "cancelled", "skipped", etc.
	// Empty if Status != "completed".
	Conclusion string `json:"conclusion"`
	// RunID is the GitHub Actions workflow run ID (numeric string).
	RunID string `json:"runId,omitempty"`
	// LogURL is the direct link to the run's log page.
	LogURL string `json:"logUrl,omitempty"`
	// ConclusionDetail is an optional human-readable summary of the failure cause.
	ConclusionDetail string `json:"conclusionDetail,omitempty"`
}

// CIState represents a snapshot of the CI check state for a PR at a given tick.
type CIState struct {
	// PRNumber is the GitHub pull request number being watched.
	PRNumber int `json:"prNumber"`
	// Branch is the head branch of the PR.
	Branch string `json:"branch"`
	// RequiredFailed holds required checks that reached a failed terminal state.
	RequiredFailed []CheckResult `json:"requiredFailed,omitempty"`
	// RequiredPending holds required checks still in progress.
	RequiredPending []CheckResult `json:"requiredPending,omitempty"`
	// RequiredPassed is the count of required checks that passed (success/skipped).
	RequiredPassed int `json:"requiredPassed"`
	// AuxiliaryFailed holds auxiliary checks that failed (advisory, does not block).
	AuxiliaryFailed []CheckResult `json:"auxiliaryFailed,omitempty"`
}

// FailedCheck is the forward-stable schema for a failed required check,
// intended for consumption by Wave 3 expert-debug prompt injection.
type FailedCheck struct {
	// Name is the check context name.
	Name string `json:"name"`
	// RunID is the GitHub Actions workflow run ID (numeric string).
	RunID string `json:"runId,omitempty"`
	// LogURL is the direct URL to the failed run log.
	LogURL string `json:"logUrl,omitempty"`
	// ConclusionDetail is a human-readable summary of the failure.
	ConclusionDetail string `json:"conclusionDetail,omitempty"`
}

// Handoff is the structured metadata package produced when the CI watch loop
// detects required check failures. It is serialized to JSON and piped to the
// Wave 3 expert-debug invocation prompt.
//
// JSON shape stability: fields are tagged for stable serialization across
// Wave 3 consumers. Do not rename or remove fields without a major version bump.
type Handoff struct {
	// PRNumber is the GitHub pull request number.
	PRNumber int `json:"prNumber"`
	// Branch is the head branch name.
	Branch string `json:"branch"`
	// FailedChecks is the slice of required checks that failed.
	// Auxiliary failures are NOT included here.
	FailedChecks []FailedCheck `json:"failedChecks"`
	// AuxiliaryFailCount is the number of auxiliary checks that failed
	// (informational; does not block merge).
	AuxiliaryFailCount int `json:"auxiliaryFailCount"`
	// TotalRequired is the total count of required checks tracked.
	TotalRequired int `json:"totalRequired"`
}

// NewHandoff constructs a Handoff from a CIState snapshot.
// Only required failures are included in FailedChecks; auxiliary failures
// are counted but excluded from the blocking list.
func NewHandoff(state CIState) Handoff {
	failed := make([]FailedCheck, 0, len(state.RequiredFailed))
	for _, cr := range state.RequiredFailed {
		failed = append(failed, FailedCheck{
			Name:             cr.Name,
			RunID:            cr.RunID,
			LogURL:           cr.LogURL,
			ConclusionDetail: cr.ConclusionDetail,
		})
	}

	totalRequired := state.RequiredPassed + len(state.RequiredFailed) + len(state.RequiredPending)

	return Handoff{
		PRNumber:           state.PRNumber,
		Branch:             state.Branch,
		FailedChecks:       failed,
		AuxiliaryFailCount: len(state.AuxiliaryFailed),
		TotalRequired:      totalRequired,
	}
}

// FormatStatusUpdate returns a plain-text, single-line CI status summary
// suitable for orchestrator transcript output. No ANSI codes; max 200 chars/line.
// Example: "[ci-watch] PR #785: required 4/6 pass, 2 pending; advisory 0 fail"
func FormatStatusUpdate(state CIState) string {
	totalRequired := state.RequiredPassed + len(state.RequiredFailed) + len(state.RequiredPending)
	auxFail := len(state.AuxiliaryFailed)

	// Build the base message components.
	msg := "[ci-watch] PR #" + itoa(state.PRNumber) + ": required " +
		itoa(state.RequiredPassed) + "/" + itoa(totalRequired) + " pass"

	if len(state.RequiredFailed) > 0 {
		msg += ", " + itoa(len(state.RequiredFailed)) + " failed"
	}
	if len(state.RequiredPending) > 0 {
		msg += ", " + itoa(len(state.RequiredPending)) + " pending"
	}

	msg += "; advisory " + itoa(auxFail) + " fail"

	// Truncate to 200 chars to honour the protocol constraint.
	if len(msg) > 200 {
		msg = msg[:197] + "..."
	}
	return msg
}

// itoa is a minimal integer-to-string helper to avoid strconv import churn.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
