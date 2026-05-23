//go:build integration

// Package pipeline: integration tests for Path A (Claude Design handoff bundle).
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 4 (T4-02).
//
// Verification:
// 1. Read artifacts from the testdata/path_a_handoff/ handoff bundle.
// 2. Run dtcg.Validate against tokens.json — valid samples must produce no errors.
// 3. Corrupt tokens.json must cause the validator to raise errors.
// 4. Record Path A in path-selection.json.
package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/design/dtcg"
)

// loadTestdataTokens reads testdata/path_a_handoff/tokens.json and returns it as a map.
func loadTestdataTokens(t *testing.T, filename string) map[string]any {
	t.Helper()

	// Build the testdata path relative to the test execution directory.
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

// TestPathA_HandoffBundleArtifactsExist verifies that Path A handoff bundle artifact files exist.
func TestPathA_HandoffBundleArtifactsExist(t *testing.T) {
	// Verify the reserved file paths defined in constitution §3.2.
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

// TestPathA_ValidTokensPassValidator: a valid tokens.json must pass the validator.
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

// TestPathA_CorruptTokensBlockDownstream: the validator must block a corrupt tokens.json.
func TestPathA_CorruptTokensBlockDownstream(t *testing.T) {
	// Corrupt tokens.json: unknown category + missing $value.
	corruptTokens := map[string]any{
		"color-bad": map[string]any{
			"$type":  "color",
			"$value": "not-a-hex-value", // invalid hex format
		},
		"unknown-cat": map[string]any{
			"$type":  "foobar", // unknown category
			"$value": "something",
		},
		"missing-value": map[string]any{
			"$type": "dimension",
			// $value missing
		},
	}

	report, err := dtcg.Validate(corruptTokens)
	if err != nil {
		t.Fatalf("Validate() 실행 실패: %v", err)
	}

	// Corrupt data, so Valid must be false.
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

// TestPathA_PathSelectionWrittenOnImport: after Path A processing, "A" is recorded in path-selection.json.
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

// TestPathA_ValidatorGateBeforeCodeGen: token validation -> code generation allowed flow.
// Simulates the DTCG gate pattern used by expert-frontend.
func TestPathA_ValidatorGateBeforeCodeGen(t *testing.T) {
	tokens := loadTestdataTokens(t, "tokens.json")

	// Simulate the validation flow performed by expert-frontend.
	report, err := dtcg.Validate(tokens)
	if err != nil {
		t.Fatalf("검증기 실행 실패: %v", err)
	}

	if !report.Valid {
		// Block code generation — real implementations return DTCGValidationFailure to the orchestrator.
		t.Logf("코드 생성 차단됨 (오류 %d개)", len(report.Errors))
		t.FailNow()
	}

	// Validation passed: code generation allowed.
	t.Logf("검증 통과 (%d개 토큰) — 코드 생성 허용", report.TokenCount)
}
