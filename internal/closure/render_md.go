package closure

import (
	"fmt"
	"sort"
	"strings"
)

// RenderMarkdown renders the report's Markdown form from the JSON model —
// the struct is its only input, so the two forms cannot disagree
// (design.md §A.3). Sections appear in the REQ-CLOSURE-002 order; every value
// is the JSON value's text.
func RenderMarkdown(r *Report) string {
	var b strings.Builder
	p := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }

	p("# Closure report — card %s (SPEC %s)", r.Card, r.SpecID)
	p("")
	p("head: `%s` · generated: %s · mode: %s · second_review: %s",
		valueOr(r.HeadSHA, "not observed"), valueOr(r.GeneratedAt, "not observed"),
		valueOr(r.Mode, "not observed"), valueOr(r.SecondReviewPolicy, "not observed"))

	p("")
	p("## Summary")
	p("")
	p("- verify state: %s", r.Summary.VerifyState)
	if r.Summary.Terminal != nil {
		p("- terminal: %t", *r.Summary.Terminal)
	} else {
		p("- terminal: not observed")
	}
	for _, reason := range r.Summary.VerifyReasons {
		p("- reason: %s", reason)
	}

	p("")
	p("## Kickoff")
	p("")
	k := r.Kickoff
	p("- method: %s · signer kind: %s", k.Method, k.SignerKind)
	if k.ReceiptPresent {
		p("- requested_decider: %s · effective_decider: %s", k.RequestedDecider, k.EffectiveDecider)
		p("- fallback.applied: %t", k.FallbackApplied)
		if k.FallbackReason != "" {
			p("- fallback.reason: %s", k.FallbackReason)
		}
		p("- llm_answer: %s", answerLine(k.LLMAnswer, k.LLMConfidence))
		p("- jev_answer: %s", answerLine(k.JevAnswer, k.JevConfidence))
		p("- outcome: %s", k.Outcome)
	} else if k.Method == "receipt" {
		p("- receipt: missing")
	}
	p("- verify state: %s", k.VerifyState)

	p("")
	p("## Contract Reconciliation")
	p("")
	rec := r.Reconciliation
	p("- recorded acceptance: %s (%s ACs) · measured: %s (%d ACs)",
		rec.RecordedAcceptanceSHA, rec.RecordedACCount, valueOr(rec.MeasuredAcceptanceSHA, "not observed"), rec.MeasuredACCount)
	p("- verify state: %s", rec.VerifyState)
	for _, reason := range rec.VerifyReasons {
		p("- reason: %s", reason)
	}
	if len(rec.Rows) == 0 {
		p("- (no acceptance-criterion rows)")
	}
	for _, row := range rec.Rows {
		p("- %s: %s", row.ID, row.Status)
		if row.Evidence != "" {
			p("  - evidence: %s", row.Evidence)
		}
	}

	p("")
	p("## Invariants")
	p("")
	if len(r.Invariants.Rows) == 0 {
		p("- (no invariants)")
	}
	for _, row := range r.Invariants.Rows {
		p("- %s (%s): %s", row.Invariant, row.Kind, row.Result)
	}

	p("")
	p("## Ownership")
	p("")
	for _, rec := range r.Ownership.Records {
		p("- %s/%s %s (%d) — decider: %s", rec.Kind, rec.Class, rec.Status, rec.Occurrences, valueOr(rec.Decider, "not recorded"))
	}
	p("- ownership-violation count: %s", r.Ownership.Count.Text())

	p("")
	p("## New APIs")
	p("")
	if !r.NewAPIs.Observed && len(r.NewAPIs.Additions) == 0 {
		p("- not observed")
	}
	for _, a := range r.NewAPIs.Additions {
		p("- %s %s (%s)", a.Kind, a.Name, a.Path)
	}
	if !r.NewAPIs.Observed && len(r.NewAPIs.Additions) > 0 {
		p("- (live comparison not observed; the additions above come from records)")
	}

	p("")
	p("## Escalations")
	p("")
	if len(r.Escalations.Records) == 0 && len(r.Escalations.Unreadable) == 0 {
		p("- (no escalation records)")
	}
	for _, rec := range r.Escalations.Records {
		p("- %s/%s %s (%d) — contract: %s · decider: %s",
			rec.Kind, rec.Class, rec.Status, rec.Occurrences,
			valueOr(rec.ContractRef, "not recorded"), valueOr(rec.Decider, "not recorded"))
	}
	for _, name := range r.Escalations.Unreadable {
		p("- unreadable: %s", name)
	}
	p("- needs-decision: %s", r.Escalations.NeedsDecision)

	p("")
	p("## First Verdict")
	p("")
	fv := r.FirstVerdict
	p("- run_status: %s", fv.RunStatus)
	p("- ac_pass_count: %s · ac_fail_count: %s", fv.ACPassCount, fv.ACFailCount)
	p("- run_commit_sha: %s", fv.RunCommitSHA)
	if fv.CountMismatch {
		p("- COUNT MISMATCH: reported pass+fail does not sum to the measured AC count")
	}

	p("")
	p("## Second Verdict")
	p("")
	sv := r.SecondVerdict
	if sv.State == SecondReviewNotRequired {
		p("- not required")
	} else {
		switch {
		case sv.State == SecondReviewNotPerformed:
			p("NOT PERFORMED — %s", valueOr(sv.Cause, "unknown cause"))
		case sv.State == SecondReviewStale:
			p("STALE — superseded by %s", valueOr(sv.SupersedingCommit, "unknown commit"))
		case sv.Verdict == "fail":
			p("FAILED")
		default:
			p("PERFORMED — %s", valueOr(sv.Verdict, ""))
		}
		for _, b := range sv.Backends {
			p("- %s: %s", b.Backend, b.Verdict)
		}
		p("- contract second_model: %s", sv.ContractSecondModel)
		if sv.SubstituteBackend {
			p("- SUBSTITUTE BACKEND: the performing backend differs from the contract's second_model")
		}
	}
	if sv.Warning != "" {
		p("- WARNING: %s", sv.Warning)
	}

	p("")
	p("## Plan-Audit Binding")
	p("")
	pa := r.PlanAuditBinding
	p("- state: %s", pa.State)
	if pa.File != "" {
		p("- file: %s", pa.File)
	}
	if pa.Verdict != "" {
		p("- verdict: %s (expected %s)", pa.Verdict, valueOr(pa.Expected, "not recorded"))
	}

	p("")
	p("## Not Performed")
	p("")
	if len(r.NotPerformed) == 0 {
		p("- (nothing)")
	}
	for _, np := range r.NotPerformed {
		p("- %s: %s", np.Item, np.Detail)
	}

	p("")
	p("## Residual Risk")
	p("")
	if len(r.ResidualRisk) == 0 {
		p("- (none stated)")
	}
	for _, rr := range r.ResidualRisk {
		p("- %s: %s", rr.Item, rr.Detail)
	}

	p("")
	p("## Human Verdict")
	p("")
	hv := r.HumanVerdict
	if hv.State == HumanVerdictNoneRecorded {
		p("- none recorded")
	} else {
		p("- %s (%s)", hv.Verdict, hv.State)
		p("- operator: %s · at: %s", valueOr(hv.Operator, "unknown"), valueOr(hv.RecordedAt, "unknown"))
	}

	p("")
	p("## Sources")
	p("")
	if len(r.Sources) == 0 {
		p("- (none)")
	}
	for _, role := range sortedStringKeys(r.Sources) {
		path := r.Sources[role]
		if path == "" {
			path = "not read"
		}
		p("- %s: %s", role, path)
	}
	return b.String()
}

func sortedStringKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func valueOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func answerLine(answer string, confidence *float64) string {
	if answer == "" || answer == NotRecorded {
		return "not recorded"
	}
	if confidence == nil {
		return answer
	}
	return fmt.Sprintf("%s (confidence %.2f)", answer, *confidence)
}
