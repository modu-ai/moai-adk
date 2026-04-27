package dtcg_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// TestIsFrozen_Violations: FROZEN 영역 파일 경로는 반드시 차단돼야 함.
// REQ-DPL-011: 설계 헌법 §2, §3.1-3.3, §5, §11, §12 + 브랜드 컨텍스트 파일 보호.
func TestIsFrozen_Violations(t *testing.T) {
	t.Parallel()

	frozenPaths := []string{
		// 헌법 파일 자체
		".claude/rules/moai/design/constitution.md",
		// 브랜드 컨텍스트 파일 (§3.1 Brand Context)
		".moai/project/brand/brand-voice.md",
		".moai/project/brand/visual-identity.md",
		".moai/project/brand/target-audience.md",
		// 브랜드 디렉토리 하위 모든 파일
		".moai/project/brand/custom-file.md",
		// 헌법의 모든 FROZEN 섹션에 해당하는 경로들
		".claude/rules/moai/design/constitution.md#2-frozen-vs-evolvable-zones",
		".claude/rules/moai/design/constitution.md#safety-architecture",
		".claude/rules/moai/design/constitution.md#gan-loop-contract",
		".claude/rules/moai/design/constitution.md#evaluator-leniency-prevention",
	}

	for _, path := range frozenPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if !dtcg.IsFrozen(path) {
				t.Errorf("IsFrozen(%q) = false; FROZEN 영역이어야 함", path)
			}
		})
	}
}

// TestIsFrozen_Allowed: FROZEN 영역이 아닌 파일 경로는 허용돼야 함.
func TestIsFrozen_Allowed(t *testing.T) {
	t.Parallel()

	allowedPaths := []string{
		// 구현 파일들 (FROZEN 아님)
		"internal/design/dtcg/validator.go",
		"internal/design/dtcg/categories/color.go",
		".claude/skills/moai-workflow-design-import/SKILL.md",
		".claude/skills/moai-design-system/SKILL.md",
		".moai/design/tokens.json",
		".moai/design/components.json",
		// EVOLVABLE 섹션 (§6, §7, §8, §9, §10)
		".claude/rules/moai/design/CHANGELOG.md",
		// 일반 규칙 파일
		".claude/rules/moai/core/moai-constitution.md",
		// 사용자 파일
		"CLAUDE.md",
		"internal/template/templates/CLAUDE.md",
	}

	for _, path := range allowedPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			if dtcg.IsFrozen(path) {
				t.Errorf("IsFrozen(%q) = true; 허용돼야 함", path)
			}
		})
	}
}

// TestBlockWrite_FrozenPath: FROZEN 경로에 쓰기 시도 시 오류 반환.
func TestBlockWrite_FrozenPath(t *testing.T) {
	t.Parallel()

	frozenPaths := []string{
		".claude/rules/moai/design/constitution.md",
		".moai/project/brand/brand-voice.md",
		".moai/project/brand/visual-identity.md",
	}

	for _, path := range frozenPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			err := dtcg.BlockWrite(path, "test reason")
			if err == nil {
				t.Errorf("BlockWrite(%q) = nil; 오류 반환해야 함", path)
			}
			// 오류 메시지에 "frozen" 포함 여부 확인
			if !strings.Contains(strings.ToLower(err.Error()), "frozen") {
				t.Errorf("BlockWrite(%q) 오류 메시지에 'frozen' 없음: %v", path, err)
			}
		})
	}
}

// TestBlockWrite_AllowedPath: 허용 경로에 대한 BlockWrite는 nil 반환.
func TestBlockWrite_AllowedPath(t *testing.T) {
	t.Parallel()

	allowedPaths := []string{
		"internal/design/dtcg/validator.go",
		".moai/design/tokens.json",
		".claude/skills/moai-design-system/SKILL.md",
	}

	for _, path := range allowedPaths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			err := dtcg.BlockWrite(path, "test write")
			if err != nil {
				t.Errorf("BlockWrite(%q) = %v; nil 반환해야 함", path, err)
			}
		})
	}
}

// TestIsFrozen_ConfigBypass: config에서 FROZEN 영역을 우회할 수 없어야 함.
// [HARD]: FROZEN guard는 hardcoded - config bypass 차단 테스트.
func TestIsFrozen_ConfigBypass(t *testing.T) {
	t.Parallel()

	// FROZEN 영역에 대한 우회 시도 - 환경변수나 설정으로 변경 불가
	// constitution.md는 항상 FROZEN
	path := ".claude/rules/moai/design/constitution.md"
	if !dtcg.IsFrozen(path) {
		t.Errorf("IsFrozen(%q) = false; config bypass 불가 - 항상 FROZEN이어야 함", path)
	}

	// 브랜드 디렉토리는 항상 FROZEN
	brandPath := ".moai/project/brand/any-file.md"
	if !dtcg.IsFrozen(brandPath) {
		t.Errorf("IsFrozen(%q) = false; 브랜드 디렉토리는 항상 FROZEN이어야 함", brandPath)
	}
}
