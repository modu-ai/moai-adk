package sign_test

import (
	"errors"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

func storeEvents(t *testing.T, p *signtest.Project) []receipt.Line {
	t.Helper()
	s, err := receipt.Open(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	ev, err := s.Events()
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

// TestSignRecordsEvent checks that every signature appends exactly one event
// of the right kind, carrying the signature's seal, and that a failing store
// leaves the contract unwritten (AC-GR-021 signer half).
func TestSignRecordsEvent(t *testing.T) {
	t.Run("human", func(t *testing.T) {
		p := signtest.New(t)
		seams, rec := humanSeams(signtest.SpecID)
		mustSign(t, p.Options(), seams, rec)
		ev := storeEvents(t, p)
		if len(ev) != 1 || ev[0].Kind != receipt.KindSignHuman {
			t.Fatalf("events = %+v, want one sign-human", ev)
		}
		var e receipt.SignEvent
		if err := ev[0].Decode(&e); err != nil {
			t.Fatal(err)
		}
		c := decodeSigned(t, p, signtest.SpecID)
		if e.Seal != c.Signature.Seal || e.Card != signtest.Card || e.Spec != signtest.SpecID ||
			e.SignerKind != "human" || e.Method != contract.MethodInteractiveTTY ||
			e.ContractSHA256 != c.Signature.ContractSHA256 || e.AcceptanceSHA256 != c.Signature.AcceptanceSHA256 {
			t.Errorf("event %+v does not match signature %+v", e, c.Signature)
		}
	})
	t.Run("receipt", func(t *testing.T) {
		p := signtest.New(t)
		opts := p.ReceiptOptions("llm", "llm")
		p.WriteReceipt(p.Receipt(opts, nil))
		seams, rec := signtest.Seams(false, noMarkers)
		mustSign(t, opts, seams, rec)
		ev := storeEvents(t, p)
		if len(ev) != 1 || ev[0].Kind != receipt.KindSignReceipt {
			t.Fatalf("events = %+v, want one sign-receipt", ev)
		}
		var e receipt.SignEvent
		_ = ev[0].Decode(&e)
		c := decodeSigned(t, p, signtest.SpecID)
		if e.SignerKind != "llm" || e.ReceiptSHA256 != c.Signature.Receipt.SHA256 || e.Seal != c.Signature.Seal {
			t.Errorf("event %+v does not match signature", e)
		}
	})
	t.Run("reseal", func(t *testing.T) {
		p := signtest.New(t)
		seams, rec := humanSeams(signtest.SpecID)
		mustSign(t, p.Options(), seams, rec)
		p.WriteFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile), signtest.AcceptanceThree)
		o := p.Options()
		o.Resign = true
		seams, rec = humanSeams(signtest.SpecID)
		mustSign(t, o, seams, rec)
		ev := storeEvents(t, p)
		if len(ev) != 2 || ev[1].Kind != receipt.KindReseal {
			t.Fatalf("events = %+v, want sign-human then reseal", ev)
		}
		var e receipt.SignEvent
		_ = ev[1].Decode(&e)
		if e.Supersedes == "" || e.Seal != decodeSigned(t, p, signtest.SpecID).Signature.Seal {
			t.Errorf("reseal event %+v", e)
		}
	})
	t.Run("store-failure-leaves-contract-unwritten", func(t *testing.T) {
		p := signtest.New(t)
		snap := p.Snapshot()
		seams, _ := humanSeams(signtest.SpecID)
		seams.RecordEvent = func(string, string, receipt.SignEvent) error { return errors.New("store down") }
		res, err := sign.Sign(p.Options(), seams)
		if err == nil {
			t.Fatalf("Sign succeeded with a failing store (result %+v)", res)
		}
		p.AssertUnchanged(t, snap)
		if _, err := os.Stat(p.Path(signtest.SpecRel(signtest.SpecID, contract.ContractFile))); err != nil {
			t.Fatal(err)
		}
		if c := decodeSigned(t, p, signtest.SpecID); c.Signature != nil {
			t.Error("contract carries a signature although the event was not recorded")
		}
	})
}

// TestSignInterimRuleFollowsDoctrine replaces the A1 interim-rule test row:
// at every head it injects both doctrine states. False: an effective llm+jev
// receipt whose answers both approve is refused as receipt_requires_human.
// True: the same receipt signs, and llm+jev receipts recording reject or
// human are refused as receipt_rejected and receipt_requires_human
// (AC-GR-017).
func TestSignInterimRuleFollowsDoctrine(t *testing.T) {
	jev := func(outcome, llm, jevAns string) func(*contract.KickoffReceipt) {
		return func(r *contract.KickoffReceipt) {
			r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm+jev"
			r.LLMAnswer.Answer = llm
			r.JevAnswer = signtest.JevAnswer(jevAns, 0.71)
			r.Outcome = outcome
		}
	}
	cases := []struct {
		name     string
		doctrine bool
		mutate   func(*contract.KickoffReceipt)
		want     string // "" = signed
	}{
		{"false/both-approve", false, jev("approve", "approve", "approve"), contract.RefuseReceiptRequiresHuman},
		{"true/both-approve", true, jev("approve", "approve", "approve"), ""},
		{"true/reject", true, jev("reject", "reject", "reject"), contract.RefuseReceiptRejected},
		{"true/human", true, jev("human", "approve", "reject"), contract.RefuseReceiptRequiresHuman},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := signtest.New(t)
			opts := p.ReceiptOptions("llm+jev", "llm+jev")
			opts.JevDoctrineAmended = tc.doctrine
			p.WriteReceipt(p.Receipt(opts, tc.mutate))
			seams, rec := signtest.Seams(false, noMarkers)
			if tc.want == "" {
				mustSign(t, opts, seams, rec)
				assertSignedValid(t, p, signtest.SpecID, opts)
				return
			}
			assertRefuses(t, p, opts, seams, rec, tc.want)
		})
	}
}
