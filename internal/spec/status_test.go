package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidStatuses verifies that all valid status values are defined
func TestValidStatuses(t *testing.T) {
	validStatuses := []string{
		"draft",
		"planned",
		"in-progress",
		"implemented",
		"completed",
		"superseded",
	}

	for _, status := range validStatuses {
		if !IsValidStatus(status) {
			t.Errorf("IsValidStatus(%q) = false, want true", status)
		}
	}

	// Test invalid status
	if IsValidStatus("invalid") {
		t.Error("IsValidStatus(\"invalid\") = true, want false")
	}
}

// TestUpdateStatus_YAMLFormat tests updating status in YAML frontmatter (Format A/B)
func TestUpdateStatus_YAMLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `---
id: SPEC-TEST-001
version: "1.0.0"
status: draft
created: "2026-04-27"
---

# Test SPEC

This is a test SPEC document.
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Test updating status
	if err := UpdateStatus(specDir, "completed"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	// Verify the update
	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec file: %v", err)
	}

	content := string(updatedContent)
	if !strings.Contains(content, "status: completed") {
		t.Errorf("status not updated to completed, got:\n%s", content)
	}

	// Verify other content is preserved
	if !strings.Contains(content, "id: SPEC-TEST-001") {
		t.Error("id field not preserved")
	}
	if !strings.Contains(content, "# Test SPEC") {
		t.Error("content after frontmatter not preserved")
	}
}

// TestUpdateStatus_YAMLNoStatus tests adding status to YAML frontmatter when missing
func TestUpdateStatus_YAMLNoStatus(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-002")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `---
id: SPEC-TEST-002
version: "1.0.0"
created: "2026-04-27"
---

# Test SPEC
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "planned"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec file: %v", err)
	}

	content := string(updatedContent)
	if !strings.Contains(content, "status: planned") {
		t.Errorf("status not added, got:\n%s", content)
	}
}

// TestUpdateStatus_MarkdownListFormat tests updating status in Markdown list format (Format D)
func TestUpdateStatus_MarkdownListFormat(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-003")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `# Test SPEC

- **Status**: draft
- **Author**: GOOS
- **Priority**: P1
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "implemented"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec file: %v", err)
	}

	content := string(updatedContent)
	if !strings.Contains(content, "- **Status**: implemented") {
		t.Errorf("status not updated in Markdown list, got:\n%s", content)
	}
	if !strings.Contains(content, "- **Author**: GOOS") {
		t.Error("other list items not preserved")
	}
}

// TestUpdateStatus_TableFormatKorean tests updating status in Korean table format (Format E)
func TestUpdateStatus_TableFormatKorean(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-004")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `# Test SPEC

| 항목 | 값 |
|------|-----|
| 상태 | draft |
| 우선순위 | P1 |
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "completed"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec file: %v", err)
	}

	content := string(updatedContent)
	if !strings.Contains(content, "| 상태 | completed |") {
		t.Errorf("status not updated in Korean table, got:\n%s", content)
	}
}

// TestUpdateStatus_TableFormatEnglish tests updating status in English table format (Format E)
func TestUpdateStatus_TableFormatEnglish(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-005")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `# Test SPEC

| Field | Value |
|-------|-------|
| Status | draft |
| Priority | P1 |
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "in-progress"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec file: %v", err)
	}

	content := string(updatedContent)
	if !strings.Contains(content, "| Status | in-progress |") {
		t.Errorf("status not updated in English table, got:\n%s", content)
	}
}

// TestUpdateStatus_NoFrontmatter tests adding YAML frontmatter when none exists (Format F)
func TestUpdateStatus_NoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-006")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `# Test SPEC

This SPEC has no frontmatter.
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "draft"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updatedContent, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read updated spec file: %v", err)
	}

	content := string(updatedContent)
	if !strings.Contains(content, "---") {
		t.Error("YAML frontmatter delimiters not added")
	}
	if !strings.Contains(content, "status: draft") {
		t.Error("status not added to new frontmatter")
	}
	if !strings.Contains(content, "# Test SPEC") {
		t.Error("original content not preserved after frontmatter")
	}
}

// TestUpdateStatus_SpecNotFound tests error handling when spec.md doesn't exist
func TestUpdateStatus_SpecNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-NONEXISTENT")
	// Don't create the file

	err := UpdateStatus(specDir, "completed")
	if err == nil {
		t.Error("UpdateStatus should return error when spec.md not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message should mention 'not found', got: %v", err)
	}
}

// TestUpdateStatus_InvalidStatus tests error handling for invalid status values
func TestUpdateStatus_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-007")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	originalContent := `---
id: SPEC-TEST-007
status: draft
---
`
	if err := os.WriteFile(specPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	err := UpdateStatus(specDir, "invalid-status")
	if err == nil {
		t.Error("UpdateStatus should return error for invalid status")
	}
}

// TestParseStatus_YAMLFormat tests reading status from YAML frontmatter
func TestParseStatus_YAMLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-008")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
id: SPEC-TEST-008
status: implemented
---
# Test
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "implemented" {
		t.Errorf("ParseStatus = %q, want %q", status, "implemented")
	}
}

// TestParseStatus_MarkdownListFormat tests reading status from Markdown list
func TestParseStatus_MarkdownListFormat(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-009")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `# Test

- **Status**: completed
- **Author**: GOOS
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "completed" {
		t.Errorf("ParseStatus = %q, want %q", status, "completed")
	}
}

// TestParseStatus_TableFormat tests reading status from table format
func TestParseStatus_TableFormat(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-010")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `# Test

| 상태 | draft |
| 우선순위 | P1 |
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "draft" {
		t.Errorf("ParseStatus = %q, want %q", status, "draft")
	}
}

// TestUpdateStatus_EmptyFrontmatter tests adding status to empty YAML frontmatter
func TestUpdateStatus_EmptyFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-011")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
---
# No fields
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "draft"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updated, _ := os.ReadFile(specPath)
	if !strings.Contains(string(updated), "status: draft") {
		t.Errorf("status not added to empty frontmatter, got:\n%s", string(updated))
	}
}

// TestUpdateStatus_KoreanMarkdownList tests updating status in Korean Markdown list
func TestUpdateStatus_KoreanMarkdownList(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-012")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `# Test

- **상태**: draft
- **작성자**: GOOS
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	if err := UpdateStatus(specDir, "planned"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	updated, _ := os.ReadFile(specPath)
	if !strings.Contains(string(updated), "- **상태**: planned") {
		t.Errorf("Korean markdown list status not updated, got:\n%s", string(updated))
	}
}

// TestParseStatus_NoStatus tests ParseStatus when no status field exists
func TestParseStatus_NoStatus(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-TEST-013")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `# Plain document

No status anywhere.
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	_, err := ParseStatus(specDir)
	if err == nil {
		t.Error("ParseStatus should return error when no status found")
	}
}

// SPEC-STATUS-DRYRUN-001 AC-001/AC-002 (i): a frontmattered SPEC whose body
// carries a version-history table header `| Version | Date | Status | Notes |`
// must resolve from the frontmatter status — never from the header cell.
func TestParseStatus_FrontmatterAnchored_HistoryTableHeader(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-DEMO-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
id: SPEC-DEMO-001
status: completed
---

## HISTORY

| Version | Date | Status | Notes |
|---|---|---|---|
| 0.1.0 | 2026-01-01 | draft | initial |
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "completed" {
		t.Errorf("ParseStatus = %q, want %q (frontmatter must win over body table header)", status, "completed")
	}
}

// SPEC-STATUS-DRYRUN-001 AC-002 (ii): the backticked-prose variant — a body
// line merely mentioning the header shape inside a code span must never be
// returned as the SPEC status.
func TestParseStatus_FrontmatterAnchored_BacktickedProse(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-PROSE-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
id: SPEC-PROSE-001
status: draft
---

## HISTORY

- HISTORY is kept as a bullet list, not a ` + "`| Version | Date | Status | Notes |`" + ` table.
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "draft" {
		t.Errorf("ParseStatus = %q, want %q (backticked prose must never be the status)", status, "draft")
	}
}

// SPEC-STATUS-DRYRUN-001 REQ-002 edge: a frontmatter block that exists but
// carries no status: key falls back to the legacy body formats.
func TestParseStatus_FrontmatterWithoutStatusKey_FallsBackToBody(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-FM-NOSTATUS-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
id: SPEC-FM-NOSTATUS-001
title: "No status key"
---

| Field | Value |
|-------|-------|
| Status | draft |
| Priority | P1 |
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "draft" {
		t.Errorf("ParseStatus = %q, want %q (frontmatter without status key must fall back to body)", status, "draft")
	}
}

// SPEC-STATUS-DRYRUN-001 edge case: malformed (unterminated) frontmatter is
// treated as no frontmatter; the legacy fallback applies and nothing panics.
func TestParseStatus_MalformedFrontmatter_FallsBackToBody(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-BAD-FM-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `---
id: SPEC-BAD-FM-001
status: draft

| Status | implemented |
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "implemented" {
		t.Errorf("ParseStatus = %q, want %q (unterminated frontmatter must fall back to body table)", status, "implemented")
	}
}

// SPEC-STATUS-DRYRUN-001 AC-005 (read): legacy English body table without any
// frontmatter block still resolves via the fallback path.
func TestParseStatus_LegacyTableFallback_NoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, ".moai", "specs", "SPEC-LEGACY-TBL-001")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specPath := filepath.Join(specDir, "spec.md")
	content := `# Test SPEC

| Field | Value |
|-------|-------|
| Status | draft |
| Priority | P1 |
`
	if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	status, err := ParseStatus(specDir)
	if err != nil {
		t.Fatalf("ParseStatus failed: %v", err)
	}
	if status != "draft" {
		t.Errorf("ParseStatus = %q, want %q (legacy fallback must survive)", status, "draft")
	}
}

// SPEC-STATUS-DRYRUN-001 AC-004 (REQ-003): updating a frontmattered SPEC
// touches ONLY the frontmatter status line — the body (history table rows,
// the `Notes` header cell, backticked prose) is byte-identical.
func TestUpdateStatus_FrontmatterAnchored_BodyByteInvariance(t *testing.T) {
	frontmatter := "---\nid: SPEC-DEMO-001\nstatus: completed\n---\n"
	body := "\n## HISTORY\n\n| Version | Date | Status | Notes |\n|---|---|---|---|\n| 0.1.0 | 2026-01-01 | draft | initial |\n\n- HISTORY is kept as a bullet list, not a `| Version | Date | Status | Notes |` table.\n"
	content := frontmatter + body

	updated, err := updateStatusInContent(content, "implemented")
	if err != nil {
		t.Fatalf("updateStatusInContent failed: %v", err)
	}

	if !strings.Contains(updated, "status: implemented") {
		t.Errorf("frontmatter status not updated, got:\n%s", updated)
	}

	// The body after the closing frontmatter delimiter must be byte-identical.
	lines := strings.Split(updated, "\n")
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		t.Fatalf("updated content lost its frontmatter closing delimiter:\n%s", updated)
	}
	updatedBody := strings.Join(lines[closingIdx+1:], "\n")
	if updatedBody != body {
		t.Errorf("body was modified by the status update:\n--- want ---\n%s\n--- got ---\n%s", body, updatedBody)
	}
	if strings.Contains(updated, "| Version | Date | Status | implemented |") {
		t.Error("history table header cell was clobbered")
	}
	if strings.Contains(updated, "Status | implemented |` table") {
		t.Error("backticked prose was clobbered")
	}
}

// SPEC-STATUS-DRYRUN-001 edge case: a frontmatter block lacking a status: key
// gains the key INSIDE the existing block — never a second block, never a
// stray key appended to the body.
func TestUpdateStatus_FrontmatterInsert_IntoExistingBlock(t *testing.T) {
	frontmatter := "---\nid: SPEC-FM-NOSTATUS-001\ntitle: \"No status key\"\n---\n"
	body := "\n| Field | Value |\n|-------|-------|\n| Status | draft |\n"
	content := frontmatter + body

	updated, err := updateStatusInContent(content, "in-progress")
	if err != nil {
		t.Fatalf("updateStatusInContent failed: %v", err)
	}

	if !strings.Contains(updated, "status: in-progress") {
		t.Errorf("status not inserted, got:\n%s", updated)
	}

	// Insertion must land inside the frontmatter block (before its closing ---).
	lines := strings.Split(updated, "\n")
	closingIdx := -1
	for i, line := range lines {
		if i > 0 && strings.TrimSpace(line) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		t.Fatalf("updated content has no frontmatter closing delimiter:\n%s", updated)
	}
	foundInside := false
	for _, line := range lines[:closingIdx] {
		if strings.HasPrefix(line, "status: in-progress") {
			foundInside = true
		}
	}
	if !foundInside {
		t.Errorf("status key not inserted inside the frontmatter block:\n%s", updated)
	}

	// Body stays byte-identical.
	updatedBody := strings.Join(lines[closingIdx+1:], "\n")
	if updatedBody != body {
		t.Errorf("body was modified by the status insertion:\n--- want ---\n%s\n--- got ---\n%s", body, updatedBody)
	}
}

// TestSpecIDPattern tests the SPEC-ID extraction regex pattern
func TestSpecIDPattern(t *testing.T) {
	// This test documents the expected pattern
	// Pattern: SPEC-[A-Z0-9]+-[0-9]+
	validPatterns := []string{
		"SPEC-STATUS-AUTO-001",
		"SPEC-TEST-123",
		"SPEC-AUTH-001",
		"SPEC-V3R2-CON-001",
	}

	// The actual extraction will be tested in the hook handler tests
	// This just documents the expected pattern format
	for _, pattern := range validPatterns {
		if len(pattern) < 5 {
			t.Errorf("pattern %q too short", pattern)
		}
		if !strings.HasPrefix(pattern, "SPEC-") {
			t.Errorf("pattern %q doesn't start with SPEC-", pattern)
		}
	}
}
