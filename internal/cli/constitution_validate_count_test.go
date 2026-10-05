package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// The clean-run line reported "(0 entries checked)" from a literal, so it read
// the same whether one entry or a hundred were checked. Card t988.

// TestRenderValidateTextReportsCheckedCount prints the measured counts.
func TestRenderValidateTextReportsCheckedCount(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	renderValidateText(&buf, constitution.ValidationResult{
		Status:       constitution.ValidateStatusOK,
		CheckedCount: 97,
		TotalCount:   101,
		RetiredCount: 4,
	})

	got := buf.String()
	if !strings.Contains(got, "97 of 101 entries checked") {
		t.Errorf("output does not report the measured counts:\n%s", got)
	}
	if strings.Contains(got, "(0 entries checked)") {
		t.Errorf("output still carries the hardcoded count:\n%s", got)
	}
}

// TestRenderValidateTextReportsZeroForEmptyRegistry is the control: zero is the
// right answer here, and it must come from the result rather than a literal.
// Together with the test above, no constant satisfies both.
func TestRenderValidateTextReportsZeroForEmptyRegistry(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	renderValidateText(&buf, constitution.ValidationResult{
		Status:       constitution.ValidateStatusOK,
		CheckedCount: 0,
		TotalCount:   0,
	})

	got := buf.String()
	if !strings.Contains(got, "0 of 0 entries checked") {
		t.Errorf("empty registry must report zero of zero:\n%s", got)
	}
}
