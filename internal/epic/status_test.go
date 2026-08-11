package epic

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/spec"
)

// TestBuildEpicStatus_BASFixture is the characterization test against the BAS
// epic fixture: 3 covered (M0/M1/M4) + 3 orphan (M2/M3/M5) + 0 untracked.
// AC-ES-005: orphan_mx = ["M2","M3","M5"].
func TestBuildEpicStatus_BASFixture(t *testing.T) {
	root := t.TempDir()
	// Three covered SPECs (status varies to exercise the classifier).
	writeSpecFixture(t, root, "SPEC-NAVIGATOR-SYNC-001", "Navigator Sync (BAS M0) — graph layer", "completed")
	writeSpecFixture(t, root, "SPEC-NAVIGATOR-SYNC-002", "Navigator Sync (BAS M4) — 4-tier map", "completed")
	writeSpecFixture(t, root, "SPEC-NAVIGATOR-SYNC-003", "Navigator Sync (BAS M1) — Detect hook", "in-progress")
	// A design report so orphan detection fires.
	html := designReportFixtureHTML()
	writeRaw(t, root, ".moai/reports/navigator-redesign-bas-20260805.html", html)

	// Add a non-empty sync_commit_sha to progress.md of the completed SPECs so
	// the done classifier (completed + non-empty sha) triggers.
	writeRaw(t, root, ".moai/specs/SPEC-NAVIGATOR-SYNC-001/progress.md",
		"## §E.4 Sync-phase Audit-Ready Signal\nsync_commit_sha: \"abc123\"\n")
	writeRaw(t, root, ".moai/specs/SPEC-NAVIGATOR-SYNC-002/progress.md",
		"## §E.4 Sync-phase Audit-Ready Signal\nsync_commit_sha: \"def456\"\n")

	got, err := BuildEpicStatus("NAVIGATOR-SYNC", Options{BaseDir: root})
	if err != nil {
		t.Fatalf("BuildEpicStatus: %v", err)
	}
	if got.Epic != "NAVIGATOR-SYNC" {
		t.Errorf("Epic = %q, want NAVIGATOR-SYNC", got.Epic)
	}
	if got.EpicToken != "BAS" {
		t.Errorf("EpicToken = %q, want BAS", got.EpicToken)
	}
	if len(got.OrphanMx) != 3 {
		t.Fatalf("OrphanMx = %v, want [M2 M3 M5]", got.OrphanMx)
	}
	wantOrphans := map[string]bool{"M2": true, "M3": true, "M5": true}
	for _, o := range got.OrphanMx {
		if !wantOrphans[o] {
			t.Errorf("unexpected orphan %q", o)
		}
	}
	if got.Done != 2 {
		t.Errorf("Done = %d, want 2 (M0+M4 completed with sync sha)", got.Done)
	}
	if got.Total != 6 {
		t.Errorf("Total = %d, want 6", got.Total)
	}
	// pct = round(2/6 * 100) = 33
	if got.Pct != 33 {
		t.Errorf("Pct = %d, want 33", got.Pct)
	}
	// M1 is in-progress → status "in-progress", covered true
	var m1 *MilestoneEntry
	for i := range got.Milestones {
		if got.Milestones[i].ID == "M1" {
			m1 = &got.Milestones[i]
		}
	}
	if m1 == nil {
		t.Fatalf("M1 missing from milestones")
	}
	if m1.Status != "in-progress" {
		t.Errorf("M1 Status = %q, want in-progress", m1.Status)
	}
	if !m1.Covered {
		t.Errorf("M1 should be covered")
	}
}

// TestBuildEpicStatus_NoDesignReport_OmitsOrphan verifies AC-ES-005b: without
// a design report, orphan_mx is omitted entirely (the field is nil → not
// marshaled in JSON).
func TestBuildEpicStatus_NoDesignReport_OmitsOrphan(t *testing.T) {
	root := t.TempDir()
	writeSpecFixture(t, root, "SPEC-X-001", "(EPICX M0) foo", "completed")

	got, err := BuildEpicStatus("X", Options{BaseDir: root})
	if err != nil {
		t.Fatalf("BuildEpicStatus: %v", err)
	}
	if got.OrphanMx != nil {
		t.Errorf("OrphanMx = %v, want nil (no design report)", got.OrphanMx)
	}
	if got.DesignReport != "" {
		t.Errorf("DesignReport = %q, want empty", got.DesignReport)
	}
}

// TestJoinStatus_Classification verifies REQ-ES-006: per-Mx status join.
func TestJoinStatus_Classification(t *testing.T) {
	records := []spec.DocRecord{
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-A-001", Title: "(T M0)", Status: "completed"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-A-002", Title: "(T M1)", Status: "in-progress"}},
		{Frontmatter: spec.SPECFrontmatter{ID: "SPEC-A-003", Title: "(T M2)", Status: "draft"}},
	}
	mxMap := map[string]string{
		"M0": "SPEC-A-001",
		"M1": "SPEC-A-002",
		"M2": "SPEC-A-003",
	}
	// sync_sha map: only M0's SPEC has a sha → only M0 is "done".
	syncSha := map[string]string{"SPEC-A-001": "abc123"}
	ms := JoinStatus(records, mxMap, syncSha, nil)
	if len(ms) != 3 {
		t.Fatalf("expected 3 milestones, got %d", len(ms))
	}
	wantStatus := map[string]string{"M0": "done", "M1": "in-progress", "M2": "planned"}
	for _, m := range ms {
		if wantStatus[m.ID] != m.Status {
			t.Errorf("%s Status = %q, want %q", m.ID, m.Status, wantStatus[m.ID])
		}
	}
}

// TestComputePct_DivideByZero verifies REQ-ES-013 / AC-ES-007: total=0 → pct=0.
func TestComputePct_DivideByZero(t *testing.T) {
	if got := computePct(0, 0); got != 0 {
		t.Errorf("computePct(0,0) = %d, want 0", got)
	}
	if got := computePct(5, 0); got != 0 {
		t.Errorf("computePct(5,0) = %d, want 0", got)
	}
	if got := computePct(3, 6); got != 50 {
		t.Errorf("computePct(3,6) = %d, want 50", got)
	}
}

// TestBuildEpicStatus_EmptyPrefix verifies AC-ES-003b shape: empty match →
// empty epic with pct 0, not an error.
func TestBuildEpicStatus_EmptyPrefix(t *testing.T) {
	root := t.TempDir()
	// unrelated SPEC, not matching prefix
	writeSpecFixture(t, root, "SPEC-OTHER-001", "(T M0) foo", "completed")

	got, err := BuildEpicStatus("NONEXIST", Options{BaseDir: root})
	if err != nil {
		t.Fatalf("BuildEpicStatus empty: %v", err)
	}
	if got.Done != 0 || got.Total != 0 || got.Pct != 0 {
		t.Errorf("empty epic counts = done=%d total=%d pct=%d, want all 0", got.Done, got.Total, got.Pct)
	}
	if len(got.Milestones) != 0 {
		t.Errorf("empty epic milestones len = %d, want 0", len(got.Milestones))
	}
}

func designReportFixtureHTML() string {
	return "<html><body><h2>7. slice</h2><table><tbody>" +
		"<tr><td>M0 graph</td><td>x</td><td>—</td><td>L</td></tr>" +
		"<tr><td>M1 detect</td><td>x</td><td>M0</td><td>M</td></tr>" +
		"<tr><td>M2 route</td><td>x</td><td>M0</td><td>M</td></tr>" +
		"<tr><td>M3 fix</td><td>x</td><td>M1+M2</td><td>L</td></tr>" +
		"<tr><td>M4 4-tier</td><td>x</td><td>M0</td><td>L</td></tr>" +
		"<tr><td>M5 brownfield</td><td>x</td><td>M4</td><td>M</td></tr>" +
		"</tbody></table></body></html>"
}

func writeRaw(t *testing.T, root, relPath, content string) {
	t.Helper()
	full := root + "/" + relPath
	mustWrite(t, full, content)
}
