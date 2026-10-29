package identity

import (
	"fmt"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const head = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// procs is the fake process table: pid -> start fingerprint, and who is alive.
type procs struct {
	start map[int]string
	live  map[int]bool
}

func (p *procs) probe(pid int) (string, homestate.ProcessIdentityState) {
	if !p.live[pid] {
		return "", homestate.ProcessIdentityDead
	}
	return p.start[pid], homestate.ProcessIdentityLive
}

type rig struct {
	t     *testing.T
	book  *Book
	procs *procs
	issue core.IssueRef
	next  int
}

// newRig has Issue 1948 claimed by the writer session "w-sess" (pid 100).
func newRig(t *testing.T) *rig {
	t.Helper()
	st, err := claim.OpenWithLimit(t.TempDir(), 10)
	if err != nil {
		t.Fatal(err)
	}
	pr := &procs{start: map[int]string{}, live: map[int]bool{}}
	r := &rig{t: t, procs: pr, book: &Book{Store: st, Probe: pr.probe}, next: 200}
	r.issue, _ = core.NewIssueRef(1948)
	if _, err := st.Acquire(r.issue, "w-sess", "d1", "run-1"); err != nil {
		t.Fatal(err)
	}
	r.session("w-sess", SessionAgent, 100)
	return r
}

func (r *rig) session(id string, kind SessionKind, pid int) Session {
	r.t.Helper()
	r.procs.start[pid], r.procs.live[pid] = fmt.Sprintf("start-%d", pid), true
	s := Session{ID: id, Kind: kind, PID: pid, Start: r.procs.start[pid], Run: "run-1"}
	if err := r.book.PutSession(s); err != nil {
		r.t.Fatal(err)
	}
	return s
}

// producer makes a session of the given kind, optionally with a grant of role.
func (r *rig) producer(kind SessionKind, role Kind) Session {
	r.next++
	s := r.session(fmt.Sprintf("p-%d", r.next), kind, r.next)
	if role != "" {
		g := Grant{ID: fmt.Sprintf("g-%d", r.next), Role: role, Issue: r.issue.Number, Session: s.ID, PID: s.PID, Start: s.Start, AuthBinding: "binding-" + s.ID, Writer: ControlService}
		if err := r.book.PutGrant(g); err != nil {
			r.t.Fatal(err)
		}
	}
	return s
}

func (r *rig) evidence(name string, kind Kind, by Session, operator bool) runner.Result {
	r.t.Helper()
	ev := Evidence{Name: name, Head: head, Kind: kind, Session: by.ID, PID: by.PID, Operator: operator}
	if err := r.book.PutEvidence(ev); err != nil {
		r.t.Fatal(err)
	}
	return runner.Result{Kind: runnerKind(kind), Name: name, Head: head}
}

func (r *rig) attribute(res runner.Result) runner.Attribution {
	return r.book.AttributeIssue(r.issue, res)
}

var evidenceKinds = []Kind{CodeReview, SecurityReview, Audit, QA, WriterHandoff, CoordinatorAcceptance}

// REQ-AFI-006 / AC-AFI-008: the producer x evidence-kind matrix of design 5, 42
// cells plus the unregistered producer for each kind (48 cases): 8 are allowed,
// 34 are role-kind-mismatch, 6 are not-registered.
func TestTheProducerEvidenceMatrix(t *testing.T) {
	type producer struct {
		name    string
		make    func(r *rig) (Session, bool) // the session and whether it is an operator record
		allowed map[Kind]string              // evidence kind -> role displayed on success
	}
	producers := []producer{
		{"code-review role", func(r *rig) (Session, bool) { return r.producer(SessionReview, CodeReview), false }, map[Kind]string{CodeReview: core.RoleReviewer}},
		{"security-review role", func(r *rig) (Session, bool) { return r.producer(SessionReview, SecurityReview), false }, map[Kind]string{SecurityReview: core.RoleReviewer}},
		{"audit role", func(r *rig) (Session, bool) { return r.producer(SessionReview, Audit), false }, map[Kind]string{Audit: core.RoleAuditor}},
		{"qa role", func(r *rig) (Session, bool) { return r.producer(SessionReview, QA), false }, map[Kind]string{QA: core.RoleQA}},
		{"current attempt's agent", func(r *rig) (Session, bool) {
			return Session{ID: "w-sess", Kind: SessionAgent, PID: 100, Start: "start-100"}, false
		}, map[Kind]string{WriterHandoff: core.RoleAgent}},
		{"registered coordinator", func(r *rig) (Session, bool) { return r.producer(SessionCoordinator, ""), false }, map[Kind]string{CoordinatorAcceptance: core.RoleCoordinator}},
		{"operator record", func(r *rig) (Session, bool) { return r.producer(SessionReview, ""), true }, map[Kind]string{CodeReview: core.RoleReviewer, SecurityReview: core.RoleReviewer}},
	}
	allowed, mismatch, unreg := 0, 0, 0
	for _, p := range producers {
		for _, k := range evidenceKinds {
			r := newRig(t)
			sess, op := p.make(r)
			res := r.evidence(fmt.Sprintf("ev-%s-%s", p.name, k), k, sess, op)
			got := r.attribute(res)
			if role, ok := p.allowed[k]; ok {
				allowed++
				if !got.Attributed || got.Role != role {
					t.Errorf("%s x %s: want attributed as %s, got %+v", p.name, k, role, got)
				}
			} else {
				mismatch++
				if got.Attributed || got.Reason != ReasonRoleKindMismatch {
					t.Errorf("%s x %s: want %s, got %+v", p.name, k, ReasonRoleKindMismatch, got)
				}
			}
		}
	}
	for _, k := range evidenceKinds { // a producer that is neither registered nor an operator record
		r := newRig(t)
		unreg++
		got := r.attribute(runner.Result{Kind: runnerKind(k), Name: "ev-nobody-" + string(k), Head: head})
		if got.Attributed || got.Reason != ReasonNotRegistered {
			t.Errorf("unregistered x %s: want %s, got %+v", k, ReasonNotRegistered, got)
		}
	}
	if allowed != 8 || mismatch != 34 || unreg != 6 {
		t.Fatalf("the matrix is 8 + 34 + 6, got %d + %d + %d", allowed, mismatch, unreg)
	}
}

// REQ-AFI-005/006: whoever wrote an attempt on the Issue cannot produce review,
// audit, QA or coordinator evidence for it, grant or no grant.
func TestAWriterOfTheIssueIsNeverItsReviewerAuditorQAOrCoordinator(t *testing.T) {
	for _, c := range []struct {
		name string
		kind Kind
		role Kind
		sk   SessionKind
	}{
		{"reviewer", CodeReview, CodeReview, SessionReview},
		{"auditor", Audit, Audit, SessionReview},
		{"qa", QA, QA, SessionReview},
		{"coordinator", CoordinatorAcceptance, "", SessionCoordinator},
	} {
		r := newRig(t)
		s := r.producer(c.sk, c.role)
		// the session wrote an attempt on this Issue: a claim event under its id
		if _, err := r.book.Store.Update(r.issue, "w-sess", 1, func(rec *claim.Record) error {
			rec.AddEvent(claim.Event{Kind: "repair-claim", Session: s.ID})
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		got := r.attribute(r.evidence("ev-writer-"+c.name, c.kind, s, false))
		if got.Attributed || got.Reason != ReasonWasWriter {
			t.Errorf("%s who was a writer: %+v", c.name, got)
		}
	}
}

// REQ-AFI-004: the process that registered the evidence is the one bound to the
// session and it is still that process.
func TestAMismatchedOrDeadProcessIsAnIdentityMismatch(t *testing.T) {
	r := newRig(t)
	s := r.producer(SessionReview, Audit)
	res := r.evidence("ev-live", Audit, s, false)
	if got := r.attribute(res); !got.Attributed {
		t.Fatalf("setup: %+v", got)
	}
	r.procs.live[s.PID] = false // the process is gone
	if got := r.attribute(res); got.Attributed || got.Reason != ReasonIdentityMismatch {
		t.Errorf("a dead process: %+v", got)
	}
	r.procs.live[s.PID] = true
	r.procs.start[s.PID] = "start-reused" // the pid was reused by another process
	if got := r.attribute(res); got.Attributed || got.Reason != ReasonIdentityMismatch {
		t.Errorf("a reused pid: %+v", got)
	}
	r.procs.start[s.PID] = s.Start
	other := r.producer(SessionReview, Audit)
	copied := Evidence{Name: "ev-copied", Head: head, Kind: Audit, Session: s.ID, PID: other.PID} // a session id copied by another process
	if err := r.book.PutEvidence(copied); err != nil {
		t.Fatal(err)
	}
	if got := r.attribute(runner.Result{Kind: runner.KindAudit, Name: "ev-copied", Head: head}); got.Attributed || got.Reason != ReasonIdentityMismatch {
		t.Errorf("a copied session id: %+v", got)
	}
}

// REQ-AFI-001: only a grant the ControlService wrote, with its authentication
// binding, for this Issue, confers a role.
func TestOnlyAControlServiceGrantForThisIssueConfersARole(t *testing.T) {
	cases := map[string]func(r *rig, s Session) Grant{
		"no grant": nil,
		"written by someone else": func(r *rig, s Session) Grant {
			return Grant{ID: "g-x", Role: Audit, Issue: r.issue.Number, Session: s.ID, PID: s.PID, Start: s.Start, AuthBinding: "b", Writer: "agent-9"}
		},
		"no authentication binding": func(r *rig, s Session) Grant {
			return Grant{ID: "g-x", Role: Audit, Issue: r.issue.Number, Session: s.ID, PID: s.PID, Start: s.Start, Writer: ControlService}
		},
		"a grant for another Issue": func(r *rig, s Session) Grant {
			return Grant{ID: "g-x", Role: Audit, Issue: r.issue.Number + 1, Session: s.ID, PID: s.PID, Start: s.Start, AuthBinding: "b", Writer: ControlService}
		},
	}
	for name, mk := range cases {
		r := newRig(t)
		s := r.producer(SessionReview, "")
		if mk != nil {
			if err := r.book.PutGrant(mk(r, s)); err != nil {
				t.Fatal(err)
			}
		}
		got := r.attribute(r.evidence("ev-"+name, Audit, s, false))
		if got.Attributed || got.Reason != ReasonGrantNotInRef {
			t.Errorf("%s: %+v", name, got)
		}
	}
}

// §7: a grant whose binding no longer matches the session (a restart, a reused pid)
// is re-verified and refused, and nothing is deleted.
func TestAGrantThatNoLongerMatchesItsSessionIsReverifiedAndRefused(t *testing.T) {
	r := newRig(t)
	s := r.producer(SessionReview, "")
	g := Grant{ID: "g-stale", Role: Audit, Issue: r.issue.Number, Session: s.ID, PID: s.PID + 500, Start: "start-other", AuthBinding: "b", Writer: ControlService}
	if err := r.book.PutGrant(g); err != nil {
		t.Fatal(err)
	}
	got := r.attribute(r.evidence("ev-stale", Audit, s, false))
	if got.Attributed || got.Reason != ReasonGrantReverifyFails {
		t.Errorf("a stale grant: %+v", got)
	}
	if _, err := r.book.Store.GetDoc("grant", "g-stale"); err != nil {
		t.Errorf("the grant must be kept, not deleted: %v", err)
	}
}

// A role grant for another kind does not stretch to this one, and the evidence
// asked about must be the kind it was registered as.
func TestAGrantForOneKindDoesNotCoverAnotherAndTheAskedKindMustMatch(t *testing.T) {
	r := newRig(t)
	s := r.producer(SessionReview, QA)
	if got := r.attribute(r.evidence("ev-as-audit", Audit, s, false)); got.Attributed || got.Reason != ReasonRoleKindMismatch {
		t.Errorf("a QA grant used for audit evidence: %+v", got)
	}
	res := r.evidence("ev-qa", QA, s, false)
	res.Kind = runner.KindAudit // asked as audit evidence, registered as QA
	if got := r.attribute(res); got.Attributed || got.Reason != ReasonRoleKindMismatch {
		t.Errorf("asked as another kind than registered: %+v", got)
	}
}

// The lookup reads only: every document and the claim record keep their revisions.
func TestTheLookupChangesNothing(t *testing.T) {
	r := newRig(t)
	s := r.producer(SessionReview, Audit)
	res := r.evidence("ev-ro", Audit, s, false)
	before, _ := r.book.Store.Read(r.issue)
	revs := map[string]int64{}
	for _, k := range []string{"session", "grant", "evidence"} {
		ds, _ := r.book.Store.ListDocs(k)
		for _, d := range ds {
			revs[k+"/"+d.Name] = d.Rev
		}
	}
	for i := 0; i < 3; i++ {
		_ = r.attribute(res)
		_ = r.attribute(runner.Result{Kind: runner.KindAudit, Name: "nobody", Head: head})
	}
	after, _ := r.book.Store.Read(r.issue)
	if after.Seq != before.Seq || after.Rev != before.Rev {
		t.Errorf("the claim record moved: %d -> %d", before.Seq, after.Seq)
	}
	for _, k := range []string{"session", "grant", "evidence"} {
		ds, _ := r.book.Store.ListDocs(k)
		for _, d := range ds {
			if revs[k+"/"+d.Name] != d.Rev {
				t.Errorf("%s/%s changed", k, d.Name)
			}
		}
	}
}

// Through the runner: with the lookup wired, the merge gate passes when the review,
// the audit and the coordinator are attributed, and records producer-not-allowed
// when one is not.
type ownIssues struct{}

func (ownIssues) IssueFacts(core.IssueRef) (runner.IssueFacts, error) {
	return runner.IssueFacts{RequestedRepo: "acme/widgets", ActualRepo: "acme/widgets", UserID: 7, AuthorID: 7, Owner: core.OwnerNone}, nil
}

func TestTheMergeGateUsesTheLookupThroughTheRunner(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 10)
	pr := &procs{start: map[int]string{}, live: map[int]bool{}}
	book := &Book{Store: st, Probe: pr.probe}
	r := runner.New(st)
	r.Attributor, r.Facts = book, ownIssues{}
	ref, _ := core.NewIssueRef(1948)
	tk, err := r.Claim(ref, "w-sess", "d1", "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Submit(tk, head, false); err != nil {
		t.Fatal(err)
	}
	rg := &rig{t: t, procs: pr, book: book, issue: ref, next: 300}
	rg.session("w-sess", SessionAgent, 100)
	rev := rg.producer(SessionReview, CodeReview)
	aud := rg.producer(SessionReview, Audit)
	crd := rg.producer(SessionCoordinator, "")
	v := runner.Verification{Head: head, Results: []runner.Result{
		{Kind: runner.KindCheck, Name: "build", Head: head, Verdict: runner.VerdictPass},
		rg.evidence("code-review", CodeReview, rev, false),
		rg.evidence("independent-audit", Audit, aud, false),
		{Kind: runner.KindHandoffWriter, Name: "writer", Head: head, Verdict: runner.VerdictPass, Session: tk.Holder, Epoch: tk.Epoch},
		rg.evidence("coordinator", CoordinatorAcceptance, crd, false),
	}}
	for i := range v.Results {
		v.Results[i].Verdict = runner.VerdictPass
	}
	rec, err := r.Verify(tk, v)
	if err != nil || rec.State != core.StateReady {
		t.Fatalf("all attributed: %+v %v", rec.Gate, err)
	}
}

func TestTheMergeGateRefusesAnUnattributedProducerThroughTheRunner(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 10)
	pr := &procs{start: map[int]string{}, live: map[int]bool{}}
	book := &Book{Store: st, Probe: pr.probe}
	r := runner.New(st)
	r.Attributor, r.Facts = book, ownIssues{}
	ref, _ := core.NewIssueRef(1948)
	tk, _ := r.Claim(ref, "w-sess", "d1", "run-1")
	_, _ = r.StartAttempt(tk, "attempt-1", core.DeliveryExecuted)
	_, _ = r.Submit(tk, head, false)
	rg := &rig{t: t, procs: pr, book: book, issue: ref, next: 300}
	rg.session("w-sess", SessionAgent, 100)
	rev := rg.producer(SessionReview, CodeReview)
	crd := rg.producer(SessionCoordinator, "")
	v := runner.Verification{Head: head, Results: []runner.Result{
		{Kind: runner.KindCheck, Name: "build", Head: head, Verdict: runner.VerdictPass},
		rg.evidence("code-review", CodeReview, rev, false),
		{Kind: runner.KindAudit, Name: "independent-audit", Head: head, Verdict: runner.VerdictPass}, // never registered
		{Kind: runner.KindHandoffWriter, Name: "writer", Head: head, Verdict: runner.VerdictPass, Session: tk.Holder, Epoch: tk.Epoch},
		rg.evidence("coordinator", CoordinatorAcceptance, crd, false),
	}}
	v.Results[1].Verdict, v.Results[4].Verdict = runner.VerdictPass, runner.VerdictPass
	rec, err := r.Verify(tk, v)
	if err == nil && rec.State == core.StateReady {
		t.Fatalf("an unregistered audit passed the gate")
	}
	if rec.Gate == nil || rec.Gate.Reason != core.ReasonProducerNotAllowed {
		t.Errorf("want a producer-not-allowed gate, got %+v (err %v)", rec.Gate, err)
	}
}
