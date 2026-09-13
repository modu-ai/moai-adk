package cli

// Card t708 (SPEC-GATEWAY-ENVELOPE-REPAIR-001, AC-EVR-012): the operator
// documentation must exist and carry the section checklist — the repair
// shape, procedure, bounds, refusal conditions, forensics locations, and
// the non-recoverable shapes with their sanctioned paths.

import (
	"os"
	"strings"
	"testing"
)

func TestGatewayEnvelopeRepairOperatorDocExistsWithSections(t *testing.T) {
	raw, err := os.ReadFile("../../.moai/docs/gateway-envelope-repair.md")
	if err != nil {
		t.Fatalf("operator documentation missing: %v", err)
	}
	doc := string(raw)
	for _, section := range []string{
		"CauseReasoning",
		"--repair-envelope",
		"--resume",
		"단일 시도",
		"비파괴",
		"원본 보존", // aside-before-replace
		".moai-repair-aside",
		"repair/<UUID>.json",
		"source-gone",
		"digest",
		"already-attempted",
		"fork",
		"SPEC-GATEWAY-WEDGE-REROOT-001",
		"t707",
	} {
		if !strings.Contains(doc, section) {
			t.Fatalf("operator documentation lost required section/token %q", section)
		}
	}
}
