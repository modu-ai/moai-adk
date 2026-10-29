package review

import "testing"

func reviewRec(t *testing.T, provider, role, session string, kind Kind, v Verdict) Record {
	t.Helper()
	r := completed(t, v, headA, session)
	r.Kind, r.Provider, r.Producer.Role = kind, provider, role
	return r
}

// design §11.1: with Codex not opted in, a same-HEAD review from a
// registered non-writer session (not GLM, not Codex) or an operator record
// fills the code / security review requirement.
func TestAdmitNeutral(t *testing.T) {
	writers := []string{"s-w1", "s-w0"} // current and earlier attempt
	cases := []struct {
		name  string
		rec   Record
		optIn bool
		want  Reason
	}{
		{"claude reviewer", reviewRec(t, ProviderClaude, "code-reviewer", "s-r", KindCodeReview, VerdictPass), false, ReasonNone},
		{"operator record", reviewRec(t, ProviderOperator, "operator", "s-op", KindSecurityReview, VerdictPass), false, ReasonNone},
		{"operator provider without operator role", reviewRec(t, ProviderOperator, "worker", "s-op", KindSecurityReview, VerdictPass), false, ReasonProducer},
		{"glm refused", reviewRec(t, ProviderGLM, "code-reviewer", "s-r", KindCodeReview, VerdictPass), false, ReasonProducer},
		{"codex refused when not opted in", reviewRec(t, ProviderCodex, "code-reviewer", "s-r", KindCodeReview, VerdictPass), false, ReasonProducer},
		{"codex allowed when opted in", reviewRec(t, ProviderCodex, "code-reviewer", "s-r", KindCodeReview, VerdictPass), true, ReasonNone},
		{"current writer", reviewRec(t, ProviderClaude, "code-reviewer", "s-w1", KindCodeReview, VerdictPass), false, ReasonProducer},
		{"earlier-attempt writer", reviewRec(t, ProviderClaude, "code-reviewer", "s-w0", KindCodeReview, VerdictPass), false, ReasonProducer},
		{"audit kind is not a review", reviewRec(t, ProviderClaude, "auditor", "s-r", KindAudit, VerdictPass), false, ReasonProducer},
		{"debt still refused", reviewRec(t, ProviderClaude, "code-reviewer", "s-r", KindCodeReview, VerdictPassWithDebt), false, ReasonDebt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rec.AdmitNeutral(headA, writers, tc.optIn); got != tc.want {
				t.Fatalf("AdmitNeutral = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAdmitNeutral_NoKnownWriters(t *testing.T) {
	rec := reviewRec(t, ProviderClaude, "code-reviewer", "s-r", KindCodeReview, VerdictPass)
	if got := rec.AdmitNeutral(headA, nil, false); got != ReasonNone {
		t.Fatalf("AdmitNeutral = %q, want admitted", got)
	}
	rec.Producer.Session = ""
	if got := rec.AdmitNeutral(headA, nil, false); got != ReasonProducer {
		t.Fatalf("anonymous producer = %q, want %q", got, ReasonProducer)
	}
}

// One session holds one role: code review and security review differ.
func TestSeparateReviewers(t *testing.T) {
	code := reviewRec(t, ProviderClaude, "code-reviewer", "s-a", KindCodeReview, VerdictPass)
	sec := reviewRec(t, ProviderClaude, "security-reviewer", "s-b", KindSecurityReview, VerdictPass)
	if !SeparateReviewers(code, sec) {
		t.Error("distinct sessions must be accepted")
	}
	sec.Producer.Session = "s-a"
	if SeparateReviewers(code, sec) {
		t.Error("one session cannot hold both review roles")
	}
}
