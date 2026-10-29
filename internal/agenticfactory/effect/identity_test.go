package effect

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// fakeIdentity answers with numeric IDs and records what it was asked. It has no write method at all.
type fakeIdentity struct {
	actor    int64
	actorErr error
	issues   map[int]int64 // issue number -> author id
	comments map[int64]int64
	asked    []string
}

func (f *fakeIdentity) CurrentUserID(context.Context) (int64, error) {
	f.asked = append(f.asked, "user")
	return f.actor, f.actorErr
}

func (f *fakeIdentity) IssueAuthorID(_ context.Context, repo string, n int) (int64, error) {
	f.asked = append(f.asked, fmt.Sprintf("issue %s#%d", repo, n))
	id, ok := f.issues[n]
	if !ok {
		return 0, errors.New("lookup failed")
	}
	return id, nil
}

func (f *fakeIdentity) CommentAuthorID(_ context.Context, repo string, c int64) (int64, error) {
	f.asked = append(f.asked, fmt.Sprintf("comment %s/%d", repo, c))
	id, ok := f.comments[c]
	if !ok {
		return 0, errors.New("lookup failed")
	}
	return id, nil
}

// testJudge is a test double that follows the stated policy. The real judging function is lane 18's.
func testJudge(op IssueOp, s Subject) Decision {
	if !s.ActorKnown || !s.AuthorKnown {
		return Decision{Reason: ReasonIdentityUnknown}
	}
	if s.ActorID != s.AuthorID {
		// someone else's Issue: only a separate, explicitly set grant for exactly this operation allows it
		if s.ForeignAllowed && op != OpCommentEdit && op != OpIssueOther {
			return Decision{Allow: true}
		}
		return Decision{Reason: ReasonNotIssueAuthor}
	}
	switch op {
	case OpIssueBodyEdit, OpCommentEdit:
		return Decision{Allow: true}
	}
	return Decision{Reason: ReasonNotIssueAuthor} // close, labels, assignees, other: not part of what is allowed
}

const me, other = int64(1001), int64(2002)

func enforcer(f *fakeIdentity) *Enforcer { return &Enforcer{Identity: f, Judge: testJudge} }

func world() *fakeIdentity {
	return &fakeIdentity{actor: me, issues: map[int]int64{7: me, 8: other}, comments: map[int64]int64{501: me, 502: other}}
}

func reasonOf(t *testing.T, err error) RefusalReason {
	t.Helper()
	var re *RefusedError
	if !errors.As(err, &re) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	return re.Reason
}

func TestIssueBodyEditOnlyOnAnIssueWeAuthored(t *testing.T) {
	e := enforcer(world())
	if err := e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(`{"body":"new text"}`)); err != nil {
		t.Errorf("own issue refused: %v", err)
	}
	err := e.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(`{"body":"new text"}`))
	if r := reasonOf(t, err); r != ReasonNotIssueAuthor {
		t.Errorf("someone else's issue: reason = %s", r)
	}
}

func TestChangesToSomeoneElsesIssueAreRefusedWhateverTheField(t *testing.T) {
	e := enforcer(world())
	for name, body := range map[string]string{
		"close":     `{"state":"closed"}`,
		"labels":    `{"labels":["x"]}`,
		"assignee":  `{"assignee":"someone"}`,
		"assignees": `{"assignees":["a"]}`,
		"mixed":     `{"body":"x","state":"closed"}`,
		"title":     `{"title":"x"}`,
	} {
		err := e.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(body))
		if r := reasonOf(t, err); r != ReasonNotIssueAuthor {
			t.Errorf("%s: reason = %s", name, r)
		}
	}
	// closing even our own issue is not part of the allowed set here
	if err := e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(`{"state":"closed"}`)); err == nil {
		t.Error("closing was allowed through the body-edit rule")
	}
}

func TestCommentEditOnlyForACommentWeWrote(t *testing.T) {
	e := enforcer(world())
	if err := e.Check(t.Context(), "PATCH", "/repos/o/r/issues/comments/501", nil, []byte(`{"body":"x"}`)); err != nil {
		t.Errorf("own comment refused: %v", err)
	}
	if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/comments/502", nil, nil)); r != ReasonNotIssueAuthor {
		t.Errorf("someone else's comment: reason = %s", r)
	}
}

func TestUnknownIdentityRefusesAndParksAsIdentityUnknown(t *testing.T) {
	for name, f := range map[string]*fakeIdentity{
		"actor lookup fails":  {actorErr: errors.New("boom"), issues: map[int]int64{7: me}, comments: map[int64]int64{501: me}},
		"actor id zero":       {actor: 0, issues: map[int]int64{7: me}, comments: map[int64]int64{501: me}},
		"actor id negative":   {actor: -5, issues: map[int]int64{7: me}, comments: map[int64]int64{501: me}},
		"author lookup fails": {actor: me, issues: map[int]int64{}, comments: map[int64]int64{}},
		"author id zero":      {actor: me, issues: map[int]int64{7: 0}, comments: map[int64]int64{501: 0}},
	} {
		e := enforcer(f)
		if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(`{"body":"x"}`))); r != ReasonIdentityUnknown {
			t.Errorf("%s (issue): reason = %s", name, r)
		}
		if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/comments/501", nil, nil)); r != ReasonIdentityUnknown {
			t.Errorf("%s (comment): reason = %s", name, r)
		}
	}
}

func TestNothingWiredFailsClosedButPlainWritesStillPass(t *testing.T) {
	for name, e := range map[string]*Enforcer{
		"nil enforcer": nil, "no identity": {Judge: testJudge}, "no judge": {Identity: world()}, "empty": {},
	} {
		if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/comments/501", nil, nil)); r != ReasonIdentityUnknown {
			t.Errorf("%s: comment edit reason = %s", name, r)
		}
		if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(`{"body":"x"}`))); r != ReasonIdentityUnknown {
			t.Errorf("%s: issue edit reason = %s", name, r)
		}
		for _, ok := range [][2]string{{"POST", "/repos/o/r/issues/8/comments"}, {"POST", "/repos/o/r/issues"}, {"PATCH", "/repos/o/r/labels/p0"}, {"GET", "/repos/o/r/issues/8"}} {
			if err := e.Check(t.Context(), ok[0], ok[1], nil, []byte(`{}`)); err != nil {
				t.Errorf("%s: %s %s refused: %v", name, ok[0], ok[1], err)
			}
		}
	}
}

func TestCommentCreateOnSomeoneElsesIssueNeedsNoIdentity(t *testing.T) {
	f := world()
	if err := enforcer(f).Check(t.Context(), "POST", "/repos/o/r/issues/8/comments", nil, []byte(`{"body":"x"}`)); err != nil {
		t.Fatalf("comment on another's issue refused: %v", err)
	}
	if len(f.asked) != 0 {
		t.Errorf("a plain comment create made %d identity lookups", len(f.asked))
	}
}

func TestStaticGuardStillAppliesAndNeverLetsTheJudgeOverrideIt(t *testing.T) {
	allowAll := &Enforcer{Identity: world(), Judge: func(IssueOp, Subject) Decision { return Decision{Allow: true} }}
	for _, c := range [][2]string{
		{"PUT", "/repos/o/r/branches/main/protection"},
		{"POST", "/repos/o/r/hooks"},
		{"DELETE", "/repos/o/r/issues/7"},
		{"PATCH", "/repos/o/r/issues/7/labels"},
	} {
		if err := allowAll.Check(t.Context(), c[0], c[1], nil, []byte(`{"body":"x"}`)); !errors.Is(err, ErrRefused) {
			t.Errorf("%s %s: err = %v", c[0], c[1], err)
		}
	}
}

func TestAJudgeThatAllowsWithUnknownIDsIsNotTrusted(t *testing.T) {
	f := &fakeIdentity{actorErr: errors.New("x"), issues: map[int]int64{7: me}}
	e := &Enforcer{Identity: f, Judge: func(IssueOp, Subject) Decision { return Decision{Allow: true} }}
	if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(`{"body":"x"}`))); r != ReasonIdentityUnknown {
		t.Errorf("reason = %s", r)
	}
	bogus := &Enforcer{Identity: world(), Judge: func(IssueOp, Subject) Decision { return Decision{Reason: "whatever"} }}
	if r := reasonOf(t, bogus.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(`{"body":"x"}`))); r != ReasonIdentityUnknown {
		t.Errorf("an unrecognised reason must still refuse, as identity-unknown: %s", r)
	}
}

func TestBadBodiesAndRoutesAreMalformedOrNotAllowlisted(t *testing.T) {
	e := enforcer(world())
	for _, body := range []string{"", "not json", "[]", "{}", "null"} {
		if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(body))); r != ReasonMalformed {
			t.Errorf("body %q: reason = %s", body, r)
		}
	}
	for _, p := range []string{"/repos/o/r/issues/abc", "/repos/o/r/issues/0", "/repos/o/r/issues/comments/abc", "/repos/o/r/issues/7/comments/9"} {
		if err := e.Check(t.Context(), "PATCH", p, nil, []byte(`{"body":"x"}`)); err == nil {
			t.Errorf("%s was allowed", p)
		}
	}
}

func TestLookupsAskForTheNumbersTheRouteNames(t *testing.T) {
	f := world()
	e := enforcer(f)
	_ = e.Check(t.Context(), "PATCH", "/repos/o/r/issues/7", nil, []byte(`{"body":"x"}`))
	_ = e.Check(t.Context(), "PATCH", "/repos/o/r/issues/comments/501", nil, nil)
	want := []string{"user", "issue o/r#7", "user", "comment o/r/501"}
	if fmt.Sprint(f.asked) != fmt.Sprint(want) {
		t.Errorf("asked %v, want %v", f.asked, want)
	}
}

func TestIdentityRefusalParksTheIntentWithItsReason(t *testing.T) {
	for _, reason := range []RefusalReason{ReasonNotIssueAuthor, ReasonIdentityUnknown} {
		s := &fakeSender{err: &RefusedError{Reason: reason, Method: "PATCH", Path: "/repos/o/r/issues/8"}}
		lk := &fakeLookup{}
		ex := newExec(s, lk, NewMemIntentStore())
		in, _ := ex.Execute(t.Context(), req())
		if in.Status != StatusParked || in.ParkReason != string(reason) || lk.calls != 0 {
			t.Errorf("%s: in=%+v lookups=%d", reason, in, lk.calls)
		}
		if again, _ := ex.Execute(t.Context(), req()); again.Status != StatusParked || s.calls != 1 {
			t.Errorf("%s: sent again (%d)", reason, s.calls)
		}
	}
}

func TestCheckRequestAloneNeverLetsAConditionalRoutePass(t *testing.T) {
	for _, c := range [][2]string{{"PATCH", "/repos/o/r/issues/comments/501"}, {"PATCH", "/repos/o/r/issues/7"}} {
		if r := reasonOf(t, CheckRequest(c[0], c[1], nil)); r != ReasonIdentityUnknown {
			t.Errorf("%s %s: reason = %s", c[0], c[1], r)
		}
	}
}

// The collaboration permissions are separate, default off, and each opens only its own operation in
// its own repository.
func TestForeignWritesNeedTheExactOperationInTheExactRepository(t *testing.T) {
	bodies := map[string]struct {
		body string
		op   IssueOp
	}{
		"body": {`{"body":"x"}`, OpIssueBodyEdit}, "close": {`{"state":"closed"}`, OpIssueClose},
		"labels": {`{"labels":["x"]}`, OpIssueLabels}, "assignees": {`{"assignees":["a"]}`, OpIssueAssignees},
	}
	for name, c := range bodies {
		e := enforcer(world()) // Foreign is nil: nothing allowed
		if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(c.body))); r != ReasonNotIssueAuthor {
			t.Errorf("%s with no setting: reason = %s", name, r)
		}
	}
	for have, hc := range bodies {
		for name, c := range bodies {
			e := enforcer(world())
			e.Foreign = func(repo string, op IssueOp) bool { return repo == "o/r" && op == hc.op }
			err := e.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(c.body))
			if (name == have) != (err == nil) {
				t.Errorf("allowed %s, attempted %s: err = %v", have, name, err)
			}
		}
	}
	other := enforcer(world())
	other.Foreign = func(repo string, op IssueOp) bool { return repo == "elsewhere/x" }
	if err := other.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(`{"body":"x"}`)); err == nil {
		t.Error("a permission for another repository allowed a write here")
	}
}

// A comment edit needs no setting and no setting can widen it to someone else's comment.
func TestNoSettingWidensACommentEditToSomeoneElsesComment(t *testing.T) {
	e := enforcer(world())
	e.Foreign = func(string, IssueOp) bool { return true }
	if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/comments/502", nil, nil)); r != ReasonNotIssueAuthor {
		t.Errorf("reason = %s", r)
	}
}

// The settings answer is only data for the judge. Without a judge, allowing everything changes nothing.
func TestSettingsWithoutAJudgeStillFailClosed(t *testing.T) {
	e := &Enforcer{Identity: world(), Foreign: func(string, IssueOp) bool { return true }}
	if r := reasonOf(t, e.Check(t.Context(), "PATCH", "/repos/o/r/issues/8", nil, []byte(`{"body":"x"}`))); r != ReasonIdentityUnknown {
		t.Errorf("reason = %s", r)
	}
}
