package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeRecords is a test helper that writes UsageRecords as JSONL to a file.
func writeRecords(t *testing.T, filePath string, records []UsageRecord) {
	t.Helper()
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("writeRecords: failed to open %s: %v", filePath, err)
	}
	defer func() { _ = f.Close() }()

	for _, r := range records {
		line, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("writeRecords: failed to marshal record: %v", err)
		}
		_, _ = f.Write(append(line, '\n'))
	}
}

// TestReport_AggregatesBySkillID verifies that the report correctly aggregates
// usage counts by skill_id.
func TestReport_AggregatesBySkillID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	records := []UsageRecord{
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: OutcomeSuccess, SessionID: "s1"},
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: OutcomeSuccess, SessionID: "s2"},
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: OutcomeError, SessionID: "s3"},
		{Timestamp: now, SkillID: "moai-workflow-run", Outcome: OutcomeSuccess, SessionID: "s4"},
	}

	dayFile := filepath.Join(telDir, "usage-"+now.Format("2006-01-02")+".jsonl")
	writeRecords(t, dayFile, records)

	report, err := GenerateReport(dir, 30)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	if report == nil {
		t.Fatal("GenerateReport() returned nil report")
	}

	// Find stats for moai-workflow-tdd
	var tddStats *SkillStats
	for i := range report.Skills {
		if report.Skills[i].SkillID == "moai-workflow-tdd" {
			tddStats = &report.Skills[i]
			break
		}
	}

	if tddStats == nil {
		t.Fatal("report missing stats for moai-workflow-tdd")
	}

	if tddStats.Uses != 3 {
		t.Errorf("moai-workflow-tdd Uses = %d, want 3", tddStats.Uses)
	}
	if tddStats.Success != 2 {
		t.Errorf("moai-workflow-tdd Success = %d, want 2", tddStats.Success)
	}
	if tddStats.Error != 1 {
		t.Errorf("moai-workflow-tdd Error = %d, want 1", tddStats.Error)
	}
}

// TestReport_IdentifiesUnderutilizedSkills verifies that skills with fewer than
// 3 uses in the time window are reported as underutilized.
func TestReport_IdentifiesUnderutilizedSkills(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	records := []UsageRecord{
		// well-used skill (3 or more uses)
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: OutcomeSuccess, SessionID: "s1"},
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: OutcomeSuccess, SessionID: "s2"},
		{Timestamp: now, SkillID: "moai-workflow-tdd", Outcome: OutcomeSuccess, SessionID: "s3"},
		// underutilized skill (< 3 uses)
		{Timestamp: now, SkillID: "moai-design-craft", Outcome: OutcomeUnknown, SessionID: "s4"},
		{Timestamp: now, SkillID: "moai-design-craft", Outcome: OutcomeUnknown, SessionID: "s5"},
	}

	dayFile := filepath.Join(telDir, "usage-"+now.Format("2006-01-02")+".jsonl")
	writeRecords(t, dayFile, records)

	report, err := GenerateReport(dir, 30)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	// Check underutilized list
	found := false
	for _, u := range report.Underutilized {
		if u.SkillID == "moai-design-craft" {
			found = true
			if u.Uses != 2 {
				t.Errorf("moai-design-craft underutilized Uses = %d, want 2", u.Uses)
			}
		}
	}
	if !found {
		t.Error("expected moai-design-craft in underutilized, not found")
	}

	// Well-used skill should not be in underutilized
	for _, u := range report.Underutilized {
		if u.SkillID == "moai-workflow-tdd" {
			t.Errorf("moai-workflow-tdd should not be in underutilized (has 3 uses)")
		}
	}
}

// TestReport_ShowsCoOccurrencePatterns verifies that skills used in the same
// session are identified as co-occurring.
func TestReport_ShowsCoOccurrencePatterns(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	// Session s1 uses both moai-workflow-tdd and moai-workflow-spec
	// Session s2 also uses both
	// Session s3 uses only moai-workflow-tdd
	records := []UsageRecord{
		{Timestamp: now, SkillID: "moai-workflow-tdd", SessionID: "s1", Outcome: OutcomeSuccess},
		{Timestamp: now, SkillID: "moai-workflow-spec", SessionID: "s1", Outcome: OutcomeSuccess},
		{Timestamp: now, SkillID: "moai-workflow-tdd", SessionID: "s2", Outcome: OutcomeSuccess},
		{Timestamp: now, SkillID: "moai-workflow-spec", SessionID: "s2", Outcome: OutcomeSuccess},
		{Timestamp: now, SkillID: "moai-workflow-tdd", SessionID: "s3", Outcome: OutcomeSuccess},
	}

	dayFile := filepath.Join(telDir, "usage-"+now.Format("2006-01-02")+".jsonl")
	writeRecords(t, dayFile, records)

	report, err := GenerateReport(dir, 30)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	// Check co-occurrence
	found := false
	for _, co := range report.CoOccurrences {
		// Order-independent check
		if (co.SkillA == "moai-workflow-tdd" && co.SkillB == "moai-workflow-spec") ||
			(co.SkillA == "moai-workflow-spec" && co.SkillB == "moai-workflow-tdd") {
			found = true
			if co.Count != 2 {
				t.Errorf("co-occurrence count = %d, want 2", co.Count)
			}
		}
	}
	if !found {
		t.Error("expected co-occurrence between moai-workflow-tdd and moai-workflow-spec")
	}
}

// TestReport_HandlesEmptyDirectory verifies that GenerateReport returns an empty
// report (not an error) when no telemetry data exists.
func TestReport_HandlesEmptyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Don't create telemetry directory

	report, err := GenerateReport(dir, 30)
	if err != nil {
		t.Fatalf("GenerateReport() with empty dir error = %v, want nil", err)
	}
	if report == nil {
		t.Fatal("GenerateReport() returned nil, want empty report")
	}
	if len(report.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(report.Skills))
	}
}

// TestReport_RespectsTimeWindow verifies that records outside the time window
// are excluded from the report.
func TestReport_RespectsTimeWindow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	old := now.AddDate(0, 0, -40) // 40 days ago, outside 30-day window

	// Write recent record
	recentFile := filepath.Join(telDir, "usage-"+now.Format("2006-01-02")+".jsonl")
	writeRecords(t, recentFile, []UsageRecord{
		{Timestamp: now, SkillID: "recent-skill", Outcome: OutcomeSuccess, SessionID: "s1"},
	})

	// Write old record
	oldFile := filepath.Join(telDir, "usage-"+old.Format("2006-01-02")+".jsonl")
	writeRecords(t, oldFile, []UsageRecord{
		{Timestamp: old, SkillID: "old-skill", Outcome: OutcomeSuccess, SessionID: "s2"},
	})

	report, err := GenerateReport(dir, 30)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	// Only recent skill should appear
	for _, s := range report.Skills {
		if s.SkillID == "old-skill" {
			t.Error("old-skill should not appear in 30-day report")
		}
	}

	foundRecent := false
	for _, s := range report.Skills {
		if s.SkillID == "recent-skill" {
			foundRecent = true
		}
	}
	if !foundRecent {
		t.Error("recent-skill should appear in 30-day report")
	}
}

// TestReport_String verifies that Report.String() produces human-readable output
// containing key headers and skill IDs.
func TestReport_String(t *testing.T) {
	t.Parallel()

	report := &Report{
		Days: 30,
		Skills: []SkillStats{
			{SkillID: "moai-workflow-tdd", Uses: 45, Success: 38, Partial: 5, Error: 2},
		},
		Underutilized: []UnderutilizedSkill{
			{SkillID: "moai-design-craft", Uses: 2},
		},
		CoOccurrences: []CoOccurrence{
			{SkillA: "moai-workflow-tdd", SkillB: "moai-workflow-ddd", Count: 18},
		},
	}

	s := report.String()

	if !strings.Contains(s, "Skill Usage Report") {
		t.Error("expected 'Skill Usage Report' header in output")
	}
	if !strings.Contains(s, "moai-workflow-tdd") {
		t.Error("expected skill ID in output")
	}
	if !strings.Contains(s, "moai-design-craft") {
		t.Error("expected underutilized skill in output")
	}
	if !strings.Contains(s, "moai-workflow-ddd") {
		t.Error("expected co-occurrence skill in output")
	}
}

// TestReport_StringEmpty verifies that an empty report produces a sensible message.
func TestReport_StringEmpty(t *testing.T) {
	t.Parallel()

	report := &Report{Days: 30}
	s := report.String()

	if !strings.Contains(s, "No telemetry data") {
		t.Errorf("expected 'No telemetry data' message for empty report, got: %s", s)
	}
}

// TestGenerateReport_SkipsNonJSONLFiles verifies that non-JSONL files in the
// telemetry directory are ignored during report generation.
func TestGenerateReport_SkipsNonJSONLFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".moai", "evolution", "telemetry")
	if err := os.MkdirAll(telDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a non-JSONL file
	if err := os.WriteFile(filepath.Join(telDir, "README.md"), []byte("readme"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a valid telemetry file
	now := time.Now().UTC()
	dayFile := filepath.Join(telDir, "usage-"+now.Format("2006-01-02")+".jsonl")
	writeRecords(t, dayFile, []UsageRecord{
		{Timestamp: now, SkillID: "test-skill", Outcome: OutcomeSuccess, SessionID: "s1"},
	})

	report, err := GenerateReport(dir, 30)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}
	if len(report.Skills) != 1 {
		t.Errorf("expected 1 skill, got %d", len(report.Skills))
	}
}
