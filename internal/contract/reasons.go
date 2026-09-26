package contract

import "slices"

// Verify reason codes — the closed set of design.md § Verify Reason Codes.
// Emitted in `verify --json` / `show --json` as `reasons: [...]`, sorted and
// de-duplicated. Adding a code is a schema amendment, not a local change.
const (
	ReasonSchemaInvalid               = "schema_invalid"
	ReasonSpecIDMismatch              = "spec_id_mismatch"
	ReasonUnsigned                    = "unsigned"
	ReasonContractDigestMismatch      = "contract_digest_mismatch"
	ReasonAcceptanceMissing           = "acceptance_missing"
	ReasonAcceptanceHashMismatch      = "acceptance_hash_mismatch"
	ReasonACCountMismatch             = "ac_count_mismatch"
	ReasonACCountAmbiguous            = "ac_count_ambiguous"
	ReasonActionsEmpty                = "actions_empty"
	ReasonUnknownAction               = "unknown_action"
	ReasonForbiddenAction             = "forbidden_action"
	ReasonPushDevelopDisabled         = "push_develop_disabled"
	ReasonSecondReviewMissing         = "second_review_missing"
	ReasonEscalateOnIncomplete        = "escalate_on_incomplete"
	ReasonOwnershipInvalid            = "ownership_invalid"
	ReasonInvariantUnresolved         = "invariant_unresolved"
	ReasonBudgetInvalid               = "budget_invalid"
	ReasonReobserveIncomplete         = "reobserve_incomplete"
	ReasonPlanAuditNotPassing         = "plan_audit_not_passing"
	ReasonReceiptMismatch             = "receipt_mismatch"
	ReasonSignatureSealMismatch       = "signature_seal_mismatch"
	ReasonSignatureInconsistent       = "signature_inconsistent"
	ReasonSignatureAcceptanceMismatch = "signature_acceptance_mismatch"
	ReasonCardInvalid                 = "card_invalid"
)

var reasonCodes = []string{
	ReasonSchemaInvalid,
	ReasonSpecIDMismatch,
	ReasonUnsigned,
	ReasonContractDigestMismatch,
	ReasonAcceptanceMissing,
	ReasonAcceptanceHashMismatch,
	ReasonACCountMismatch,
	ReasonACCountAmbiguous,
	ReasonActionsEmpty,
	ReasonUnknownAction,
	ReasonForbiddenAction,
	ReasonPushDevelopDisabled,
	ReasonSecondReviewMissing,
	ReasonEscalateOnIncomplete,
	ReasonOwnershipInvalid,
	ReasonInvariantUnresolved,
	ReasonBudgetInvalid,
	ReasonReobserveIncomplete,
	ReasonPlanAuditNotPassing,
	ReasonReceiptMismatch,
	ReasonSignatureSealMismatch,
	ReasonSignatureInconsistent,
	ReasonSignatureAcceptanceMismatch,
}

// ReasonCodes returns the closed set of verify reason codes in design order.
// The returned slice is a fresh copy.
func ReasonCodes() []string {
	return slices.Clone(reasonCodes)
}

// IsReasonCode reports whether s belongs to the closed reason-code set.
func IsReasonCode(s string) bool {
	return slices.Contains(reasonCodes, s)
}

// reasonSet accumulates reason codes; sorted returns them sorted and
// de-duplicated, never nil.
type reasonSet map[string]struct{}

func (s reasonSet) add(code string) { s[code] = struct{}{} }

func (s reasonSet) sorted() []string {
	out := make([]string, 0, len(s))
	for code := range s {
		out = append(out, code)
	}
	slices.Sort(out)
	return out
}
