package kickoff_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
)

// TestKickoffCheck is the plan→run gate verdict over seventeen fixtures
// (AC-GR-016). Each expected reason is the representative reason under the
// fixed priority; the reason list carries every reason that holds.
func TestKickoffCheck(t *testing.T) {
	type tc struct {
		name     string
		setup    func(f *fx) kickoff.CheckInput
		pass     bool
		reason   string
		alsoHave []string
		state    string // expected verify state when non-empty
	}
	cases := []tc{
		{"01_human_signed", func(f *fx) kickoff.CheckInput {
			f.signHuman(true)
			return f.checkIn("human", false)
		}, true, "", nil, ""},
		{"02_agent_written_receipt", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, false)
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonReceiptNotIssued, nil, ""},
		{"03_llm_approve_active", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			return f.checkIn("llm", true)
		}, true, "", nil, ""},
		{"04_llm_jev_approve_active", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm+jev", "llm+jev", jevApprove, true)
			return f.checkIn("llm+jev", true)
		}, true, "", nil, ""},
		{"05_fallback_receipt", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm+jev", "llm", fallbackLLM, true)
			return f.checkIn("llm+jev", true)
		}, true, "", nil, ""},
		{"06_signer_kind_jev", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			f.forge(func(s *contract.Signature) { s.SignerKind = "jev" }, true, true)
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonDeciderNotPermitted, []string{kickoff.ReasonNotSignedValid}, ""},
		{"07_inactive", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			return f.checkIn("llm", false)
		}, false, kickoff.ReasonInactive, nil, ""},
		{"08_revoke_event", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			if _, err := f.store.AppendEvent(receipt.KindRevoke, receipt.RevokeEvent{Spec: signtest.SpecID, Card: signtest.Card, Seal: f.seal()}); err != nil {
				f.t.Fatal(err)
			}
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonRevoked, nil, ""},
		{"09_human_without_event", func(f *fx) kickoff.CheckInput {
			f.signHuman(false)
			return f.checkIn("human", false)
		}, false, kickoff.ReasonSignatureNotRecorded, nil, ""},
		{"10_configured_decider_differs", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm+jev", "llm", fallbackLLM, true)
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonDeciderMismatch, nil, ""},
		{"11_card_mismatch", func(f *fx) kickoff.CheckInput {
			f.signHuman(true)
			in := f.checkIn("human", false)
			in.Card = "t9999"
			return in
		}, false, kickoff.ReasonCardMismatch, nil, ""},
		{"12_forged_reject_receipt", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			body := f.p.Receipt(f.p.ReceiptOptions("llm", "llm"), func(r *contract.KickoffReceipt) {
				r.LLMAnswer.Answer, r.Outcome = "reject", "reject"
			})
			f.p.WriteReceipt(body)
			f.issue(body)
			f.forge(func(s *contract.Signature) { s.Receipt.SHA256 = receipt.SHA256Hex(body) }, true, true)
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonReceiptNotApproved, nil, contract.StateSignedValid},
		{"13_forged_human_receipt", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm+jev", "llm+jev", jevApprove, true)
			body := f.p.Receipt(f.p.ReceiptOptions("llm+jev", "llm+jev"), func(r *contract.KickoffReceipt) {
				jevApprove(r)
				r.JevAnswer = signtest.JevAnswer("reject", 0.7)
				r.Outcome = "human"
			})
			f.p.WriteReceipt(body)
			f.issue(body)
			f.forge(func(s *contract.Signature) { s.Receipt.SHA256 = receipt.SHA256Hex(body) }, true, true)
			return f.checkIn("llm+jev", true)
		}, false, kickoff.ReasonReceiptNotApproved, nil, contract.StateSignedValid},
		{"14_seal_mismatch", func(f *fx) kickoff.CheckInput {
			f.signHuman(true)
			f.forge(func(s *contract.Signature) { s.Seal = strings.Repeat("0", 64) }, false, false)
			return f.checkIn("human", false)
		}, false, kickoff.ReasonNotSignedValid, nil, contract.StateSignedInvalid},
		{"15_unsigned", func(f *fx) kickoff.CheckInput {
			return f.checkIn("human", false)
		}, false, kickoff.ReasonNotSignedValid, nil, contract.StateUnsigned},
		{"16_effective_llm_signer_llm_jev", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			f.forge(func(s *contract.Signature) { s.SignerKind = "llm+jev" }, true, true)
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonDeciderMismatch, nil, ""},
		{"17_revoke_record_only", func(f *fx) kickoff.CheckInput {
			f.signReceipt("llm", "llm", nil, true)
			fp := escalation.Fingerprint("revoke-operator", f.seal())
			rec := escalation.Record{SchemaVersion: 1, Card: signtest.Card, Spec: signtest.SpecID,
				Kind: escalation.KindRevoke, Class: "revoke-operator", Fingerprint: fp,
				Status: escalation.StatusResolved, Decider: "human", Occurrences: 1,
				Observation: "revoked", Options: []string{"re-sign", "abandon"}}
			data, err := rec.Marshal()
			if err != nil {
				f.t.Fatal(err)
			}
			path := escalation.RecordPath(f.p.Root, signtest.Card, "revoke-operator", fp, 0)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				f.t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0o644); err != nil {
				f.t.Fatal(err)
			}
			return f.checkIn("llm", true)
		}, false, kickoff.ReasonRevoked, nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFx(t)
			in := c.setup(f)
			snap := f.p.Snapshot()
			evBefore, _ := f.store.Events()
			res, err := kickoff.Check(in)
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			if res.Pass != c.pass || res.Reason != c.reason {
				t.Fatalf("pass=%v reason=%q reasons=%v state=%s, want pass=%v reason=%q",
					res.Pass, res.Reason, res.Reasons, res.VerifyState, c.pass, c.reason)
			}
			for _, r := range c.alsoHave {
				if !slices.Contains(res.Reasons, r) {
					t.Errorf("reasons %v lack %q", res.Reasons, r)
				}
			}
			if c.reason != "" && !slices.Contains(res.Reasons, c.reason) {
				t.Errorf("reasons %v lack the representative %q", res.Reasons, c.reason)
			}
			if !slices.IsSorted(res.Reasons) || len(slices.Compact(slices.Clone(res.Reasons))) != len(res.Reasons) {
				t.Errorf("reasons %v not sorted and de-duplicated", res.Reasons)
			}
			if c.state != "" && res.VerifyState != c.state {
				t.Errorf("verify state %q, want %q", res.VerifyState, c.state)
			}
			if res.AutonomousKickoffEnabled != in.Enabled || res.JevDoctrineAmended != in.DoctrineAmended {
				t.Errorf("activation fields %v/%v, want %v/%v", res.AutonomousKickoffEnabled, res.JevDoctrineAmended, in.Enabled, in.DoctrineAmended)
			}
			f.p.AssertUnchanged(t, snap)
			if evAfter, _ := f.store.Events(); len(evAfter) != len(evBefore) {
				t.Error("kickoff-check wrote to the store")
			}
		})
	}
}
