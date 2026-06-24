// Package harness — applier.go tests.
// REQ-HL-003: EnrichDescription frontmatter modification verification.
// REQ-HL-004: InjectTrigger dedup + feature flag verification.
// REQ-HL-005: Apply() snapshot + safety pipeline verification (Phase 4).
package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────
// Test fixtures
// ─────────────────────────────────────────────

// skillFixture is the fixture content for SKILL.md tests.
// Golden fixture for frontmatter preservation verification.
const skillFixture = `---
name: harness-test
description: original description here
triggers:
  - keyword: "harness test"
  - keyword: "test harness"
metadata:
  version: "1.0.0"
  author: "tester"
---

# My Test Harness Skill

This is the skill body content.
It should remain byte-identical after EnrichDescription.

## Section 1

Some content here.
`

// ─────────────────────────────────────────────
// EnrichDescription tests (T-P2-04)
// ─────────────────────────────────────────────

// TestEnrichDescription_UpdatesDescriptionOnly verifies only description is modified and the rest is preserved.
func TestEnrichDescription_UpdatesDescriptionOnly(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	originalBody := extractBody(skillFixture)

	a := NewApplier()
	heuristicNote := "harness frequently triggered"
	if err := a.EnrichDescription(skillPath, heuristicNote); err != nil {
		t.Fatalf("EnrichDescription 오류: %v", err)
	}

	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}

	// Verify the heuristic note was added to the description
	if !strings.Contains(string(content), "# heuristic: "+heuristicNote) {
		t.Errorf("description에 heuristic note 없음:\n%s", content)
	}

	// Verify body is byte-identical
	newBody := extractBody(string(content))
	if newBody != originalBody {
		t.Errorf("body 변경됨:\noriginal: %q\nnew:      %q", originalBody, newBody)
	}
}

// TestEnrichDescription_PreservesAllOtherFrontmatterFields verifies name, triggers, metadata, etc. are preserved.
func TestEnrichDescription_PreservesAllOtherFrontmatterFields(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)

	a := NewApplier()
	if err := a.EnrichDescription(skillPath, "test note"); err != nil {
		t.Fatalf("EnrichDescription 오류: %v", err)
	}

	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("파일 읽기 실패: %v", err)
	}
	text := string(content)

	// Verify other frontmatter fields are preserved
	if !strings.Contains(text, "name: harness-test") {
		t.Error("name 필드 손실")
	}
	if !strings.Contains(text, `keyword: "harness test"`) {
		t.Error("triggers.keyword 손실")
	}
	if !strings.Contains(text, `version: "1.0.0"`) {
		t.Error("metadata.version 손실")
	}
	if !strings.Contains(text, `author: "tester"`) {
		t.Error("metadata.author 손실")
	}
}

// TestEnrichDescription_Idempotent verifies that calling twice with the same note does not add duplicates.
func TestEnrichDescription_Idempotent(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	a := NewApplier()
	note := "idempotent test note"

	// First call
	if err := a.EnrichDescription(skillPath, note); err != nil {
		t.Fatalf("1차 EnrichDescription 오류: %v", err)
	}

	content1, _ := os.ReadFile(skillPath)

	// Second call (same note)
	if err := a.EnrichDescription(skillPath, note); err != nil {
		t.Fatalf("2차 EnrichDescription 오류: %v", err)
	}

	content2, _ := os.ReadFile(skillPath)

	// Contents must be identical
	if !bytes.Equal(content1, content2) {
		t.Errorf("두 번째 호출 후 내용 변경:\nfirst:  %q\nsecond: %q", content1, content2)
	}
}

// TestEnrichDescription_FileNotExist verifies an error is returned when the file does not exist.
func TestEnrichDescription_FileNotExist(t *testing.T) {
	t.Parallel()

	a := NewApplier()
	err := a.EnrichDescription("/nonexistent/path/SKILL.md", "note")
	if err == nil {
		t.Error("없는 파일에서 오류 없음")
	}
}

// TestEnrichDescription_NoFrontmatter verifies that files without frontmatter are handled without error.
func TestEnrichDescription_NoFrontmatter(t *testing.T) {
	t.Parallel()

	noFM := "# Plain Markdown\n\nNo frontmatter here.\n"
	skillPath := writeSkillFixture(t, noFM)

	a := NewApplier()
	// Without frontmatter, either return an error or pass through unchanged.
	// This implementation returns an error.
	err := a.EnrichDescription(skillPath, "note")
	if err == nil {
		t.Error("frontmatter 없는 파일에서 오류 없음")
	}
}

// ─────────────────────────────────────────────
// InjectTrigger tests (T-P2-05)
// ─────────────────────────────────────────────

// TestInjectTrigger_FeatureFlagOff verifies that when the feature flag is OFF, no actual file change occurs.
func TestInjectTrigger_FeatureFlagOff(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	originalContent, _ := os.ReadFile(skillPath)

	a := NewApplier()
	// Feature flag defaults to OFF — actual file write must not happen
	if err := a.InjectTrigger(skillPath, "new-keyword"); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	newContent, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, newContent) {
		t.Error("feature flag OFF임에도 파일이 변경됨")
	}
}

// TestInjectTrigger_DeduplicatesKeywords verifies that existing keywords are not added as duplicates.
// Uses an Applier with the feature flag set to ON.
func TestInjectTrigger_DeduplicatesKeywords(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)

	// Use an Applier with feature flag ON
	a := newApplierWithWritesEnabled()

	// Try to inject an existing keyword ("harness test") again
	existingKeyword := "harness test"
	if err := a.InjectTrigger(skillPath, existingKeyword); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	text := string(content)

	// The "harness test" keyword must appear only once
	count := strings.Count(text, `keyword: "harness test"`)
	if count != 1 {
		t.Errorf(`"harness test" 키워드 횟수 = %d, want 1`, count)
	}
}

// TestInjectTrigger_AddsNewKeyword verifies that a new keyword is added to the triggers list.
func TestInjectTrigger_AddsNewKeyword(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	a := newApplierWithWritesEnabled()

	newKeyword := "brand-new-trigger"
	if err := a.InjectTrigger(skillPath, newKeyword); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	if !strings.Contains(string(content), `keyword: "brand-new-trigger"`) {
		t.Error("새 키워드가 추가되지 않음")
	}
}

// TestInjectTrigger_Idempotent verifies that calling twice with the same keyword does not add duplicates.
func TestInjectTrigger_Idempotent(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	a := newApplierWithWritesEnabled()

	kw := "idempotent-kw"

	if err := a.InjectTrigger(skillPath, kw); err != nil {
		t.Fatalf("1차 InjectTrigger 오류: %v", err)
	}
	if err := a.InjectTrigger(skillPath, kw); err != nil {
		t.Fatalf("2차 InjectTrigger 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	count := strings.Count(string(content), `keyword: "`+kw+`"`)
	if count != 1 {
		t.Errorf("키워드 중복: %d회 존재, want 1", count)
	}
}

// TestInjectTrigger_PreservesBody verifies that the body is byte-identical after InjectTrigger.
func TestInjectTrigger_PreservesBody(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	originalBody := extractBody(skillFixture)
	a := newApplierWithWritesEnabled()

	if err := a.InjectTrigger(skillPath, "test-trigger"); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	newBody := extractBody(string(content))
	if newBody != originalBody {
		t.Errorf("body 변경됨:\noriginal: %q\nnew:      %q", originalBody, newBody)
	}
}

// ─────────────────────────────────────────────
// splitFrontmatterBody additional coverage tests
// ─────────────────────────────────────────────

// TestSplitFrontmatterBody_NoClosingDelimiter verifies an error is returned when the closing --- is missing.
func TestSplitFrontmatterBody_NoClosingDelimiter(t *testing.T) {
	t.Parallel()

	// Frontmatter start present but no closing
	content := "---\nname: test\ndescription: broken\n"
	_, _, err := splitFrontmatterBody(content)
	if err == nil {
		t.Error("종료 구분자 없음: 오류 미반환")
	}
}

// TestSplitFrontmatterBody_BodyEmpty verifies the case where only frontmatter exists and body is empty.
func TestSplitFrontmatterBody_BodyEmpty(t *testing.T) {
	t.Parallel()

	// No body after the closing ---
	content := "---\nname: test\n---"
	fm, body, err := splitFrontmatterBody(content)
	if err != nil {
		t.Fatalf("오류: %v", err)
	}
	if !strings.Contains(fm, "name: test") {
		t.Errorf("fm에 name 없음: %q", fm)
	}
	if body != "" {
		t.Errorf("body = %q, want empty", body)
	}
}

// TestEnrichDescription_BlockScalar verifies handling when description is in block-scalar form.
func TestEnrichDescription_BlockScalar(t *testing.T) {
	t.Parallel()

	// description: | form (block scalar)
	blockFixture := `---
name: my-block-skill
description: |
  This is a multiline
  description content.
triggers:
  - keyword: "test"
---

Body content here.
`
	skillPath := writeSkillFixture(t, blockFixture)
	a := NewApplier()

	// Block-scalar form must be handled without error
	if err := a.EnrichDescription(skillPath, "block test note"); err != nil {
		t.Fatalf("EnrichDescription 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	// File must remain valid after processing
	if len(content) == 0 {
		t.Error("파일 내용 없음")
	}
}

// TestInjectTrigger_NoTriggersSection verifies that the file is unchanged when the triggers section is absent.
func TestInjectTrigger_NoTriggersSection(t *testing.T) {
	t.Parallel()

	noTriggersFixture := `---
name: no-triggers
description: a skill without triggers
---

Body here.
`
	skillPath := writeSkillFixture(t, noTriggersFixture)
	originalContent, _ := os.ReadFile(skillPath)

	a := newApplierWithWritesEnabled()
	// No change if triggers is missing
	if err := a.InjectTrigger(skillPath, "new-kw"); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	newContent, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, newContent) {
		t.Error("triggers 없는 파일이 변경됨")
	}
}

// ─────────────────────────────────────────────
// SafetyEvaluator stub for Apply() tests
// ─────────────────────────────────────────────

// stubEvaluator is a SafetyEvaluator implementation for tests.
type stubEvaluator struct {
	decision Decision
	err      error
}

func (s *stubEvaluator) Evaluate(_ Proposal, _ []Session) (Decision, error) {
	return s.decision, s.err
}

// approvedEvaluator is a stub that always returns DecisionApproved.
func approvedEvaluator() SafetyEvaluator {
	return &stubEvaluator{decision: Decision{Kind: DecisionApproved}}
}

// rejectedEvaluator is a stub that returns a Layer 1 rejection.
func rejectedEvaluator() SafetyEvaluator {
	return &stubEvaluator{decision: Decision{
		Kind:       DecisionRejected,
		RejectedBy: 1,
		Reason:     "L1 FROZEN Guard: 테스트 거부",
	}}
}

// pendingEvaluator is a stub that returns pending_approval.
func pendingEvaluator(proposalID string) SafetyEvaluator {
	return &stubEvaluator{decision: Decision{
		Kind: DecisionPendingApproval,
		OversightProposal: &OversightProposal{
			ProposalID: proposalID,
			Question:   "이 변경을 적용하시겠습니까?",
			Options: []OversightOption{
				{Label: "승인 (권장)", Value: "approve", Recommended: true, Description: "변경을 적용합니다"},
				{Label: "거부", Value: "reject", Description: "변경을 취소합니다"},
			},
		},
	}}
}

// ─────────────────────────────────────────────
// Apply() tests (T-P4-01)
// REQ-HL-005, REQ-HL-009
// ─────────────────────────────────────────────

// TestApply_SnapshotPrecedesWrite verifies that a snapshot is created before the file write.
// [HARD] Snapshot creation MUST precede the file write.
func TestApply_SnapshotPrecedesWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("SKILL.md 생성 실패: %v", err)
	}

	snapshotBase := filepath.Join(dir, "snapshots")
	a := NewApplier()
	proposal := Proposal{
		ID:               "test-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "snapshot test note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply 오류: %v", err)
	}

	// Verify the snapshot directory was created
	entries, err := os.ReadDir(snapshotBase)
	if err != nil {
		t.Fatalf("스냅샷 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Error("스냅샷 디렉토리가 비어 있음 — snapshot 생성 안 됨")
	}

	// Verify that manifest.json exists
	found := false
	for _, e := range entries {
		manifestPath := filepath.Join(snapshotBase, e.Name(), "manifest.json")
		if _, statErr := os.Stat(manifestPath); statErr == nil {
			found = true
			break
		}
	}
	if !found {
		t.Error("manifest.json이 snapshot 디렉토리에 없음")
	}
}

// TestApply_RejectedByFrozenGuard verifies Apply returns an error on a rejection decision.
// [HARD] Apply() must call evaluator.Evaluate() first; reject on Decision = Reject.
func TestApply_RejectedByFrozenGuard(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	snapshotBase := filepath.Join(dir, "snapshots")

	a := NewApplier()
	proposal := Proposal{
		ID:               "frozen-001",
		TargetPath:       ".claude/skills/moai-test-skill/SKILL.md",
		FieldKey:         "description",
		NewValue:         "should not apply",
		PatternKey:       "test:test:ctx",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, rejectedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Error("거부 결정에서 Apply 성공 — 오류가 반환되어야 함")
	}
	if !strings.Contains(err.Error(), "rejected") && !strings.Contains(err.Error(), "거부") {
		t.Errorf("오류 메시지에 거부 원인 없음: %v", err)
	}

	// No snapshot must be created (abort before write on rejection)
	if _, statErr := os.Stat(snapshotBase); statErr == nil {
		entries, _ := os.ReadDir(snapshotBase)
		if len(entries) > 0 {
			t.Error("거부 결정임에도 스냅샷이 생성됨")
		}
	}
}

// TestApply_SnapshotFailAborts verifies the write is aborted when snapshot directory creation fails.
// [HARD] If snapshot fails, abort write.
func TestApply_SnapshotFailAborts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("SKILL.md 생성 실패: %v", err)
	}

	// Create snapshotBase as a file to force directory creation failure
	snapshotBase := filepath.Join(dir, "snapshots-file")
	if err := os.WriteFile(snapshotBase, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("스냅샷 기반 파일 생성 실패: %v", err)
	}

	originalContent, _ := os.ReadFile(skillPath)

	a := NewApplier()
	proposal := Proposal{
		ID:               "snap-fail-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "should not write",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Error("snapshot 실패 시 오류가 반환되어야 함")
	}

	// File must not be modified
	newContent, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, newContent) {
		t.Error("snapshot 실패 후 파일이 변경됨 — abort가 동작하지 않음")
	}
}

// TestApply_PendingApprovalReturnsOversightPayload verifies that in the pending_approval state
// ApplyPendingError and its payload are returned.
func TestApply_PendingApprovalReturnsOversightPayload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("SKILL.md 생성 실패: %v", err)
	}

	snapshotBase := filepath.Join(dir, "snapshots")
	a := NewApplier()
	proposal := Proposal{
		ID:               "pending-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "pending note",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	// pending_approval must be of type ApplyPendingError
	err := a.Apply(proposal, pendingEvaluator("pending-001"), snapshotBase, []Session{})
	if err == nil {
		t.Error("pending_approval 상태에서 오류 없이 성공 — ApplyPendingError가 반환되어야 함")
	}

	var pendingErr *ApplyPendingError
	if !isPendingError(err, &pendingErr) {
		t.Errorf("오류 타입이 *ApplyPendingError가 아님: %T", err)
	} else {
		if pendingErr.OversightPayload == nil {
			t.Error("OversightPayload가 nil")
		}
		if pendingErr.OversightPayload.ProposalID != "pending-001" {
			t.Errorf("ProposalID = %q, want %q", pendingErr.OversightPayload.ProposalID, "pending-001")
		}
	}
}

// TestApply_ManifestContainsExpectedFields verifies that manifest.json contains the expected fields.
func TestApply_ManifestContainsExpectedFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("SKILL.md 생성 실패: %v", err)
	}

	snapshotBase := filepath.Join(dir, "snapshots")
	a := NewApplier()
	proposal := Proposal{
		ID:               "manifest-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "manifest test",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply 오류: %v", err)
	}

	// Read manifest.json
	entries, _ := os.ReadDir(snapshotBase)
	if len(entries) == 0 {
		t.Fatal("스냅샷 디렉토리 없음")
	}

	manifestPath := filepath.Join(snapshotBase, entries[0].Name(), "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest.json 읽기 실패: %v", err)
	}

	var manifest snapshotManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("manifest.json 파싱 실패: %v", err)
	}

	if manifest.ProposalID != "manifest-001" {
		t.Errorf("ProposalID = %q, want %q", manifest.ProposalID, "manifest-001")
	}
	if len(manifest.Files) == 0 {
		t.Error("manifest.files가 비어 있음")
	}
	if manifest.CreatedAt.IsZero() {
		t.Error("manifest.created_at가 zero 값")
	}
}

// isPendingError is a helper that checks whether err is *ApplyPendingError.
func isPendingError(err error, target **ApplyPendingError) bool {
	if err == nil {
		return false
	}
	if pe, ok := err.(*ApplyPendingError); ok {
		*target = pe
		return true
	}
	return false
}

// TestApplyPendingError_Error verifies the ApplyPendingError.Error() method.
func TestApplyPendingError_Error(t *testing.T) {
	t.Parallel()

	t.Run("with_payload", func(t *testing.T) {
		t.Parallel()
		err := &ApplyPendingError{
			OversightPayload: &OversightProposal{ProposalID: "test-123"},
		}
		msg := err.Error()
		if !strings.Contains(msg, "test-123") {
			t.Errorf("Error() = %q, want proposal_id included", msg)
		}
	})

	t.Run("nil_payload", func(t *testing.T) {
		t.Parallel()
		err := &ApplyPendingError{OversightPayload: nil}
		msg := err.Error()
		if msg == "" {
			t.Error("Error() 반환값이 비어 있음")
		}
	})
}

// TestRestoreSnapshot_RestoresByteIdentical verifies RestoreSnapshot restores the file byte-identically.
func TestRestoreSnapshot_RestoresByteIdentical(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	originalContent := []byte(skillFixture)

	// Create the original file
	targetPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(targetPath, originalContent, 0o644); err != nil {
		t.Fatalf("원본 파일 생성 실패: %v", err)
	}

	// Apply creates a snapshot and then modifies the file
	snapshotBase := filepath.Join(dir, "snapshots")
	a := NewApplier()
	proposal := Proposal{
		ID:               "restore-test-001",
		TargetPath:       targetPath,
		FieldKey:         "description",
		NewValue:         "modified value",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply 오류: %v", err)
	}

	// Verify the file was modified after Apply
	after, _ := os.ReadFile(targetPath)
	if bytes.Equal(originalContent, after) {
		t.Skip("EnrichDescription이 변경을 만들지 않은 경우 (fixture에 이미 해당 값 포함)")
	}

	// Find the snapshot directory
	entries, _ := os.ReadDir(snapshotBase)
	if len(entries) == 0 {
		t.Fatal("스냅샷 없음")
	}
	snapshotDir := filepath.Join(snapshotBase, entries[0].Name())

	// Restore via RestoreSnapshot
	if err := RestoreSnapshot(snapshotDir); err != nil {
		t.Fatalf("RestoreSnapshot 오류: %v", err)
	}

	// Verify byte-identical restoration
	restored, _ := os.ReadFile(targetPath)
	if !bytes.Equal(originalContent, restored) {
		t.Errorf("복원 후 byte-identical 불일치:\ngot:  %q\nwant: %q", string(restored), string(originalContent))
	}
}

// TestRestoreSnapshot_InvalidManifest verifies an error is returned when manifest.json is missing.
func TestRestoreSnapshot_InvalidManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Directory with no manifest.json
	err := RestoreSnapshot(dir)
	if err == nil {
		t.Error("manifest.json 없는 디렉토리에서 오류가 반환되어야 함")
	}
}

// TestApply_UnsupportedFieldKey verifies an error is returned for an unknown fieldKey.
func TestApply_UnsupportedFieldKey(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("SKILL.md 생성 실패: %v", err)
	}

	snapshotBase := filepath.Join(dir, "snapshots")
	a := NewApplier()
	proposal := Proposal{
		ID:               "unsupported-001",
		TargetPath:       skillPath,
		FieldKey:         "nonexistent_field",
		NewValue:         "value",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Error("지원하지 않는 fieldKey에서 오류가 반환되어야 함")
	}
	if !strings.Contains(err.Error(), "nonexistent_field") {
		t.Errorf("오류 메시지에 fieldKey 없음: %v", err)
	}
}

// TestWritePromotion_ZeroTime verifies that the current time is set automatically when Ts is zero.
func TestWritePromotion_ZeroTime(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	promoPath := filepath.Join(dir, "tier-promotions.jsonl")
	l := NewLearner(promoPath)

	// Set Ts to the zero value
	promo := Promotion{
		PatternKey:       "moai_subcommand:/moai plan",
		FromTier:         TierObservation.String(),
		ToTier:           TierHeuristic.String(),
		ObservationCount: 3,
		Confidence:       0.80,
	}
	// Ts is the zero value

	if err := l.WritePromotion(promo); err != nil {
		t.Fatalf("WritePromotion 오류: %v", err)
	}

	data, _ := os.ReadFile(promoPath)
	if !strings.Contains(string(data), "\"ts\":") {
		t.Error("ts 필드가 기록되지 않음")
	}
}

// ─────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────

// writeSkillFixture creates a SKILL.md file in a temp directory and returns its path.
func writeSkillFixture(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("SKILL.md 픽스처 생성 실패: %v", err)
	}
	return skillPath
}

// ─────────────────────────────────────────────
// M4 — in-Apply non-regression gate tests
// SPEC-HARNESS-REGRESSION-GATE-001 REQ-RG-007/008/009/010/014
// ─────────────────────────────────────────────

// seqMeasurer is a Measurer that returns a scripted sequence of triples (one per
// call). The first call is the baseline measure, the second the candidate measure.
// When an entry's err is non-nil, that call returns the measurement-exec error
// (drives the fail-closed path, REQ-RG-014).
type seqMeasurer struct {
	calls []measureCall
	idx   int
}

type measureCall struct {
	triple MetricTriple
	err    error
}

func (m *seqMeasurer) Measure(projectRoot string) (MetricTriple, error) {
	if m.idx >= len(m.calls) {
		return MetricTriple{}, nil
	}
	c := m.calls[m.idx]
	m.idx++
	return c.triple, c.err
}

// newGateApplier builds an Applier wired with a manifest, an injected baseline
// store, and an injected measurer — the in-Apply gate is active only when both
// the measurer and the baseline store are non-nil.
func newGateApplier(manifestPath, baselinePath string, m Measurer) *Applier {
	a := newApplierWithManifest(manifestPath)
	a.measurer = m
	a.baselineStore = NewBaselineStore(baselinePath)
	return a
}

// seedBaseline writes a baseline file so the gate treats the next Apply as a
// subsequent run (hasPriorBaseline=true) — enabling the regression block path.
// Without a seed, the first gated Apply adopts the candidate without blocking
// (REQ-RG-005 first-run no-block).
func seedBaseline(t *testing.T, baselinePath string, triple MetricTriple) {
	t.Helper()
	if err := NewBaselineStore(baselinePath).Save(triple); err != nil {
		t.Fatalf("seed baseline: %v", err)
	}
}

// TestApply_Regression_NonRegressing_Keeps verifies a non-regressing Apply
// (Δ ≥ 0 for all dimensions) keeps the change, updates the baseline store, and
// writes the existing M6 "approved" lineage entry (REQ-RG-007, REQ-RG-009).
func TestApply_Regression_NonRegressing_Keeps(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	// baseline then candidate, both equal (Δ=0 — markdown-only typical case).
	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)

	proposal := Proposal{
		ID:               "nonreg-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply (non-regressing) must succeed: %v", err)
	}

	// Baseline store updated to candidate.
	got, present, err := a.baselineStore.Load()
	if err != nil {
		t.Fatalf("baseline Load: %v", err)
	}
	if !present {
		t.Error("baseline store not written after non-regressing apply")
	}
	if got != (MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}) {
		t.Errorf("baseline = %+v, want candidate triple", got)
	}

	// Lineage entry is "approved" (M6 path preserved).
	entries, lerr := LoadManifest(manifestPath)
	if lerr != nil {
		t.Fatalf("LoadManifest: %v", lerr)
	}
	if len(entries) != 1 || entries[0].Decision != "approved" {
		t.Errorf("lineage entries = %+v, want one approved entry", entries)
	}
}

// TestApply_Regression_Blocks_RollsBack verifies a regressing Apply rolls the
// file back to its original bytes and returns ApplyRegressionError (REQ-RG-008).
func TestApply_Regression_Blocks_RollsBack(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	originalContent, _ := os.ReadFile(skillPath)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	// Seed a prior baseline so the gate enables the block path (not first-run).
	seedBaseline(t, baselinePath, MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0})
	// baseline measure good, candidate measure regressed (coverage dropped + lint up).
	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 80.0, LintCount: 3}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)

	proposal := Proposal{
		ID:               "reg-block-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Fatal("regressing apply must return an error")
	}
	var regErr *ApplyRegressionError
	if !asRegressionError(err, &regErr) {
		t.Fatalf("error type = %T, want *ApplyRegressionError", err)
	}
	if len(regErr.Regressed) == 0 {
		t.Error("ApplyRegressionError.Regressed must list the regressed dimensions")
	}

	// File rolled back to original bytes.
	after, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, after) {
		t.Errorf("file not rolled back:\n got: %q\nwant: %q", after, originalContent)
	}

	// Baseline store NOT updated on a blocked apply — it still holds the seeded triple.
	stored, present, _ := a.baselineStore.Load()
	if !present {
		t.Fatal("seeded baseline unexpectedly removed")
	}
	if stored != (MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}) {
		t.Errorf("baseline store must NOT be updated on a blocked apply; got %+v", stored)
	}
}

// TestApply_Outcome_RolledBack_RecordError verifies the F2 contract
// (SPEC-HARNESS-OUTCOME-ERRJOIN-001): when the in-Apply regression gate reaches
// the rolled-back branch AND the additive outcome-record write itself fails, the
// returned error MUST still carry the typed *ApplyRegressionError signal (so a
// caller's errors.As(err, &*ApplyRegressionError) keeps working) AND the
// outcome-record error MUST also be reachable.
//
// The deliberately-failing observer is wired with a logPath whose PARENT is an
// EXISTING REGULAR FILE, so os.MkdirAll(filepath.Dir(logPath)) inside
// RecordExtendedEvent fails ("observer: 디렉토리 생성 실패 …") and recordOutcome
// returns a non-nil error — driving the rolled-back branch's error-compose path.
func TestApply_Outcome_RolledBack_RecordError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	originalContent, _ := os.ReadFile(skillPath)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	// Wire a deliberately-failing observer: logPath's parent is an EXISTING
	// REGULAR FILE, so os.MkdirAll(filepath.Dir(logPath)) fails inside
	// RecordExtendedEvent and recordOutcome returns a non-nil error.
	regularFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(regularFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed regular file: %v", err)
	}
	logPath := filepath.Join(regularFile, "usage-log.jsonl") // parent is a regular file
	failingObserver := NewObserver(logPath)

	// Seed a prior baseline so the gate enables the block path (not first-run),
	// and script a regressed candidate (coverage drop + lint up) to reach the
	// rolled-back branch — same setup as TestApply_Regression_Blocks_RollsBack.
	seedBaseline(t, baselinePath, MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0})
	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 80.0, LintCount: 3}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m).WithOutcomeObserver(failingObserver)

	proposal := Proposal{
		ID:               "errjoin-rollback-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Fatal("rolled-back apply with a failing outcome-record write must return an error")
	}

	// (a) The target file is rolled back to its original bytes — the rollback +
	//     block decision is unaffected by the outcome-record write failure.
	after, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, after) {
		t.Errorf("file not rolled back:\n got: %q\nwant: %q", after, originalContent)
	}

	// (b) F2 contract: errors.As MUST reach the typed *ApplyRegressionError. This
	//     assertion is RED before the fix (the pre-fix branch returns only the
	//     fmt.Errorf wrapper whose single unwrap target is the observer error) and
	//     GREEN after errors.Join(regErr, oerr). NOTE: errors.As is used directly
	//     — NOT the asRegressionError helper, whose direct type assertion an
	//     errors.Join value correctly fails.
	var regErr *ApplyRegressionError
	if !errors.As(err, &regErr) {
		t.Fatalf("errors.As must reach *ApplyRegressionError on the joined error; got %T: %v", err, err)
	}
	if len(regErr.Regressed) == 0 {
		t.Error("ApplyRegressionError.Regressed must list the regressed dimensions")
	}

	// (c) The outcome-record error MUST also be reachable — neither signal is
	//     suppressed. The observer's os.MkdirAll failure surfaces as the
	//     "디렉토리 생성 실패" / "observer:" substring.
	if !strings.Contains(err.Error(), "디렉토리 생성 실패") && !strings.Contains(err.Error(), "observer:") {
		t.Errorf("outcome-record error must remain reachable; got: %v", err)
	}
}

// TestApply_Regression_AppendsBlockedLineage verifies a blocked apply appends a
// "regression-blocked" lineage entry carrying the regressed-dimension summary
// (REQ-RG-010).
func TestApply_Regression_AppendsBlockedLineage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	seedBaseline(t, baselinePath, MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0})
	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 95, Coverage: 87.0, LintCount: 0}}, // tests dropped
	}}
	a := newGateApplier(manifestPath, baselinePath, m)

	proposal := Proposal{
		ID:               "reg-lineage-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	_ = a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})

	entries, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("lineage len = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Decision != "regression-blocked" {
		t.Errorf("Decision = %q, want regression-blocked", e.Decision)
	}
	if !strings.Contains(e.Reason, "tests_passed") {
		t.Errorf("Reason = %q, want it to summarize regressed dimension tests_passed", e.Reason)
	}
}

// TestApply_Regression_MeasurementError_FailsClosed verifies that a measurement
// step that cannot execute (stubbed exec error) causes a fail-closed wrapped
// error and does NOT keep the change (REQ-RG-014).
func TestApply_Regression_MeasurementError_FailsClosed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	originalContent, _ := os.ReadFile(skillPath)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	// First measure (baseline) good, candidate measure errors (build broke).
	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{err: errMeasureExec},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)

	proposal := Proposal{
		ID:               "reg-failclosed-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{})
	if err == nil {
		t.Fatal("measurement exec error must fail closed (return an error)")
	}
	// Must NOT be kept: file rolled back AND baseline not updated.
	after, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, after) {
		t.Error("fail-closed must roll the change back")
	}
	_, present, _ := a.baselineStore.Load()
	if present {
		t.Error("fail-closed must NOT update the baseline store")
	}
}

// TestApply_Regression_ForbiddenFilesUntouched verifies the gate does not create
// or modify the forbidden runtime files (REQ-RG-006, C11).
func TestApply_Regression_ForbiddenFilesUntouched(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := writeLineageFixture(t)
	snapshotBase := filepath.Join(dir, "snapshots")
	manifestPath := filepath.Join(dir, "learning-history", "manifest.jsonl")
	baselinePath := filepath.Join(dir, "harness", "measurements-baseline.yaml")

	harnessDir := filepath.Join(dir, "harness")
	if err := os.MkdirAll(harnessDir, 0o755); err != nil {
		t.Fatalf("mkdir harness: %v", err)
	}
	forbidden := map[string]string{
		"usage-log.jsonl":       `{"x":1}`,
		"observations.yaml":     "obs: 1\n",
		"tier-promotions.jsonl": `{"p":1}`,
	}
	for name, content := range forbidden {
		if err := os.WriteFile(filepath.Join(harnessDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}

	m := &seqMeasurer{calls: []measureCall{
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
		{triple: MetricTriple{TestsPassed: 100, Coverage: 87.0, LintCount: 0}},
	}}
	a := newGateApplier(manifestPath, baselinePath, m)

	proposal := Proposal{
		ID:               "reg-forbidden-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "harness frequently triggered",
		PatternKey:       "moai_subcommand:/moai plan:ctx001",
		Tier:             TierAutoUpdate,
		ObservationCount: 10,
	}

	if err := a.Apply(proposal, approvedEvaluator(), snapshotBase, []Session{}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	for name, want := range forbidden {
		got, _ := os.ReadFile(filepath.Join(harnessDir, name))
		if string(got) != want {
			t.Errorf("forbidden file %s was modified: got %q, want %q", name, got, want)
		}
	}
}

// asRegressionError reports whether err is *ApplyRegressionError.
func asRegressionError(err error, target **ApplyRegressionError) bool {
	if err == nil {
		return false
	}
	if re, ok := err.(*ApplyRegressionError); ok {
		*target = re
		return true
	}
	return false
}

// errMeasureExec is a sentinel measurement-exec error for the fail-closed test.
var errMeasureExec = &measureExecError{}

type measureExecError struct{}

func (e *measureExecError) Error() string { return "measure: simulated build error" }

// extractBody extracts the body of a SKILL.md after the frontmatter (---...---).
func extractBody(content string) string {
	// Frontmatter spans from the first --- to the second ---
	lines := strings.Split(content, "\n")
	inFM := false
	fmClosed := false
	var bodyLines []string

	for _, line := range lines {
		if !inFM && line == "---" {
			inFM = true
			continue
		}
		if inFM && line == "---" {
			fmClosed = true
			inFM = false
			continue
		}
		if fmClosed {
			bodyLines = append(bodyLines, line)
		}
	}
	return strings.Join(bodyLines, "\n")
}
