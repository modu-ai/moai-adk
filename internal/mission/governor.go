package mission

type BoundaryDecision struct {
	Proceed         bool
	Blocked         bool
	AskUserQuestion bool
	Report          string
}

// SuppressAutoMissionQuestions encodes the approved-auto boundary: inside the
// sealed scope work proceeds with no prompt; outside it stops with a blocker
// report and no side effect. It never manufactures new authority.
func SuppressAutoMissionQuestions(insideSealedScope bool, reason string) BoundaryDecision {
	if insideSealedScope {
		return BoundaryDecision{Proceed: true}
	}
	if reason == "" {
		reason = "outside_sealed_scope"
	}
	return BoundaryDecision{Blocked: true, Report: reason}
}
