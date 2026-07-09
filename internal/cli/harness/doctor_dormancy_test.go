// Package harness — doctor dormancy check tests.
// SPEC-HARNESS-RATCHET-REWIRE-001 REQ-HRR-008 / AC-HRR-010: the doctor command
// emits a pipeline-dormancy warning when tier-promotions.jsonl contains ≥1
// promotion AND .moai/harness/proposals/ is absent.
package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedPromotions writes tier-promotions.jsonl under dir/.moai/harness/learning-history/
// with the supplied number of promotion lines.
func seedPromotions(t *testing.T, dir string, n int) {
	t.Helper()
	histDir := filepath.Join(dir, ".moai", "harness", "learning-history")
	if err := os.MkdirAll(histDir, 0o755); err != nil {
		t.Fatalf("mkdir learning-history: %v", err)
	}
	line := `{"ts":"2026-07-09T10:00:00Z","pattern_key":"agent_invocation:Bash:hash1","from_tier":"","to_tier":"rule","observation_count":5,"confidence":1}` + "\n"
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(line)
	}
	if err := os.WriteFile(filepath.Join(histDir, "tier-promotions.jsonl"), []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write promotions: %v", err)
	}
}

// createProposalsDir creates dir/.moai/harness/proposals/ so the dormancy
// check's "proposals dir absent" condition is false.
func createProposalsDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "harness", "proposals"), 0o755); err != nil {
		t.Fatalf("mkdir proposals: %v", err)
	}
}

// hasDormancyFinding reports whether the report carries the pipeline-dormancy
// finding (the REQ-HRR-008 warning).
func hasDormancyFinding(report DoctorReport) bool {
	for _, f := range report.Findings {
		if strings.Contains(strings.ToLower(f.Message), "dormanc") ||
			strings.Contains(strings.ToLower(f.Axis), "pipeline") {
			return true
		}
	}
	return false
}

// TestDoctor_DormancyWarning_PromotionsNoProposals covers AC-HRR-010 (positive
// case): promotions ≥1 AND proposals dir absent → dormancy warning emitted.
func TestDoctor_DormancyWarning_PromotionsNoProposals(t *testing.T) {
	dir := t.TempDir()
	seedPromotions(t, dir, 3)
	// proposals dir deliberately absent.

	report, err := Doctor(dir)
	if err != nil {
		// Doctor returns an error only when ERROR findings exist; dormancy is
		// INFO so err should be nil here.
		t.Fatalf("Doctor error: %v", err)
	}
	if !hasDormancyFinding(report) {
		t.Errorf("expected dormancy warning (promotions exist + proposals dir absent); findings=%v", report.Findings)
	}
}

// TestDoctor_NoDormancyWarning_ProposalsPresent covers AC-HRR-010 (negative
// case): when the proposals dir IS present, the dormancy warning is NOT emitted.
func TestDoctor_NoDormancyWarning_ProposalsPresent(t *testing.T) {
	dir := t.TempDir()
	seedPromotions(t, dir, 3)
	createProposalsDir(t, dir)

	report, err := Doctor(dir)
	if err != nil {
		t.Fatalf("Doctor error: %v", err)
	}
	if hasDormancyFinding(report) {
		t.Errorf("dormancy warning emitted even though proposals dir exists; findings=%v", report.Findings)
	}
}

// TestDoctor_NoDormancyWarning_NoPromotions covers AC-HRR-010 (no-promotions
// case): when tier-promotions.jsonl is absent/empty, no dormancy warning even
// if proposals dir is absent (nothing is starved).
func TestDoctor_NoDormancyWarning_NoPromotions(t *testing.T) {
	dir := t.TempDir()
	// No promotions seeded; no proposals dir.

	report, err := Doctor(dir)
	if err != nil {
		t.Fatalf("Doctor error: %v", err)
	}
	if hasDormancyFinding(report) {
		t.Errorf("dormancy warning emitted with zero promotions; findings=%v", report.Findings)
	}
}
