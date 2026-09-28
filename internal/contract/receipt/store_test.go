package receipt_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/revoke"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/paths"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

func open(t *testing.T, p *signtest.Project) *receipt.Store {
	t.Helper()
	s, err := receipt.Open(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func events(t *testing.T, s *receipt.Store) []receipt.Line {
	t.Helper()
	ev, err := s.Events()
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

func signHuman(t *testing.T, p *signtest.Project, resign bool) {
	t.Helper()
	o := p.Options()
	o.Resign = resign
	seams, rec := signtest.Seams(true, map[string]string{}, signtest.SpecID)
	if res, err := sign.Sign(o, seams); err != nil || res.Refusal != "" {
		t.Fatalf("sign: %v %s\n%s", err, res.Refusal, rec.Out.String())
	}
}

func policy(p *signtest.Project) contract.Policy {
	o := p.Options()
	return contract.Policy{SecondReview: o.SecondReview, PushDevelop: o.PushDevelop, Mode: "contract", BudgetDefault: o.BudgetDefault}
}

// decideReady makes the fixture's decide preconditions hold: a SPEC commit
// carrying an author trailer and a bound PASS plan-audit report.
func decideReady(t *testing.T, p *signtest.Project) {
	t.Helper()
	spec := ".moai/specs/" + signtest.SpecID
	p.WriteFile(spec+"/research.md", "# research\n")
	p.Git("add", "-A")
	p.Git("commit", "-q", "-m", "docs: plan\n\nAuthored-By-Agent: manager-spec\n\n🗿 MoAI\n")
	h, err := runtime.NewInMemoryCache().ComputeHash(p.Path(spec))
	if err != nil {
		t.Fatal(err)
	}
	p.WriteFile(".moai/reports/"+signtest.Card+"/plan-audit-1.md",
		fmt.Sprintf("Verdict: PASS\nOverall Score: 0.90\nplan_artifact_hash: %s\n", h))
}

func decideIn(p *signtest.Project, s *receipt.Store, decider string, jr jev.Result) kickoff.DecideInput {
	j, _ := json.Marshal(map[string]any{"agent": "lead-session", "model": "m", "answer": "approve",
		"confidence": 0.8, "reason": "ok", "reason_refs": []string{"contract.yaml:2"}})
	return kickoff.DecideInput{
		Root: p.Root, SpecID: signtest.SpecID, Card: signtest.Card, Judgement: j,
		Config:          kickoff.Config{Mode: "contract", Decider: decider, JevEnabled: true, JevMinConfidence: 0.5},
		Policy:          policy(p),
		DoctrineAmended: true, Store: s,
		Jev: func(context.Context, jev.Request) jev.Result { return jr },
	}
}

// TestEventStore checks that every signing event kind lands as exactly one
// line, that the chain detects an edited middle line, and how the consumers
// cover a deleted last line (AC-GR-021).
func TestEventStore(t *testing.T) {
	t.Run("one-line-per-event-and-chain", func(t *testing.T) {
		p := signtest.New(t)
		s := open(t, p)
		want := []string{}
		step := func(kind string, do func()) {
			t.Helper()
			before := len(events(t, s))
			do()
			ev := events(t, s)
			if len(ev) != before+1 || ev[len(ev)-1].Kind != kind {
				t.Fatalf("%s: events %d→%d (last %q)", kind, before, len(ev), ev[len(ev)-1].Kind)
			}
			want = append(want, kind)
		}
		decideReady(t, p)
		step(receipt.KindDecide, func() {
			r0, _ := s.Receipts()
			if _, err := kickoff.Decide(decideIn(p, s, "llm", jev.Result{})); err != nil {
				t.Fatal(err)
			}
			r1, _ := s.Receipts()
			if len(r1) != len(r0)+1 {
				t.Errorf("decide added %d receipt lines, want 1", len(r1)-len(r0))
			}
		})
		step(receipt.KindSignReceipt, func() {
			o := p.ReceiptOptions("llm", "llm")
			seams, rec := signtest.Seams(false, map[string]string{})
			if res, err := sign.Sign(o, seams); err != nil || res.Refusal != "" {
				t.Fatalf("receipt sign: %v %s\n%s", err, res.Refusal, rec.Out.String())
			}
		})
		step(receipt.KindRevoke, func() {
			if _, err := revoke.Revoke(revoke.Options{Root: p.Root, SpecID: signtest.SpecID, Card: signtest.Card}, revoke.Seams{}); err != nil {
				t.Fatal(err)
			}
		})
		p.WriteFile(signtest.SpecRel(signtest.SpecID, contract.AcceptanceFile), signtest.AcceptanceThree)
		step(receipt.KindReseal, func() { signHuman(t, p, true) })

		q := signtest.New(t)
		step(receipt.KindSignHuman, func() {
			sq := open(t, q)
			before := len(events(t, sq))
			signHuman(t, q, false)
			if len(events(t, sq)) != before+1 {
				t.Fatal("human signature added no event")
			}
			// Mirror the line into this store's expectation bookkeeping.
			if _, err := s.AppendEvent(receipt.KindSignHuman, receipt.SignEvent{Spec: "mirror"}); err != nil {
				t.Fatal(err)
			}
		})
		if err := s.Verify(); err != nil {
			t.Fatalf("(i) Verify: %v", err)
		}
		t.Logf("kinds recorded: %v", want)

		// (ii) edit a middle line.
		path := filepath.Join(s.Dir, receipt.EventsFile)
		data, _ := os.ReadFile(path)
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		mid := len(lines) / 2
		lines[mid] = strings.Replace(lines[mid], `"at":"2`, `"at":"3`, 1)
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		err := s.Verify()
		var ce *receipt.ChainError
		if !errors.As(err, &ce) || ce.Line != mid+1 || !errors.Is(err, receipt.ErrIntegrity) {
			t.Fatalf("(ii) Verify after editing line %d = %v", mid+1, err)
		}
		t.Logf("(ii) observed: %v", err)
	})
	t.Run("deleted-sign-event", func(t *testing.T) {
		p := signtest.New(t)
		signHuman(t, p, false)
		s := open(t, p)
		if err := os.WriteFile(filepath.Join(s.Dir, receipt.EventsFile), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		res, err := kickoff.Check(kickoff.CheckInput{Root: p.Root, SpecID: signtest.SpecID, Card: signtest.Card,
			Config: kickoff.Config{Mode: "contract", Decider: "human"}, Policy: policy(p), Store: s})
		if err != nil || res.Reason != kickoff.ReasonSignatureNotRecorded {
			t.Fatalf("(iii) reason %q err %v, want signature-not-recorded", res.Reason, err)
		}
	})
	t.Run("deleted-revoke-event", func(t *testing.T) {
		t.Setenv(paths.EnvHome, t.TempDir())
		p := signtest.New(t)
		decideReady(t, p)
		signHuman(t, p, false)
		if _, err := revoke.Revoke(revoke.Options{Root: p.Root, SpecID: signtest.SpecID, Card: signtest.Card}, revoke.Seams{}); err != nil {
			t.Fatal(err)
		}
		s := open(t, p)
		path := filepath.Join(s.Dir, receipt.EventsFile)
		data, _ := os.ReadFile(path)
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if err := os.WriteFile(path, []byte(strings.Join(lines[:len(lines)-1], "\n")+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		res, err := kickoff.Decide(decideIn(p, s, "llm", jev.Result{}))
		if err != nil || res.Outcome != "human" || res.Reason != "precondition:e" {
			t.Fatalf("(iv) outcome %q reason %q err %v, want human precondition:e", res.Outcome, res.Reason, err)
		}
	})
	t.Run("fallback-receipt-line", func(t *testing.T) {
		p := signtest.New(t)
		decideReady(t, p)
		s := open(t, p)
		res, err := kickoff.Decide(decideIn(p, s, "llm+jev", jev.Result{Availability: jev.Unreachable}))
		if err != nil || res.Rule != "R2" {
			t.Fatalf("(vi) rule %q err %v", res.Rule, err)
		}
		lines, _ := s.Receipts()
		var rd receipt.ReceiptData
		if err := lines[len(lines)-1].Decode(&rd); err != nil {
			t.Fatal(err)
		}
		r, err := contract.DecodeKickoffReceipt([]byte(rd.Body))
		if err != nil {
			t.Fatal(err)
		}
		if r.RequestedDecider != "llm+jev" || r.EffectiveDecider != "llm" || r.Fallback == nil || !r.Fallback.Applied ||
			r.Fallback.Reason == nil || *r.Fallback.Reason != "jev_call_failed" {
			t.Errorf("(vi) receipt line %+v", r)
		}
		if !strings.HasPrefix(s.Dir, os.Getenv(paths.EnvHome)) || strings.HasPrefix(s.Dir, p.Root) {
			t.Errorf("store %s is not under MOAI_HOME outside the work tree", s.Dir)
		}
	})
}
