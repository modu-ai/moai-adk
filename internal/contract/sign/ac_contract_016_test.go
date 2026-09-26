package sign_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// TestAC_CONTRACT_016 implements AC-CONTRACT-016: sign-time refusals on the
// human path (a)-(g) and receipt-path refusals (h)-(t). Every refusal must
// name its code and leave every fixture file byte-identical.
func TestAC_CONTRACT_016(t *testing.T) {
	humanCases := []struct {
		name  string
		setup func(p *signtest.Project, seams *sign.Seams)
		want  string
	}{
		{"a_draft_acceptance_stale", func(p *signtest.Project, _ *sign.Seams) {
			p.AddSpec(signtest.SpecID, signtest.Draft{RecordedSHA256: strings.Repeat("0", 64)}, signtest.Acceptance)
		}, contract.RefuseDraftAcceptanceStale},
		{"b_ac_count_ambiguous", func(p *signtest.Project, _ *sign.Seams) {
			p.AddSpec(signtest.SpecID, signtest.Draft{}, signtest.AcceptanceAmbiguous)
		}, contract.RefuseACCountAmbiguous},
		{"c_plan_audit_not_passing", func(p *signtest.Project, _ *sign.Seams) {
			p.AddSpec(signtest.SpecID, signtest.Draft{Verdict: "FAIL"}, signtest.Acceptance)
		}, contract.RefusePlanAuditNotPassing},
		{"d_verify_failed_forbidden_action", func(p *signtest.Project, _ *sign.Seams) {
			p.AddSpec(signtest.SpecID, signtest.Draft{Actions: []string{"push-main"}}, signtest.Acceptance)
		}, contract.RefuseVerifyFailed},
		{"e_git_identity_missing", func(_ *signtest.Project, seams *sign.Seams) {
			seams.GitIdentity = func(string) (string, string, error) { return signtest.OperatorName, "", nil }
		}, contract.RefuseGitIdentityMissing},
		{"f_ac_count_zero", func(p *signtest.Project, _ *sign.Seams) {
			p.AddSpec(signtest.SpecID, signtest.Draft{}, signtest.AcceptanceZero)
		}, contract.RefuseACCountZero},
	}
	for _, tc := range humanCases {
		t.Run(tc.name, func(t *testing.T) {
			p := signtest.New(t)
			seams, rec := humanSeams(signtest.SpecID)
			tc.setup(p, &seams)
			res := assertRefuses(t, p, p.Options(), seams, rec, tc.want)
			if tc.want == contract.RefuseVerifyFailed {
				if len(res.Reasons) == 0 || !strings.Contains(strings.Join(res.Reasons, ","), contract.ReasonForbiddenAction) {
					t.Errorf("verify_failed reasons = %v, want forbidden_action", res.Reasons)
				}
				if !strings.Contains(rec.Out.String(), contract.ReasonForbiddenAction) {
					t.Errorf("output does not print the verify reasons:\n%s", rec.Out.String())
				}
			}
		})
	}

	t.Run("g_already_signed", func(t *testing.T) {
		p := signtest.New(t)
		seams, rec := humanSeams(signtest.SpecID)
		mustSign(t, p.Options(), seams, rec)
		assertSignedValid(t, p, signtest.SpecID, p.Options())
		seams2, rec2 := humanSeams(signtest.SpecID)
		assertRefuses(t, p, p.Options(), seams2, rec2, contract.RefuseAlreadySigned)
	})

	jevApprove := func(r *contract.KickoffReceipt) { r.JevAnswer = signtest.JevAnswer("approve", 0.71) }
	receiptCases := []struct {
		name            string
		decider, signer string
		deciderJevSole  bool
		receiptPath     string // "" = the fixed path
		mutate          func(r *contract.KickoffReceipt)
		want            string
	}{
		{"h_llm_answer_without_reason_refs", "llm", "llm", false, "", func(r *contract.KickoffReceipt) {
			r.LLMAnswer.ReasonRefs = nil
		}, contract.RefuseReceiptInvalid},
		{"i_silent_fallback", "llm+jev", "llm", false, "", func(r *contract.KickoffReceipt) {
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm"
			r.Fallback = &contract.ReceiptFallback{Applied: false}
		}, contract.RefuseReceiptSignerMismatch},
		{"j_acceptance_input_mismatch", "llm", "llm", false, "", func(r *contract.KickoffReceipt) {
			r.Inputs.AcceptanceSHA256 = strings.Repeat("0", 64)
		}, contract.RefuseReceiptInputMismatch},
		{"k_llm_jev_without_jev_answer", "llm+jev", "llm+jev", false, "", func(r *contract.KickoffReceipt) {
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm+jev"
		}, contract.RefuseReceiptInvalid},
		{"l_fallback_reason_outside_set", "llm+jev", "llm", false, "", func(r *contract.KickoffReceipt) {
			reason := "jev_timeout"
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm"
			r.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &reason}
		}, contract.RefuseReceiptInvalid},
		{"m_approve_without_llm_approve", "llm+jev", "llm+jev", false, "", func(r *contract.KickoffReceipt) {
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm+jev"
			r.LLMAnswer.Answer = "reject"
			jevApprove(r)
			r.Outcome = "approve"
		}, contract.RefuseReceiptInvalid},
		{"n_outcome_human", "llm", "llm", false, "", func(r *contract.KickoffReceipt) {
			r.Outcome = "human"
		}, contract.RefuseReceiptRequiresHuman},
		{"o_receipt_outside_fixed_path", "llm", "llm", false, "OUTSIDE", nil, contract.RefuseReceiptInvalid},
		{"p_outcome_reject", "llm", "llm", false, "", func(r *contract.KickoffReceipt) {
			r.LLMAnswer.Answer = "reject"
			r.Outcome = "reject"
		}, contract.RefuseReceiptRejected},
		{"q_unrequested_fallback", "llm", "llm", false, "", func(r *contract.KickoffReceipt) {
			reason := "jev_call_failed"
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm"
			r.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &reason}
		}, contract.RefuseReceiptSignerMismatch},
		{"r_configured_decider_jev", "jev", "llm", true, "", nil, contract.RefuseKickoffDeciderJevSole},
		{"s_llm_with_jev_answer", "llm", "llm", false, "", jevApprove, contract.RefuseReceiptInvalid},
		{"t_llm_jev_both_approve_interim_rule", "llm+jev", "llm+jev", false, "", func(r *contract.KickoffReceipt) {
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm+jev"
			jevApprove(r)
		}, contract.RefuseReceiptRequiresHuman},
	}
	for _, tc := range receiptCases {
		t.Run(tc.name, func(t *testing.T) {
			p := signtest.New(t)
			opts := p.ReceiptOptions(tc.decider, tc.signer)
			opts.DeciderJevSole = tc.deciderJevSole
			opts.JevEnabled = true
			receipt := p.Receipt(opts, tc.mutate)
			p.WriteReceipt(receipt)
			if tc.receiptPath == "OUTSIDE" {
				p.WriteFile("receipt.json", string(receipt))
				opts.ReceiptPath = p.Path("receipt.json")
			}
			seams, rec := signtest.Seams(false, noMarkers)
			assertRefuses(t, p, opts, seams, rec, tc.want)
			if rec.ReadLines != 0 {
				t.Errorf("receipt path read %d confirmation lines, want 0", rec.ReadLines)
			}
		})
	}
}
