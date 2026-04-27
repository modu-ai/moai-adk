// Package dtcg: W3C Design Tokens Community Group (DTCG) 2025.10 사양 기반 토큰 검증기.
//
// 소스: https://tr.designtokens.org/format/ (Editor's Draft 2025-10)
// SPEC: SPEC-V3R3-DESIGN-PIPELINE-001 Phase 3
//
// 사용 예시:
//
//	tokens := map[string]any{
//	    "color-primary": map[string]any{
//	        "$type":  "color",
//	        "$value": "#0066ff",
//	    },
//	}
//	report, err := dtcg.Validate(tokens)
//	if err != nil { ... }
//	if !report.Valid { ... report.Errors ... }
package dtcg

import "fmt"

// ValidationError: 특정 토큰 경로에서 발생한 검증 오류.
// DTCG 2025.10 §9 검증 규칙 위반을 나타낸다.
type ValidationError struct {
	// TokenPath: 오류가 발생한 토큰 경로 (예: "group.subgroup.token-name")
	TokenPath string
	// Category: 해당 토큰의 $type 값 (예: "color", "dimension")
	Category string
	// Rule: 위반된 규칙 설명 (예: "invalid hex color format")
	Rule string
	// Value: 오류를 유발한 실제 값 (디버깅용)
	Value any
}

// Error: ValidationError 문자열 표현.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("토큰 '%s' (%s): %s (값: %v)", e.TokenPath, e.Category, e.Rule, e.Value)
	}
	return fmt.Sprintf("토큰 '%s' (%s): %s", e.TokenPath, e.Category, e.Rule)
}

// ValidationWarning: 토큰 사용을 차단하지 않지만 주의가 필요한 경고.
// 예: 브랜드 컨텍스트와 충돌하는 색상 값.
type ValidationWarning struct {
	// TokenPath: 경고가 발생한 토큰 경로
	TokenPath string
	// Category: 해당 토큰의 $type 값
	Category string
	// Message: 경고 메시지
	Message string
}

// Warning: ValidationWarning 문자열 표현.
func (w *ValidationWarning) Warning() string {
	return fmt.Sprintf("경고: 토큰 '%s' (%s): %s", w.TokenPath, w.Category, w.Message)
}

// Report: 전체 토큰 집합에 대한 검증 결과 보고서.
// DTCG 검증기가 반환하는 최상위 결과 타입.
type Report struct {
	// Valid: 모든 토큰이 오류 없이 검증을 통과했는지 여부
	Valid bool
	// Errors: 검증 오류 목록 (Valid == false일 때 최소 1개 이상)
	Errors []*ValidationError
	// Warnings: 비차단 경고 목록
	Warnings []*ValidationWarning
	// TokenCount: 검증된 토큰 수
	TokenCount int
}

// HasErrors: 검증 오류가 존재하는지 확인한다.
func (r *Report) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings: 경고가 존재하는지 확인한다.
func (r *Report) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// AddError: 보고서에 검증 오류를 추가하고 Valid를 false로 설정한다.
func (r *Report) AddError(err *ValidationError) {
	r.Errors = append(r.Errors, err)
	r.Valid = false
}

// AddWarning: 보고서에 경고를 추가한다.
func (r *Report) AddWarning(w *ValidationWarning) {
	r.Warnings = append(r.Warnings, w)
}
