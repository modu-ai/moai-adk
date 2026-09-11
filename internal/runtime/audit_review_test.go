package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeReviewFixture(t *testing.T, dir, specID, iteration, hash string) string {
	t.Helper()
	path := filepath.Join(dir, specID+"-review-"+iteration+".md")
	content := "Verdict: PASS\nOverall Score: 0.85\nAuditor Version: plan-auditor/v7\nPlan Artifact Hash: " + hash + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestFileAuditCacheIgnoresDateHistory proves that a date change does not
// force a second auditor call when the final review-N snapshot and artifact
// hash are unchanged.
func TestFileAuditCacheIgnoresDateHistory(t *testing.T) {
	projectDir := t.TempDir()
	reportDir := filepath.Join(projectDir, ".moai", "reports", "plan-audit")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specDir := filepath.Join(projectDir, ".moai", "specs", "SPEC-CACHE-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("# stable\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := NewFileAuditCache(projectDir, reportDir)
	hash, err := cache.ComputeHash(specDir)
	if err != nil {
		t.Fatal(err)
	}
	reviewPath := writeReviewFixture(t, reportDir, "SPEC-CACHE-001", "1", hash)
	// This is a history record from a different date and must not be read as
	// the cache source.
	dateHistory := filepath.Join(reportDir, "SPEC-CACHE-001-2099-12-31.md")
	if err := os.WriteFile(dateHistory, []byte("## Audit Run 1\n- verdict: FAIL\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	auditor := &countingReviewAuditor{}
	gate := &GateConfig{
		SpecID:     "SPEC-CACHE-001",
		SpecDir:    specDir,
		ProjectDir: projectDir,
		Auditor:    auditor,
		Cache:      cache,
		Reporter:   &mockReporter{},
		Clock:      FakeClock{FixedTime: time.Date(2099, 12, 31, 12, 0, 0, 0, time.UTC)},
	}
	result, err := gate.Invoke(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Verdict != VerdictPass || !result.CacheHit {
		t.Fatalf("result = %+v, want PASS cache hit", result)
	}
	if auditor.calls != 0 {
		t.Fatalf("auditor calls = %d, want 0 on date-only change", auditor.calls)
	}
	if result.ReportPath != reviewPath {
		t.Fatalf("report path = %q, want final review %q", result.ReportPath, reviewPath)
	}
}

type countingReviewAuditor struct{ calls int }

func (a *countingReviewAuditor) Audit(context.Context, string) (Verdict, string, error) {
	a.calls++
	return VerdictPass, "", nil
}

func TestResolveLatestPlanAuditUsesHighestIterationAndSameSnapshotFields(t *testing.T) {
	dir := t.TempDir()
	writeReviewFixture(t, dir, "SPEC-LATEST-001", "1", "old")
	latest := writeReviewFixture(t, dir, "SPEC-LATEST-001", "2", "new")

	snapshot, err := ResolveLatestPlanAudit(dir, "SPEC-LATEST-001")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot == nil || snapshot.ReportPath != latest || snapshot.PlanArtifactHash != "new" || snapshot.AuditorVersion != "plan-auditor/v7" || !snapshot.ScorePresent {
		t.Fatalf("snapshot = %+v, want latest review metadata", snapshot)
	}
}

func TestResolveLatestPlanAuditRejectsMalformedFinalReview(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SPEC-BAD-001-review-1.md")
	if err := os.WriteFile(path, []byte("Overall Score: nope\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveLatestPlanAudit(dir, "SPEC-BAD-001")
	if err == nil || !strings.Contains(err.Error(), "parse overall score") {
		t.Fatalf("error = %v, want malformed score error", err)
	}
}
