// Package sign signs SPEC autonomy contracts (`moai contract sign`;
// SPEC-AUTONOMY-CONTRACT-001 REQ-CONTRACT-010..013, 019, 021, 022, 024).
//
// It sits beside the pure verification core internal/contract and carries
// every side effect signing needs — the terminal check, the confirmation
// read, the git subprocesses, the atomic write — so that the core stays free
// of os/exec. It imports internal/contract, internal/atomicfile, the standard
// library, and gopkg.in/yaml.v3; it never imports internal/config (callers,
// including configuration tests, map their effective workflow.autonomy values
// into Options), and internal/contract never imports it.
//
// # Public API
//
//	Sign(opts Options, seams Seams) (Result, error)  // design.md § Signing Flow
//	Options   // invocation + effective policy (mode, batch_sign, decider, …)
//	Seams     // IsTTY, Getenv, ReadLine, Out, Now, GitIdentity, GitHead, NewBatchID, WriteFile
//	Result    // Refusal code, SpecID, Cause, Reasons (verify_failed), Written / Unwritten
//	KickoffNotice   // REQ-CONTRACT-019 notice text
//	ErrUsage, ErrInternal
//
// # Outcomes
//
//   - Signed: Result.Refusal == "", Written lists every contract, the output
//     ends with KickoffNotice (both paths, both modes). Exit 0.
//   - Refused: Result.Refusal is one code of contract.SignRefusalCodes and
//     no file changed. Output line: `refused <code>[ (<SPEC-ID>)]: <cause>`,
//     followed by `reasons: <code>, …` for verify_failed. Exit 1.
//   - Error: a usage error (wraps ErrUsage), an I/O error (missing
//     contract.yaml, a path escaping the SPEC directory, a git failure), an
//     internal error (wraps ErrInternal: the produced file did not verify
//     signed-valid; nothing was written), or a write failure (Result carries
//     the exact Written / Unwritten split). Exit 2.
//
// # Flow (design.md § Signing Flow, in order)
//
//  1. Path gates. Human path (Signer "" or "human"): agent_marker (the
//     first Options.AgentMarkers variable set to a non-empty value, named in
//     the output), not_tty, then batch_disabled for more than one distinct
//     ID when BatchSign is false — before any prompt. Receipt path ("llm" /
//     "llm+jev"): batch_non_human, mode_not_contract, kickoff_decider_jev_sole.
//  2. Per SPEC (every SPEC before any prompt or write): load and strictly
//     decode contract.yaml; already_signed / not_signed by signature state
//     and --resign; measure acceptance.md (ac_count_ambiguous, ac_count_zero,
//     draft_acceptance_stale for an unsigned draft whose recorded hash
//     differs); write the measured binding and fill an absent budget from
//     BudgetDefault; verify the would-be contract — any reason other than
//     unsigned and plan_audit_not_passing refuses verify_failed (reasons
//     printed); plan_audit_not_passing alone refuses with that code.
//  3. git user.name / user.email (git_identity_missing when either is
//     empty) and HEAD.
//  4. Human path: the summary (ID, acceptance hash prefix and AC count —
//     `old → new` on --resign — actions, budget, second model, plan-audit
//     verdict marked "(carried)" on --resign), the token (the SPEC ID, or
//     `sign N contracts` in batch mode), one ReadLine; confirmation_mismatch.
//     Receipt path: --receipt relativized against ProjectRoot, then
//     contract.ValidateKickoffReceipt and contract.ReceiptOutcome (the
//     interim A1 rule refuses an effective llm+jev receipt with
//     receipt_requires_human). No terminal, no ReadLine.
//  5. The signature: signer_kind, operator, signed_at (UTC RFC 3339),
//     head_sha, contract_sha256 (digest of the final body), acceptance_sha256,
//     method, receipt {kickoff-receipt.json, raw SHA-256, file}, batch_id
//     (batch), supersedes (--resign: the previous contract_sha256), and the
//     seal computed last. Each produced file is re-decoded and verified
//     signed-valid before any write; files are then written atomically, one
//     at a time.
//
// The author's text is kept: the acceptance binding, a missing budget, and
// the signature are spliced into the original lines, so comments, blank
// lines, and quoting survive. A document shape the splice does not handle
// (for example a flow-style acceptance mapping) is re-encoded from its
// yaml.Node tree instead, which keeps comments but not blank lines. Either
// rendering is used only when it decodes to exactly the intended body.
package sign
