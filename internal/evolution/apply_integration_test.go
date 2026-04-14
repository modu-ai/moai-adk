package evolution_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

const sampleSkillWithStructure = `# Skill Header

Introduction content that must be preserved.

<!-- moai:evolvable-start id="best-practices" -->
Initial best practice content.
<!-- moai:evolvable-end -->

Footer content that must also be preserved.
`

// TestApplyProposal_PreservesSurroundingContent는 ApplyProposal이 존 주변의
// 헤더/푸터 내용을 보존하는지 확인하는 통합 테스트이다.
// CRITICAL 3: MergeEvolvableZones 오용(파일 파괴) 수정 검증
func TestApplyProposal_PreservesSurroundingContent(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	relPath := ".claude/skills/moai-lang-go/SKILL.md"
	fullPath := filepath.Join(projectRoot, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(sampleSkillWithStructure), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}

	proposal := &evolution.ProposedChange{
		TargetFile: relPath,
		ZoneID:     "best-practices",
		Addition:   "- Always pass context.Context as first argument.\n",
	}

	if err := evolution.ApplyProposal(projectRoot, proposal); err != nil {
		t.Fatalf("ApplyProposal: %v", err)
	}

	updated, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read updated file: %v", err)
	}

	updatedStr := string(updated)

	// 헤더가 보존되어야 함
	if !strings.Contains(updatedStr, "# Skill Header") {
		t.Error("ApplyProposal: 파일 헤더가 사라짐 (CRITICAL 3 회귀)")
	}

	// 소개 내용이 보존되어야 함
	if !strings.Contains(updatedStr, "Introduction content that must be preserved.") {
		t.Error("ApplyProposal: 소개 내용이 사라짐 (CRITICAL 3 회귀)")
	}

	// 푸터가 보존되어야 함
	if !strings.Contains(updatedStr, "Footer content that must also be preserved.") {
		t.Error("ApplyProposal: 푸터가 사라짐 (CRITICAL 3 회귀)")
	}

	// 원본 존 내용이 보존되어야 함
	if !strings.Contains(updatedStr, "Initial best practice content.") {
		t.Error("ApplyProposal: 원본 존 내용이 사라짐")
	}

	// 추가된 내용이 있어야 함
	if !strings.Contains(updatedStr, "Always pass context.Context as first argument.") {
		t.Error("ApplyProposal: 추가된 내용이 없음")
	}
}
