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
	bind := RepairBinding{Issue: mustIssue(t, 1948), Branch: "WT-agentic-factory-review", HolderSession: "lane-19", Epoch: 2}
	cases := []struct {
		name string
		res  RepairResult
		want error
	}{
		{"same issue branch holder", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: bind.HolderSession, Epoch: 2, Head: headB}, nil},
		{"other branch", RepairResult{Issue: bind.Issue, Branch: "WT-other", HolderSession: bind.HolderSession, Epoch: 2, Head: headB}, ErrRepairBranch},
		{"other holder", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: "lane-20", Epoch: 2, Head: headB}, core.ErrForeign},
		{"empty holder", RepairResult{Issue: bind.Issue, Branch: bind.Branch, Epoch: 2, Head: headB}, core.ErrForeign},
		{"other issue", RepairResult{Issue: mustIssue(t, 1949), Branch: bind.Branch, HolderSession: bind.HolderSession, Epoch: 2, Head: headB}, ErrRepairIssue},
		{"previous writer epoch", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: bind.HolderSession, Epoch: 1, Head: headB}, ErrRepairEpoch},
		{"invalid head", RepairResult{Issue: bind.Issue, Branch: bind.Branch, HolderSession: bind.HolderSession, Epoch: 2, Head: "abc"}, core.ErrHeadInvalid},
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
	bind := RepairBinding{Issue: mustIssue(t, 1948), Branch: "b", HolderSession: "sess-111", DisplayName: "agent-1", Epoch: 1}
	same := RepairResult{Issue: bind.Issue, Branch: "b", HolderSession: "sess-222", DisplayName: "agent-1", Epoch: 1, Head: headA}
	if err := bind.Accept(same); !errors.Is(err, core.ErrForeign) {
		t.Fatalf("same display name, different session: %v, want ErrForeign", err)
	}
	renamed := RepairResult{Issue: bind.Issue, Branch: "b", HolderSession: "sess-111", DisplayName: "agent-7", Epoch: 1, Head: headA}
	if err := bind.Accept(renamed); err != nil {
		t.Fatalf("same session, renamed display: %v", err)
	}
}

// A binding with no branch or holder accepts nothing: it cannot be a bypass.
func TestRepairBinding_IncompleteAcceptsNothing(t *testing.T) {
	res := RepairResult{Issue: mustIssue(t, 1948), Branch: "b", HolderSession: "h", Epoch: 1, Head: headA}
	for _, bind := range []RepairBinding{
		{Issue: res.Issue, Branch: "", HolderSession: "h", Epoch: 1},
		{Issue: res.Issue, Branch: "b", HolderSession: "", Epoch: 1},
		{Issue: res.Issue, Branch: "b", HolderSession: "h"},
		{Branch: "b", HolderSession: "h", Epoch: 1},
	} {
		if err := bind.Accept(res); err == nil {
			t.Errorf("incomplete binding %+v accepted a result", bind)
		}
	}
}

// S46 E06: the repair is re-reserved on the original Issue/branch and given to
// an available agent; an agent already holding another Issue is never
// assigned, so its current work is not overwritten or made a second write.
func TestAssignRepair_AvailableAgentOnly(t *testing.T) {
	bind := RepairBinding{Issue: mustIssue(t, 1948), Branch: "b", HolderSession: "sess-new", Epoch: 2}
	other := mustIssue(t, 1950)
	holding := func(m map[string]core.IssueRef) func(string) (core.IssueRef, bool) {
		return func(s string) (core.IssueRef, bool) { r, ok := m[s]; return r, ok }
	}
	cases := []struct {
		name string
		held map[string]core.IssueRef
		want error
	}{
		{"idle agent", nil, nil},
		{"agent holds the same issue", map[string]core.IssueRef{"sess-new": bind.Issue}, nil},
		{"agent busy on another issue", map[string]core.IssueRef{"sess-new": other}, ErrAgentBusy},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := bind.Assign(holding(tc.held))
			if tc.want == nil && err != nil || tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("Assign = %v, want %v", err, tc.want)
			}
		})
	}
	if err := (RepairBinding{}).Assign(holding(nil)); err == nil {
		t.Error("incomplete binding must not assign")
	}
}
