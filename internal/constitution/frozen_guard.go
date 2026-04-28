package constitution

// frozenGuard는 FrozenGuard interface의 구현이다.
// Frozen zone rule amendment를 차단한다.
type frozenGuard struct{}

// NewFrozenGuard는 FrozenGuard를 생성한다.
func NewFrozenGuard() FrozenGuard {
	return &frozenGuard{}
}

// Check는 proposal이 Frozen zone rule을 수정하는지 확인한다.
// Frozen→Evolvable demotion은 Evidence로 허용.
// SPEC-V3R2-CON-002 REQ-CON-002-004 Layer 1 구현.
func (g *frozenGuard) Check(proposal *AmendmentProposal, currentZone Zone) error {
	if currentZone != ZoneFrozen {
		// Evolvable zone은 통과
		return nil
	}

	// Frozen zone: Evidence 필수
	if proposal.Evidence == "" {
		return &ErrFrozenAmendment{
			RuleID: proposal.RuleID,
			Reason: "Frozen zone rule 수정에는 Evidence(증거)가 필수이다. Frozen→Evolvable demotion 사유를 설명하라.",
		}
	}

	// Evidence가 있는 Frozen zone 수정은 허용 (demotion 가정)
	// 실제 zone 변경 여부는 registry 업데이트 시 검증
	return nil
}

// frozenGuard는 FrozenGuard interface를 만족한다.
var _ FrozenGuard = (*frozenGuard)(nil)
