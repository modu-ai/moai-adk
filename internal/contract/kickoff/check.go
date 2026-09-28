package kickoff

import (
	"errors"
	"fmt"
	"slices"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/revoke"
)

// reasonPriority orders the kickoff-check reasons; the first that holds is
// the representative reason.
var reasonPriority = []string{
	ReasonCardMismatch,
	ReasonDeciderNotPermitted,
	ReasonNotSignedValid,
	ReasonSignatureNotRecorded,
	ReasonReceiptNotIssued,
	ReasonReceiptNotApproved,
	ReasonDeciderMismatch,
	ReasonRevoked,
	ReasonInactive,
}

// revokeReader is the A3 revoke reader kickoff-check consults.
var revokeReader = revoke.Blocked

// Check evaluates the plan→run gate for one SPEC and card. It writes nothing.
// The gate passes only when the contract's card matches, verify reports
// signed-valid, the store records a signing event for the current seal, no
// revocation (store event or revoke record) covers that seal, and the
// signature is either a human interactive signature or an approve receipt
// moai issued whose deciders match the signature and the configuration while
// autonomous Kickoff is active. Every reason that holds is reported, sorted;
// Reason is the representative one under reasonPriority. An error (exit 2)
// is a usage, I/O, or store integrity failure.
//
// @MX:ANCHOR: [AUTO] The contract-mode plan→run gate.
// @MX:REASON: the kickoff-check CLI, the orchestrator's stage-boundary
// check, and the activation-order test all decide on its verdict; its reason
// priority is a published contract.
func Check(in CheckInput) (CheckResult, error) {
	res := CheckResult{
		Mode: in.Config.Mode, Decider: in.Config.Decider,
		AutonomousKickoffEnabled: in.Enabled, JevDoctrineAmended: in.DoctrineAmended,
		Reasons: []string{}, VerifyReasons: []string{},
	}
	if !contract.ValidSpecID(in.SpecID) || in.Card == "" || in.Root == "" {
		return res, fmt.Errorf("%w: need a SPEC ID, a card id, and a project root", ErrUsage)
	}
	st := in.Store
	if st == nil {
		var err error
		if st, err = receipt.Open(in.Root); err != nil {
			return res, err
		}
	}
	if err := st.Verify(); err != nil {
		return res, err
	}
	events, err := st.Events()
	if err != nil {
		return res, err
	}
	receipts, err := st.Receipts()
	if err != nil {
		return res, err
	}

	dir, err := contract.ResolveSpecDir(in.Root, in.SpecID)
	if err != nil {
		return res, fmt.Errorf("%w: %v", ErrUsage, err)
	}
	cin, err := contract.LoadDir(dir)
	holds := map[string]bool{}
	var c *contract.Contract
	switch {
	case errors.Is(err, contract.ErrContractMissing):
		res.VerifyState = contract.StateUnsigned
		holds[ReasonNotSignedValid] = true
	case err != nil:
		return res, err
	default:
		cin.Policy = in.Policy
		cin.RegistryRuleIDs, cin.RegistryFrozenFiles = in.RegistryRuleIDs, in.RegistryFrozenFiles
		cin.SpecStatus = in.SpecStatus
		rep := contract.Verify(cin)
		res.VerifyState, res.VerifyReasons = rep.State, rep.Reasons
		c = rep.Contract
		if rep.Card != in.Card {
			holds[ReasonCardMismatch] = true
		}
		if rep.State != contract.StateSignedValid {
			holds[ReasonNotSignedValid] = true
		}
	}
	if c != nil && c.Signature != nil {
		sig := c.Signature
		res.SignerKind = sig.SignerKind
		if err := judgeSignature(in, sig, events, receipts, holds); err != nil {
			return res, err
		}
	}

	for r, ok := range holds {
		if ok {
			res.Reasons = append(res.Reasons, r)
		}
	}
	slices.Sort(res.Reasons)
	for _, r := range reasonPriority {
		if holds[r] {
			res.Reason = r
			break
		}
	}
	res.Pass = res.Reason == ""
	return res, nil
}

// judgeSignature adds the reasons a present signature earns.
func judgeSignature(in CheckInput, sig *contract.Signature, events, receipts []receipt.Line, holds map[string]bool) error {
	recorded, err := receipt.SignatureRecorded(events, in.SpecID, in.Card, sig.Seal)
	if err != nil {
		return err
	}
	if !recorded {
		holds[ReasonSignatureNotRecorded] = true
	}
	revoked, err := receipt.Revoked(events, in.SpecID, in.Card, sig.Seal)
	if err != nil {
		return err
	}
	// A reader error is a block, never a silent pass.
	if blocked, _ := revokeReader(in.Root, in.Card, in.SpecID, sig.Seal); blocked || revoked {
		holds[ReasonRevoked] = true
	}

	switch {
	case sig.SignerKind == "jev":
		holds[ReasonDeciderNotPermitted] = true
	case sig.SignerKind == contract.SignerHuman && sig.Method == contract.MethodInteractiveTTY:
		// The human path needs nothing beyond the recorded, unrevoked seal.
	case (sig.SignerKind == contract.DeciderLLM || sig.SignerKind == contract.DeciderLLMJev) &&
		sig.Method == contract.MethodReceipt && sig.Receipt != nil:
		rd, err := receipt.FindReceipt(receipts, in.SpecID, in.Card, sig.Receipt.SHA256)
		if err != nil {
			return err
		}
		if rd == nil {
			holds[ReasonReceiptNotIssued] = true
		} else {
			r, err := contract.DecodeKickoffReceipt([]byte(rd.Body))
			if err != nil {
				holds[ReasonReceiptNotIssued] = true
			} else {
				if r.Outcome != "approve" {
					holds[ReasonReceiptNotApproved] = true
				}
				if r.EffectiveDecider != sig.SignerKind || r.RequestedDecider != in.Config.Decider {
					holds[ReasonDeciderMismatch] = true
				}
			}
		}
		if !in.Enabled {
			holds[ReasonInactive] = true
		}
	default:
		holds[ReasonNotSignedValid] = true
	}
	return nil
}
