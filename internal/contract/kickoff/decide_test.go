package kickoff_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/kickoff"
	"github.com/modu-ai/moai-adk/internal/contract/receipt"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
	"github.com/modu-ai/moai-adk/internal/escalation"
	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/runtime"
)

const specRel = ".moai/specs/" + signtest.SpecID

// dfx is a decide fixture.
type dfx struct {
	*fx
	jevCalls int
	jevRes   jev.Result
}

type auditReport struct {
	name    string // report file name under .moai/reports/<card>/
	verdict string
	score   string
	hash    string // "" = current plan-artifact hash; "-" = omit the line
}

// newDecide builds a project whose SPEC directory was touched by a commit
// carrying `Authored-By-Agent: <author>` followed by a blank line and a
// signature line (so git's trailer parser sees nothing), and whose card
// evidence path holds the given plan-audit reports.
func newDecide(t *testing.T, author string, reports ...auditReport) *dfx {
	t.Helper()
	f := &dfx{fx: newFx(t)}
	f.p.WriteFile(specRel+"/research.md", "# research\n\nMeasured.\n")
	f.p.Git("add", "-A")
	msg := "docs: plan artifacts\n\ncard: " + signtest.Card + "\n"
	if author != "" {
		msg += "Authored-By-Agent: " + author + "\n"
	}
	msg += "\n🗿 MoAI\n"
	f.p.Git("commit", "-q", "-m", msg)
	if len(reports) == 0 {
		reports = []auditReport{{name: "plan-audit-1.md", verdict: "PASS", score: "0.90"}}
	}
	for _, r := range reports {
		f.writeReport(r)
	}
	f.jevRes = jevAnswer("approve", 0.8)
	return f
}

func (f *dfx) hash() string {
	h, err := runtime.NewInMemoryCache().ComputeHash(f.p.Path(specRel))
	if err != nil {
		f.t.Fatal(err)
	}
	return h
}

func (f *dfx) writeReport(r auditReport) {
	var b strings.Builder
	fmt.Fprintf(&b, "# plan-audit\n\nVerdict: %s\nOverall Score: %s\n", r.verdict, r.score)
	switch r.hash {
	case "":
		fmt.Fprintf(&b, "plan_artifact_hash: %s\n", f.hash())
	case "-":
	default:
		fmt.Fprintf(&b, "plan_artifact_hash: %s\n", r.hash)
	}
	f.p.WriteFile(".moai/reports/"+signtest.Card+"/"+r.name, b.String())
}

func jevAnswer(choice string, p float64) jev.Result {
	return jev.Result{Availability: jev.Available, Answers: []jev.Answer{{
		QuestionID: "start", Kind: jev.KindChoice, Choice: choice, Probability: p,
	}}}
}

func judgement(agent, answer string) []byte {
	b, _ := json.Marshal(map[string]any{
		"agent": agent, "model": "claude", "answer": answer, "confidence": 0.8,
		"reason": "The contract matches the plan.", "reason_refs": []string{"contract.yaml:2"},
	})
	return b
}

func (f *dfx) input(decider, answer string, doctrine bool) kickoff.DecideInput {
	return kickoff.DecideInput{
		Root: f.p.Root, SpecID: signtest.SpecID, Card: signtest.Card,
		Judgement:       judgement("lead-session", answer),
		Config:          kickoff.Config{Mode: "contract", Decider: decider, JevEnabled: true, JevMinConfidence: 0.5},
		Policy:          f.policy(),
		DoctrineAmended: doctrine,
		Store:           f.store,
		Now:             func() time.Time { return time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC) },
		Jev: func(context.Context, jev.Request) jev.Result {
			f.jevCalls++
			return f.jevRes
		},
	}
}

func (f *dfx) counts() (receipts, events int) {
	r, err := f.store.Receipts()
	if err != nil {
		f.t.Fatal(err)
	}
	e, err := f.store.Events()
	if err != nil {
		f.t.Fatal(err)
	}
	return len(r), len(e)
}

func (f *dfx) receiptFile() ([]byte, bool) {
	data, err := os.ReadFile(f.p.Path(signtest.ReceiptRel()))
	return data, err == nil
}

func (f *dfx) lastEvent() receipt.DecideEvent {
	f.t.Helper()
	ev, err := f.store.Events()
	if err != nil || len(ev) == 0 {
		f.t.Fatalf("no events (%v)", err)
	}
	var d receipt.DecideEvent
	if err := ev[len(ev)-1].Decode(&d); err != nil {
		f.t.Fatal(err)
	}
	return d
}

// mustDecide runs Decide and fails on an error.
func (f *dfx) mustDecide(in kickoff.DecideInput) kickoff.DecideResult {
	f.t.Helper()
	res, err := kickoff.Decide(in)
	if err != nil {
		f.t.Fatalf("Decide: %v", err)
	}
	return res
}

// assertNoDecision checks a non-deciding path: outcome human with the reason,
// no Jev call, no receipt line, no receipt file, exactly one new event.
func (f *dfx) assertNoDecision(res kickoff.DecideResult, reason string, r0, e0 int) {
	f.t.Helper()
	if res.Outcome != "human" || res.Reason != reason {
		f.t.Errorf("outcome %q reason %q, want human %q", res.Outcome, res.Reason, reason)
	}
	if f.jevCalls != 0 {
		f.t.Errorf("Jev called %d times", f.jevCalls)
	}
	r1, e1 := f.counts()
	if r1 != r0 || e1 != e0+1 {
		f.t.Errorf("receipts %d→%d events %d→%d, want receipts unchanged and one event", r0, r1, e0, e1)
	}
	if _, ok := f.receiptFile(); ok {
		f.t.Error("kickoff-receipt.json written on a non-deciding path")
	}
}

// assertReceipt checks a deciding path wrote a receipt line, a decide event,
// and a kickoff-receipt.json byte-identical to the line's body.
func (f *dfx) assertReceipt(r0, e0 int) *contract.KickoffReceipt {
	f.t.Helper()
	r1, e1 := f.counts()
	if r1 != r0+1 || e1 != e0+1 {
		f.t.Fatalf("receipts %d→%d events %d→%d, want +1/+1", r0, r1, e0, e1)
	}
	data, ok := f.receiptFile()
	if !ok {
		f.t.Fatal("kickoff-receipt.json not written")
	}
	lines, _ := f.store.Receipts()
	var rd receipt.ReceiptData
	if err := lines[len(lines)-1].Decode(&rd); err != nil {
		f.t.Fatal(err)
	}
	if rd.Body != string(data) || rd.BodySHA256 != receipt.SHA256Hex(data) {
		f.t.Error("kickoff-receipt.json differs from the issued receipt body")
	}
	r, err := contract.DecodeKickoffReceipt(data)
	if err != nil {
		f.t.Fatalf("receipt does not strictly decode: %v", err)
	}
	return r
}

// TestDecidePreconditions covers preconditions (a)-(f) (AC-GR-018).
func TestDecidePreconditions(t *testing.T) {
	type pc struct {
		name   string
		setup  func(t *testing.T) *dfx
		reason string // "" = all preconditions hold
	}
	withContract := func(mut func(string) string) func(*dfx) {
		return func(f *dfx) {
			rel := signtest.SpecRel(signtest.SpecID, contract.ContractFile)
			f.p.WriteFile(rel, mut(string(f.p.ReadFile(rel))))
		}
	}
	build := func(author string, reports []auditReport, mods ...func(*dfx)) func(t *testing.T) *dfx {
		return func(t *testing.T) *dfx {
			f := newDecide(t, author, reports...)
			for _, m := range mods {
				m(f)
			}
			return f
		}
	}
	pass := []auditReport{{name: "plan-audit-1.md", verdict: "PASS", score: "0.90"}}
	cases := []pc{
		{"a_verdict_fail", build("manager-spec", []auditReport{{name: "plan-audit-1.md", verdict: "FAIL", score: "0.90"}}), "precondition:a"},
		{"a_score_annotated", build("manager-spec", []auditReport{{name: "plan-audit-1.md", verdict: "PASS", score: "0.90 (Tier L threshold 0.85)"}}), "precondition:a"},
		{"a_hash_missing", build("manager-spec", []auditReport{{name: "plan-audit-1.md", verdict: "PASS", score: "0.90", hash: "-"}}), "precondition:a"},
		{"a_hash_stale", build("manager-spec", pass, func(f *dfx) {
			f.p.WriteFile(specRel+"/spec.md", string(f.p.ReadFile(specRel+"/spec.md"))+"\nEdited after the audit.\n")
		}), "precondition:a"},
		{"a_debt_below_threshold", build("manager-spec", []auditReport{{name: "plan-audit-1.md", verdict: "PASS-WITH-DEBT", score: "0.84"}}), "precondition:a"},
		{"a_debt_above_threshold", build("manager-spec", []auditReport{{name: "plan-audit-1.md", verdict: "PASS-WITH-DEBT", score: "0.86"}}), ""},
		{"b_clarification_marker", build("manager-spec", pass, func(f *dfx) {
			f.p.WriteFile(specRel+"/plan.md", "# plan\n\n[NEEDS CLARIFICATION: database]\n")
			f.writeReport(pass[0])
		}), "precondition:b"},
		{"c_verify_reason", build("manager-spec", pass, withContract(func(s string) string {
			return strings.Replace(s, "  - commit\n", "  - commit\n  - frobnicate\n", 1)
		})), "precondition:c"},
		{"d_forbidden_action", build("manager-spec", pass, withContract(func(s string) string {
			return strings.Replace(s, "  - commit\n", "  - commit\n  - push-main\n", 1)
		})), "precondition:d"},
		{"e_open_contract_record", build("manager-spec", pass, func(f *dfx) {
			writeRecord(f.t, f.p.Root, escalation.Record{SchemaVersion: 1, Card: signtest.Card, Spec: signtest.SpecID,
				Kind: escalation.KindContract, Class: escalation.ClassAcceptanceChange,
				Fingerprint: escalation.Fingerprint(escalation.ClassAcceptanceChange, "x"),
				Status:      escalation.StatusOpen, Occurrences: 1, Observation: "o", Options: []string{"a", "b"}})
		}), "precondition:e"},
		{"e_revoke_record_only", build("manager-spec", pass, func(f *dfx) {
			f.signHuman(true)
			f.writeReport(pass[0])
			writeRecord(f.t, f.p.Root, escalation.Record{SchemaVersion: 1, Card: signtest.Card, Spec: signtest.SpecID,
				Kind: escalation.KindRevoke, Class: "revoke-operator",
				Fingerprint: escalation.Fingerprint("revoke-operator", f.seal()),
				Status:      escalation.StatusResolved, Decider: "human", Occurrences: 1, Observation: "o", Options: []string{"a", "b"}})
			if nd, _ := escalation.NeedsDecision(f.p.Root, signtest.Card); nd {
				f.t.Fatal("premise: A2 NeedsDecision must be false for a revoke-only card")
			}
		}), "precondition:e"},
		{"f_frozen_ownership", build("manager-spec", pass, withContract(func(s string) string {
			return strings.Replace(s, "    - \"internal/fixture/**\"\n", "    - \"internal/fixture/**\"\n    - \"**/CLAUDE.md\"\n", 1)
		})), "precondition:f"},
		{"all_hold", build("manager-spec", []auditReport{
			{name: "plan-audit-1.md", verdict: "FAIL", score: "0.40"},
			{name: "plan-audit-iter2.md", verdict: "PASS", score: "0.90"},
		}), ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := c.setup(t)
			r0, e0 := f.counts()
			res := f.mustDecide(f.input("llm+jev", "approve", false))
			if c.reason != "" {
				f.assertNoDecision(res, c.reason, r0, e0)
				return
			}
			if f.jevCalls != 1 {
				t.Errorf("Jev called %d times, want 1", f.jevCalls)
			}
			r := f.assertReceipt(r0, e0)
			if c.name == "all_hold" {
				want := ".moai/reports/" + signtest.Card + "/plan-audit-iter2.md"
				if r.Inputs == nil || r.Inputs.PlanAuditReport == nil || r.Inputs.PlanAuditReport.Path != want {
					t.Errorf("plan_audit_report %+v, want %s", r.Inputs, want)
				}
			}
		})
	}
}

func writeRecord(t *testing.T, root string, r escalation.Record) {
	t.Helper()
	data, err := r.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	path := escalation.RecordPath(root, r.Card, r.Class, r.Fingerprint, 0)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// decideRule runs one rule case and checks the outcome, the reason, and the
// recorded rule.
func decideRule(t *testing.T, decider, llm string, jevRes *jev.Result, doctrine bool, outcome, reason, rule string) (*dfx, kickoff.DecideResult) {
	t.Helper()
	f := newDecide(t, "manager-spec")
	if jevRes != nil {
		f.jevRes = *jevRes
	}
	r0, e0 := f.counts()
	res := f.mustDecide(f.input(decider, llm, doctrine))
	if res.Outcome != outcome || res.Reason != reason || res.Rule != rule {
		t.Fatalf("outcome %q reason %q rule %q, want %q %q %q", res.Outcome, res.Reason, res.Rule, outcome, reason, rule)
	}
	r := f.assertReceipt(r0, e0)
	if r.Outcome != outcome || r.RequestedDecider != decider {
		t.Errorf("receipt outcome %q requested %q", r.Outcome, r.RequestedDecider)
	}
	ev := f.lastEvent()
	if ev.Rule != rule || ev.DeciderAgent != "lead-session" || len(ev.AuthorTrailers) == 0 {
		t.Errorf("decide event %+v", ev)
	}
	return f, res
}

func jr(choice string) *jev.Result { r := jevAnswer(choice, 0.8); return &r }

// TestDecideRuleCrossCheckAgree — R1, both approve (AC-GR-019).
func TestDecideRuleCrossCheckAgree(t *testing.T) {
	f, _ := decideRule(t, "llm+jev", "approve", jr("approve"), true, "approve", "", "R1")
	if f.jevCalls != 1 {
		t.Errorf("Jev calls %d", f.jevCalls)
	}
}

// TestDecideRuleCrossCheckBothReject — R1, both reject (AC-GR-019).
func TestDecideRuleCrossCheckBothReject(t *testing.T) {
	decideRule(t, "llm+jev", "reject", jr("reject"), true, "reject", "", "R1")
}

// TestDecideRuleCrossCheckDisagree — R1 disagreement or escalate (AC-GR-019).
func TestDecideRuleCrossCheckDisagree(t *testing.T) {
	for _, c := range [][2]string{
		{"approve", "reject"}, {"approve", "escalate"}, {"reject", "approve"},
		{"escalate", "approve"}, {"reject", "escalate"},
	} {
		t.Run(c[0]+"+"+c[1], func(t *testing.T) {
			decideRule(t, "llm+jev", c[0], jr(c[1]), true, "human", "cross-check-disagree", "R1")
		})
	}
}

// TestDecideRuleJevBeforeAmendment — R3 (AC-GR-019).
func TestDecideRuleJevBeforeAmendment(t *testing.T) {
	decideRule(t, "llm+jev", "approve", jr("approve"), false, "human", "jev-doctrine-not-amended", "R3")
}

// TestDecideRuleLLMAlone — R4 (AC-GR-019).
func TestDecideRuleLLMAlone(t *testing.T) {
	for _, c := range [][2]string{{"approve", "approve"}, {"reject", "reject"}, {"escalate", "human"}} {
		t.Run(c[0], func(t *testing.T) {
			f, _ := decideRule(t, "llm", c[0], nil, true, c[1], "", "R4")
			if f.jevCalls != 0 {
				t.Errorf("Jev called %d times under decider llm", f.jevCalls)
			}
		})
	}
}

// TestDecideRuleJevAloneRefused — R5 (AC-GR-019).
func TestDecideRuleJevAloneRefused(t *testing.T) {
	f := newDecide(t, "manager-spec")
	r0, e0 := f.counts()
	in := f.input("jev", "approve", true)
	in.Config.DeciderJevSole = true
	_, err := kickoff.Decide(in)
	if !errors.Is(err, kickoff.ErrDeciderJevRefused) || !errors.Is(err, kickoff.ErrUsage) {
		t.Fatalf("err = %v, want decider-jev-refused (usage)", err)
	}
	if r1, e1 := f.counts(); r1 != r0 || e1 != e0 || f.jevCalls != 0 {
		t.Errorf("R5 wrote or called Jev: receipts %d→%d events %d→%d jev %d", r0, r1, e0, e1, f.jevCalls)
	}
}

// TestDecideAuthorExclusion — the SPEC author cannot decide (AC-GR-019).
func TestDecideAuthorExclusion(t *testing.T) {
	cases := []struct {
		name, author, agent, reason string
	}{
		{"declared_manager_spec", "manager-spec", "manager-spec", "author-decider-conflict"},
		{"trailer_author", "lead-session", "lead-session", "author-decider-conflict"},
		{"no_trailer", "", "lead-session", "author-check-unmeasured"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newDecide(t, c.author)
			if c.author != "" {
				out := f.p.Git("log", "-1", "--format=%(trailers:key=Authored-By-Agent,valueonly)")
				if strings.TrimSpace(out) != "" {
					t.Fatalf("premise: git's trailer parser must not see the trailer, got %q", out)
				}
			}
			r0, e0 := f.counts()
			in := f.input("llm", "approve", true)
			in.Judgement = judgement(c.agent, "approve")
			res := f.mustDecide(in)
			f.assertNoDecision(res, c.reason, r0, e0)
			ev := f.lastEvent()
			if ev.DeciderAgent != c.agent {
				t.Errorf("event decider agent %q", ev.DeciderAgent)
			}
		})
	}
}

// TestDecideJevFallback — R2, Jev-side failures fall back to llm alone
// (AC-GR-025).
func TestDecideJevFallback(t *testing.T) {
	causes := []struct {
		reason string
		res    jev.Result
		off    bool
	}{
		{"jev_call_failed", jev.Result{Availability: jev.Unreachable, Condition: "dial failed"}, false},
		{"jev_key_missing", jev.Result{Availability: jev.NoCredential}, false},
		{"jev_disabled", jev.Result{Availability: jev.Disabled}, true},
		{"jev_malformed_response", jev.Result{Availability: jev.Available, Answers: []jev.Answer{{QuestionID: "start", Kind: jev.KindChoice, Choice: "maybe", Probability: 0.9}}}, false},
		{"jev_low_confidence", jevAnswer("approve", 0.49), false},
	}
	for _, c := range causes {
		for _, llm := range []string{"approve", "reject"} {
			t.Run(c.reason+"/"+llm, func(t *testing.T) {
				f := newDecide(t, "manager-spec")
				f.jevRes = c.res
				r0, e0 := f.counts()
				in := f.input("llm+jev", llm, true)
				in.Config.JevEnabled = !c.off
				res := f.mustDecide(in)
				if res.Outcome != llm || res.Rule != "R2" || res.FallbackReason != c.reason {
					t.Fatalf("outcome %q rule %q fallback %q, want %q R2 %q", res.Outcome, res.Rule, res.FallbackReason, llm, c.reason)
				}
				r := f.assertReceipt(r0, e0)
				if r.RequestedDecider != "llm+jev" || r.EffectiveDecider != "llm" || r.Fallback == nil ||
					!r.Fallback.Applied || r.Fallback.Reason == nil || *r.Fallback.Reason != c.reason || r.JevAnswer != nil {
					t.Errorf("receipt %+v", r)
				}
				lines, _ := f.store.Receipts()
				var rd receipt.ReceiptData
				_ = lines[len(lines)-1].Decode(&rd)
				if c.reason == "jev_low_confidence" && len(rd.JevResponse) == 0 {
					t.Error("low-confidence raw response not kept in the receipt line")
				}
			})
		}
	}
}
