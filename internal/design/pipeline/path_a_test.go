//go:build integration

// Package pipeline: Path A (Claude Design 핸드오프 번들) 통합 테스트.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-02).
//
// 검증:
// 1. testdata/path_a_handoff/ 핸드오프 번들에서 아티팩트 읽기
// 2. tokens.json에 dtcg.Validate 실행 — 유효 샘플은 오류 없어야 함
// 3. corrupt tokens.json은 검증기가 오류를 발생시켜야 함
// 4. path-selection.json에 Path A 기록
package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// loadTestdataTokens: testdata/path_a_handoff/tokens.json을 읽어 map으로 반환한다.
func loadTestdataTokens(t *testing.T, filename string) map[string]any {
	t.Helper()

	// 테스트 실행 디렉토리에서 testdata 경로 구성
	path := filepath.Join("testdata", "path_a_handoff", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("testdata 읽기 실패 (%s): %v", path, err)
	}

	var tokens map[string]any
	if err := json.Unmarshal(data, &tokens); err != nil {
		t.Fatalf("tokens.json 파싱 실패: %v", err)
	}

	return tokens
}

// TestPathA_HandoffBundleArtifactsExist: Path A 핸드오프 번들 아티팩트 파일 존재 검증.
func TestPathA_HandoffBundleArtifactsExist(t *testing.T) {
	// constitution §3.2의 reserved file paths 검증
	required := []string{
		filepath.Join("testdata", "path_a_handoff", "tokens.json"),
		filepath.Join("testdata", "path_a_handoff", "components.json"),
		filepath.Join("testdata", "path_a_handoff", "assets"),
		filepath.Join("testdata", "path_a_handoff", "import-warnings.json"),
	}

	for _, path := range required {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("필수 아티팩트 없음: %s", path)
		}
	}
}

// TestPathA_ValidTokensPassValidator: 유효한 tokens.json은 검증기를 통과해야 한다.
func TestPathA_ValidTokensPassValidator(t *testing.T) {
	tokens := loadTestdataTokens(t, "tokens.json")

	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("Validate() 실행 실패: %v", err)
	}

	if !report.Valid {
		t.Errorf("유효한 tokens.json이 검증 실패. 오류 수: %d", len(report.Errors))
		for _, e := range report.Errors {
			t.Logf("  - %s", e.Error())
		}
	}

	if report.TokenCount == 0 {
		t.Error("토큰이 0개 집계됨 — tokens.json이 비어있거나 파싱 오류")
	}
}

// TestPathA_CorruptTokensBlockDownstream: 손상된 tokens.json은 검증기가 차단해야 한다.
func TestPathA_CorruptTokensBlockDownstream(t *testing.T) {
	// 손상된 tokens.json: 알 수 없는 카테고리 + $value 누락
	corruptTokens := map[string]any{
		"color-bad": map[string]any{
			"$type":  "color",
			"$value": "not-a-hex-value", // 잘못된 hex 형식
		},
		"unknown-cat": map[string]any{
			"$type":  "foobar", // 알 수 없는 카테고리
			"$value": "something",
		},
		"missing-value": map[string]any{
			"$type": "dimension",
			// $value 누락
		},
	}

	report, err := dtcg.Validate(corruptTokens)
	if err != nil {
		t.Fatalf("Validate() 실행 실패: %v", err)
	}

	// 손상된 데이터이므로 Valid는 false여야 함
	if report.Valid {
		t.Error("손상된 tokens.json이 검증을 통과함 — 다운스트림 차단이 되지 않음")
	}

	if len(report.Errors) == 0 {
		t.Error("오류가 없음 — 검증기가 손상 감지 실패")
	}

	t.Logf("감지된 오류 %d개:", len(report.Errors))
	for _, e := range report.Errors {
		t.Logf("  - %s", e.Error())
	}
}

// TestPathA_PathSelectionWrittenOnImport: Path A 처리 후 path-selection.json에 "A" 기록.
func TestPathA_PathSelectionWrittenOnImport(t *testing.T) {
	dir := t.TempDir()

	ps := PathSelection{
		Path:               "A",
		BrandContextLoaded: false,
		SpecID:             "SPEC-V3R3-TEST-001",
		Timestamp:          time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		SessionID:          "test-session-path-a",
	}

	if err := WritePathSelection(dir, ps); err != nil {
		t.Fatalf("WritePathSelection 실패: %v", err)
	}

	got, err := ReadPathSelection(dir)
	if err != nil {
		t.Fatalf("ReadPathSelection 실패: %v", err)
	}

	if got.Path != "A" {
		t.Errorf("Path = %q, want 'A'", got.Path)
	}
	if got.SpecID != ps.SpecID {
		t.Errorf("SpecID = %q, want %q", got.SpecID, ps.SpecID)
	}
}

// TestPathA_ValidatorGateBeforeCodeGen: 토큰 검증 → 코드 생성 허용 흐름 검증.
// expert-frontend의 DTCG 게이트 패턴을 시뮬레이션한다.
func TestPathA_ValidatorGateBeforeCodeGen(t *testing.T) {
	tokens := loadTestdataTokens(t, "tokens.json")

	// expert-frontend가 수행하는 검증 흐름 시뮬레이션
	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("검증기 실행 실패: %v", err)
	}

	if !report.Valid {
		// 코드 생성 차단 — 실제 구현에서는 오케스트레이터에 DTCGValidationFailure 반환
		t.Logf("코드 생성 차단됨 (오류 %d개)", len(report.Errors))
		t.FailNow()
	}

	// 검증 통과: 코드 생성 허용
	t.Logf("검증 통과 (%d개 토큰) — 코드 생성 허용", report.TokenCount)
}
