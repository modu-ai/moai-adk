package constitution

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// humanOversight는 HumanOversight interface의 구현이다.
// CLI에서 terminal Y/N prompt로 승인을 받는다.
type humanOversight struct {
	// reader는 표준 입력 reader이다.
	reader *bufio.Reader
	// writer는 표준 출력 writer이다.
	writer *bufio.Writer
}

// NewHumanOversight는 HumanOversight를 생성한다.
func NewHumanOversight() HumanOversight {
	return &humanOversight{
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

// Approve는 사용자에게 proposal diff를 보여고 승인을 요청한다.
// Dry-run mode에서는 항상 true 반환 (실제 승인 없음).
// SPEC-V3R2-CON-002 REQ-CON-002-009 Layer 5 구현.
func (h *humanOversight) Approve(proposal *AmendmentProposal, dryRun bool) (bool, error) {
	// Dry-run: 자동 승인
	if dryRun {
		return true, nil
	}

	// Proposal diff 출력
	if err := h.printDiff(proposal); err != nil {
		return false, err
	}

	// Y/N prompt
	for {
		if err := h.writer.Flush(); err != nil {
			return false, fmt.Errorf("출력 버퍼 플러시 오류: %w", err)
		}
		fmt.Print("\n승인하시겠습니까? (Y/N): ")
		response, err := h.reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("입력 읽기 오류: %w", err)
		}

		response = strings.TrimSpace(strings.ToUpper(response))
		switch response {
		case "Y", "YES":
			return true, nil
		case "N", "NO":
			return false, nil
		}

		fmt.Println("Y 또는 N을 입력하세요.")
	}
}

// printDiff는 proposal의 변경 사항을 출력한다.
func (h *humanOversight) printDiff(proposal *AmendmentProposal) error {
	fmt.Printf("\n=== Constitutional Amendment Proposal ===\n")
	fmt.Printf("Rule ID: %s\n", proposal.RuleID)
	fmt.Printf("\n--- Before ---\n%s\n", proposal.Before)
	fmt.Printf("\n+++ After +++\n%s\n", proposal.After)

	if proposal.Evidence != "" {
		fmt.Printf("\nEvidence: %s\n", proposal.Evidence)
	}

	// Canary 결과
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

	// 모순 탐지 결과
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

// humanOversight는 HumanOversight interface를 만족한다.
var _ HumanOversight = (*humanOversight)(nil)
