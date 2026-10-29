package runner

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

const ownRepo = "acme/widgets"

// hostWith serves the own Issue (1948, the one the agent works on) and the given
// others; authors maps Issue number to author id (7 is the acting user).
func hostWith(authors map[int]int64) stubFacts {
	out := stubFacts{}
	for n, a := range authors {
		out[n] = IssueFacts{RequestedRepo: ownRepo, ActualRepo: ownRepo, UserID: 7, AuthorID: a, Owner: core.OwnerNone}
	}
	return out
}

func ready(t *testing.T, facts Facts, c core.Collaboration) (*Runner, Ticket) {
	t.Helper()
	r, ref := newRunner(t)
	r.Facts, r.Collab = facts, c
	tk := toVerifying(t, r, ref)
	if rec, err := r.Verify(tk, passAllFor(tk, headA)); err != nil || rec.State != core.StateReady {
		t.Fatalf("setup: %+v %v", rec, err)
	}
	return r, tk
}

func TestLandNeedsAPreflightOnThisHead(t *testing.T) {
	r, tk := ready(t, ownIssues{}, core.Collaboration{})
	if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Fatalf("land without a preflight: %v", err)
	}
	if rec, err := r.PreflightPR(tk, headA, ownRepo, "Refs #1948"); err != nil || string(rec.PreflightHead) != headA {
		t.Fatalf("preflight: %+v %v", rec, err)
	}
	if rec, err := r.Land(tk, headA); err != nil || rec.State != core.StateLanding {
		t.Fatalf("land after the preflight: %+v %v", rec, err)
	}
}

func TestPreflightForAnotherHeadIsNotEnough(t *testing.T) {
	r, tk := ready(t, ownIssues{}, core.Collaboration{})
	if _, err := r.PreflightPR(tk, headB, ownRepo, "Refs #1948"); !errors.Is(err, core.ErrStaleHead) {
		t.Errorf("preflight on a head that is not the candidate: %v", err)
	}
	if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrEvidenceIncomplete) {
		t.Errorf("land: %v", err)
	}
}

func TestPreflightOnlyWhenReady(t *testing.T) {
	r, ref := newRunner(t)
	tk := toVerifying(t, r, ref)
	if _, err := r.PreflightPR(tk, headA, ownRepo, "Refs #1"); !errors.Is(err, core.ErrTransitionInvalid) {
		t.Errorf("preflight while verifying: %v", err)
	}
}

func TestPlainReferencesAndOwnIssuesPass(t *testing.T) {
	host := hostWith(map[int]int64{1948: 7, 2001: 7, 3000: 99})
	r, tk := ready(t, host, core.Collaboration{})
	texts := []string{"Closes #1948", "Fixes #2001, refs #3000", "Related to #3000"}
	if rec, err := r.PreflightPR(tk, headA, ownRepo, texts...); err != nil || rec.PreflightHead == "" {
		t.Fatalf("own issues and plain references: %+v %v", rec, err)
	}
}

// A closing keyword aimed at someone else's Issue stops the PR before it is published.
func TestForeignClosingKeywordBlocksThePR(t *testing.T) {
	host := hostWith(map[int]int64{1948: 7, 3000: 99})
	for name, texts := range map[string][]string{
		"in the PR description": {"Fixes #3000"},
		"in a commit message":   {"Refs #1948", "wip\n\ncloses #3000"},
		"upper case":            {"RESOLVES: #3000"},
	} {
		t.Run(name, func(t *testing.T) {
			r, tk := ready(t, host, core.Collaboration{})
			rec, err := r.PreflightPR(tk, headA, ownRepo, texts...)
			if !errors.Is(err, core.ErrForeignClosing) {
				t.Fatalf("want ErrForeignClosing, got %v", err)
			}
			if rec.State != core.StateReady || rec.Refusal == nil || rec.Refusal.Reason != core.ReasonForeignClosing || rec.PreflightHead != "" || rec.Park != nil {
				t.Fatalf("a refusal leaves the state alone and records itself: %+v", rec)
			}
			if !strings.Contains(rec.Refusal.Detail, "#3000") || !strings.Contains(rec.Refusal.Detail, "Refs") {
				t.Errorf("the refusal must name the Issue and the fix: %q", rec.Refusal.Detail)
			}
			if last := rec.History[len(rec.History)-1]; last.Kind != "preflight-refused" || !strings.Contains(last.Note, "#3000") {
				t.Errorf("history: %+v", last)
			}
			if _, err := r.Land(tk, headA); !errors.Is(err, core.ErrEvidenceIncomplete) {
				t.Errorf("the PR must not be published after a refusal: %v", err)
			}
		})
	}
}

func TestRewordedTextAfterTheBlockLetsThePRThrough(t *testing.T) {
	host := hostWith(map[int]int64{1948: 7, 3000: 99})
	r, tk := ready(t, host, core.Collaboration{})
	if _, err := r.PreflightPR(tk, headA, ownRepo, "Fixes #3000"); err == nil {
		t.Fatal("want a refusal")
	}
	// no resume step any more: the next preflight on reworded text clears the refusal
	rec, err := r.PreflightPR(tk, headA, ownRepo, "Refs #3000")
	if err != nil || rec.Refusal != nil {
		t.Fatalf("reworded: %+v %v", rec, err)
	}
	if rec, err := r.Land(tk, headA); err != nil || rec.State != core.StateLanding {
		t.Errorf("land: %+v %v", rec, err)
	}
}

// What cannot be checked blocks: another repository, an unknown author, no host lookup.
func TestWhatCannotBeCheckedBlocks(t *testing.T) {
	host := hostWith(map[int]int64{1948: 7, 3000: 0})
	cases := map[string]struct {
		facts Facts
		texts []string
	}{
		"another repository":    {host, []string{"Fixes other-org/other#5"}},
		"url to another repo":   {host, []string{"closes https://github.com/other-org/other/issues/5"}},
		"author not known":      {host, []string{"Fixes #3000"}},
		"issue not found":       {host, []string{"Fixes #4040"}},
		"no host lookup at all": {nil, []string{"Fixes #3000"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setup := c.facts
			if setup == nil {
				setup = host // the claim itself needs a host; only the preflight lacks one
			}
			r, tk := ready(t, setup, core.Collaboration{})
			r.Facts = c.facts
			if _, err := r.PreflightPR(tk, headA, ownRepo, c.texts...); !errors.Is(err, core.ErrForeignClosing) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

// Closing another author's Issue by keyword is allowed only where closing it is
// allowed by the permission decision itself: no separate rule exists.
func TestCollaborationClosePermissionLetsAForeignKeywordThrough(t *testing.T) {
	host := hostWith(map[int]int64{1948: 7, 3000: 99})
	c := core.Collaboration{Provider: "github", Enabled: true, AllowedRepositories: []string{ownRepo}}
	r, tk := ready(t, host, c)
	if _, err := r.PreflightPR(tk, headA, ownRepo, "Fixes #3000"); !errors.Is(err, core.ErrForeignClosing) {
		t.Fatalf("the option alone must not allow it: %v", err)
	}
	c.Close = true
	r, tk = ready(t, host, c)
	if _, err := r.PreflightPR(tk, headA, ownRepo, "Fixes #3000"); err != nil {
		t.Errorf("with the close permission: %v", err)
	}
}

func TestPreflightWithoutTextsPasses(t *testing.T) {
	r, tk := ready(t, ownIssues{}, core.Collaboration{})
	if _, err := r.PreflightPR(tk, headA, ownRepo); err != nil {
		t.Errorf("no text, no keyword: %v", err)
	}
}
