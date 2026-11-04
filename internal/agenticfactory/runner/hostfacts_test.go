package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/claim"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func hostWithAuthor(author int64) HostFacts {
	return HostFacts{
		Repo:    "acme/widgets",
		Timeout: time.Second,
		User:    func(context.Context) (int64, error) { return 7, nil },
		Issue: func(_ context.Context, repo string, issue int) (HostIssue, error) {
			return HostIssue{AuthorID: author, ActualRepo: repo}, nil
		},
		Owner: func(core.IssueRef) (core.OwnerState, error) { return core.OwnerNone, nil },
	}
}

// The adapter turns what a host identity lookup says into the facts the
// permission decision takes: the repository asked about, the one answered for, the
// acting user, the Issue's author and who owns the work.
func TestHostFactsAnswersTheFactsTheGuardNeeds(t *testing.T) {
	ref, _ := core.NewIssueRef(1948)
	h := hostWithAuthor(99)
	h.Issue = func(_ context.Context, repo string, issue int) (HostIssue, error) {
		if repo != "acme/widgets" || issue != 1948 {
			t.Errorf("the lookup asked for %s#%d", repo, issue)
		}
		return HostIssue{AuthorID: 99, ActualRepo: "acme/gadgets", Redirected: true}, nil
	}
	got, err := h.IssueFacts(context.Background(), ref)
	if err != nil {
		t.Fatal(err)
	}
	want := IssueFacts{RequestedRepo: "acme/widgets", ActualRepo: "acme/gadgets", Redirected: true, UserID: 7, AuthorID: 99, Owner: core.OwnerNone}
	if got != want {
		t.Errorf("facts = %+v, want %+v", got, want)
	}
}

// Every shortfall is an error, which the guard reads as an unknown identity.
func TestHostFactsFailsClosed(t *testing.T) {
	ref, _ := core.NewIssueRef(1948)
	boom := errors.New("host down")
	cases := map[string]func(*HostFacts){
		"no user lookup":    func(h *HostFacts) { h.User = nil },
		"no issue lookup":   func(h *HostFacts) { h.Issue = nil },
		"no owner lookup":   func(h *HostFacts) { h.Owner = nil },
		"no repository":     func(h *HostFacts) { h.Repo = "" },
		"no time limit":     func(h *HostFacts) { h.Timeout = 0 },
		"user lookup fails": func(h *HostFacts) { h.User = func(context.Context) (int64, error) { return 0, boom } },
		"issue lookup fails": func(h *HostFacts) {
			h.Issue = func(context.Context, string, int) (HostIssue, error) { return HostIssue{}, boom }
		},
		"owner lookup fails":  func(h *HostFacts) { h.Owner = func(core.IssueRef) (core.OwnerState, error) { return "", boom } },
		"user id not numeric": func(h *HostFacts) { h.User = func(context.Context) (int64, error) { return 0, nil } },
		"author id missing": func(h *HostFacts) {
			h.Issue = func(context.Context, string, int) (HostIssue, error) {
				return HostIssue{ActualRepo: "acme/widgets"}, nil
			}
		},
	}
	for name, mut := range cases {
		h := hostWithAuthor(99)
		mut(&h)
		if _, err := h.IssueFacts(context.Background(), ref); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// A lookup that never answers is cut off by the time limit.
func TestHostFactsLookupsAreTimeLimited(t *testing.T) {
	ref, _ := core.NewIssueRef(1948)
	h := hostWithAuthor(99)
	h.Timeout = 30 * time.Millisecond
	h.Issue = func(ctx context.Context, _ string, _ int) (HostIssue, error) {
		<-ctx.Done()
		return HostIssue{}, ctx.Err()
	}
	if _, err := h.IssueFacts(context.Background(), ref); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("a lookup that never answers: %v", err)
	}
}

// Through the Runner: the author may claim their own Issue, another author's is
// refused, and a failing host is an unknown identity.
func TestHostFactsDrivesTheRunnerGuard(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 10)
	r := New(st)
	r.Collab = core.Collaboration{}
	ref, _ := core.NewIssueRef(1948)
	r.Facts = hostWithAuthor(7)
	if _, err := r.Claim(ref, "sess-a", "d1", "run-1"); err != nil {
		t.Fatalf("own Issue: %v", err)
	}
	other, _ := core.NewIssueRef(2000)
	r.Facts = hostWithAuthor(99)
	if _, err := r.Claim(other, "sess-a", "d2", "run-1"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Errorf("another author's Issue: %v", err)
	}
	h := hostWithAuthor(7)
	h.User = func(context.Context) (int64, error) { return 0, errors.New("down") }
	r.Facts = h
	third, _ := core.NewIssueRef(3000)
	if _, err := r.Claim(third, "sess-a", "d3", "run-1"); !errors.Is(err, core.ErrIdentityUnknown) {
		t.Errorf("a failing host: %v", err)
	}
}

// StoreOwner reads who owns the work from the claim record.
func TestStoreOwnerReadsTheClaimRecord(t *testing.T) {
	st, _ := claim.OpenWithLimit(t.TempDir(), 10)
	owner := StoreOwner(st)
	ref, _ := core.NewIssueRef(1948)
	if got, err := owner(ref); err != nil || got != core.OwnerNone {
		t.Errorf("no record: %q %v", got, err)
	}
	rec, _ := st.Acquire(ref, "sess-a", "d1", "run-1")
	if got, err := owner(ref); err != nil || got != core.OwnerActive {
		t.Errorf("held: %q %v", got, err)
	}
	if _, err := st.UpdateRelease(ref, "sess-a", rec.Epoch, func(r *claim.Record) error {
		r.State = core.StateLanding
		r.Receipt = &claim.Receipt{PR: 1, Head: core.HeadSHA("a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"), Branch: "WT-x", Worktree: "/wt"}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got, err := owner(ref); err != nil || got != core.OwnerHandedOver {
		t.Errorf("released with a receipt: %q %v", got, err)
	}
}
