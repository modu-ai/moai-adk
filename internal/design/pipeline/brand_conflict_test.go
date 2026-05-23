// Package pipeline: unit tests for the brand-conflict checker.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-05).
package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExtractBrandColors_WithColors: normal extraction from a visual-identity.md containing hex colors.
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

	// Expect 4 extracted colors.
	if len(colors) != 4 {
		t.Errorf("색상 수 = %d, want 4", len(colors))
	}

	// Verify normalized lowercase hex values.
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

// TestExtractBrandColors_TBDOnly: returns an empty slice when only _TBD_ placeholders are present.
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

// TestExtractBrandColors_FileNotExist: returns nil, nil when the file is missing.
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

// TestCheckBrandConflicts_NoConflict: no warning when tokens.json colors exist in the brand palette.
func TestCheckBrandConflicts_NoConflict(t *testing.T) {
	t.Parallel()

	brandColors := []BrandColor{
		{HexValue: "#1d4ed8"},
		{HexValue: "#7c3aed"},
	}

	tokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#1D4ED8", // case differs but matches after normalization
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

// TestCheckBrandConflicts_WithConflict: a color absent from the brand palette triggers a warning.
func TestCheckBrandConflicts_WithConflict(t *testing.T) {
	t.Parallel()

	brandColors := []BrandColor{
		{HexValue: "#1d4ed8"},
	}

	tokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#FF0000", // color not in the brand palette
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

// TestCheckBrandConflicts_NonColorSkipped: $type other than color is excluded from the check.
func TestCheckBrandConflicts_NonColorSkipped(t *testing.T) {
	t.Parallel()

	brandColors := []BrandColor{
		{HexValue: "#1d4ed8"},
	}

	tokens := map[string]any{
		// dimension token is unrelated to brand colors — no warning expected.
		"spacing-md": map[string]any{
			"$type":  "dimension",
			"$value": "16px",
		},
		// color token is in the brand palette — no warning expected.
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

// TestCheckBrandConflicts_EmptyBrandColors: no warnings when brand colors are absent.
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

// TestRunBrandConflictCheck_Integration: integration test with a synthesized visual-identity.md + tokens.
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

// TestRunBrandConflictCheck_FileNotExist: returns nil when visual-identity.md is missing.
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
