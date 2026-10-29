package core

import "fmt"

// The six reasons S46 names only for PARKED (comment 6489610367: E02 dependency,
// E04 stop_unknown, E11 unknown-effect, E14 pending-effect, E24 contract-change,
// E26 start_unknown). S46 spells stop_unknown and start_unknown with
// underscores; the values follow the canon, not SPEC 0.6.1's hyphens.
const (
	ReasonDependency     BlockReason = "dependency"
	ReasonStopUnknown    BlockReason = "stop_unknown"
	ReasonUnknownEffect  BlockReason = "unknown-effect"
	ReasonPendingEffect  BlockReason = "pending-effect"
	ReasonContractChange BlockReason = "contract-change"
	ReasonStartUnknown   BlockReason = "start_unknown"
)

// ReasonFromCaller marks a transition outcome whose overlay reason S46 leaves to
// the caller's classification (E15 cleanup unknown, E19, E23 inconclusive).
const ReasonFromCaller BlockReason = "(caller-supplied)"

// ReasonClassified marks E19: the reason must be one of ClassifiedReasons, the
// result of classifying the real error (rate limit, permission, budget ...).
const ReasonClassified BlockReason = "(classified)"

var canonParkReasons = []BlockReason{
	ReasonDependency, ReasonStopUnknown, ReasonUnknownEffect, ReasonPendingEffect, ReasonContractChange, ReasonStartUnknown,
}

// canonUnconfirmedParkReasons are overlay reasons the code uses that S46 does not
// name. operator-hold is one: S46 has the CLI pause/resume only (operator
// decision OD-12 is open).
var canonUnconfirmedParkReasons = []BlockReason{ReasonOperatorHold}

// parkable are the reasons that may sit in an overlay besides the canon six:
// the E19 classifications (quota, permission, budget, environment) and the
// control reasons of the UNAUDITED SPEC 0.6.1 §2.1a column "P". It rests on that
// SPEC column, not on the canon: do not extend it until the SPEC is audited.
// Refusal (R) and merge-gate (G) reasons never become an overlay.
var parkable = map[BlockReason]bool{
	ReasonNotInstalled: true, ReasonConnectionFailed: true, ReasonNotLoggedIn: true, ReasonNotConfigured: true,
	ReasonAuthFailed: true, ReasonPermissionDenied: true, ReasonRateLimited: true, ReasonUnclassified: true,
	ReasonCapacityShort: true, ReasonBudgetExhausted: true, ReasonStoreUnavailable: true, ReasonUserControl: true,
	ReasonOperatorAckMissing: true,
}

func CanonParkReasons() []BlockReason { return append([]BlockReason(nil), canonParkReasons...) }

func CanonUnconfirmedParkReasons() []BlockReason {
	return append([]BlockReason(nil), canonUnconfirmedParkReasons...)
}

// IsCanonParkReason reports whether S46 itself names r as a PARKED reason.
func IsCanonParkReason(r BlockReason) bool {
	for _, c := range canonParkReasons {
		if c == r {
			return true
		}
	}
	return false
}

func parkableReason(r BlockReason) bool {
	if parkable[r] || IsCanonParkReason(r) {
		return true
	}
	for _, c := range canonUnconfirmedParkReasons {
		if c == r {
			return true
		}
	}
	return false
}

// Park is the PARKED overlay: one reason, its evidence and the resume condition.
type Park struct {
	Reason   BlockReason `json:"reason"`
	Evidence string      `json:"evidence"`
	Resume   string      `json:"resume_condition"`
}

func (p Park) Validate() error {
	if !parkableReason(p.Reason) {
		return fmt.Errorf("%w: %q cannot be a PARKED reason", ErrReasonNotAllowed, p.Reason)
	}
	if p.Evidence == "" || p.Resume == "" {
		return fmt.Errorf("%w: a park needs evidence and a resume condition", ErrEvidenceIncomplete)
	}
	return nil
}

// unknownFamily are the overlay reasons that record an unknown outcome or state.
// They are never overwritten by another reason, only resumed.
func unknownFamily(r BlockReason) bool {
	switch r {
	case ReasonUnknownEffect, ReasonPendingEffect, ReasonStopUnknown, ReasonStartUnknown:
		return true
	}
	return false
}

// classified is the set of overlay reasons that classify an observed error:
// quota, permission, budget, environment and store failures. The forge Kinds map
// into it (Limited -> rate-limited, Denied -> permission-denied, Unclassified ->
// unclassified-403, see mirror). Control reasons such as operator-hold or
// dependency have their own edges and are not here.
var classified = []BlockReason{
	ReasonNotInstalled, ReasonConnectionFailed, ReasonNotLoggedIn, ReasonNotConfigured, ReasonAuthFailed,
	ReasonPermissionDenied, ReasonRateLimited, ReasonUnclassified, ReasonCapacityShort, ReasonBudgetExhausted,
	ReasonStoreUnavailable,
}

func ClassifiedReasons() []BlockReason { return append([]BlockReason(nil), classified...) }

func isClassified(r BlockReason) bool {
	for _, c := range classified {
		if c == r {
			return true
		}
	}
	return false
}

// IsParkable reports whether r may sit in a PARKED overlay. Gate (G) and
// refusal (R) reasons may not.
func IsParkable(r BlockReason) bool { return parkableReason(r) }
