package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/intake"
)

// Duplicate work before an Issue is created (lane-18 f9f44e5ecc): the search only
// reads; your own Issue is reused, another author's stops the request and no Issue
// is made in its place; an identity that cannot be told stops it too. Candidate
// ids are numeric host ids and the fake host's answers are test inputs.

func blockedOf(t *testing.T, r *intakeRig, key string) *claim.Blocked {
	t.Helper()
	req, err := r.st.GetRequest(key)
	if err != nil {
		t.Fatal(err)
	}
	if req.Blocked == nil {
		t.Fatalf("%s has no block recorded (stage %s)", key, req.Stage)
	}
	return req.Blocked
}

// No duplicate: the Issue is created, after the search ran exactly once.
func TestIntakeDuplicates_NoneFound_CreatesTheIssue(t *testing.T) {
	r := newRig(t, rigCfg)
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if req, err := r.in.Advance(context.Background(), "req-1"); err != nil || req.Stage != claim.StageIssueConfirmed {
		t.Fatalf("advance: %+v, %v", req, err)
	}
	if p, _ := r.pub.counts(); p != 1 {
		t.Errorf("PublishIssue calls = %d, want 1", p)
	}
	if r.pub.dupLookups != 1 {
		t.Errorf("duplicate searches = %d, want 1 (read-only, before the create)", r.pub.dupLookups)
	}
}

// An existing Issue of mine is reused: no second Issue is created.
func TestIntakeDuplicates_OwnIssue_IsReused_NothingCreated(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.dups = []intake.Candidate{{Issue: 4242, AuthorID: me, Ref: "issue/4242"}}
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	req, err := r.in.Advance(context.Background(), "req-1")
	if err != nil || req.Stage != claim.StageIssueConfirmed || req.Issue != 4242 {
		t.Fatalf("advance: %+v, %v; want ISSUE_CONFIRMED on the existing #4242", req, err)
	}
	if p, _ := r.pub.counts(); p != 0 {
		t.Errorf("PublishIssue calls = %d, want 0: the work already has my Issue", p)
	}
}

// Another author's Issue for the same work stops the request: no Issue is
// created in its place, the block names the candidate and says how to proceed,
// and it is not a retry.
func TestIntakeDuplicates_ForeignIssue_BlocksAndCreatesNothing(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.dups = []intake.Candidate{{Issue: 4343, AuthorID: them, Ref: "issue/4343"}}
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	_, err := r.in.Advance(context.Background(), "req-1")
	if !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Fatalf("advance: %v, want ErrNotIssueAuthor", err)
	}
	b := blockedOf(t, r, "req-1")
	if b.Reason != core.ReasonNotIssueAuthor || b.Candidate != 4343 || b.Retryable || b.Unknown {
		t.Errorf("block = %+v; want not-issue-author naming #4343, neither retryable nor unknown", b)
	}
	if !strings.Contains(b.Resume, "4343") {
		t.Errorf("resume %q does not point at the existing Issue", b.Resume)
	}
	if p, _ := r.pub.counts(); p != 0 {
		t.Errorf("PublishIssue calls = %d, want 0: no Issue may be made in place of another author's", p)
	}
	if req, _ := r.st.GetRequest("req-1"); req.Stage != claim.StageAccepted {
		t.Errorf("stage = %s, want ACCEPTED (nothing was confirmed)", req.Stage)
	}
}

// Another author with the very same title or reference is still another author:
// only the numeric id decides.
func TestIntakeDuplicates_SameRefDifferentID_IsForeign(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.dups = []intake.Candidate{{Issue: 4242, AuthorID: me + 1, Ref: "issue/4242"}}
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if _, err := r.in.Advance(context.Background(), "req-1"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Fatalf("advance: %v, want ErrNotIssueAuthor", err)
	}
	if p, _ := r.pub.counts(); p != 0 {
		t.Errorf("PublishIssue calls = %d, want 0", p)
	}
}

// An unknown author (id 0) or an unknown authenticated user blocks as
// identity-unknown, retryable once the fact can be established; nothing is created.
func TestIntakeDuplicates_UnknownIdentity_BlocksAsIdentityUnknown(t *testing.T) {
	cases := map[string]func(*fakePub){
		"candidate author unknown": func(p *fakePub) { p.dups = []intake.Candidate{{Issue: 4343, AuthorID: 0}} },
		"user id 0":                func(p *fakePub) { p.user = 0 },
		"user lookup fails":        func(p *fakePub) { p.userErr = errors.New("auth lookup failed") },
	}
	for name, mod := range cases {
		t.Run(name, func(t *testing.T) {
			r := newRig(t, rigCfg)
			mod(r.pub)
			_, _ = r.in.Accept("req-1", []byte("p"), 1)
			if _, err := r.in.Advance(context.Background(), "req-1"); !errors.Is(err, core.ErrIdentityUnknown) {
				t.Fatalf("advance: %v, want ErrIdentityUnknown", err)
			}
			b := blockedOf(t, r, "req-1")
			if b.Reason != core.ReasonIdentityUnknown || !b.Retryable {
				t.Errorf("block = %+v; want identity-unknown and retryable", b)
			}
			if p, _ := r.pub.counts(); p != 0 {
				t.Errorf("PublishIssue calls = %d, want 0", p)
			}
		})
	}
}

// A failed search is a retryable block, not a license to create blindly.
func TestIntakeDuplicates_SearchFails_BlocksRetryable_ThenProceedsWhenItWorks(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.dupErr = errors.New("search unavailable")
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	if _, err := r.in.Advance(context.Background(), "req-1"); err == nil {
		t.Fatal("advance succeeded with the duplicate search down")
	}
	if b := blockedOf(t, r, "req-1"); !b.Retryable {
		t.Errorf("block = %+v; want retryable", b)
	}
	if p, _ := r.pub.counts(); p != 0 {
		t.Fatalf("PublishIssue calls = %d, want 0 while the search is down", p)
	}
	r.pub.dupErr = nil
	if req, err := r.in.Advance(context.Background(), "req-1"); err != nil || req.Stage != claim.StageIssueConfirmed {
		t.Errorf("after the search works: %+v, %v", req, err)
	}
}

// When both exist, my own Issue is reused whatever order the search listed them in.
func TestIntakeDuplicates_OwnAmongForeign_IsReused(t *testing.T) {
	r := newRig(t, rigCfg)
	r.pub.dups = []intake.Candidate{{Issue: 4343, AuthorID: them}, {Issue: 4242, AuthorID: me}}
	_, _ = r.in.Accept("req-1", []byte("p"), 1)
	req, err := r.in.Advance(context.Background(), "req-1")
	if err != nil || req.Issue != 4242 {
		t.Fatalf("advance: %+v, %v; want my #4242 reused", req, err)
	}
}
