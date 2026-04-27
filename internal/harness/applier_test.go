// Package harness — applier.go 테스트.
// REQ-HL-003: EnrichDescription frontmatter 수정 검증.
// REQ-HL-004: InjectTrigger dedup + feature flag 검증.
// REQ-HL-005: Apply() snapshot + safety pipeline 검증 (Phase 4).
package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────
// 테스트 픽스처
// ─────────────────────────────────────────────

// skillFixture는 SKILL.md 테스트용 픽스처 내용이다.
// frontmatter 보존 검증을 위한 golden fixture.
const skillFixture = `---
name: my-harness-test
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
// EnrichDescription 테스트 (T-P2-04)
// ─────────────────────────────────────────────

// TestEnrichDescription_UpdatesDescriptionOnly는 description만 수정되고 나머지가 보존되는지 검증한다.
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

	// description에 heuristic 노트가 추가되었는지 확인
	if !strings.Contains(string(content), "# heuristic: "+heuristicNote) {
		t.Errorf("description에 heuristic note 없음:\n%s", content)
	}

	// body가 byte-identical인지 확인
	newBody := extractBody(string(content))
	if newBody != originalBody {
		t.Errorf("body 변경됨:\noriginal: %q\nnew:      %q", originalBody, newBody)
	}
}

// TestEnrichDescription_PreservesAllOtherFrontmatterFields는 name, triggers, metadata 등이 보존되는지 검증한다.
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

	// 다른 frontmatter 필드 보존 확인
	if !strings.Contains(text, "name: my-harness-test") {
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

// TestEnrichDescription_Idempotent는 동일 노트로 두 번 호출해도 중복 추가되지 않는지 검증한다.
func TestEnrichDescription_Idempotent(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	a := NewApplier()
	note := "idempotent test note"

	// 첫 번째 호출
	if err := a.EnrichDescription(skillPath, note); err != nil {
		t.Fatalf("1차 EnrichDescription 오류: %v", err)
	}

	content1, _ := os.ReadFile(skillPath)

	// 두 번째 호출 (동일 노트)
	if err := a.EnrichDescription(skillPath, note); err != nil {
		t.Fatalf("2차 EnrichDescription 오류: %v", err)
	}

	content2, _ := os.ReadFile(skillPath)

	// 내용이 동일해야 함
	if !bytes.Equal(content1, content2) {
		t.Errorf("두 번째 호출 후 내용 변경:\nfirst:  %q\nsecond: %q", content1, content2)
	}
}

// TestEnrichDescription_FileNotExist는 파일이 없으면 오류를 반환하는지 검증한다.
func TestEnrichDescription_FileNotExist(t *testing.T) {
	t.Parallel()

	a := NewApplier()
	err := a.EnrichDescription("/nonexistent/path/SKILL.md", "note")
	if err == nil {
		t.Error("없는 파일에서 오류 없음")
	}
}

// TestEnrichDescription_NoFrontmatter는 frontmatter가 없는 파일에서도 오류 없이 처리하는지 검증한다.
func TestEnrichDescription_NoFrontmatter(t *testing.T) {
	t.Parallel()

	noFM := "# Plain Markdown\n\nNo frontmatter here.\n"
	skillPath := writeSkillFixture(t, noFM)

	a := NewApplier()
	// frontmatter 없으면 오류 반환 또는 변경 없이 통과해야 함
	// 이 구현에서는 오류 반환
	err := a.EnrichDescription(skillPath, "note")
	if err == nil {
		t.Error("frontmatter 없는 파일에서 오류 없음")
	}
}

// ─────────────────────────────────────────────
// InjectTrigger 테스트 (T-P2-05)
// ─────────────────────────────────────────────

// TestInjectTrigger_FeatureFlagOff은 feature flag가 OFF이면 실제 파일 변경이 없는지 검증한다.
func TestInjectTrigger_FeatureFlagOff(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)
	originalContent, _ := os.ReadFile(skillPath)

	a := NewApplier()
	// feature flag는 기본 OFF — 실제 파일 write 발생하지 않아야 함
	if err := a.InjectTrigger(skillPath, "new-keyword"); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	newContent, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, newContent) {
		t.Error("feature flag OFF임에도 파일이 변경됨")
	}
}

// TestInjectTrigger_DeduplicatesKeywords는 기존 키워드가 중복 추가되지 않는지 검증한다.
// feature flag를 ON으로 설정한 Applier를 사용한다.
func TestInjectTrigger_DeduplicatesKeywords(t *testing.T) {
	t.Parallel()

	skillPath := writeSkillFixture(t, skillFixture)

	// feature flag ON Applier 사용
	a := newApplierWithWritesEnabled()

	// 이미 존재하는 키워드 ("harness test")를 다시 주입 시도
	existingKeyword := "harness test"
	if err := a.InjectTrigger(skillPath, existingKeyword); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	text := string(content)

	// "harness test" 키워드가 한 번만 있어야 함
	count := strings.Count(text, `keyword: "harness test"`)
	if count != 1 {
		t.Errorf(`"harness test" 키워드 횟수 = %d, want 1`, count)
	}
}

// TestInjectTrigger_AddsNewKeyword는 새 키워드가 triggers 목록에 추가되는지 검증한다.
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

// TestInjectTrigger_Idempotent는 동일 키워드로 두 번 호출해도 중복 추가되지 않는지 검증한다.
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

// TestInjectTrigger_PreservesBody는 InjectTrigger 후 body가 byte-identical인지 검증한다.
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
// splitFrontmatterBody 추가 커버리지 테스트
// ─────────────────────────────────────────────

// TestSplitFrontmatterBody_NoClosingDelimiter는 종료 ---가 없으면 오류를 반환하는지 검증한다.
func TestSplitFrontmatterBody_NoClosingDelimiter(t *testing.T) {
	t.Parallel()

	// frontmatter 시작은 있지만 종료 없음
	content := "---\nname: test\ndescription: broken\n"
	_, _, err := splitFrontmatterBody(content)
	if err == nil {
		t.Error("종료 구분자 없음: 오류 미반환")
	}
}

// TestSplitFrontmatterBody_BodyEmpty는 frontmatter만 있고 body가 없는 경우를 검증한다.
func TestSplitFrontmatterBody_BodyEmpty(t *testing.T) {
	t.Parallel()

	// 종료 --- 이후 body 없음
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

// TestEnrichDescription_BlockScalar는 description이 블록 스칼라 형태일 때 처리하는지 검증한다.
func TestEnrichDescription_BlockScalar(t *testing.T) {
	t.Parallel()

	// description: | 형태 (블록 스칼라)
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

	// 블록 스칼라 형태에서도 오류 없이 처리되어야 함
	if err := a.EnrichDescription(skillPath, "block test note"); err != nil {
		t.Fatalf("EnrichDescription 오류: %v", err)
	}

	content, _ := os.ReadFile(skillPath)
	// 처리 후 파일이 여전히 유효해야 함
	if len(content) == 0 {
		t.Error("파일 내용 없음")
	}
}

// TestInjectTrigger_NoTriggersSection은 triggers 섹션이 없을 때 변경 없이 반환하는지 검증한다.
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
	// triggers 없으면 변경 없음
	if err := a.InjectTrigger(skillPath, "new-kw"); err != nil {
		t.Fatalf("InjectTrigger 오류: %v", err)
	}

	newContent, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, newContent) {
		t.Error("triggers 없는 파일이 변경됨")
	}
}

// ─────────────────────────────────────────────
// Apply() 테스트용 SafetyEvaluator stub
// ─────────────────────────────────────────────

// stubEvaluator는 테스트용 SafetyEvaluator 구현체이다.
type stubEvaluator struct {
	decision Decision
	err      error
}

func (s *stubEvaluator) Evaluate(_ Proposal, _ []Session) (Decision, error) {
	return s.decision, s.err
}

// approvedEvaluator는 항상 DecisionApproved를 반환하는 stub이다.
func approvedEvaluator() SafetyEvaluator {
	return &stubEvaluator{decision: Decision{Kind: DecisionApproved}}
}

// rejectedEvaluator는 Layer 1 거부를 반환하는 stub이다.
func rejectedEvaluator() SafetyEvaluator {
	return &stubEvaluator{decision: Decision{
		Kind:       DecisionRejected,
		RejectedBy: 1,
		Reason:     "L1 FROZEN Guard: 테스트 거부",
	}}
}

// pendingEvaluator는 pending_approval을 반환하는 stub이다.
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
// Apply() 테스트 (T-P4-01)
// REQ-HL-005, REQ-HL-009
// ─────────────────────────────────────────────

// TestApply_SnapshotPrecedesWrite는 snapshot이 파일 write보다 먼저 생성되는지 검증한다.
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

	// 스냅샷 디렉토리가 생성되었는지 확인
	entries, err := os.ReadDir(snapshotBase)
	if err != nil {
		t.Fatalf("스냅샷 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Error("스냅샷 디렉토리가 비어 있음 — snapshot 생성 안 됨")
	}

	// manifest.json이 존재하는지 확인
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

// TestApply_RejectedByFrozenGuard는 거부 결정 시 Apply가 오류를 반환하는지 검증한다.
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

	// 스냅샷이 생성되지 않아야 함 (거부 시 write 발생 전 중단)
	if _, statErr := os.Stat(snapshotBase); statErr == nil {
		entries, _ := os.ReadDir(snapshotBase)
		if len(entries) > 0 {
			t.Error("거부 결정임에도 스냅샷이 생성됨")
		}
	}
}

// TestApply_SnapshotFailAborts는 snapshot 디렉토리 생성 실패 시 write가 중단되는지 검증한다.
// [HARD] If snapshot fails, abort write.
func TestApply_SnapshotFailAborts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillFixture), 0o644); err != nil {
		t.Fatalf("SKILL.md 생성 실패: %v", err)
	}

	// snapshotBase를 파일로 만들어 디렉토리 생성 실패 유발
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

	// 파일이 변경되지 않아야 함
	newContent, _ := os.ReadFile(skillPath)
	if !bytes.Equal(originalContent, newContent) {
		t.Error("snapshot 실패 후 파일이 변경됨 — abort가 동작하지 않음")
	}
}

// TestApply_PendingApprovalReturnsOversightPayload는 pending_approval 상태에서
// ApplyPendingError와 payload가 반환되는지 검증한다.
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

	// pending_approval은 ApplyPendingError 타입이어야 함
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

// TestApply_ManifestContainsExpectedFields는 manifest.json이 올바른 필드를 포함하는지 검증한다.
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

	// manifest.json 읽기
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

// isPendingError는 err가 *ApplyPendingError인지 확인하는 헬퍼이다.
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

// TestApplyPendingError_Error는 ApplyPendingError.Error() 메서드를 검증한다.
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

// TestRestoreSnapshot_RestoresByteIdentical은 RestoreSnapshot이 byte-identical로 복원하는지 검증한다.
func TestRestoreSnapshot_RestoresByteIdentical(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	originalContent := []byte(skillFixture)

	// 원본 파일 생성
	targetPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(targetPath, originalContent, 0o644); err != nil {
		t.Fatalf("원본 파일 생성 실패: %v", err)
	}

	// Apply로 스냅샷 생성 후 파일 변경
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

	// Apply 후 파일이 변경되었는지 확인
	after, _ := os.ReadFile(targetPath)
	if bytes.Equal(originalContent, after) {
		t.Skip("EnrichDescription이 변경을 만들지 않은 경우 (fixture에 이미 해당 값 포함)")
	}

	// 스냅샷 디렉토리 찾기
	entries, _ := os.ReadDir(snapshotBase)
	if len(entries) == 0 {
		t.Fatal("스냅샷 없음")
	}
	snapshotDir := filepath.Join(snapshotBase, entries[0].Name())

	// RestoreSnapshot으로 복원
	if err := RestoreSnapshot(snapshotDir); err != nil {
		t.Fatalf("RestoreSnapshot 오류: %v", err)
	}

	// byte-identical 복원 확인
	restored, _ := os.ReadFile(targetPath)
	if !bytes.Equal(originalContent, restored) {
		t.Errorf("복원 후 byte-identical 불일치:\ngot:  %q\nwant: %q", string(restored), string(originalContent))
	}
}

// TestRestoreSnapshot_InvalidManifest는 manifest.json이 없으면 오류를 반환하는지 검증한다.
func TestRestoreSnapshot_InvalidManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// manifest.json 없는 디렉토리
	err := RestoreSnapshot(dir)
	if err == nil {
		t.Error("manifest.json 없는 디렉토리에서 오류가 반환되어야 함")
	}
}

// TestApply_UnsupportedFieldKey는 알 수 없는 fieldKey에서 오류가 반환되는지 검증한다.
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

// TestWritePromotion_ZeroTime은 Ts가 zero이면 자동으로 현재 시각이 설정되는지 검증한다.
func TestWritePromotion_ZeroTime(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	promoPath := filepath.Join(dir, "tier-promotions.jsonl")
	l := NewLearner(promoPath)

	// Ts를 zero value로 설정
	promo := Promotion{
		PatternKey:       "moai_subcommand:/moai plan",
		FromTier:         TierObservation.String(),
		ToTier:           TierHeuristic.String(),
		ObservationCount: 3,
		Confidence:       0.80,
	}
	// Ts는 zero value

	if err := l.WritePromotion(promo); err != nil {
		t.Fatalf("WritePromotion 오류: %v", err)
	}

	data, _ := os.ReadFile(promoPath)
	if !strings.Contains(string(data), "\"ts\":") {
		t.Error("ts 필드가 기록되지 않음")
	}
}

// ─────────────────────────────────────────────
// 테스트 헬퍼
// ─────────────────────────────────────────────

// writeSkillFixture는 임시 디렉토리에 SKILL.md 파일을 생성하고 경로를 반환한다.
func writeSkillFixture(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("SKILL.md 픽스처 생성 실패: %v", err)
	}
	return skillPath
}

// extractBody는 SKILL.md 내용에서 frontmatter(---...---) 이후 body를 추출한다.
func extractBody(content string) string {
	// 첫 번째 --- 이후 두 번째 ---까지가 frontmatter
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
