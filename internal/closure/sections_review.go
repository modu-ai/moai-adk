package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// ─── Second Verdict (REQ-CLOSURE-013 / REQ-CLOSURE-014) ───

func buildSecondVerdict(r *Report, in BuildInput) {
	sec := SecondVerdictSection{
		Backends:            []SecondReviewBackendView{},
		ContractSecondModel: orNotRecorded(secondModelOf(in)),
	}
	if in.SecondReviewPolicy == PolicyOff {
		sec.State = SecondReviewNotRequired
		r.SecondVerdict = sec
		return
	}
	if in.SecondReviewPolicy == PolicyAdvisory {
		sec.Warning = "second_review is advisory: a missing or failed second review does not stop a push"
	}
	st := SelectSecondReview(SecondReviewInput{
		Records:             in.SecondReviews,
		ContractCard:        in.Card,
		ContractDigest:      recordedDigestOf(in),
		ContractSecondModel: secondModelOf(in),
		EvalCommit:          in.EvalCommit,
		Facts:               in.Facts,
		WriteGlobs:          writeGlobsOf(in),
		SpecID:              in.SpecID,
	})
	sec.State = st.State
	sec.Cause = st.Cause
	sec.Verdict = st.Verdict
	sec.Backends = st.Backends
	sec.SubstituteBackend = st.SubstituteBackend
	sec.SupersedingCommit = st.SupersedingCommit
	if st.CurrencyUndetermined {
		sec.Warning = joinWarnings(sec.Warning,
			"the currency rule could not run; the review's currency is undetermined")
	}
	switch st.State {
	case SecondReviewNotPerformed:
		r.AddNotPerformed(NotPerformedSecondReviewNotPerformed, "cause: "+st.Cause)
	case SecondReviewStale:
		r.AddNotPerformed(NotPerformedSecondReviewStale, "superseded by "+st.SupersedingCommit)
	}
	// Skipped lines (unknown schema_version) never count as performed and are
	// listed.
	for _, s := range in.SecondReviewSkipped {
		r.AddNotPerformed(NotPerformedSecondReviewNotPerformed, "second-review line skipped: "+s)
	}
	r.SecondVerdict = sec
}

func secondModelOf(in BuildInput) string {
	if in.Verify.Contract != nil && in.Verify.Contract.Review != nil {
		return in.Verify.Contract.Review.SecondModel
	}
	return ""
}

func recordedDigestOf(in BuildInput) string {
	if in.Verify.Contract != nil && in.Verify.Contract.Signature != nil {
		return in.Verify.Contract.Signature.ContractSHA256
	}
	return ""
}

func writeGlobsOf(in BuildInput) []string {
	if in.Verify.Contract != nil && in.Verify.Contract.Ownership != nil {
		return in.Verify.Contract.Ownership.Write
	}
	return nil
}

func joinWarnings(a, b string) string {
	if a == "" {
		return b
	}
	return a + "; " + b
}

// ─── Plan-Audit Binding (REQ-CLOSURE-011) ───

func buildPlanAuditBinding(r *Report, in BuildInput) {
	sec := PlanAuditBindingSection{}
	expected := ""
	if in.Verify.Contract != nil && in.Verify.Contract.PlanAudit != nil {
		expected = in.Verify.Contract.PlanAudit.Verdict
	}
	sec.Expected = expected

	ev := EvidenceFor(in.Home, in.Card)
	receiptRef := receiptPlanAuditRef(in)
	path, wantHash := FindPlanAuditReport(ev, receiptRef, in.SpecID)
	if path == "" {
		sec.State = PlanAuditSelfReported
		r.AddNotPerformed(NotPerformedPlanAuditSelfReported, "no plan-audit report found in the card evidence home")
		r.PlanAuditBinding = sec
		return
	}
	sec.File = path
	data, ok, err := readFileOrEmpty(path)
	if err != nil || !ok {
		sec.State = PlanAuditSelfReported
		r.AddNotPerformed(NotPerformedPlanAuditSelfReported, "plan-audit report unreadable: "+path)
		r.PlanAuditBinding = sec
		return
	}
	// Receipt-named file: its bytes must match the receipt's recorded hash.
	if wantHash != "" {
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != wantHash {
			sec.State = PlanAuditMismatch
			r.AddNotPerformed(NotPerformedPlanAuditMismatch, "receipt-named file hash differs from the receipt: "+path)
			r.PlanAuditBinding = sec
			return
		}
	}
	line, ok := auditreceipt.ParseVerdictLine(string(data))
	if !ok || line.SpecID != in.SpecID {
		sec.State = PlanAuditSelfReported
		r.AddNotPerformed(NotPerformedPlanAuditSelfReported, "no audit verdict line naming this SPEC in "+filepath.Base(path))
		r.PlanAuditBinding = sec
		return
	}
	sec.Verdict = line.Verdict
	if expected != "" && line.Verdict != expected {
		sec.State = PlanAuditMismatch
		r.AddNotPerformed(NotPerformedPlanAuditMismatch, "verdict "+line.Verdict+" differs from the contract's "+expected)
		r.PlanAuditBinding = sec
		return
	}
	sec.State = PlanAuditBound
	r.PlanAuditBinding = sec
}

// receiptPlanAuditRef returns the receipt's plan_audit_report reference when
// the receipt is present.
func receiptPlanAuditRef(in BuildInput) *ReceiptFileRef {
	if !in.ReceiptPresent || len(in.Receipt) == 0 {
		return nil
	}
	var kr struct {
		Inputs *struct {
			PlanAuditReport *ReceiptFileRef `json:"plan_audit_report"`
		} `json:"inputs"`
	}
	if err := json.Unmarshal(in.Receipt, &kr); err != nil || kr.Inputs == nil {
		return nil
	}
	return kr.Inputs.PlanAuditReport
}

// ─── Human Verdict (REQ-CLOSURE-022) ───

func buildHumanVerdict(r *Report, in BuildInput) {
	sec := HumanVerdictSection{State: HumanVerdictNoneRecorded}
	latest := LatestVerdict(in.Verdicts)
	if latest == nil {
		r.AddNotPerformed(NotPerformedHumanVerdictNone, "no verdict recorded")
		r.HumanVerdict = sec
		return
	}
	sec.Verdict = latest.Verdict
	if latest.Operator.Name != "" || latest.Operator.Email != "" {
		sec.Operator = strings.TrimSpace(latest.Operator.Name + " <" + latest.Operator.Email + ">")
	}
	sec.RecordedAt = latest.RecordedAt
	sec.ReportSHA256 = latest.ReportSHA256
	if latest.ReportSHA256 != "" && latest.ReportSHA256 == in.PreviousReportHash {
		sec.State = HumanVerdictCurrent
	} else {
		sec.State = HumanVerdictStale
		r.AddNotPerformed(NotPerformedHumanVerdictStale, "the recorded verdict names a different report hash")
	}
	r.HumanVerdict = sec
}

// ─── Residual Risk ───

// buildResidualRisk states the residual risks the report itself derives from
// observed states (spec.md §H's machine-observable subset).
func buildResidualRisk(r *Report, in BuildInput) {
	if r.NewAPIs.Observed {
		r.AddResidualRisk("new-api-comparison-heuristic",
			"the class-4 comparison is a heuristic; an empty list is evidence of no detected addition, not of no addition")
	}
	if r.SecondVerdict.SubstituteBackend {
		r.AddResidualRisk("substitute-second-model",
			"the performing backend differs from the contract's review.second_model")
	}
	if r.SecondVerdict.State == SecondReviewStale {
		r.AddResidualRisk("stale-second-review",
			"the remedy is a fresh second review of this card, then a fresh closure report")
	}
}

// PreviousReportHashOf hashes the closure report on disk in the canonical
// form (generated_at masked — CanonicalReportHash), so a verdict recorded
// against the rendered report stays current across byte-identical rebuilds.
func PreviousReportHashOf(path string) string {
	data, ok, err := readFileOrEmpty(path)
	if err != nil || !ok {
		return ""
	}
	var r Report
	if err := json.Unmarshal(data, &r); err != nil {
		return ""
	}
	return CanonicalReportHash(&r)
}
