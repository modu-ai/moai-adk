package mission

type FinalizationInput struct {
	Current            MissionState
	Revoked            bool
	InFlightEffects    int
	RequiredEvidence   map[string]bool
	MainLandedAncestry bool
	AgentIdle          bool
	LeaseExpired       bool
}

type FinalizationResult struct {
	State             MissionState
	AllowNewWork      bool
	ReconcileInFlight bool
	Reason            string
}

func FinalizeOrRevokeMission(in FinalizationInput) FinalizationResult {
	if in.Revoked {
		return FinalizationResult{State: StateRevoking, ReconcileInFlight: in.InFlightEffects > 0, Reason: "revoked"}
	}
	complete := len(in.RequiredEvidence) > 0 && in.MainLandedAncestry
	for _, satisfied := range in.RequiredEvidence {
		if !satisfied {
			complete = false
		}
	}
	if complete {
		return FinalizationResult{State: StateCompleted, Reason: "authoritative_evidence_satisfied"}
	}
	state := in.Current
	if state == "" {
		state = StateRunning
	}
	return FinalizationResult{State: state, AllowNewWork: state == StateRunning, Reason: "evidence_pending"}
}
