package core

import "fmt"

// gateReasons are the merge-gate (G) reasons of SPEC 0.6.1 section 2.1a: they
// record that the evidence does not yet allow the next step. They are not
// PARKED overlays: the stage stays where it is and a Gate record names why.
var gateReasons = []BlockReason{ReasonCheckNotPassed, ReasonDebtPresent, ReasonVerdictNotPass, ReasonProducerNotAllowed}

func GateReasons() []BlockReason { return append([]BlockReason(nil), gateReasons...) }

func IsGateReason(r BlockReason) bool {
	for _, g := range gateReasons {
		if g == r {
			return true
		}
	}
	return false
}

// Gate is the merge-gate record on a stage. Recording the same gate again for
// the same HEAD and evidence changes nothing else; only a new HEAD or new
// evidence can turn it into a pass.
type Gate struct {
	Reason BlockReason `json:"reason"`
	Detail string      `json:"detail"`
	Resume string      `json:"resume_condition"`
	Head   HeadSHA     `json:"head,omitempty"`
}

func (g Gate) Validate() error {
	if !IsGateReason(g.Reason) {
		return fmt.Errorf("%w: %q is not a gate reason", ErrReasonNotAllowed, g.Reason)
	}
	if g.Detail == "" || g.Resume == "" {
		return fmt.Errorf("%w: a gate needs a detail and a resume condition", ErrEvidenceIncomplete)
	}
	return nil
}

// Refusal records that an operation was refused (SPEC section 2.1a column R).
// The state does not change; the last refusal is kept for the reader.
type Refusal struct {
	Reason BlockReason `json:"reason"`
	Detail string      `json:"detail"`
}

func (r Refusal) Validate() error {
	if !r.Reason.Valid() {
		return fmt.Errorf("%w: %q", ErrReasonUnknown, r.Reason)
	}
	if r.Detail == "" {
		return fmt.Errorf("%w: a refusal needs a detail", ErrEvidenceIncomplete)
	}
	return nil
}
