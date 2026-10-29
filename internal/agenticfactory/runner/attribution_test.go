package runner

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// fakeAttributor stands in for the read-only attribution lookup of
// SPEC-AGENTIC-FACTORY-IDENTITY-001: it returns the role it was told per result name.
type fakeAttributor struct{ roles map[string]string }

func (f fakeAttributor) Attribute(res Result) Attribution {
	role, ok := f.roles[res.Name]
	if !ok {
		return Attribution{Reason: "no such producer"}
	}
	return Attribution{Attributed: true, Role: role, Session: "producer-" + res.Name, Attempt: "attempt-9", Epoch: 3}
}

func goodRoles() fakeAttributor {
	return fakeAttributor{roles: map[string]string{
		"code-review": core.RoleReviewer, "independent-audit": core.RoleAuditor,
		"coordinator": core.RoleCoordinator, "qa": core.RoleQA,
	}}
}

func verifyWith(t *testing.T, a Attributor, mutate func(*Verification)) claimRecord {
	t.Helper()
	r, ref := newRunner(t)
	r.Attributor = a
	tk := toVerifying(t, r, ref)
	v := passAllFor(tk, headA)
	if mutate != nil {
		mutate(&v)
	}
	rec, err := r.Verify(tk, v)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

// While the identity SPEC is incomplete no review, audit, QA or coordinator
// evidence can count: without an attribution source everything is refused.
func TestWithoutAnAttributionSourceEvidenceIsRefused(t *testing.T) {
	rec := verifyWith(t, nil, nil)
	if rec.State != core.StateVerifying || gateReason(rec) != core.ReasonProducerNotAllowed || rec.Park != nil {
		t.Fatalf("record: %+v", rec)
	}
	if !strings.Contains(rec.Gate.Detail, core.AttributionLocked) {
		t.Errorf("the lock must be named in the gate detail: %q", rec.Gate.Detail)
	}
}

func TestCorrectRolesLetEvidenceCount(t *testing.T) {
	rec := verifyWith(t, goodRoles(), nil)
	if rec.State != core.StateReady {
		t.Fatalf("record: %+v", rec)
	}
}

func TestEveryProducerCheckedKindNeedsItsOwnRole(t *testing.T) {
	cases := map[string]string{
		"code-review":       core.RoleAuditor,  // an auditor cannot supply the review
		"independent-audit": core.RoleReviewer, // nor a reviewer the audit
		"coordinator":       core.RoleAgent,    // nor the executing agent the coordinator handoff
	}
	for name, wrong := range cases {
		a := goodRoles()
		a.roles[name] = wrong
		if rec := verifyWith(t, a, nil); rec.State != core.StateVerifying || gateReason(rec) != core.ReasonProducerNotAllowed {
			t.Errorf("%s attributed as %s: %+v", name, wrong, rec)
		}
	}
	// A producer the lookup does not know is unattributable.
	a := goodRoles()
	delete(a.roles, "independent-audit")
	if rec := verifyWith(t, a, nil); gateReason(rec) != core.ReasonProducerNotAllowed {
		t.Errorf("unknown producer: %+v", rec)
	}
}

func TestQAEvidenceNeedsTheQARoleWhenPresent(t *testing.T) {
	add := func(v *Verification) {
		v.Results = append(v.Results, Result{Kind: KindQA, Name: "qa", Head: headA, Verdict: VerdictPass})
	}
	if rec := verifyWith(t, goodRoles(), add); rec.State != core.StateReady {
		t.Fatalf("qa by the qa role: %+v", rec)
	}
	a := goodRoles()
	a.roles["qa"] = core.RoleReviewer
	if rec := verifyWith(t, a, add); gateReason(rec) != core.ReasonProducerNotAllowed {
		t.Errorf("qa by the wrong role: %+v", rec)
	}
}

// The role a producer writes inside its own evidence is never used: only the
// lookup decides.
func TestSelfDeclaredRoleInTheEvidenceIsIgnored(t *testing.T) {
	a := goodRoles()
	a.roles["code-review"] = core.RoleAgent
	rec := verifyWith(t, a, func(v *Verification) {
		for i := range v.Results {
			if v.Results[i].Kind == KindReview {
				v.Results[i].Session = "i-am-the-reviewer"
			}
		}
	})
	if gateReason(rec) != core.ReasonProducerNotAllowed {
		t.Errorf("a self-declared session decided the verdict: %+v", rec)
	}
}

// The writer handoff is judged by the holder and epoch the attempt was issued
// to, not by a lookup.
func TestWriterHandoffMustComeFromTheIssuedHolderAndEpoch(t *testing.T) {
	for name, mut := range map[string]func(*Result){
		"another session": func(res *Result) { res.Session = "someone-else" },
		"an older epoch":  func(res *Result) { res.Epoch-- },
		"no session":      func(res *Result) { res.Session = "" },
	} {
		rec := verifyWith(t, goodRoles(), func(v *Verification) {
			for i := range v.Results {
				if v.Results[i].Kind == KindHandoffWriter {
					mut(&v.Results[i])
				}
			}
		})
		if rec.State != core.StateVerifying || gateReason(rec) != core.ReasonProducerNotAllowed {
			t.Errorf("%s: %+v", name, rec)
		}
	}
}

// Evidence from an unattributed producer is not even allowed to send the Issue
// back to implementing: only counted evidence moves it.
func TestUnattributedFailDoesNotCountAsAFail(t *testing.T) {
	rec := verifyWith(t, nil, func(v *Verification) {
		for i := range v.Results {
			if v.Results[i].Kind == KindReview {
				v.Results[i].Verdict = VerdictFail
			}
		}
	})
	if rec.State != core.StateVerifying || gateReason(rec) != core.ReasonProducerNotAllowed {
		t.Errorf("record: %+v", rec)
	}
}

func TestChecksNeedNoAttribution(t *testing.T) {
	// Only checks fail here; with a good lookup the verdict mapping is unchanged.
	rec := verifyWith(t, goodRoles(), func(v *Verification) { v.Results[0].Verdict = "skipped" })
	if gateReason(rec) != core.ReasonCheckNotPassed {
		t.Errorf("record: %+v", rec)
	}
	_ = errors.New
}

func gateReason(rec claim.Record) core.BlockReason {
	if rec.Gate == nil {
		return ""
	}
	return rec.Gate.Reason
}
