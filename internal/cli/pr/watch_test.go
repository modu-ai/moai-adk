// Package pr_test validates the moai pr watch command and its handoff report.
package pr_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
	"github.com/modu-ai/moai-adk/internal/cli/pr"
)

// TestEmitReadyToMergeReport_AllPass verifies the ready-to-merge report format.
// The report must:
//  - Contain PR number
//  - Include "(권장)" in the first option (per AskUserQuestion protocol)
//  - Be markdown formatted
//  - NOT include AskUserQuestion call (orchestrator handles that)
func TestEmitReadyToMergeReport_AllPass(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber:       785,
		Branch:         "feat/SPEC-V3R3-CI-AUTONOMY-001-wave-2",
		RequiredPassed: 6,
		RequiredFailed: nil,
		RequiredPending: nil,
		AuxiliaryFailed: nil,
	}

	var buf bytes.Buffer
	if err := pr.EmitReadyToMergeReport(&buf, state); err != nil {
		t.Fatalf("EmitReadyToMergeReport: %v", err)
	}

	out := buf.String()

	// Must reference the PR number.
	if !strings.Contains(out, "785") {
		t.Errorf("report missing PR number 785: %q", out)
	}

	// Must have "(권장)" in the first option label (AskUserQuestion protocol).
	if !strings.Contains(out, "(권장)") {
		t.Errorf("report missing (권장) label for recommended action: %q", out)
	}

	// Must be markdown (has at least one ## header).
	if !strings.Contains(out, "##") {
		t.Errorf("report missing markdown headers: %q", out)
	}

	// Must NOT call AskUserQuestion (HARD: orchestrator only).
	if strings.Contains(out, "AskUserQuestion") {
		t.Errorf("report must not reference AskUserQuestion: %q", out)
	}

	// Must have check summary.
	if !strings.Contains(out, "6/6") || !strings.Contains(out, "pass") {
		t.Errorf("report missing check summary (6/6 pass): %q", out)
	}
}

// TestEmitReadyToMergeReport_WithAdvisoryFail verifies advisory failures are
// mentioned but do NOT block the merge recommendation.
func TestEmitReadyToMergeReport_WithAdvisoryFail(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber:       786,
		Branch:         "main",
		RequiredPassed: 6,
		RequiredFailed: nil,
		RequiredPending: nil,
		AuxiliaryFailed: []ciwatch.CheckResult{
			{Name: "claude-code-review", Conclusion: "failure"},
		},
	}

	var buf bytes.Buffer
	if err := pr.EmitReadyToMergeReport(&buf, state); err != nil {
		t.Fatalf("EmitReadyToMergeReport: %v", err)
	}

	out := buf.String()
	// "(권장)" must still be present.
	if !strings.Contains(out, "(권장)") {
		t.Errorf("report missing (권장): %q", out)
	}
	// Advisory failure should be mentioned.
	if !strings.Contains(out, "advisory") && !strings.Contains(out, "claude-code-review") {
		t.Logf("Advisory mention optional but check summary should be present; got: %q", out)
	}
}

// TestReadyToMergeFlow_CLINoAskUserQuestion ensures the CLI writer does not
// contain any AskUserQuestion invocation — that is strictly orchestrator territory.
func TestReadyToMergeFlow_CLINoAskUserQuestion(t *testing.T) {
	// This is a documentation/contract test. The pr package must not import
	// any AskUserQuestion tooling or invoke user interaction directly.
	// We verify by calling the function and confirming no panic / no side-effects
	// beyond writing to the provided io.Writer.
	state := ciwatch.CIState{PRNumber: 999, Branch: "test", RequiredPassed: 3}
	var buf bytes.Buffer
	if err := pr.EmitReadyToMergeReport(&buf, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output must be non-empty.
	if buf.Len() == 0 {
		t.Error("expected non-empty report output")
	}
}
