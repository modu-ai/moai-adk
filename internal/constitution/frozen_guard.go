package constitution

// frozenGuard is an implementation of the FrozenGuard interface.
// Blocks Frozen zone rule amendments.
type frozenGuard struct{}

// NewFrozenGuard creates a FrozenGuard.
func NewFrozenGuard() FrozenGuard {
	return &frozenGuard{}
}

// Check checks if the proposal modifies a Frozen zone rule.
// Frozen→Evolvable demotion is allowed with Evidence.
// SPEC-V3R2-CON-002 REQ-CON-002-004 Layer 1 implementation.
func (g *frozenGuard) Check(proposal *AmendmentProposal, currentZone Zone) error {
	if currentZone != ZoneFrozen {
		// Evolvable zone passes
		return nil
	}

	// Frozen zone: Evidence required
	if proposal.Evidence == "" {
		return &ErrFrozenAmendment{
			RuleID: proposal.RuleID,
			Reason: "Frozen zone rule modification requires Evidence. Explain the Frozen→Evolvable demotion reason.",
		}
	}

	// Frozen zone modification with Evidence is allowed (assuming demotion)
	// Actual zone change is verified during registry update
	return nil
}

// frozenGuard satisfies the FrozenGuard interface.
var _ FrozenGuard = (*frozenGuard)(nil)
