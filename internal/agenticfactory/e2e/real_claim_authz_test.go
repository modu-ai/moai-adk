package e2e

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/runner"
)

// Runner claim-entry authorization and the PR closing-keyword preflight
// (lane-18 f9f44e5ecc). The host is a fake with per-Issue facts; no real Issue
// is read or written.

func factsFor(owner core.OwnerState, authorID int64) *fakeFacts {
	f := runner.IssueFacts{RequestedRepo: repoA, ActualRepo: repoA, UserID: me, AuthorID: authorID, Owner: owner}
	return &fakeFacts{fallback: &f}
}

// claimWith runs one claim on a fresh store under the given host facts and
// collaboration; it returns the error and whether any record exists afterwards.
func claimWith(t *testing.T, facts runner.Facts, c core.Collaboration) (error, bool) {
	t.Helper()
	st, ref, run := realStore(t)
	r := wired(st)
	r.Facts, r.Collab = facts, c
	_, err := r.Claim(ref, "w1", "d-1", run)
	_, loadErr := st.Load(ref)
	return err, loadErr == nil
}

func TestRunnerClaim_OwnIssue_Admitted(t *testing.T) {
	err, rec := claimWith(t, factsFor(core.OwnerNone, me), core.Collaboration{})
	if err != nil || !rec {
		t.Fatalf("own Issue: err=%v record=%v, want admitted with a record", err, rec)
	}
}

// Another author's Issue is refused by default, and a refusal writes no record.
func TestRunnerClaim_ForeignIssue_DefaultRefused_NoRecord(t *testing.T) {
	err, rec := claimWith(t, factsFor(core.OwnerNone, them), core.Collaboration{Provider: "github"})
	if !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Fatalf("foreign Issue, collaboration off: %v, want ErrNotIssueAuthor", err)
	}
	if rec {
		t.Error("a refused claim left a record behind")
	}
}

func TestRunnerClaim_Collaboration_Cases(t *testing.T) {
	on := func(mod func(*core.Collaboration)) core.Collaboration {
		return collab(func(c *core.Collaboration) {
			c.Enabled, c.AllowedRepositories = true, []string{repoA}
			if mod != nil {
				mod(c)
			}
		})
	}
	cases := []struct {
		name  string
		facts runner.Facts
		c     core.Collaboration
		want  error // nil = admitted
	}{
		{"option and list, unowned", factsFor(core.OwnerNone, them), on(nil), nil},
		{"option and list, handed over", factsFor(core.OwnerHandedOver, them), on(nil), nil},
		{"owner active", factsFor(core.OwnerActive, them), on(nil), core.ErrOwnerUnconfirmed},
		{"owner unknown", factsFor(core.OwnerUnknown, them), on(nil), core.ErrOwnerUnconfirmed},
		{"option without a list", factsFor(core.OwnerNone, them), collab(func(c *core.Collaboration) { c.Enabled = true }), core.ErrRepoNotAllowed},
		{"list without the option", factsFor(core.OwnerNone, them), collab(func(c *core.Collaboration) { c.AllowedRepositories = []string{repoA} }), core.ErrNotIssueAuthor},
		{"repository off the list", factsFor(core.OwnerNone, them), collab(func(c *core.Collaboration) {
			c.Enabled, c.AllowedRepositories = true, []string{repoB}
		}), core.ErrRepoNotAllowed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err, rec := claimWith(t, tc.facts, tc.c)
			if tc.want == nil {
				if err != nil || !rec {
					t.Fatalf("err=%v record=%v, want admitted", err, rec)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
			if rec {
				t.Error("a refused claim left a record behind")
			}
		})
	}
}

// Identity that cannot be established is never guessed: no host lookup, a failed
// lookup, a redirected answer and a zero id all stop the claim, with no record.
func TestRunnerClaim_UnknownIdentity_Refused_NoRecord(t *testing.T) {
	zeroUser := runner.IssueFacts{RequestedRepo: repoA, ActualRepo: repoA, UserID: 0, AuthorID: me, Owner: core.OwnerNone}
	zeroAuthor := runner.IssueFacts{RequestedRepo: repoA, ActualRepo: repoA, UserID: me, AuthorID: 0, Owner: core.OwnerNone}
	for name, facts := range map[string]runner.Facts{
		"no host lookup":     nil,
		"lookup fails":       &fakeFacts{err: errors.New("host unreachable")},
		"user id zero":       &fakeFacts{fallback: &zeroUser},
		"issue author zero":  &fakeFacts{fallback: &zeroAuthor},
		"lookup has no data": &fakeFacts{},
	} {
		t.Run(name, func(t *testing.T) {
			err, rec := claimWith(t, facts, core.Collaboration{})
			if !errors.Is(err, core.ErrIdentityUnknown) {
				t.Errorf("err = %v, want ErrIdentityUnknown", err)
			}
			if rec {
				t.Error("a refused claim left a record behind")
			}
		})
	}
}

// Claiming the same Issue twice never produces a second holder or a new epoch.
func TestRunnerClaim_Duplicate_RefusedHolderAndEpochUnchanged(t *testing.T) {
	st, ref, run := realStore(t)
	r := wired(st)
	first, err := r.Claim(ref, "w1", "d-1", run)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := st.Load(ref)
	if _, err := r.Claim(ref, "w2", "d-2", run); err == nil {
		t.Fatal("a second holder claimed an Issue that is already claimed")
	}
	after, _ := st.Load(ref)
	if after.Holder != "w1" || after.Epoch != first.Epoch || after.Epoch != before.Epoch {
		t.Errorf("record changed by a refused duplicate: holder=%s epoch=%d (was %d)", after.Holder, after.Epoch, before.Epoch)
	}
}

// The repair claim goes through the same permission decision.
func TestRunnerClaimRepair_ForeignIssue_Refused(t *testing.T) {
	st, ref, run := realStore(t)
	r := wired(st)
	r.Facts, r.Collab = factsFor(core.OwnerNone, them), core.Collaboration{Provider: "github"}
	if _, err := r.ClaimRepair(ref, "w1", "d-1", run, "WT-repair"); !errors.Is(err, core.ErrNotIssueAuthor) {
		t.Errorf("repair claim on another author's Issue: %v, want ErrNotIssueAuthor", err)
	}
	if _, err := st.Load(ref); err == nil {
		t.Error("a refused repair claim left a record behind")
	}
}

// --- PR closing-keyword preflight ---

const foreignN = 99

// readyTicket returns a runner and a ticket at READY, with the host knowing my
// own Issue and Issue #99 as another author's.
func readyTicket(t *testing.T) (*runner.Runner, runner.Ticket) {
	t.Helper()
	r, tk := newRunner(t)
	own := ownIssueFacts()
	r.Facts = &fakeFacts{byIssue: map[int]runner.IssueFacts{
		tk.Issue.Number: own,
		foreignN:        foreignIssueFacts(core.OwnerNone),
	}}
	for _, step := range []func() error{
		func() error { _, err := r.StartAttempt(tk, "att-1", core.DeliveryExecuted); return err },
		func() error { _, err := r.Submit(tk, headA, false); return err },
		func() error { _, err := r.Verify(tk, passAll(tk, headA)); return err },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	return r, tk
}

func TestPreflightPR_OwnIssueAndRefs_Pass_ThenLandWorks(t *testing.T) {
	for name, text := range map[string]string{
		"refs only":       "Refs #99",
		"closes my Issue": "Closes #1948",
		"no keyword":      "tidy up the tests",
	} {
		t.Run(name, func(t *testing.T) {
			r, tk := readyTicket(t)
			rec, err := r.PreflightPR(tk, headA, repoA, text)
			if err != nil || string(rec.PreflightHead) != headA {
				t.Fatalf("preflight: err=%v preflight-head=%q", err, rec.PreflightHead)
			}
			if _, err := r.Land(tk, headA); err != nil {
				t.Errorf("land after a passing preflight: %v", err)
			}
		})
	}
}

// A closing keyword aimed at another author's Issue blocks the PR with the
// foreign-closing-keyword reason, and Land stays closed.
func TestPreflightPR_ForeignClosingKeyword_Blocks(t *testing.T) {
	for _, text := range []string{"Closes #99", "fixes #99", "Resolved #99"} {
		t.Run(text, func(t *testing.T) {
			r, tk := readyTicket(t)
			rec, err := r.PreflightPR(tk, headA, repoA, text)
			if !errors.Is(err, core.ErrForeignClosing) {
				t.Fatalf("err = %v, want ErrForeignClosing", err)
			}
			if rec.State != core.StateBlocked || rec.Reason != core.ReasonForeignClosing {
				t.Errorf("record = state %s reason %s, want blocked/foreign-closing-keyword", rec.State, rec.Reason)
			}
			if rec.PreflightHead != "" {
				t.Error("a blocked preflight left a preflight head")
			}
			if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrTransitionInvalid) {
				t.Errorf("land while blocked: %v, want ErrTransitionInvalid", err)
			}
		})
	}
}

// Closing another author's Issue is allowed exactly where collaboration grants
// close on a listed repository: the same decision as any other close.
func TestPreflightPR_ForeignClose_PassesOnlyWithTheCloseKey(t *testing.T) {
	r, tk := readyTicket(t)
	r.Collab = collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories, c.Close = true, []string{repoA}, true
	})
	if _, err := r.PreflightPR(tk, headA, repoA, "Closes #99"); err != nil {
		t.Errorf("close key on a listed repository: %v", err)
	}
	r2, tk2 := readyTicket(t)
	r2.Collab = collab(func(c *core.Collaboration) {
		c.Enabled, c.AllowedRepositories, c.EditBody = true, []string{repoA}, true
	})
	if _, err := r2.PreflightPR(tk2, headA, repoA, "Closes #99"); !errors.Is(err, core.ErrForeignClosing) {
		t.Errorf("other key set, no close key: %v, want ErrForeignClosing", err)
	}
}

// A reference into another repository, or one that cannot be checked, blocks.
func TestPreflightPR_OtherRepoOrUnknown_Blocks(t *testing.T) {
	r, tk := readyTicket(t)
	if _, err := r.PreflightPR(tk, headA, repoA, "Closes Org/Other#5"); !errors.Is(err, core.ErrForeignClosing) {
		t.Errorf("other repository: %v, want ErrForeignClosing", err)
	}
	r2, tk2 := readyTicket(t)
	if _, err := r2.PreflightPR(tk2, headA, repoA, "Closes #12345"); !errors.Is(err, core.ErrForeignClosing) {
		t.Errorf("Issue the host cannot answer for: %v, want ErrForeignClosing", err)
	}
	r3, tk3 := readyTicket(t)
	r3.Facts = nil
	if _, err := r3.PreflightPR(tk3, headA, repoA, "Closes #99"); !errors.Is(err, core.ErrForeignClosing) {
		t.Errorf("no host lookup: %v, want ErrForeignClosing", err)
	}
}

// Commit messages are text too: every text passed is checked.
func TestPreflightPR_EveryTextIsChecked(t *testing.T) {
	r, tk := readyTicket(t)
	_, err := r.PreflightPR(tk, headA, repoA, "Refs #99", "fix(x): tidy\n\nCloses #99")
	if !errors.Is(err, core.ErrForeignClosing) {
		t.Errorf("keyword in a later commit message: %v, want ErrForeignClosing", err)
	}
}

func TestPreflightPR_StaleHead_And_LandWithoutPreflight(t *testing.T) {
	r, tk := readyTicket(t)
	if _, err := r.PreflightPR(tk, headB, repoA, "Refs #1"); !errors.Is(err, core.ErrStaleHead) {
		t.Errorf("preflight on another head: %v, want ErrStaleHead", err)
	}
	if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("land with no preflight: %v, want ErrEvidenceIncomplete", err)
	}
	if _, err := r.PreflightPR(tk, "not-a-sha", repoA, "Refs #1"); err == nil {
		t.Error("a malformed head was accepted")
	}
}

// --- Own Issue: labels and assignees are never opened by a permission key ---

func TestAuthz_OwnIssue_LabelAndAssignee_RefusedEvenWithEveryKey(t *testing.T) {
	for name, c := range map[string]core.Collaboration{
		"defaults": collab(nil),
		"every key": collab(func(c *core.Collaboration) {
			c.Enabled, c.AllowedRepositories = true, []string{repoA}
			c.EditBody, c.Close, c.Labels, c.Assignees = true, true, true, true
		}),
		"label and assignee keys only": collab(func(c *core.Collaboration) { c.Labels, c.Assignees = true, true }),
	} {
		t.Run(name, func(t *testing.T) {
			for _, op := range []core.Op{core.OpLabel, core.OpAssignee} {
				if got := c.Authorize(ownIssue(op)); got != core.DenyPermissionNotGranted {
					t.Errorf("own Issue %s -> %s, want permission-not-granted", op, got)
				}
			}
			// Control: the other own-Issue operations stay open, so this is not "refuse everything".
			for _, op := range []core.Op{core.OpEditBody, core.OpClose, core.OpCommentCreate, core.OpTakeOver} {
				if got := c.Authorize(ownIssue(op)); got != core.Allow {
					t.Errorf("own Issue %s -> %s, want allow", op, got)
				}
			}
		})
	}
}
