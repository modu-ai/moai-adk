// audit_test.go — TDD coverage for audit engine.
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 AC-LSG-002, AC-LSG-007, AC-LSG-009, AC-LSG-013, AC-LSG-017.
package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// auditFixtureSpec represents a SPEC fixture to write under a base dir.
type auditFixtureSpec struct {
	id          string
	specMD      string
	progressMD  string // empty = do not create progress.md (V2.x signal)
}

// buildAuditFixture writes fixtures under tempDir/.moai/specs/<id>/ and returns
// the absolute path of tempDir for use as opts.BaseDir.
func buildAuditFixture(t *testing.T, fixtures []auditFixtureSpec) string {
	t.Helper()
	tempDir := t.TempDir()
	for _, f := range fixtures {
		dir := filepath.Join(tempDir, ".moai", "specs", f.id)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(f.specMD), 0644); err != nil {
			t.Fatal(err)
		}
		if f.progressMD != "" {
			if err := os.WriteFile(filepath.Join(dir, "progress.md"), []byte(f.progressMD), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return tempDir
}

// makeSpecMD assembles a minimal valid spec.md frontmatter with the given status + era + dates.
func makeSpecMD(id, status, era, created string) string {
	frontmatter := "---\nid: " + id + "\n" +
		"title: \"" + id + " title\"\n" +
		"version: \"0.1.0\"\n" +
		"status: " + status + "\n" +
		"created: " + created + "\n" +
		"updated: " + created + "\n" +
		"author: Test\n" +
		"priority: P1\n" +
		"phase: \"v3.0.0\"\n" +
		"module: \"internal/test\"\n" +
		"lifecycle: spec-anchored\n" +
		"tags: \"test\"\n"
	if era != "" {
		frontmatter += "era: \"" + era + "\"\n"
	}
	frontmatter += "---\n\n# Body\n"
	return frontmatter
}

// AC-LSG-002 — Era classification 5 buckets: fixture set covering each era.
func TestAudit_EraClassification5Buckets(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:     "SPEC-V2X-001",
			specMD: makeSpecMD("SPEC-V2X-001", "implemented", "", "2025-12-15"),
			// no progress.md → V2.x (H-1)
		},
		{
			id:         "SPEC-V3R2-001",
			specMD:     makeSpecMD("SPEC-V3R2-001", "implemented", "", "2026-02-15"),
			progressMD: "# Progress\n\n## Section A\nnotes only\n", // no §E.* markers → V3R2-R4 (H-2)
		},
		{
			id:         "SPEC-V3R5-001",
			specMD:     makeSpecMD("SPEC-V3R5-001", "implemented", "", "2026-03-15"),
			progressMD: "## §E.2 Sync-phase\nsync_commit_sha:\n", // empty sync_commit_sha → V3R5 (H-3)
		},
		{
			id:         "SPEC-V3R6-001",
			specMD:     makeSpecMD("SPEC-V3R6-001", "completed", "", "2026-05-25"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc1234\n## §E.5 Mx\nmx_commit_sha: def5678\n",
		},
		{
			id:         "SPEC-UNCLASSIFIED-001",
			specMD:     makeSpecMD("SPEC-UNCLASSIFIED-001", "draft", "", "2026-02-01"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc123\n", // has sync SHA, no mx, no v3.0 phase=YES but date 2026-02-01 < threshold
		},
	}
	// Override phase for unclassified case to make truly ambiguous
	fixtures[4].specMD = strings.Replace(fixtures[4].specMD, `phase: "v3.0.0"`, `phase: "legacy"`, 1)

	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir, IncludeGrandfathered: true})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	if result.TotalSpecs != 5 {
		t.Errorf("TotalSpecs = %d, want 5", result.TotalSpecs)
	}

	// Verify each era was observed (via DriftFindings + Grandfathered count).
	eraCounts := make(map[string]int)
	for _, f := range result.DriftFindings {
		eraCounts[f.Era]++
	}
	for _, era := range []string{"V2.x", "V3R2-R4", "V3R5", "V3R6", "unclassified"} {
		if eraCounts[era] == 0 {
			t.Errorf("era %s not observed in findings; got eraCounts=%v", era, eraCounts)
		}
	}

	// Grandfathered count: V2.x + V3R2-R4 + V3R5 = 3
	if result.Grandfathered != 3 {
		t.Errorf("Grandfathered = %d, want 3", result.Grandfathered)
	}
	// ModernEraClean: V3R6 with no drift = 1
	if result.ModernEraClean != 1 {
		t.Errorf("ModernEraClean = %d, want 1", result.ModernEraClean)
	}
}

// AC-LSG-007 — JSON schema validation: result must marshal/unmarshal cleanly with required fields.
func TestAudit_JSONSchema(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:         "SPEC-V3R6-CLEAN-001",
			specMD:     makeSpecMD("SPEC-V3R6-CLEAN-001", "completed", "V3R6", "2026-05-25"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n",
		},
	}
	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	requiredFields := []string{"audited_at", "total_specs", "grandfathered", "modern_era_clean", "drift_findings"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("required field %q missing in JSON output; got: %s", field, jsonBytes)
		}
	}
}

// AC-LSG-009 — Y_Y_Y_Y_StatusDrift detection: V3R6 with all phase markers + SHAs but status != completed.
func TestAudit_Y4StatusDriftDetection(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:         "SPEC-V3R6-DRIFT-001",
			specMD:     makeSpecMD("SPEC-V3R6-DRIFT-001", "implemented", "V3R6", "2026-05-25"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc1234\n## §E.5 Mx\nmx_commit_sha: def5678\n",
		},
	}

	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir, FilterEra: "V3R6"})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	found := false
	for _, f := range result.DriftFindings {
		if f.FindingType == FindingY_Y_Y_Y_StatusDrift {
			found = true
			if f.Severity != "MUST-FIX" {
				t.Errorf("Y_Y_Y_Y_StatusDrift severity = %q, want MUST-FIX", f.Severity)
			}
			if !strings.Contains(f.Remediation, "moai spec close") {
				t.Errorf("Remediation should reference moai spec close; got %q", f.Remediation)
			}
			if !strings.Contains(f.Remediation, "--backfill-only") {
				t.Errorf("Remediation should suggest --backfill-only; got %q", f.Remediation)
			}
		}
	}
	if !found {
		t.Error("Y_Y_Y_Y_StatusDrift finding not emitted for drift fixture")
	}
}

// AC-LSG-009 — Y_N_N_Y drift: sync present, mx absent + status != completed.
func TestAudit_Y_N_N_Y_DriftDetection(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:         "SPEC-V3R6-Y_N_N_Y-001",
			specMD:     makeSpecMD("SPEC-V3R6-Y_N_N_Y-001", "implemented", "V3R6", "2026-05-25"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc\n", // mx absent
		},
	}
	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	found := false
	for _, f := range result.DriftFindings {
		if f.FindingType == FindingY_N_N_Y {
			found = true
		}
	}
	if !found {
		t.Error("Y_N_N_Y finding not emitted")
	}
}

// AC-LSG-009 — Y_Y_N_Y drift: both sections present but mx_commit_sha missing.
func TestAudit_Y_Y_N_Y_DriftDetection(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:         "SPEC-V3R6-Y_Y_N_Y-001",
			specMD:     makeSpecMD("SPEC-V3R6-Y_Y_N_Y-001", "implemented", "V3R6", "2026-05-25"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha:\n", // empty mx
		},
	}
	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}
	found := false
	for _, f := range result.DriftFindings {
		if f.FindingType == FindingY_Y_N_Y {
			found = true
		}
	}
	if !found {
		t.Error("Y_Y_N_Y finding not emitted")
	}
}

// AC-LSG-013 — Era auto-detection emits EraAutoDetected INFO finding when frontmatter
// era field is absent.
func TestAudit_EraAutoDetection(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:         "SPEC-AUTODETECT-001",
			specMD:     makeSpecMD("SPEC-AUTODETECT-001", "completed", "", "2026-05-25"), // no era field
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n",
		},
	}
	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	found := false
	for _, f := range result.DriftFindings {
		if f.FindingType == FindingEraAutoDetected {
			found = true
			if f.Severity != "INFO" {
				t.Errorf("EraAutoDetected severity = %q, want INFO", f.Severity)
			}
			if f.Era != "V3R6" {
				t.Errorf("EraAutoDetected era = %q, want V3R6", f.Era)
			}
			if _, ok := f.Details["heuristic_matched"]; !ok {
				t.Error("EraAutoDetected details missing heuristic_matched key")
			}
		}
	}
	if !found {
		t.Error("EraAutoDetected finding not emitted for SPEC without era field")
	}
}

// AC-LSG-017 — Backward compatibility: pre-V3R6 SPECs classify as era_final + no drift finding.
func TestAudit_GrandfatherClause_NoDriftForPreV3R6(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:     "SPEC-V2X-LEGACY-001",
			specMD: makeSpecMD("SPEC-V2X-LEGACY-001", "implemented", "", "2025-10-01"),
			// no progress.md
		},
	}
	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	if result.Grandfathered != 1 {
		t.Errorf("Grandfathered = %d, want 1", result.Grandfathered)
	}

	// Must NOT emit MUST-FIX findings for grandfathered SPECs.
	for _, f := range result.DriftFindings {
		if f.SpecID == "SPEC-V2X-LEGACY-001" && f.Severity == "MUST-FIX" {
			t.Errorf("grandfathered SPEC should not emit MUST-FIX; got %v", f)
		}
	}
}

// AC-LSG-017 — IncludeGrandfathered surfaces V2.x SPECs as INFO findings.
func TestAudit_IncludeGrandfathered(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:     "SPEC-V2X-INCLUDE-001",
			specMD: makeSpecMD("SPEC-V2X-INCLUDE-001", "implemented", "", "2025-10-01"),
		},
	}
	baseDir := buildAuditFixture(t, fixtures)
	result, err := Audit(AuditOptions{BaseDir: baseDir, IncludeGrandfathered: true})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	found := false
	for _, f := range result.DriftFindings {
		if f.FindingType == "Grandfathered" && f.SpecID == "SPEC-V2X-INCLUDE-001" {
			found = true
			if f.Severity != "INFO" {
				t.Errorf("Grandfathered severity = %q, want INFO", f.Severity)
			}
		}
	}
	if !found {
		t.Error("Grandfathered finding not emitted when IncludeGrandfathered=true")
	}
}

// Status terminal exemption: superseded / archived / rejected should not emit drift.
func TestAudit_TerminalStatusNoDrift(t *testing.T) {
	t.Parallel()

	terminalStatuses := []string{"superseded", "archived", "rejected"}
	for _, status := range terminalStatuses {
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			fixtures := []auditFixtureSpec{
				{
					id:         "SPEC-TERMINAL-001",
					specMD:     makeSpecMD("SPEC-TERMINAL-001", status, "V3R6", "2026-05-25"),
					progressMD: "## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n",
				},
			}
			baseDir := buildAuditFixture(t, fixtures)
			result, err := Audit(AuditOptions{BaseDir: baseDir})
			if err != nil {
				t.Fatalf("Audit() error = %v", err)
			}
			for _, f := range result.DriftFindings {
				if f.FindingType == FindingY_Y_Y_Y_StatusDrift {
					t.Errorf("terminal status %q should not emit drift finding", status)
				}
			}
		})
	}
}

// Audit on missing .moai/specs dir returns empty result without error.
func TestAudit_MissingSpecsDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	result, err := Audit(AuditOptions{BaseDir: tempDir})
	if err != nil {
		t.Fatalf("Audit() should not error on missing dir; got %v", err)
	}
	if result.TotalSpecs != 0 {
		t.Errorf("TotalSpecs = %d, want 0", result.TotalSpecs)
	}
}

// FilterEra restricts findings to a single era.
func TestAudit_FilterEra(t *testing.T) {
	t.Parallel()

	fixtures := []auditFixtureSpec{
		{
			id:     "SPEC-V2X-FILTER-001",
			specMD: makeSpecMD("SPEC-V2X-FILTER-001", "implemented", "", "2025-10-01"),
		},
		{
			id:         "SPEC-V3R6-FILTER-001",
			specMD:     makeSpecMD("SPEC-V3R6-FILTER-001", "implemented", "V3R6", "2026-05-25"),
			progressMD: "## §E.2 Sync\nsync_commit_sha: abc\n## §E.5 Mx\nmx_commit_sha: def\n",
		},
	}
	baseDir := buildAuditFixture(t, fixtures)

	result, err := Audit(AuditOptions{BaseDir: baseDir, FilterEra: "V3R6", IncludeGrandfathered: true})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	for _, f := range result.DriftFindings {
		if f.Era != "V3R6" {
			t.Errorf("FilterEra=V3R6 leaked finding with era=%q", f.Era)
		}
	}
}

// Empty drift findings serializes as []  not null (NaN-safe JSON).
func TestAudit_EmptyDriftFindingsSerialize(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	result, err := Audit(AuditOptions{BaseDir: tempDir})
	if err != nil {
		t.Fatalf("Audit() error = %v", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}

	if !strings.Contains(string(jsonBytes), `"drift_findings":[]`) {
		t.Errorf("empty drift_findings should serialize as [], got: %s", jsonBytes)
	}
}
