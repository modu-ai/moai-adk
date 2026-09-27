package kickoff_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// fx is one kickoff fixture: a signtest project and its contract store.
type fx struct {
	t     *testing.T
	p     *signtest.Project
	store *receipt.Store
}

func newFx(t *testing.T) *fx {
	t.Helper()
	p := signtest.New(t) // isolates MOAI_HOME
	s, err := receipt.Open(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	return &fx{t: t, p: p, store: s}
}

// policy is the verify policy the fixture signs under.
func (f *fx) policy() contract.Policy {
	o := f.p.Options()
	return contract.Policy{SecondReview: o.SecondReview, PushDevelop: o.PushDevelop, Mode: "contract", BudgetDefault: o.BudgetDefault}
}

func (f *fx) checkIn(decider string, enabled bool) kickoff.CheckInput {
	return kickoff.CheckInput{
		Root: f.p.Root, SpecID: signtest.SpecID, Card: signtest.Card,
		Config:  kickoff.Config{Mode: "contract", Decider: decider, JevEnabled: true, JevMinConfidence: 0.5},
		Policy:  f.policy(),
		Enabled: enabled, Store: f.store,
	}
}

// signHuman signs on the human path; record=false skips the store event.
func (f *fx) signHuman(record bool) {
	f.t.Helper()
	seams, rec := signtest.Seams(true, map[string]string{}, signtest.SpecID)
	if !record {
		seams.RecordEvent = func(string, string, receipt.SignEvent) error { return nil }
	}
	res, err := sign.Sign(f.p.Options(), seams)
	if err != nil || res.Refusal != "" {
		f.t.Fatalf("human sign: %v %s\n%s", err, res.Refusal, rec.Out.String())
	}
}

// signReceipt writes a receipt, optionally issues it into the store, and
// signs on the receipt path with the given configured decider and signer.
func (f *fx) signReceipt(decider, signer string, mutate func(*contract.KickoffReceipt), issue bool) []byte {
	f.t.Helper()
	opts := f.p.ReceiptOptions(decider, signer)
	opts.JevDoctrineAmended = true // the fixture needs llm+jev signatures to exist
	body := f.p.Receipt(opts, mutate)
	f.p.WriteReceipt(body)
	if issue {
		f.issue(body)
	}
	seams, rec := signtest.Seams(false, map[string]string{})
	res, err := sign.Sign(opts, seams)
	if err != nil || res.Refusal != "" {
		f.t.Fatalf("receipt sign: %v %s\n%s", err, res.Refusal, rec.Out.String())
	}
	return body
}

func (f *fx) issue(body []byte) {
	f.t.Helper()
	if _, err := f.store.AppendReceipt(receipt.ReceiptData{Spec: signtest.SpecID, Card: signtest.Card,
		Body: string(body), BodySHA256: receipt.SHA256Hex(body)}); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fx) contract() *contract.Contract {
	f.t.Helper()
	c, err := contract.Decode(f.p.ReadFile(signtest.SpecRel(signtest.SpecID, contract.ContractFile)))
	if err != nil {
		f.t.Fatal(err)
	}
	return c
}

func (f *fx) seal() string { return f.contract().Signature.Seal }

// forge rewrites the signature through mutate, re-seals it when reseal is
// true, and appends a chain-valid sign-receipt event for it when event is
// true — the shape of a forger with write access to the tree and the store.
func (f *fx) forge(mutate func(*contract.Signature), reseal, event bool) {
	f.t.Helper()
	c := f.contract()
	mutate(c.Signature)
	if reseal {
		s := *c.Signature
		s.Seal = ""
		sealed, err := contract.ComputeSeal(s)
		if err != nil {
			f.t.Fatal(err)
		}
		c.Signature.Seal = sealed
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		f.t.Fatal(err)
	}
	f.p.WriteFile(signtest.SpecRel(signtest.SpecID, contract.ContractFile), string(data))
	if event {
		sig := c.Signature
		rs := ""
		if sig.Receipt != nil {
			rs = sig.Receipt.SHA256
		}
		if _, err := f.store.AppendEvent(receipt.KindSignReceipt, receipt.SignEvent{
			Spec: signtest.SpecID, Card: signtest.Card, SignerKind: sig.SignerKind, Method: sig.Method,
			Seal: sig.Seal, ContractSHA256: sig.ContractSHA256, AcceptanceSHA256: sig.AcceptanceSHA256,
			ReceiptSHA256: rs,
		}); err != nil {
			f.t.Fatal(err)
		}
	}
}

func jevApprove(r *contract.KickoffReceipt) {
	r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm+jev"
	r.JevAnswer = signtest.JevAnswer("approve", 0.71)
}

func fallbackLLM(r *contract.KickoffReceipt) {
	reason := "jev_call_failed"
	r.RequestedDecider, r.EffectiveDecider = "llm+jev", "llm"
	r.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &reason}
}
