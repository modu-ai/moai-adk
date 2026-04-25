package taxonomy_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy"
)

// TestAuditFile_MissingType verifies AC-EXT001-01:
// a file without a type key produces MEMORY_MISSING_TYPE.
func TestAuditFile_MissingType(t *testing.T) {
	t.Parallel()
	findings, err := taxonomy.AuditFile(fixturesPath("missing_type.md"))
	if err != nil {
		t.Fatalf("AuditFile error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnMissingType) {
		t.Errorf("want MEMORY_MISSING_TYPE finding; got %v", findings)
	}
}

// TestAuditFile_MissingFrontmatterKeys verifies AC-EXT001-01b:
// files missing name or description produce MEMORY_MISSING_FRONTMATTER.
func TestAuditFile_MissingFrontmatterKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		fixture  string
		wantCode taxonomy.AuditCode
	}{
		{"missing name", "missing_name.md", taxonomy.WarnMissingFrontmatter},
		{"missing description", "missing_description.md", taxonomy.WarnMissingFrontmatter},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			findings, err := taxonomy.AuditFile(fixturesPath(tt.fixture))
			if err != nil {
				t.Fatalf("AuditFile(%s) error = %v", tt.fixture, err)
			}
			if !hasCode(findings, tt.wantCode) {
				t.Errorf("want %s finding for %s; got %v", tt.wantCode, tt.fixture, findings)
			}
		})
	}
}

// TestAuditFile_FeedbackBodyMarkers verifies AC-EXT001-04a:
// feedback files are checked for the **Why:** and **How to apply:** markers.
func TestAuditFile_FeedbackBodyMarkers(t *testing.T) {
	t.Parallel()
	t.Run("pass: markers present", func(t *testing.T) {
		t.Parallel()
		findings, err := taxonomy.AuditFile(fixturesPath("feedback_testing.md"))
		if err != nil {
			t.Fatalf("AuditFile(feedback_testing.md) error = %v", err)
		}
		if hasCode(findings, taxonomy.WarnBodyStructureMissing) {
			t.Error("want no MEMORY_BODY_STRUCTURE_MISSING for well-formed feedback; got finding")
		}
	})
	t.Run("fail: markers absent", func(t *testing.T) {
		t.Parallel()
		findings, err := taxonomy.AuditFile(fixturesPath("feedback_missing_why.md"))
		if err != nil {
			t.Fatalf("AuditFile(feedback_missing_why.md) error = %v", err)
		}
		if !hasCode(findings, taxonomy.WarnBodyStructureMissing) {
			t.Errorf("want MEMORY_BODY_STRUCTURE_MISSING for feedback without markers; got %v", findings)
		}
	})
}

// TestAuditFile_ExcludedCategory verifies AC-EXT001-10:
// content matching excluded-category keywords produces MEMORY_EXCLUDED_CATEGORY.
func TestAuditFile_ExcludedCategory(t *testing.T) {
	t.Parallel()
	findings, err := taxonomy.AuditFile(fixturesPath("excluded_claude_md.md"))
	if err != nil {
		t.Fatalf("AuditFile(excluded_claude_md.md) error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnExcludedCategory) {
		t.Errorf("want MEMORY_EXCLUDED_CATEGORY for CLAUDE.md mirrored content; got %v", findings)
	}
	// Detail must name the category.
	for _, f := range findings {
		if f.Code == taxonomy.WarnExcludedCategory && f.Detail == "" {
			t.Error("MEMORY_EXCLUDED_CATEGORY finding has empty Detail; must name the category")
		}
	}
}

// TestAuditIndex_Overflow verifies AC-EXT001-03:
// a MEMORY.md with 250 lines triggers MEMORY_INDEX_OVERFLOW.
func TestAuditIndex_Overflow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "MEMORY.md")
	content := strings.Repeat("- [entry](file.md) — description\n", 250)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditIndex(path, config.DefaultMemoryIndexLineCap)
	if err != nil {
		t.Fatalf("AuditIndex error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnIndexOverflow) {
		t.Errorf("want MEMORY_INDEX_OVERFLOW for 250-line MEMORY.md; got %v", findings)
	}
	// Detail must include the line cap.
	for _, f := range findings {
		if f.Code == taxonomy.WarnIndexOverflow && !strings.Contains(f.Detail, "200") {
			t.Errorf("MEMORY_INDEX_OVERFLOW detail = %q, want to contain '200'", f.Detail)
		}
	}
}

// TestAuditIndex_EdgeExactly200 verifies no overflow at exactly 200 lines.
func TestAuditIndex_EdgeExactly200(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "MEMORY.md")
	content := strings.Repeat("- line\n", 200)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditIndex(path, config.DefaultMemoryIndexLineCap)
	if err != nil {
		t.Fatalf("AuditIndex error = %v", err)
	}
	if hasCode(findings, taxonomy.WarnIndexOverflow) {
		t.Error("want no MEMORY_INDEX_OVERFLOW for exactly 200-line MEMORY.md")
	}
}

// TestAuditIndex_EdgeOverflow verifies overflow at 201 lines.
func TestAuditIndex_EdgeOverflow(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "MEMORY.md")
	content := strings.Repeat("- line\n", 201)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := taxonomy.AuditIndex(path, config.DefaultMemoryIndexLineCap)
	if err != nil {
		t.Fatalf("AuditIndex error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnIndexOverflow) {
		t.Error("want MEMORY_INDEX_OVERFLOW for 201-line MEMORY.md")
	}
}

// TestAuditDuplicates verifies AC-EXT001-11:
// two files with the same description trigger MEMORY_DUPLICATE.
func TestAuditDuplicates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sharedDesc := "shared description for duplicate test"
	writeMemFile(t, dir, "a.md", "user", sharedDesc)
	writeMemFile(t, dir, "b.md", "feedback", sharedDesc)

	findings, err := taxonomy.AuditDuplicates(dir)
	if err != nil {
		t.Fatalf("AuditDuplicates error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnDuplicate) {
		t.Errorf("want MEMORY_DUPLICATE for same-description files; got %v", findings)
	}
	// Detail must contain both file paths.
	for _, f := range findings {
		if f.Code == taxonomy.WarnDuplicate {
			if !strings.Contains(f.Detail, "a.md") || !strings.Contains(f.Detail, "b.md") {
				t.Errorf("MEMORY_DUPLICATE detail = %q, want both paths (a.md, b.md)", f.Detail)
			}
		}
	}
}

// TestAuditDuplicates_SameDescDifferentType verifies that description-based
// duplicate detection is type-agnostic (per acceptance.md edge case).
func TestAuditDuplicates_SameDescDifferentType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sharedDesc := "oncall latency dashboard reference"
	writeMemFile(t, dir, "ref1.md", "reference", sharedDesc)
	writeMemFile(t, dir, "user1.md", "user", sharedDesc)

	findings, err := taxonomy.AuditDuplicates(dir)
	if err != nil {
		t.Fatalf("AuditDuplicates error = %v", err)
	}
	if !hasCode(findings, taxonomy.WarnDuplicate) {
		t.Errorf("want MEMORY_DUPLICATE even when types differ; got %v", findings)
	}
}

// TestAuditDuplicates_NoDuplicates verifies no false positives when descriptions differ.
func TestAuditDuplicates_NoDuplicates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeMemFile(t, dir, "a.md", "user", "unique description A")
	writeMemFile(t, dir, "b.md", "feedback", "unique description B")

	findings, err := taxonomy.AuditDuplicates(dir)
	if err != nil {
		t.Fatalf("AuditDuplicates error = %v", err)
	}
	if hasCode(findings, taxonomy.WarnDuplicate) {
		t.Error("want no MEMORY_DUPLICATE for distinct descriptions")
	}
}

// TestAuditFile_NoFindings verifies a well-formed file produces no audit findings.
func TestAuditFile_NoFindings(t *testing.T) {
	t.Parallel()
	findings, err := taxonomy.AuditFile(fixturesPath("user_role.md"))
	if err != nil {
		t.Fatalf("AuditFile(user_role.md) error = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("AuditFile(well-formed user_role.md) = %v, want no findings", findings)
	}
}

// hasCode is a helper that checks whether any finding has the given code.
func hasCode(findings []taxonomy.AuditFinding, code taxonomy.AuditCode) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}

// writeMemFile creates a minimal memory file with the given type and description in dir.
func writeMemFile(t *testing.T, dir, name, typ, description string) {
	t.Helper()
	content := fmt.Sprintf("---\nname: %s\ndescription: %s\ntype: %s\n---\nbody content\n", name, description, typ)
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("writeMemFile %s: %v", name, err)
	}
}
