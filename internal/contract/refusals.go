package contract

import "slices"

// Sign refusal codes — the closed set of design.md § Sign Refusal Codes.
// `moai contract sign` prints one of these and exits 1. Declared in the core
// so the signer and the CLI share one vocabulary; the verify reason codes
// (reasons.go) are a separate closed set.
const (
	RefuseAgentMarker           = "agent_marker"
	RefuseNotTTY                = "not_tty"
	RefuseConfirmationMismatch  = "confirmation_mismatch"
	RefuseAlreadySigned         = "already_signed"
	RefuseNotSigned             = "not_signed"
	RefuseDraftAcceptanceStale  = "draft_acceptance_stale"
	RefuseACCountAmbiguous      = "ac_count_ambiguous"
	RefuseACCountZero           = "ac_count_zero"
	RefusePlanAuditNotPassing   = "plan_audit_not_passing"
	RefuseGitIdentityMissing    = "git_identity_missing"
	RefuseBatchDisabled         = "batch_disabled"
	RefuseBatchNonHuman         = "batch_non_human"
	RefuseModeNotContract       = "mode_not_contract"
	RefuseKickoffDeciderJevSole = "kickoff_decider_jev_sole"
	RefuseReceiptInvalid        = "receipt_invalid"
	RefuseReceiptSignerMismatch = "receipt_signer_mismatch"
	RefuseReceiptInputMismatch  = "receipt_input_mismatch"
	RefuseReceiptRequiresHuman  = "receipt_requires_human"
	RefuseReceiptRejected       = "receipt_rejected"
	RefuseVerifyFailed          = "verify_failed"
)

var signRefusalCodes = []string{
	RefuseAgentMarker,
	RefuseNotTTY,
	RefuseConfirmationMismatch,
	RefuseAlreadySigned,
	RefuseNotSigned,
	RefuseDraftAcceptanceStale,
	RefuseACCountAmbiguous,
	RefuseACCountZero,
	RefusePlanAuditNotPassing,
	RefuseGitIdentityMissing,
	RefuseBatchDisabled,
	RefuseBatchNonHuman,
	RefuseModeNotContract,
	RefuseKickoffDeciderJevSole,
	RefuseReceiptInvalid,
	RefuseReceiptSignerMismatch,
	RefuseReceiptInputMismatch,
	RefuseReceiptRequiresHuman,
	RefuseReceiptRejected,
	RefuseVerifyFailed,
}

// SignRefusalCodes returns the closed sign refusal-code set in design order.
// The returned slice is a fresh copy.
func SignRefusalCodes() []string {
	return slices.Clone(signRefusalCodes)
}

// IsSignRefusalCode reports whether s belongs to the sign refusal-code set.
func IsSignRefusalCode(s string) bool {
	return slices.Contains(signRefusalCodes, s)
}
