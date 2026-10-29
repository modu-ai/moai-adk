package review

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

func mustIssue(t *testing.T, n int) core.IssueRef {
	t.Helper()
	ref, err := core.NewIssueRef(n)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// #1948 DELTA4: a repair is dispatched for the original Issue and branch and
// bound to that Issue's single holder; a result from another Issue, branch,
// or holder is not accepted.
func TestRepairBinding_Accept(t *testing.T) {
	bind := RepairBinding{Issue: mustIssue(t, 1948), Branch: "WT-agentic-factory-review", HolderSession: "lane-19"}
	cases := []struct {
		name string
		res  RepairResult
		want error
	}{
		{"same issue branch holder", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: bind.HolderSession, Head: headB}, nil},
		{"other branch", RepairResult{Issue: bind.Issue, Branch: "WT-other", HolderSession: bind.HolderSession, Head: headB}, ErrRepairBranch},
		{"other holder", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: "lane-20", Head: headB}, core.ErrForeign},
		{"empty holder", RepairResult{Issue: bind.Issue, Branch: bind.Branch, Head: headB}, core.ErrForeign},
		{"other issue", RepairResult{Issue: mustIssue(t, 1949), Branch: bind.Branch, HolderSession: bind.HolderSession, Head: headB}, ErrRepairIssue},
		{"invalid head", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: bind.HolderSession, Head: "abc"}, core.ErrHeadInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := bind.Accept(tc.res)
			if tc.want == nil && err != nil || tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("Accept = %v, want %v", err, tc.want)
			}
		})
	}
}

// The holder is compared by immutable session ID, never by display name.
func TestRepairBinding_HolderIsSessionIDNotDisplayName(t *testing.T) {
	bind := RepairBinding{Issue: mustIssue(t, 1948), Branch: "b", HolderSession: "sess-111", DisplayName: "agent-1"}
	same := RepairResult{Issue: bind.Issue, Branch: "b", HolderSession: "sess-222", DisplayName: "agent-1", Head: headA}
	if err := bind.Accept(same); !errors.Is(err, core.ErrForeign) {
		t.Fatalf("same display name, different session: %v, want ErrForeign", err)
	}
	renamed := RepairResult{Issue: bind.Issue, Branch: "b", HolderSession: "sess-111", DisplayName: "agent-7", Head: headA}
	if err := bind.Accept(renamed); err != nil {
		t.Fatalf("same session, renamed display: %v", err)
	}
}

// A binding with no branch or holder accepts nothing: it cannot be a bypass.
func TestRepairBinding_IncompleteAcceptsNothing(t *testing.T) {
	res := RepairResult{Issue: mustIssue(t, 1948), Branch: "b", HolderSession: "h", Head: headA}
	for _, bind := range []RepairBinding{
		{Issue: res.Issue, Branch: "", HolderSession: "h"},
		{Issue: res.Issue, Branch: "b", HolderSession: ""},
		{Branch: "b", HolderSession: "h"},
	} {
		if err := bind.Accept(res); err == nil {
			t.Errorf("incomplete binding %+v accepted a result", bind)
		}
	}
}
