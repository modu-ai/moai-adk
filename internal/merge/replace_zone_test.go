package merge

import (
	"strings"
	"testing"
)

// TestReplaceEvolvableZone_Basic verifies basic zone replacement with multi-line content.
func TestReplaceEvolvableZone_Basic(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"# Header",
		"",
		"Introduction paragraph.",
		"",
		`<!-- moai:evolvable-start id="best-practices" -->`,
		"Original content line 1.",
		"Original content line 2.",
		"<!-- moai:evolvable-end -->",
		"",
		"Footer content.",
	}, "\n")

	newZoneContent := "New line 1.\nNew line 2.\nNew line 3."

	result, err := ReplaceEvolvableZone(content, "best-practices", newZoneContent)
	if err != nil {
		t.Fatalf("ReplaceEvolvableZone: %v", err)
	}

	// 헤더와 푸터가 보존되어야 함
	if !strings.Contains(result, "# Header") {
		t.Error("결과에 헤더가 없음")
	}
	if !strings.Contains(result, "Introduction paragraph.") {
		t.Error("결과에 소개 문단이 없음")
	}
	if !strings.Contains(result, "Footer content.") {
		t.Error("결과에 푸터가 없음")
	}

	// 새 존 내용이 포함되어야 함
	if !strings.Contains(result, "New line 1.") {
		t.Error("결과에 새 내용이 없음")
	}
	if !strings.Contains(result, "New line 3.") {
		t.Error("결과에 새 내용 3이 없음")
	}

	// 원본 존 내용은 없어야 함
	if strings.Contains(result, "Original content") {
		t.Error("결과에 원본 내용이 남아 있음")
	}
}

// TestReplaceEvolvableZone_ErrZoneNotFound verifies ErrZoneNotFound is returned
// when the zone does not exist.
func TestReplaceEvolvableZone_ErrZoneNotFound(t *testing.T) {
	t.Parallel()

	content := "# Just a plain file\nNo markers here."

	_, err := ReplaceEvolvableZone(content, "nonexistent-zone", "new content")
	if err == nil {
		t.Fatal("ErrZoneNotFound 예상, 하지만 nil 반환")
	}
	if err != ErrZoneNotFound {
		t.Fatalf("ErrZoneNotFound 예상, 하지만 반환: %v", err)
	}
}

// TestReplaceEvolvableZone_Idempotent verifies that applying the same replacement twice
// yields the same result (idempotency).
func TestReplaceEvolvableZone_Idempotent(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"# Header",
		`<!-- moai:evolvable-start id="zone1" -->`,
		"Original",
		"<!-- moai:evolvable-end -->",
		"Footer",
	}, "\n")

	newContent := "Replacement content."

	result1, err := ReplaceEvolvableZone(content, "zone1", newContent)
	if err != nil {
		t.Fatalf("첫 번째 ReplaceEvolvableZone: %v", err)
	}

	result2, err := ReplaceEvolvableZone(result1, "zone1", newContent)
	if err != nil {
		t.Fatalf("두 번째 ReplaceEvolvableZone: %v", err)
	}

	if result1 != result2 {
		t.Errorf("멱등성 실패:\n첫 번째: %q\n두 번째: %q", result1, result2)
	}
}

// TestMergeEvolvableZones_PreservesFileStructure는 apply.go:78의 버그를 재현한다.
// MergeEvolvableZones를 (file_content, zoneID, newContent)로 잘못 호출했을 때
// 파일 헤더와 푸터가 사라지는 것을 확인한다.
//
// 이 테스트는 ReplaceEvolvableZone으로 수정하기 전 apply.go의 동작을 재현한다.
func TestMergeEvolvableZones_PreservesFileStructure(t *testing.T) {
	t.Parallel()

	// 헤더, 존, 푸터가 있는 완전한 파일
	fullFile := strings.Join([]string{
		"# Skill Header",
		"",
		"Introduction content that must be preserved.",
		"",
		`<!-- moai:evolvable-start id="best-practices" -->`,
		"Initial best practice content.",
		"<!-- moai:evolvable-end -->",
		"",
		"Footer content that must also be preserved.",
	}, "\n")

	newZoneContent := "Initial best practice content.\n- Always pass context.Context as first argument."

	// ReplaceEvolvableZone으로 올바른 대체 수행
	result, err := ReplaceEvolvableZone(fullFile, "best-practices", newZoneContent)
	if err != nil {
		t.Fatalf("ReplaceEvolvableZone: %v", err)
	}

	// 헤더와 푸터가 보존되어야 함
	if !strings.Contains(result, "# Skill Header") {
		t.Error("헤더가 보존되지 않음")
	}
	if !strings.Contains(result, "Introduction content that must be preserved.") {
		t.Error("소개 내용이 보존되지 않음")
	}
	if !strings.Contains(result, "Footer content that must also be preserved.") {
		t.Error("푸터가 보존되지 않음")
	}

	// 새 내용이 포함되어야 함
	if !strings.Contains(result, "Always pass context.Context as first argument.") {
		t.Error("추가된 내용이 없음")
	}
}
