package dtcg_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// TestValidate_ValidTokenSet: 유효한 토큰 집합 전체 검증.
func TestValidate_ValidTokenSet(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"color-primary": map[string]any{
			"$type":  "color",
			"$value": "#0066ff",
		},
		"spacing-base": map[string]any{
			"$type":  "dimension",
			"$value": map[string]any{"value": 8.0, "unit": "px"},
		},
		"font-weight-bold": map[string]any{
			"$type":  "fontWeight",
			"$value": float64(700),
		},
		"font-family-base": map[string]any{
			"$type":  "fontFamily",
			"$value": []any{"Roboto", "sans-serif"},
		},
		"duration-fast": map[string]any{
			"$type":  "duration",
			"$value": map[string]any{"value": 150.0, "unit": "ms"},
		},
		"easing-ease": map[string]any{
			"$type":  "cubicBezier",
			"$value": []any{0.25, 0.1, 0.25, 1.0},
		},
		"opacity-full": map[string]any{
			"$type":  "number",
			"$value": float64(1),
		},
		"border-style": map[string]any{
			"$type":  "strokeStyle",
			"$value": "solid",
		},
		"border-default": map[string]any{
			"$type": "border",
			"$value": map[string]any{
				"color": "#000000",
				"width": "1px",
				"style": "solid",
			},
		},
		"shadow-sm": map[string]any{
			"$type": "shadow",
			"$value": map[string]any{
				"color":   "#000000",
				"offsetX": "0px",
				"offsetY": "2px",
				"blur":    "4px",
				"spread":  "0px",
			},
		},
		"gradient-brand": map[string]any{
			"$type": "gradient",
			"$value": []any{
				map[string]any{"color": "#0066ff", "position": 0.0},
				map[string]any{"color": "#00ccff", "position": 1.0},
			},
		},
		"transition-default": map[string]any{
			"$type": "transition",
			"$value": map[string]any{
				"duration":       "300ms",
				"timingFunction": []any{0.25, 0.1, 0.25, 1.0},
			},
		},
	}

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	if !report.Valid {
		for _, e := range report.Errors {
			t.Logf("오류: %v", e)
		}
		t.Errorf("Validate() = Invalid; 유효해야 함 (오류 %d개)", len(report.Errors))
	}
	if report.TokenCount != len(tokens) {
		t.Errorf("TokenCount = %d; want %d", report.TokenCount, len(tokens))
	}
}

// TestValidate_InvalidTokens: 잘못된 토큰 검증 - 오류 집계.
func TestValidate_InvalidTokens(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"bad-color": map[string]any{
			"$type":  "color",
			"$value": "notacolor",
		},
		"bad-dimension": map[string]any{
			"$type":  "dimension",
			"$value": 42, // 숫자 타입 (잘못된 형식)
		},
		"bad-weight": map[string]any{
			"$type":  "fontWeight",
			"$value": float64(950), // 범위 초과
		},
	}

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 예상치 못한 오류: %v", err)
	}
	if report.Valid {
		t.Errorf("Validate() = Valid; 오류가 있어야 함")
	}
	if len(report.Errors) != 3 {
		t.Errorf("오류 수 = %d; want 3", len(report.Errors))
	}
}

// TestValidate_MissingType: $type 누락 토큰 처리.
func TestValidate_MissingType(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"no-type": map[string]any{
			"$value": "#0066ff",
		},
	}

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	if report.Valid {
		t.Errorf("Validate() = Valid; $type 누락 시 오류여야 함")
	}
}

// TestValidate_MissingValue: $value 누락 토큰 처리.
func TestValidate_MissingValue(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"no-value": map[string]any{
			"$type": "color",
		},
	}

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	if report.Valid {
		t.Errorf("Validate() = Valid; $value 누락 시 오류여야 함")
	}
}

// TestValidate_UnknownCategory: 알 수 없는 $type 카테고리 처리.
func TestValidate_UnknownCategory(t *testing.T) {
	t.Parallel()

	tokens := map[string]any{
		"unknown-cat": map[string]any{
			"$type":  "motionPath",
			"$value": "some-value",
		},
	}

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	if report.Valid {
		t.Errorf("Validate() = Valid; 알 수 없는 카테고리 시 오류여야 함")
	}
}

// TestValidate_EmptyTokenSet: 빈 토큰 집합 검증.
func TestValidate_EmptyTokenSet(t *testing.T) {
	t.Parallel()

	report, err := dtcg.Validate(map[string]any{})
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	if !report.Valid {
		t.Errorf("Validate() = Invalid; 빈 집합은 유효해야 함")
	}
	if report.TokenCount != 0 {
		t.Errorf("TokenCount = %d; want 0", report.TokenCount)
	}
}

// TestValidate_NilTokenSet: nil 입력 처리.
func TestValidate_NilTokenSet(t *testing.T) {
	t.Parallel()

	report, err := dtcg.Validate(nil)
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	if !report.Valid {
		t.Errorf("Validate() = Invalid; nil 입력은 유효(빈 집합)로 처리해야 함")
	}
}

// TestValidate_NonTokenEntry: 토큰이 아닌 map[string]any 항목 (중첩 그룹) 처리.
func TestValidate_NonTokenEntry(t *testing.T) {
	t.Parallel()

	// DTCG에서는 $type/$value 없는 항목은 그룹으로 취급 - 스킵
	tokens := map[string]any{
		"color-group": map[string]any{
			// $type/$value 없음 - 그룹 노드로 스킵
			"description": "색상 그룹",
		},
		"actual-color": map[string]any{
			"$type":  "color",
			"$value": "#fff",
		},
	}

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 오류: %v", err)
	}
	// 그룹 노드는 스킵, actual-color만 검증
	if !report.Valid {
		t.Errorf("Validate() = Invalid; 그룹 노드는 스킵해야 함")
	}
}

// TestValidate_PositiveGoldenFiles: testdata/positive/*.json 파일 모두 검증 통과해야 함.
// T3-10: 골든 파일 table-driven 테스트.
func TestValidate_PositiveGoldenFiles(t *testing.T) {
	t.Parallel()

	positiveDir := filepath.Join("testdata", "positive")
	entries, err := os.ReadDir(positiveDir)
	if err != nil {
		t.Fatalf("testdata/positive 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("testdata/positive 디렉토리가 비어있음")
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(positiveDir, name))
			if err != nil {
				t.Fatalf("파일 읽기 실패 (%s): %v", name, err)
			}

			var tokens map[string]any
			if err := json.Unmarshal(data, &tokens); err != nil {
				t.Fatalf("JSON 파싱 실패 (%s): %v", name, err)
			}

			report, err := dtcg.Validate(tokens)
			if err != nil {
				t.Fatalf("Validate() 오류 (%s): %v", name, err)
			}
			if !report.Valid {
				for _, e := range report.Errors {
					t.Logf("오류: %v", e)
				}
				t.Errorf("Validate(%s) = Invalid; 양성 파일은 통과해야 함 (오류 %d개)", name, len(report.Errors))
			}
		})
	}
}

// TestValidate_NegativeGoldenFiles: testdata/negative/*.json 파일 모두 검증 실패해야 함.
func TestValidate_NegativeGoldenFiles(t *testing.T) {
	t.Parallel()

	negativeDir := filepath.Join("testdata", "negative")
	entries, err := os.ReadDir(negativeDir)
	if err != nil {
		t.Fatalf("testdata/negative 디렉토리 읽기 실패: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("testdata/negative 디렉토리가 비어있음")
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(negativeDir, name))
			if err != nil {
				t.Fatalf("파일 읽기 실패 (%s): %v", name, err)
			}

			var tokens map[string]any
			if err := json.Unmarshal(data, &tokens); err != nil {
				t.Fatalf("JSON 파싱 실패 (%s): %v", name, err)
			}

			report, err := dtcg.Validate(tokens)
			if err != nil {
				t.Fatalf("Validate() 예상치 못한 오류 (%s): %v", name, err)
			}
			if report.Valid {
				t.Errorf("Validate(%s) = Valid; 음성 파일은 오류가 있어야 함", name)
			}
		})
	}
}

// BenchmarkValidate_500Tokens: 500개 토큰 검증 성능 벤치마크.
// [HARD]: 500개 토큰 검증은 100ms 미만이어야 함 (REQ-DPL-010).
func BenchmarkValidate_500Tokens(b *testing.B) {
	// 500개 토큰 생성
	tokens := make(map[string]any, 500)
	categories := []string{"color", "dimension", "fontWeight", "fontFamily", "duration", "number", "strokeStyle"}
	values := map[string]any{
		"color":       "#0066ff",
		"dimension":   map[string]any{"value": 8.0, "unit": "px"},
		"fontWeight":  float64(400),
		"fontFamily":  "Roboto",
		"duration":    map[string]any{"value": 300.0, "unit": "ms"},
		"number":      float64(1),
		"strokeStyle": "solid",
	}

	for i := range 500 {
		cat := categories[i%len(categories)]
		key := "token." + cat + "." + string(rune('a'+i%26)) + string(rune('0'+i/26%10))
		tokens[key] = map[string]any{
			"$type":  cat,
			"$value": values[cat],
		}
	}

	b.ResetTimer()
	for range b.N {
		start := time.Now()
		report, err := dtcg.Validate(tokens)
		elapsed := time.Since(start)

		if err != nil {
			b.Fatalf("Validate() 오류: %v", err)
		}
		if !report.Valid {
			b.Fatalf("Validate() = Invalid: %v", report.Errors)
		}
		// 첫 번째 반복에서만 시간 체크 (벤치마크 워밍업 제외)
		if b.N == 1 && elapsed > 100*time.Millisecond {
			b.Errorf("성능 위반: 500 토큰 검증 %v > 100ms", elapsed)
		}
	}
}
