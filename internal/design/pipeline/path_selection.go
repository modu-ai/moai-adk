// Package pipeline: /moai design 워크플로우 경로 선택 지속성 레이어.
// SPEC-V3R3-DESIGN-PIPELINE-001 Phase 2 (T2-02) 구현.
//
// PathSelection은 사용자가 선택한 설계 경로(A/B1/B2)와 관련 메타데이터를
// `.moai/design/path-selection.json`에 저장하고 읽어오는 기능을 제공한다.
// JSON 필드 순서는 struct 정의 순서대로 고정되어 동일 입력 → 동일 bytes 보장.
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PathSelectionFile: path-selection.json 파일 이름 (`.moai/design/` 하위 경로).
const PathSelectionFile = "path-selection.json"

// PathSelection: /moai design 워크플로우 경로 선택 데이터.
// JSON 태그 순서가 직렬화 필드 순서를 결정한다 (map 미사용 → 안정적 ordering).
type PathSelection struct {
	// Path: 선택된 설계 경로 ("A" | "B1" | "B2")
	Path string `json:"path"`
	// BrandContextLoaded: 브랜드 컨텍스트(.moai/project/brand/) 로드 여부
	BrandContextLoaded bool `json:"brand_context_loaded"`
	// SpecID: 연관된 SPEC 식별자 (예: "SPEC-V3R3-DESIGN-001")
	SpecID string `json:"spec_id"`
	// Timestamp: 경로 선택 시각 (RFC3339 UTC)
	Timestamp time.Time `json:"ts"`
	// SessionID: 선택이 이루어진 세션 식별자
	SessionID string `json:"session_id"`
}

// MissingFieldError: 필수 필드 누락 시 반환되는 구조화된 에러.
type MissingFieldError struct {
	Field string
}

// Error: MissingFieldError 문자열 표현.
func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("missing required field: %s", e.Field)
}

// WritePathSelection: PathSelection을 dir/PathSelectionFile에 JSON으로 저장한다.
// json.MarshalIndent를 사용하여 struct 필드 순서대로 안정적으로 직렬화한다.
// dir이 존재하지 않으면 MkdirAll로 생성한다.
func WritePathSelection(dir string, ps PathSelection) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패 (%s): %w", dir, err)
	}

	// Timestamp를 UTC로 정규화하여 결정론적 직렬화 보장
	ps.Timestamp = ps.Timestamp.UTC()

	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return fmt.Errorf("PathSelection JSON 직렬화 실패: %w", err)
	}

	dst := filepath.Join(dir, PathSelectionFile)
	if err := os.WriteFile(dst, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("path-selection.json 쓰기 실패 (%s): %w", dst, err)
	}

	return nil
}

// ReadPathSelection: dir/PathSelectionFile을 읽어 PathSelection을 반환한다.
//
// 에러 조건:
//   - 파일 없음: os.ErrNotExist 포함 에러
//   - 손상된 JSON: json 파싱 에러 (컨텍스트 포함)
//   - "path" 필드 누락 또는 빈 값: *MissingFieldError{Field: "path"}
func ReadPathSelection(dir string) (PathSelection, error) {
	src := filepath.Join(dir, PathSelectionFile)

	data, err := os.ReadFile(src)
	if err != nil {
		return PathSelection{}, fmt.Errorf("path-selection.json 읽기 실패 (%s): %w", src, err)
	}

	var ps PathSelection
	if err := json.Unmarshal(data, &ps); err != nil {
		return PathSelection{}, fmt.Errorf("path-selection.json 파싱 실패: %w", err)
	}

	// 필수 필드 검증: path
	if ps.Path == "" {
		return PathSelection{}, &MissingFieldError{Field: "path"}
	}

	return ps, nil
}
