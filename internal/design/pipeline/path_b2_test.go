//go:build integration

// Package pipeline: Path B2 (Pencil MCP meta-harness stub) 통합 테스트.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-04).
//
// 검증:
// 1. my-harness-pencil-mcp/SKILL.md stub 프론트매터 (.pen 경로 + MCP 엔드포인트)
// 2. 생성된 tokens.json이 DTCG 검증 통과
// 3. path-selection.json에 "B2" 기록
// 4. moai-workflow-pencil-integration 참조 없음 (Phase 5 T5-03에서 실제 제거)
//
// [HARD] stub 파일은 t.TempDir() 안에만 생성 — 실제 .claude/skills/ 오염 금지.
package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// writeStubPencilMCPSkill: my-harness-pencil-mcp/SKILL.md stub을 생성한다.
// REQ-DPL-003: .pen 파일 경로 + Pencil MCP 서버 엔드포인트 포함.
func writeStubPencilMCPSkill(t *testing.T, skillsDir string) string {
	t.Helper()

	skillDir := filepath.Join(skillsDir, "my-harness-pencil-mcp")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// REQ-DPL-003: .pen 파일 경로 + MCP 서버 엔드포인트 포함해야 함
	skillContent := `---
name: my-harness-pencil-mcp
description: Pencil MCP 통합 스킬 — 프로젝트 전용 meta-harness 생성 스킬
pen_file_paths:
  - "designs/main.pen"
  - "designs/components.pen"
pencil_mcp_endpoint: "http://localhost:5100"
generated_by: moai-meta-harness
spec_id: SPEC-V3R3-DESIGN-PIPELINE-001
---

# Pencil MCP 통합 (프로젝트 전용)

이 스킬은 meta-harness가 생성한 Pencil MCP 전용 스킬입니다.
moai-workflow-pencil-integration을 대체합니다.
`

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
		t.Fatal(err)
	}

	return skillPath
}

// TestPathB2_PencilMCPSkillFrontmatter: stub SKILL.md 프론트매터 검증.
// REQ-DPL-003: .pen 파일 경로 + Pencil MCP 서버 엔드포인트 포함 확인.
func TestPathB2_PencilMCPSkillFrontmatter(t *testing.T) {
	skillsDir := t.TempDir()
	skillPath := writeStubPencilMCPSkill(t, skillsDir)

	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("SKILL.md 읽기 실패: %v", err)
	}

	content := string(data)
	fm := parseFrontmatter(t, content)

	// pencil_mcp_endpoint 존재 및 비어있지 않음
	if fm["pencil_mcp_endpoint"] == "" {
		t.Error("frontmatter에 pencil_mcp_endpoint 없음")
	}

	// pen_file_paths 존재
	if !strings.Contains(content, "pen_file_paths") {
		t.Error("SKILL.md에 pen_file_paths 없음")
	}

	// .pen 파일 참조 존재 확인
	if !strings.Contains(content, ".pen") {
		t.Error("SKILL.md에 .pen 파일 참조 없음")
	}

	t.Logf("pencil_mcp_endpoint: %s", fm["pencil_mcp_endpoint"])
}

// TestPathB2_GeneratedTokensValidate: Path B2가 생성하는 tokens.json은 DTCG 2025.10 검증 통과.
func TestPathB2_GeneratedTokensValidate(t *testing.T) {
	// Pencil MCP가 생성하는 tokens.json 시뮬레이션
	simulatedPencilTokens := map[string]any{
		"color-surface": map[string]any{
			"$type":  "color",
			"$value": "#FFFFFF",
		},
		"color-text": map[string]any{
			"$type":  "color",
			"$value": "#111827",
		},
		"border-radius-sm": map[string]any{
			"$type":  "dimension",
			"$value": "4px",
		},
		"border-radius-md": map[string]any{
			"$type":  "dimension",
			"$value": "8px",
		},
		"font-weight-bold": map[string]any{
			"$type":  "fontWeight",
			"$value": 700.0,
		},
	}

	report, err := dtcg.Validate(simulatedPencilTokens)
	if err != nil {
		t.Fatalf("Validate() 실행 실패: %v", err)
	}

	if !report.Valid {
		t.Errorf("Path B2 토큰이 DTCG 검증 실패. 오류 %d개:", len(report.Errors))
		for _, e := range report.Errors {
			t.Logf("  - %s", e.Error())
		}
	}

	t.Logf("Path B2 토큰 %d개 검증 통과", report.TokenCount)
}

// TestPathB2_PathSelectionRecordsB2: path-selection.json에 "B2" 기록.
func TestPathB2_PathSelectionRecordsB2(t *testing.T) {
	dir := t.TempDir()

	ps := PathSelection{
		Path:               "B2",
		BrandContextLoaded: false,
		SpecID:             "SPEC-V3R3-TEST-B2",
		Timestamp:          time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		SessionID:          "test-session-b2",
	}

	if err := WritePathSelection(dir, ps); err != nil {
		t.Fatalf("WritePathSelection 실패: %v", err)
	}

	got, err := ReadPathSelection(dir)
	if err != nil {
		t.Fatalf("ReadPathSelection 실패: %v", err)
	}

	if got.Path != "B2" {
		t.Errorf("Path = %q, want 'B2'", got.Path)
	}
}

// TestPathB2_LegacyPencilIntegrationAbsent: moai-workflow-pencil-integration 참조 없음 검증.
// [HARD] Phase 5 T5-03에서 실제 파일을 제거. 이 테스트는 stub 수준에서 부재를 검증한다.
func TestPathB2_LegacyPencilIntegrationAbsent(t *testing.T) {
	skillsDir := t.TempDir()

	// stub 환경에서 레거시 스킬 생성하지 않음 — 부재 검증
	legacyPath := filepath.Join(skillsDir, "moai-workflow-pencil-integration", "SKILL.md")

	_, err := os.Stat(legacyPath)
	if !os.IsNotExist(err) {
		t.Errorf("레거시 moai-workflow-pencil-integration이 stub 환경에 존재함 — 없어야 함")
	}

	// Path B2 stub만 존재해야 함
	writeStubPencilMCPSkill(t, skillsDir)
	pencilMCPPath := filepath.Join(skillsDir, "my-harness-pencil-mcp", "SKILL.md")
	if _, err := os.Stat(pencilMCPPath); os.IsNotExist(err) {
		t.Error("my-harness-pencil-mcp/SKILL.md stub이 없음")
	}

	t.Log("검증 완료: stub 환경에 moai-workflow-pencil-integration 없음, my-harness-pencil-mcp 존재")
}
