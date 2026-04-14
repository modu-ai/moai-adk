package ralph

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/loop"
)

// TestClassifyFeedback_SourceAware_CompilerError tests REQ-LL-005 rule 1:
// Severity=Error + Source=compiler → ErrorLevelBlocker.
func TestClassifyFeedback_SourceAware_CompilerError(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityError, Source: "compiler", Message: "undeclared name: foo"},
		},
	}

	classified := ClassifyFeedback(fb)

	found := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelBlocker {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("compiler error must produce ErrorLevelBlocker, classified = %v", classified)
	}
}

// TestClassifyFeedback_SourceAware_StaticcheckError tests REQ-LL-005 rule 2:
// Severity=Error + Source=staticcheck SA* → ErrorLevelApproval.
func TestClassifyFeedback_SourceAware_StaticcheckSAError(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityError, Source: "staticcheck", Code: "SA1001", Message: "invalid format"},
		},
	}

	classified := ClassifyFeedback(fb)

	found := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelApproval {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("staticcheck SA* error must produce ErrorLevelApproval, classified = %v", classified)
	}
}

// TestClassifyFeedback_SourceAware_StaticcheckWarning tests REQ-LL-005 rule 3:
// Severity=Warning + Source=staticcheck → ErrorLevelAutoFix.
func TestClassifyFeedback_SourceAware_StaticcheckWarning(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityWarning, Source: "staticcheck", Message: "deprecated API"},
		},
	}

	classified := ClassifyFeedback(fb)

	found := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelAutoFix && strings.Contains(ce.Description, "staticcheck") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("staticcheck warning must produce ErrorLevelAutoFix, classified = %v", classified)
	}
}

// TestClassifyFeedback_SourceAware_InformationSkip tests REQ-LL-005 rule 4:
// Severity=Information → ErrorLevelSkip (not added to classified list).
func TestClassifyFeedback_SourceAware_InformationSkip(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityInfo, Source: "gopls", Message: "info hint"},
		},
	}

	classified := ClassifyFeedback(fb)

	// Information-only diagnostics must not produce any classified error.
	for _, ce := range classified {
		if strings.Contains(ce.Description, "information") || strings.Contains(ce.Description, "hint") {
			t.Errorf("Information/Hint diagnostics must be skipped, got: %v", ce)
		}
	}
	if len(classified) != 0 {
		t.Errorf("Information-only diagnostics must produce no classified errors, got %v", classified)
	}
}

// TestClassifyFeedback_SourceAware_HintSkip tests REQ-LL-005 rule 5:
// Severity=Hint → ErrorLevelSkip.
func TestClassifyFeedback_SourceAware_HintSkip(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityHint, Source: "gopls", Message: "rename suggestion"},
		},
	}

	classified := ClassifyFeedback(fb)

	if len(classified) != 0 {
		t.Errorf("Hint-only diagnostics must produce no classified errors, got %v", classified)
	}
}

// TestClassifyFeedbackWithConfig_LintAsInstruction tests REQ-LL-006:
// when LintAsInstruction=true, warning-severity diagnostics are classified
// as instruction-level (not gate block).
func TestClassifyFeedbackWithConfig_LintAsInstruction(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityWarning, Source: "gopls", Message: "unused var"},
		},
	}

	cfg := config.RalphConfig{LintAsInstruction: true}
	classified := ClassifyFeedbackWithConfig(fb, cfg)

	// With LintAsInstruction=true, warnings should be instruction-level
	// (ErrorLevelAutoFix or no block) rather than gate-blocking.
	for _, ce := range classified {
		if ce.Level >= ErrorLevelApproval {
			t.Errorf("LintAsInstruction=true: warning must not reach Approval level, got level=%d", ce.Level)
		}
	}
}

// TestClassifyFeedbackWithConfig_WarnAsInstruction tests REQ-LL-007:
// when WarnAsInstruction=true, warning diagnostics are injected as instruction,
// not as a gate block.
func TestClassifyFeedbackWithConfig_WarnAsInstruction(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		BuildSuccess: true,
		LSPDiagnostics: []lsp.Diagnostic{
			{Severity: lsp.SeverityWarning, Source: "staticcheck", Message: "use of deprecated func"},
		},
	}

	cfg := config.RalphConfig{WarnAsInstruction: true}
	classified := ClassifyFeedbackWithConfig(fb, cfg)

	// WarnAsInstruction=true: warning diagnostics must not gate-block.
	for _, ce := range classified {
		if ce.Level >= ErrorLevelApproval {
			t.Errorf("WarnAsInstruction=true: warning must not reach Approval level, got level=%d", ce.Level)
		}
	}
}

// TestClassifyFeedback_BackwardsCompat_IntegerMetrics tests REQ-LL-008:
// when LSPDiagnostics is empty, existing integer-based classification must work.
func TestClassifyFeedback_BackwardsCompat_IntegerMetrics(t *testing.T) {
	t.Parallel()

	fb := &loop.Feedback{
		TestsFailed:    3,
		LintErrors:     2,
		BuildSuccess:   true,
		LSPDiagnostics: nil, // empty — backwards compatibility path
	}

	classified := ClassifyFeedback(fb)

	if len(classified) == 0 {
		t.Fatal("integer-based classification must still work when LSPDiagnostics is empty")
	}

	foundLint := false
	foundTest := false
	for _, ce := range classified {
		if ce.Level == ErrorLevelAutoFix && strings.Contains(ce.Description, "lint") {
			foundLint = true
		}
		if strings.Contains(ce.Description, "test") {
			foundTest = true
		}
	}
	if !foundLint {
		t.Errorf("lint errors must produce AutoFix classification, classified = %v", classified)
	}
	if !foundTest {
		t.Errorf("test failures must produce classification, classified = %v", classified)
	}
}

// TestErrorLevelBlocker_HigherThanManual verifies ErrorLevelBlocker is the highest level.
func TestErrorLevelBlocker_HigherThanManual(t *testing.T) {
	t.Parallel()

	if ErrorLevelBlocker <= ErrorLevelManual {
		t.Errorf("ErrorLevelBlocker (%d) must be higher than ErrorLevelManual (%d)",
			ErrorLevelBlocker, ErrorLevelManual)
	}
}

// TestErrorLevelSkip_LowestLevel verifies ErrorLevelSkip is the skip sentinel.
func TestErrorLevelSkip_LowestLevel(t *testing.T) {
	t.Parallel()

	if ErrorLevelSkip >= ErrorLevelAutoFix {
		t.Errorf("ErrorLevelSkip (%d) must be lower than ErrorLevelAutoFix (%d)",
			ErrorLevelSkip, ErrorLevelAutoFix)
	}
}
