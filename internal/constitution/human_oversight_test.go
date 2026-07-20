package constitution

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func TestHumanOversight_Approve_DryRunAutoApproves(t *testing.T) {
	h := NewHumanOversight()
	ok, err := h.Approve(&AmendmentProposal{RuleID: "R", Before: "a", After: "b"}, true)
	if err != nil {
		t.Fatalf("Approve(dryRun) err = %v", err)
	}
	if !ok {
		t.Error("Approve(dryRun) should auto-approve")
	}
}

// newTestOversight constructs a humanOversight with an injected reader (for Y/N
// input) and a discard writer. printDiff writes to os.Stdout via fmt.Printf
// (not the injected writer), so the diff output appears in verbose test logs
// but does not affect assertions. All state stays in-process (no fs mutation).
func newTestOversight(input string) *humanOversight {
	return &humanOversight{
		reader: bufio.NewReader(strings.NewReader(input)),
		writer: bufio.NewWriter(io.Discard),
	}
}

func TestHumanOversight_Approve_YesApproves(t *testing.T) {
	h := newTestOversight("Y\n")
	ok, err := h.Approve(&AmendmentProposal{
		RuleID:   "R",
		Before:   "a",
		After:    "b",
		Evidence: "demotion evidence",
	}, false)
	if err != nil {
		t.Fatalf("Approve err = %v", err)
	}
	if !ok {
		t.Error("Y input should approve")
	}
}

func TestHumanOversight_Approve_NoRejects(t *testing.T) {
	h := newTestOversight("N\n")
	ok, err := h.Approve(&AmendmentProposal{RuleID: "R", Before: "a", After: "b"}, false)
	if err != nil {
		t.Fatalf("Approve err = %v", err)
	}
	if ok {
		t.Error("N input should reject")
	}
}

func TestHumanOversight_Approve_InvalidRetriesThenYes(t *testing.T) {
	h := newTestOversight("maybe\nYES\n")
	ok, err := h.Approve(&AmendmentProposal{RuleID: "R", Before: "a", After: "b"}, false)
	if err != nil {
		t.Fatalf("Approve err = %v", err)
	}
	if !ok {
		t.Error("invalid then YES should approve after retry")
	}
}

// TestHumanOversight_printDiff_Branches exercises every printDiff branch by
// feeding Approve proposals with varied CanaryResult / Contradicts / Evidence
// states. printDiff is called inside the non-dry-run Approve path; covering it
// here lifts human_oversight.go::printDiff from 0%.
func TestHumanOversight_printDiff_Branches(t *testing.T) {
	proposals := []*AmendmentProposal{
		{
			RuleID: "R1",
			Before: "a",
			After:  "b",
			CanaryResult: &CanaryResult{
				Available:      true,
				Passed:         true,
				ScoreBefore:    0.90,
				ScoreAfter:     0.85,
				MaxDrop:        0.05,
				EvaluatedSpecs: []string{"SPEC-A", "SPEC-B"},
			},
		},
		{
			RuleID:       "R2",
			Before:       "a",
			After:        "b",
			CanaryResult: &CanaryResult{Available: false, Reason: "insufficient SPECs"},
		},
		{
			RuleID: "R3",
			Before: "a",
			After:  "b",
			Contradicts: &ContradictionResult{
				Conflicts: []ConflictDetail{
					{ConflictingRuleID: "X", Description: "blocks", IsBlocking: true},
					{ConflictingRuleID: "Y", Description: "warns"},
				},
			},
		},
		{RuleID: "R4", Before: "a", After: "b"}, // no canary, no contradictions
	}
	for i, proposal := range proposals {
		h := newTestOversight("Y\n")
		if _, err := h.Approve(proposal, false); err != nil {
			t.Errorf("printDiff case %d: Approve err = %v", i, err)
		}
	}
}
