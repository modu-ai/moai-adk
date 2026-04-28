// Package pipeline: 브랜드 충돌 검사기 단위 테스트.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-05).
package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExtractBrandColors_WithColors: hex 색상이 포함된 visual-identity.md에서 정상 추출.
func TestExtractBrandColors_WithColors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	identityPath := filepath.Join(dir, "visual-identity.md")

	content := `# Visual Identity
primary: "#1D4ED8"
secondary: "#7C3AED"
accent: #F59E0B
neutral: #6B7280
`
	if err := os.WriteFile(identityPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	colors, err := ExtractBrandColors(identityPath)
	if err != nil {
		t.Fatalf("ExtractBrandColors 실패: %v", err)
	}

	// 4개 색상 추출 기대
	if len(colors) != 4 {
		t.Errorf("색상 수 = %d, want 4", len(colors))
	}

	// 정규화된 소문자 hex 검증
	hexSet := make(map[string]bool)
	for _, c := range colors {
		hexSet[c.HexValue] = true
	}
	expected := []string{"#1d4ed8", "#7c3aed", "#f59e0b", "#6b7280"}
	for _, e := range expected {
		if !hexSet[e] {
			t.Errorf("예상 색상 '%s'이 추출 결과에 없음. 결과: %v", e, colors)
		}
	}
}

// TestExtractBrandColors_TBDOnly: _TBD_ 플레이스홀더만 있는 경우 빈 슬라이스 반환.
func TestExtractBrandColors_TBDOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	identityPath := filepath.Join(dir, "visual-identity.md")

	content := `# Visual Identity
primary: _TBD_
secondary: _TBD_
`
	if err := os.WriteFile(identityPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	colors, err := ExtractBrandColors(identityPath)
	if err != nil {
		t.Fatalf("ExtractBrandColors 실패: %v", err)
	}

	if len(colors) != 0 {
		t.Errorf("_TBD_ 전용 파일에서 색상 %d개 추출됨 (0 기대)", len(colors))
	}
}

// TestExtractBrandColors_FileNotExist: 파일 없는 경우 nil, nil 반환.
func TestExtractBrandColors_FileNotExist(t *testing.T) {
	t.Parallel()

	colors, err := ExtractBrandColors("/nonexistent/path/visual-identity.md")
	if err != nil {
		t.Fatalf("파일 없는 경우 에러 반환됨: %v", err)
	}
	if colors != nil {
		t.Errorf("파일 없는 경우 nil 기대, got %v", colors)
	}
}

// TestCheckBrandConflicts_NoConflict: tokens.json 색상이 브랜드 팔레트에 있으면 경고 없음.
func TestCheckBrandConflicts_NoConflict(t *testing.T) {
	t.Parallel()

	brandColors := []BrandColor{
		{HexValue: "#1d4ed8"},
		{HexValue: "#7c3aed"},
	}

	tokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#1D4ED8", // 대소문자 달라도 정규화 후 일치
		},
		"color-secondary": map[string]any{
			"$type":  "color",
			"$value": "#7c3aed",
		},
	}

	warnings := CheckBrandConflicts(tokens, brandColors)
	if len(warnings) != 0 {
		t.Errorf("충돌 없어야 하는데 %d개 경고 발생: %v", len(warnings), warnings)
	}
}

// TestCheckBrandConflicts_WithConflict: 브랜드에 없는 색상은 경고 발생.
func TestCheckBrandConflicts_WithConflict(t *testing.T) {
	t.Parallel()

	brandColors := []BrandColor{
		{HexValue: "#1d4ed8"},
	}

	tokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#FF0000", // 브랜드에 없는 색상
		},
	}

	warnings := CheckBrandConflicts(tokens, brandColors)
	if len(warnings) != 1 {
		t.Fatalf("경고 1개 기대, got %d", len(warnings))
	}

	w := warnings[0]
	if w.TokenPath != "color-primary" {
		t.Errorf("TokenPath = %q, want 'color-primary'", w.TokenPath)
	}
	if w.Category != "brand-conflict" {
		t.Errorf("Category = %q, want 'brand-conflict'", w.Category)
	}
}

// TestCheckBrandConflicts_NonColorSkipped: color 이외 $type은 검사에서 제외.
func TestCheckBrandConflicts_NonColorSkipped(t *testing.T) {
	t.Parallel()

	brandColors := []BrandColor{
		{HexValue: "#1d4ed8"},
	}

	tokens := map[string]any{
		// dimension 토큰은 브랜드 색상과 무관 — 경고 없어야 함
		"spacing-md": map[string]any{
			"$type":  "dimension",
			"$value": "16px",
		},
		// color 토큰은 브랜드에 있음 — 경고 없어야 함
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#1d4ed8",
		},
	}

	warnings := CheckBrandConflicts(tokens, brandColors)
	if len(warnings) != 0 {
		t.Errorf("경고 없어야 하는데 %d개 발생", len(warnings))
	}
}

// TestCheckBrandConflicts_EmptyBrandColors: 브랜드 색상이 없으면 경고 없음.
func TestCheckBrandConflicts_EmptyBrandColors(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#FF0000",
		},
	}

	warnings := CheckBrandConflicts(tokens, nil)
	if len(warnings) != 0 {
		t.Errorf("브랜드 색상 없을 때 경고 없어야 함, got %d", len(warnings))
	}
}

// TestRunBrandConflictCheck_Integration: 합성 visual-identity.md + tokens로 통합 테스트.
func TestRunBrandConflictCheck_Integration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	identityPath := filepath.Join(dir, "visual-identity.md")

	content := `# Visual Identity
primary: #1D4ED8
secondary: #7C3AED
`
	if err := os.WriteFile(identityPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("일치하는 색상 — 경고 없음", func(t *testing.T) {
		tokens := map[string]any{
			"color-brand": map[string]any{
				"$type":  "color",
				"$value": "#1d4ed8",
			},
		}
		warnings, err := RunBrandConflictCheck(identityPath, tokens)
		if err != nil {
			t.Fatal(err)
		}
		if len(warnings) != 0 {
			t.Errorf("경고 없어야 함, got %d", len(warnings))
		}
	})

	t.Run("불일치 색상 — 경고 1개", func(t *testing.T) {
		tokens := map[string]any{
			"color-custom": map[string]any{
				"$type":  "color",
				"$value": "#ABCDEF",
			},
		}
		warnings, err := RunBrandConflictCheck(identityPath, tokens)
		if err != nil {
			t.Fatal(err)
		}
		if len(warnings) != 1 {
			t.Fatalf("경고 1개 기대, got %d", len(warnings))
		}
		if warnings[0].Category != "brand-conflict" {
			t.Errorf("category = %q, want 'brand-conflict'", warnings[0].Category)
		}
	})
}

// TestRunBrandConflictCheck_FileNotExist: visual-identity.md 없으면 nil 반환.
func TestRunBrandConflictCheck_FileNotExist(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"color-x": map[string]any{
			"$type":  "color",
			"$value": "#FF0000",
		},
	}
	warnings, err := RunBrandConflictCheck("/nonexistent/visual-identity.md", tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("파일 없을 때 경고 없어야 함, got %d", len(warnings))
	}
}
