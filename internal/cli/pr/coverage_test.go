package pr_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/ciwatch"
	"github.com/modu-ai/moai-adk/internal/cli/pr"
)

// TestEmitFailureHandoff verifies the JSON handoff format for T3 consumption.
func TestEmitFailureHandoff(t *testing.T) {
	state := ciwatch.CIState{
		PRNumber: 785,
		Branch:   "feat/test",
		RequiredFailed: []ciwatch.CheckResult{
			{Name: "Lint", RunID: "123", LogURL: "https://example.com/run/123"},
		},
		AuxiliaryFailed: []ciwatch.CheckResult{
			{Name: "claude-code-review", Conclusion: "failure"},
		},
	}

	var buf bytes.Buffer
	if err := pr.EmitFailureHandoff(&buf, state); err != nil {
		t.Fatalf("EmitFailureHandoff: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"prNumber":785`) {
		t.Errorf("missing prNumber in handoff: %q", out)
	}
	if !strings.Contains(out, `"Lint"`) {
		t.Errorf("missing Lint in handoff: %q", out)
	}
	if !strings.Contains(out, `"auxiliaryFailCount":1`) {
		t.Errorf("missing auxiliaryFailCount in handoff: %q", out)
	}
}

// TestEmitReadyToMergeReport_ZeroCounts verifies report with no checks.
func TestEmitReadyToMergeReport_ZeroCounts(t *testing.T) {
	state := ciwatch.CIState{PRNumber: 1, Branch: "main"}
	var buf bytes.Buffer
	if err := pr.EmitReadyToMergeReport(&buf, state); err != nil {
		t.Fatalf("EmitReadyToMergeReport: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "(권장)") {
		t.Errorf("missing (권장) in zero-count report: %q", out)
	}
}
