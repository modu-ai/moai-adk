package constitution

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// humanOversight is an implementation of the HumanOversight interface.
// It receives approval through terminal Y/N prompts in the CLI.
type humanOversight struct {
	// reader is the standard input reader.
	reader *bufio.Reader
	// writer is the standard output writer.
	writer *bufio.Writer
}

// NewHumanOversight creates a HumanOversight.
func NewHumanOversight() HumanOversight {
	return &humanOversight{
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

// Approve shows the proposal diff to the user and requests approval.
// In dry-run mode, always returns true (no actual approval).
// SPEC-V3R2-CON-002 REQ-CON-002-009 Layer 5 implementation.
func (h *humanOversight) Approve(proposal *AmendmentProposal, dryRun bool) (bool, error) {
	// Dry-run: auto-approve
	if dryRun {
		return true, nil
	}

	// Print proposal diff
	if err := h.printDiff(proposal); err != nil {
		return false, err
	}

	// Y/N prompt
	for {
		_ = h.writer.Flush()
		fmt.Print("\n승인하시겠습니까? (Y/N): ")
		response, err := h.reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("입력 읽기 오류: %w", err)
		}

		switch response = strings.TrimSpace(strings.ToUpper(response)); response {
		case "Y", "YES":
			return true, nil
		case "N", "NO":
			return false, nil
		}

		fmt.Println("Y 또는 N을 입력하세요.")
	}
}

// printDiff prints the changes of the proposal.
func (h *humanOversight) printDiff(proposal *AmendmentProposal) error {
	fmt.Printf("\n=== Constitutional Amendment Proposal ===\n")
	fmt.Printf("Rule ID: %s\n", proposal.RuleID)
	fmt.Printf("\n--- Before ---\n%s\n", proposal.Before)
	fmt.Printf("\n+++ After +++\n%s\n", proposal.After)

	if proposal.Evidence != "" {
		fmt.Printf("\nEvidence: %s\n", proposal.Evidence)
	}

	// Canary results
	if proposal.CanaryResult != nil {
		fmt.Printf("\n--- Canary Evaluation ---\n")
		if proposal.CanaryResult.Available {
			fmt.Printf("Status: %s\n", map[bool]string{true: "PASSED", false: "FAILED"}[proposal.CanaryResult.Passed])
			fmt.Printf("Score Before: %.2f\n", proposal.CanaryResult.ScoreBefore)
			fmt.Printf("Score After: %.2f\n", proposal.CanaryResult.ScoreAfter)
			fmt.Printf("Score Drop: %.2f\n", proposal.CanaryResult.MaxDrop)
			fmt.Printf("Evaluated Specs: %v\n", proposal.CanaryResult.EvaluatedSpecs)
		} else {
			fmt.Printf("Status: SKIPPED (%s)\n", proposal.CanaryResult.Reason)
		}
	}

	// Contradiction detection results
	if proposal.Contradicts != nil && len(proposal.Contradicts.Conflicts) > 0 {
		fmt.Printf("\n--- Contradiction Detection ---\n")
		for _, conflict := range proposal.Contradicts.Conflicts {
			prefix := "WARNING"
			if conflict.IsBlocking {
				prefix = "BLOCKING"
			}
			fmt.Printf("[%s] %s: %s\n", prefix, conflict.ConflictingRuleID, conflict.Description)
		}
	}

	return nil
}

// humanOversight satisfies the HumanOversight interface.
var _ HumanOversight = (*humanOversight)(nil)
