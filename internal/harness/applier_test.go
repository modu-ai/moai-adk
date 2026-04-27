// Package harness — applier.go 테스트.
// REQ-HL-003: EnrichDescription frontmatter 수정 검증.
// REQ-HL-004: InjectTrigger dedup + feature flag 검증.
package harness

import (
	"bytes"
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
