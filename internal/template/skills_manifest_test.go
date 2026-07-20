// SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 / Defect B
package template

import (
	"strings"
	"testing"
)

// TestEmbeddedMoaiSkillNames verifies the manifest reader enumerates the
// embedded moai-* skill directories and filters correctly (REQ-DFS-005).
func TestEmbeddedMoaiSkillNames(t *testing.T) {
	names, err := EmbeddedMoaiSkillNames()
	if err != nil {
		t.Fatalf("EmbeddedMoaiSkillNames: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("expected a non-empty embedded moai-* skill set")
	}
	seen := make(map[string]bool, len(names))
	for _, n := range names {
		if !strings.HasPrefix(n, "moai-") {
			t.Errorf("name %q lacks the moai- prefix", n)
		}
		if n == "moai" {
			t.Errorf("bare 'moai' skill must NOT be included (no trailing dash)")
		}
		seen[n] = true
	}
	// Spot-check the 10 skills that were absent from the retired static-22
	// allowlist (issue #1088) — they MUST now be present in the derived set.
	for _, must := range []string{
		"moai-domain-backend", "moai-domain-database", "moai-domain-frontend",
		"moai-domain-html-report", "moai-domain-humanize", "moai-harness-learner",
		"moai-ref-llm-security", "moai-ref-secops", "moai-ref-supply-chain",
		"moai-workflow-ci-loop",
	} {
		if !seen[must] {
			t.Errorf("embedded manifest missing expected core skill %q", must)
		}
	}
}
