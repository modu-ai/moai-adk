package closure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/contract"
	"github.com/modu-ai/moai-adk/internal/contract/sign"
	"github.com/modu-ai/moai-adk/internal/contract/sign/signtest"
)

// receiptKickoffBuild assembles a minimal BuildInput over a signtest project
// with the given verify report and receipt bytes (the Kickoff section reads
// no other files).
func receiptKickoffBuild(t *testing.T, root string, rep contract.Report, receipt []byte) *Report {
	t.Helper()
	in := BuildInput{
		Card: "c1", SpecID: signtest.SpecID, Home: root,
		Mode: "contract", SecondReviewPolicy: "required",
		Verify:         rep,
		Receipt:        receipt,
		ReceiptPresent: receipt != nil,
	}
	return mustBuild(t, in)
}

// signWithReceipt signs the signtest SPEC through the receipt path and
// returns the receipt file's bytes.
func signWithReceipt(t *testing.T, p *signtest.Project) ([]byte, contract.Report) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(p.Root, filepath.Dir(signtest.PlanAuditReportPath)), 0o755); err != nil {
		t.Fatalf("mkdir plan-audit dir: %v", err)
	}
	p.WriteFile(signtest.PlanAuditReportPath,
		"# plan audit\n\nAUDIT-VERDICT: PASS spec="+signtest.SpecID+" receipts=none\n")
	o := p.ReceiptOptions("llm", "llm")
	receipt := p.Receipt(o, nil)
	p.WriteFile(signtest.ReceiptRel(), string(receipt))
	seams, _ := signtest.Seams(true, map[string]string{})
	res, err := sign.Sign(o, seams)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if res.Refusal != "" {
		t.Fatalf("sign refused: %s", res.Refusal)
	}
	return receipt, p.Verify(signtest.SpecID, o)
}

// TestAC_CLOSURE_010 — kickoff display for the five receipt fixtures.
func TestAC_CLOSURE_010(t *testing.T) {
	t.Run("a: llm signed-valid", func(t *testing.T) {
		p := signtest.New(t)
		receipt, rep := signWithReceipt(t, p)
		if rep.State != contract.StateSignedValid {
			t.Fatalf("fixture verify = %q %v, want signed-valid", rep.State, rep.Reasons)
		}
		r := receiptKickoffBuild(t, p.Root, rep, receipt)
		k := r.Kickoff
		if k.Method != "receipt" || k.SignerKind != "llm" {
			t.Fatalf("method/signer = %q/%q", k.Method, k.SignerKind)
		}
		if k.RequestedDecider != "llm" || k.EffectiveDecider != "llm" || k.FallbackApplied {
			t.Fatalf("deciders/fallback = %q/%q/%t", k.RequestedDecider, k.EffectiveDecider, k.FallbackApplied)
		}
		if k.LLMAnswer != "approve" || k.LLMConfidence == nil || *k.LLMConfidence != 0.82 {
			t.Fatalf("llm answer = %q %v", k.LLMAnswer, k.LLMConfidence)
		}
		if k.Outcome != "approve" {
			t.Fatalf("outcome = %q", k.Outcome)
		}
		if k.VerifyState != contract.StateSignedValid {
			t.Fatalf("kickoff verify state = %q", k.VerifyState)
		}
	})

	overwrite := func(t *testing.T, p *signtest.Project, mutate func(*contract.KickoffReceipt)) []byte {
		t.Helper()
		o := p.ReceiptOptions("llm", "llm")
		return p.Receipt(o, mutate)
	}

	t.Run("b: llm+jev both answers human", func(t *testing.T) {
		p := signtest.New(t)
		_, _ = signWithReceipt(t, p)
		receipt := overwrite(t, p, func(kr *contract.KickoffReceipt) {
			kr.RequestedDecider = "llm+jev"
			kr.EffectiveDecider = "llm+jev"
			kr.JevAnswer = signtest.JevAnswer("escalate", 0.41)
			kr.Outcome = "human"
		})
		p.WriteFile(signtest.ReceiptRel(), string(receipt))
		rep := p.Verify(signtest.SpecID, p.ReceiptOptions("llm", "llm"))
		r := receiptKickoffBuild(t, p.Root, rep, receipt)
		k := r.Kickoff
		if k.EffectiveDecider != "llm+jev" || k.JevAnswer != "escalate" ||
			k.JevConfidence == nil || *k.JevConfidence != 0.41 || k.Outcome != "human" {
			t.Fatalf("kickoff = %+v", k)
		}
		if k.VerifyState != contract.StateSignedInvalid {
			t.Fatalf("verify state = %q, want signed-invalid (receipt overwritten after signing)", k.VerifyState)
		}
	})

	t.Run("c: fallback jev_low_confidence", func(t *testing.T) {
		p := signtest.New(t)
		_, _ = signWithReceipt(t, p)
		reason := "jev_low_confidence"
		receipt := overwrite(t, p, func(kr *contract.KickoffReceipt) {
			kr.RequestedDecider = "llm+jev"
			kr.EffectiveDecider = "llm"
			kr.Fallback = &contract.ReceiptFallback{Applied: true, Reason: &reason}
		})
		p.WriteFile(signtest.ReceiptRel(), string(receipt))
		rep := p.Verify(signtest.SpecID, p.ReceiptOptions("llm", "llm"))
		r := receiptKickoffBuild(t, p.Root, rep, receipt)
		k := r.Kickoff
		if k.RequestedDecider != "llm+jev" || k.EffectiveDecider != "llm" ||
			!k.FallbackApplied || k.FallbackReason != "jev_low_confidence" {
			t.Fatalf("kickoff = %+v", k)
		}
		if k.JevAnswer != "not recorded" {
			t.Fatalf("jev answer = %q, want not recorded", k.JevAnswer)
		}
	})

	t.Run("d: unknown field listed", func(t *testing.T) {
		p := signtest.New(t)
		receipt, rep := signWithReceipt(t, p)
		// Splice an unknown top-level field into the receipt bytes.
		spliced := strings.Replace(string(receipt), "{",
			`{"extra_answer":{"answer":"approve"},`, 1)
		r := receiptKickoffBuild(t, p.Root, rep, []byte(spliced))
		k := r.Kickoff
		if k.Outcome != "approve" || k.LLMAnswer != "approve" {
			t.Fatalf("known fields lost: %+v", k)
		}
		found := false
		for _, np := range r.NotPerformed {
			if np.Item == NotPerformedReceiptFieldUnrecognized && strings.Contains(np.Detail, "extra_answer") {
				found = true
			}
		}
		if !found {
			t.Fatalf("not_performed missing receipt-field-unrecognized extra_answer: %v", r.NotPerformed)
		}
	})

	t.Run("e: receipt file absent states missing", func(t *testing.T) {
		p := signtest.New(t)
		_, rep := signWithReceipt(t, p)
		if err := os.Remove(filepath.Join(p.Root, signtest.ReceiptRel())); err != nil {
			t.Fatalf("remove receipt: %v", err)
		}
		r := receiptKickoffBuild(t, p.Root, rep, nil)
		if r.Kickoff.Method != "receipt" || r.Kickoff.ReceiptPresent {
			t.Fatalf("kickoff = %+v, want method receipt with absent receipt", r.Kickoff)
		}
		found := false
		for _, np := range r.NotPerformed {
			if np.Item == NotPerformedReceiptMissing {
				found = true
			}
		}
		if !found {
			t.Fatalf("not_performed missing receipt-missing")
		}
	})

	t.Run("human path", func(t *testing.T) {
		f := newACFixture(t)
		in := f.input()
		r := mustBuild(t, in)
		k := r.Kickoff
		if k.Method != "interactive-tty" || k.SignerKind != "human" || k.ReceiptPresent {
			t.Fatalf("kickoff = %+v, want interactive-tty human without receipt", k)
		}
	})
}

// TestAC_CLOSURE_011 — plan-audit binding across both report streams.
func TestAC_CLOSURE_011(t *testing.T) {
	passLine := "AUDIT-VERDICT: PASS spec=" + fixSpecID + " receipts=none"
	failLine := "AUDIT-VERDICT: FAIL spec=" + fixSpecID + " receipts=none"

	t.Run("card stream wins and binds", func(t *testing.T) {
		f := newACFixture(t)
		f.writeCardPlanAudit("plan-audit-iter2.md", passLine)
		f.writeGlobalPlanAudit(fixSpecID+"-review-1.md", failLine)
		in := f.input()
		r := mustBuild(t, in)
		pa := r.PlanAuditBinding
		if pa.State != PlanAuditBound || !strings.HasSuffix(pa.File, "plan-audit-iter2.md") || pa.Verdict != "PASS" {
			t.Fatalf("binding = %+v, want bound naming plan-audit-iter2.md", pa)
		}
	})

	t.Run("global stream review-2 bound and mismatch", func(t *testing.T) {
		f := newACFixture(t)
		f.writeGlobalPlanAudit(fixSpecID+"-review-2.md", passLine)
		in := f.input()
		r := mustBuild(t, in)
		if r.PlanAuditBinding.State != PlanAuditBound ||
			!strings.HasSuffix(r.PlanAuditBinding.File, fixSpecID+"-review-2.md") {
			t.Fatalf("binding = %+v, want bound naming review-2", r.PlanAuditBinding)
		}

		f2 := newACFixture(t)
		f2.writeGlobalPlanAudit(fixSpecID+"-review-2.md", failLine)
		in2 := f2.input()
		r2 := mustBuild(f2.t, in2)
		if r2.PlanAuditBinding.State != PlanAuditMismatch {
			f2.t.Fatalf("binding = %+v, want mismatch", r2.PlanAuditBinding)
		}
	})

	t.Run("no report is self-reported", func(t *testing.T) {
		f := newACFixture(t)
		in := f.input()
		r := mustBuild(t, in)
		if r.PlanAuditBinding.State != PlanAuditSelfReported {
			t.Fatalf("binding = %+v, want self-reported", r.PlanAuditBinding)
		}
	})

	t.Run("receipt-named file hash mismatch", func(t *testing.T) {
		f := newACFixture(t)
		f.write(filepath.Join(f.CardDir, ".moai", "reports", "plan-audit", "other.md"), "changed later\n")
		receiptJSON := `{"receipt_version":1,"spec_id":"` + fixSpecID + `",` +
			`"inputs":{"plan_audit_report":{"path":".moai/reports/plan-audit/other.md","sha256":"deadbeef"}},` +
			`"llm_answer":{"answer":"approve"},"outcome":"approve"}`
		in := f.input()
		in.Receipt = []byte(receiptJSON)
		in.ReceiptPresent = true
		r := mustBuild(t, in)
		if r.PlanAuditBinding.State != PlanAuditMismatch {
			t.Fatalf("binding = %+v, want mismatch (hash differs)", r.PlanAuditBinding)
		}
	})
}

// TestAC_CLOSURE_014 — second verdict rendering.
func TestAC_CLOSURE_014(t *testing.T) {
	r := NewReport(fixCard, fixSpecID, "head", "contract", "required")

	r.SecondVerdict = SecondVerdictSection{State: SecondReviewNotPerformed, Cause: SecondReviewCauseNoRecord}
	if md := RenderMarkdown(r); !strings.Contains(md, "NOT PERFORMED") || !strings.Contains(md, "no-record") {
		t.Fatalf("not-performed render missing literals:\n%s", md)
	}

	r.SecondVerdict = SecondVerdictSection{State: SecondReviewStale, SupersedingCommit: "cafe567"}
	if md := RenderMarkdown(r); !strings.Contains(md, "STALE") || !strings.Contains(md, "cafe567") {
		t.Fatalf("stale render missing literals:\n%s", md)
	}

	r.SecondVerdict = SecondVerdictSection{State: SecondReviewPerformed, Verdict: "fail"}
	if md := RenderMarkdown(r); !strings.Contains(md, "FAILED") {
		t.Fatalf("failed render missing literal:\n%s", md)
	}

	r.SecondVerdict = SecondVerdictSection{
		State: SecondReviewPerformed, Verdict: "pass",
		Backends:            []SecondReviewBackendView{{Backend: "glm", Verdict: "pass"}},
		ContractSecondModel: "codex",
		SubstituteBackend:   true,
	}
	md := RenderMarkdown(r)
	if !strings.Contains(md, "glm") || !strings.Contains(md, "codex") || !strings.Contains(md, "SUBSTITUTE BACKEND") {
		t.Fatalf("substitute render missing literals:\n%s", md)
	}

	r.SecondVerdict = SecondVerdictSection{State: SecondReviewNotRequired}
	if md := RenderMarkdown(r); !strings.Contains(md, "not required") {
		t.Fatalf("off render missing literal:\n%s", md)
	}
}

// TestAC_CLOSURE_022 — human verdict section.
func TestAC_CLOSURE_022(t *testing.T) {
	f := newACFixture(t)

	// Build and write the report once so a hash exists to record against.
	in := f.input()
	first := mustBuild(t, in)
	if err := os.MkdirAll(f.EvidenceDir, 0o755); err != nil {
		t.Fatalf("mkdir evidence dir: %v", err)
	}
	data, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	reportPath := filepath.Join(f.EvidenceDir, ReportJSONFile)
	if err := os.WriteFile(reportPath, data, 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	recorded := CanonicalReportHash(first)

	operator := VerdictOperator{Name: "GOOS", Email: "goos@example.com"}
	in = f.input()
	in.Verdicts = []VerdictRecord{{
		SchemaVersion: 1, Card: fixCard, SpecID: fixSpecID, Verdict: "accept",
		Operator: operator, RecordedAt: "2026-09-27T09:00:00Z",
		ReportSHA256: recorded, Method: "interactive-tty",
	}}
	r := mustBuild(t, in)
	if r.HumanVerdict.State != HumanVerdictCurrent || r.HumanVerdict.Verdict != "accept" ||
		!strings.Contains(r.HumanVerdict.Operator, "GOOS") {
		t.Fatalf("human verdict = %+v, want current accept by GOOS", r.HumanVerdict)
	}

	// A different recorded hash renders stale and lists the entry.
	in.Verdicts[0].ReportSHA256 = strings.Repeat("a", 64)
	r = mustBuild(t, in)
	if r.HumanVerdict.State != HumanVerdictStale {
		t.Fatalf("state = %q, want stale", r.HumanVerdict.State)
	}
	found := false
	for _, np := range r.NotPerformed {
		if np.Item == NotPerformedHumanVerdictStale {
			found = true
		}
	}
	if !found {
		t.Fatalf("not_performed missing human-verdict-stale")
	}

	// Two lines: the later one is shown.
	in.Verdicts = append(in.Verdicts, VerdictRecord{
		SchemaVersion: 1, Card: fixCard, SpecID: fixSpecID, Verdict: "reject",
		Operator: operator, RecordedAt: "2026-09-27T10:00:00Z",
		ReportSHA256: recorded, Method: "interactive-tty",
	})
	r = mustBuild(t, in)
	if r.HumanVerdict.Verdict != "reject" {
		t.Fatalf("verdict = %q, want the later line's reject", r.HumanVerdict.Verdict)
	}
}
