//go:build integration

// Package pipeline: Path B1 (Figma 추출기 meta-harness stub) 통합 테스트.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-03).
//
// 실제 meta-harness 스킬 파일은 사용자 영역에서 생성되므로, 테스트는
// t.TempDir()으로 stub SKILL.md 파일을 생성하여 계약을 검증한다.
// 실제 .claude/skills/ 디렉토리를 오염시키지 않는다.
package pipeline

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// writeStubFigmaExtractorSkill: my-harness-figma-extractor/SKILL.md stub을 생성한다.
// meta-harness가 생성하는 파일의 프론트매터 계약을 시뮬레이션한다.
func writeStubFigmaExtractorSkill(t *testing.T, skillsDir string) string {
	t.Helper()

	skillDir := filepath.Join(skillsDir, "my-harness-figma-extractor")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// REQ-DPL-002: Figma file ID + page selectors + credential reference 포함해야 함
	skillContent := `---
name: my-harness-figma-extractor
description: Figma 디자인 추출기 — 프로젝트 전용 meta-harness 생성 스킬
figma_file_id: "ABCDEF1234567890"
figma_page_selectors:
  - "Design System"
  - "Components"
figma_credential_ref: "env:FIGMA_TOKEN"
generated_by: moai-meta-harness
spec_id: SPEC-V3R3-DESIGN-PIPELINE-001
---

# Figma 추출기 (프로젝트 전용)

이 스킬은 meta-harness가 생성한 Figma 전용 디자인 추출기입니다.
`

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
		t.Fatal(err)
	}

	return skillPath
}

// parseFrontmatter: YAML frontmatter(--- ... ---) 키-값을 파싱한다.
// 단순 key: value 형식만 지원 (YAML 배열은 별도 처리).
func parseFrontmatter(t *testing.T, content string) map[string]string {
	t.Helper()

	result := make(map[string]string)
	inFrontmatter := false
	scanner := bufio.NewScanner(strings.NewReader(content))

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			lineCount++
			if lineCount == 1 {
				inFrontmatter = true
				continue
			}
			if lineCount == 2 {
				break
			}
		}
		if !inFrontmatter {
			continue
		}

		// 간단한 key: value 파싱 (따옴표 제거)
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		result[key] = val
	}

	return result
}

// TestPathB1_FigmaExtractorSkillFrontmatter: stub SKILL.md 프론트매터 검증.
// REQ-DPL-002: Figma file ID, page selectors, credential reference 포함 확인.
func TestPathB1_FigmaExtractorSkillFrontmatter(t *testing.T) {
	skillsDir := t.TempDir()
	skillPath := writeStubFigmaExtractorSkill(t, skillsDir)

	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("SKILL.md 읽기 실패: %v", err)
	}

	content := string(data)
	fm := parseFrontmatter(t, content)

	// figma_file_id 존재 및 비어있지 않음
	if fm["figma_file_id"] == "" {
		t.Error("frontmatter에 figma_file_id 없음")
	}

	// figma_credential_ref 존재
	if fm["figma_credential_ref"] == "" {
		t.Error("frontmatter에 figma_credential_ref 없음")
	}

	// figma_page_selectors 존재 (내용에 포함)
	if !strings.Contains(content, "figma_page_selectors") {
		t.Error("SKILL.md에 figma_page_selectors 없음")
	}

	t.Logf("figma_file_id: %s", fm["figma_file_id"])
	t.Logf("figma_credential_ref: %s", fm["figma_credential_ref"])
}

// TestPathB1_GeneratedTokensValidate: Path B1이 생성하는 tokens.json은 DTCG 2025.10 검증 통과.
func TestPathB1_GeneratedTokensValidate(t *testing.T) {
	// meta-harness figma-extractor가 생성하는 tokens.json 시뮬레이션
	simulatedFigmaTokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#0F172A",
		},
		"color-accent": map[string]any{
			"$type":  "color",
			"$value": "#6366F1",
		},
		"font-heading": map[string]any{
			"$type":  "fontFamily",
			"$value": []any{"Geist", "sans-serif"},
		},
		"space-base": map[string]any{
			"$type":  "dimension",
			"$value": "4px",
		},
	}

	report, err := dtcg.Validate(simulatedFigmaTokens)
	if err != nil {
		t.Fatalf("Validate() 실행 실패: %v", err)
	}

	if !report.Valid {
		t.Errorf("Path B1 토큰이 DTCG 검증 실패. 오류 %d개:", len(report.Errors))
		for _, e := range report.Errors {
			t.Logf("  - %s", e.Error())
		}
	}

	t.Logf("Path B1 토큰 %d개 검증 통과", report.TokenCount)
}

// TestPathB1_PathSelectionRecordsB1: path-selection.json에 "B1" 기록.
func TestPathB1_PathSelectionRecordsB1(t *testing.T) {
	dir := t.TempDir()

	ps := PathSelection{
		Path:               "B1",
		BrandContextLoaded: true,
		SpecID:             "SPEC-V3R3-TEST-B1",
		Timestamp:          time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		SessionID:          "test-session-b1",
	}

	if err := WritePathSelection(dir, ps); err != nil {
		t.Fatalf("WritePathSelection 실패: %v", err)
	}

	got, err := ReadPathSelection(dir)
	if err != nil {
		t.Fatalf("ReadPathSelection 실패: %v", err)
	}

	if got.Path != "B1" {
		t.Errorf("Path = %q, want 'B1'", got.Path)
	}
	if !got.BrandContextLoaded {
		t.Error("BrandContextLoaded = false, want true")
	}
}
