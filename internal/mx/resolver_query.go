package mx

import (
	"fmt"
	"time"
)

// Query는 @MX TAG 사이드카 인덱스에 대한 구조화된 질의를 나타냅니다.
// REQ-SPC-004-001에 정의된 모든 필터를 포함합니다.
type Query struct {
	// SpecID는 태그를 특정 SPEC과 연결하는 필터입니다 (--spec 플래그).
	SpecID string

	// Kind는 태그 종류 필터입니다 (--kind 플래그).
	// 빈 문자열이면 모든 종류를 반환합니다.
	Kind TagKind

	// FanInMin은 최소 fan-in 값 필터입니다 (--fan-in-min 플래그).
	// 0이면 필터링하지 않습니다. ANCHOR 태그에만 적용됩니다.
	FanInMin int

	// Danger는 위험 카테고리 필터입니다 (--danger 플래그).
	// WARN 태그의 REASON 텍스트와 mx.yaml danger_categories: 패턴을 매칭합니다.
	Danger string

	// FilePrefix는 파일 경로 접두사 필터입니다 (--file-prefix 플래그).
	FilePrefix string

	// Since는 LastSeenAt 기준 시간 필터입니다 (--since 플래그).
	// 제로 값이면 필터링하지 않습니다.
	Since time.Time

	// Limit은 반환할 최대 태그 수입니다 (--limit 플래그, 기본 100, 최대 10000).
	Limit int

	// Offset은 페이지네이션 오프셋입니다 (--offset 플래그, 기본 0).
	Offset int

	// IncludeTests는 fan-in 계산 시 테스트 파일 참조를 포함할지 여부입니다 (--include-tests 플래그).
	// 기본값은 false (테스트 파일 참조 제외).
	IncludeTests bool
}

// DefaultLimit은 --limit 플래그의 기본값입니다 (REQ-SPC-004-007).
const DefaultLimit = 100

// MaxLimit은 --limit 플래그의 최대값입니다 (REQ-SPC-004-007).
const MaxLimit = 10000

// TagResult는 단일 @MX TAG 조회 결과를 나타냅니다.
// REQ-SPC-004-005에 정의된 JSON 스키마를 구현합니다.
type TagResult struct {
	Kind     TagKind `json:"kind"`
	File     string  `json:"file"`
	Line     int     `json:"line"`
	Body     string  `json:"body"`
	Reason   string  `json:"reason,omitempty"`
	AnchorID string  `json:"anchor_id,omitempty"`

	// CreatedBy는 태그를 생성한 주체를 나타냅니다 (에이전트 이름 또는 "human").
	CreatedBy string `json:"created_by"`

	// LastSeenAt은 이 태그가 마지막으로 스캔에서 발견된 시각입니다.
	LastSeenAt time.Time `json:"last_seen_at"`

	// FanIn은 ANCHOR 태그의 코드 참조 수입니다 (ANCHOR 전용).
	FanIn int `json:"fan_in,omitempty"`

	// FanInMethod는 fan-in 계산 방식입니다: "lsp" 또는 "textual" (ANCHOR 전용).
	FanInMethod string `json:"fan_in_method,omitempty"`

	// DangerCategory는 WARN 태그의 위험 카테고리입니다 (WARN 전용).
	DangerCategory string `json:"danger_category,omitempty"`

	// SpecAssociations는 이 태그와 연결된 SPEC ID 목록입니다.
	SpecAssociations []string `json:"spec_associations"`
}

// QueryResult는 Resolve 쿼리의 결과를 나타냅니다.
type QueryResult struct {
	// Tags는 필터에 매칭된 태그 목록입니다 (페이지네이션 적용 후).
	Tags []TagResult

	// TruncationNotice는 결과가 Limit으로 잘렸을 때 true입니다 (REQ-SPC-004-021).
	TruncationNotice bool

	// TotalCount는 페이지네이션 적용 전 총 매칭 수입니다.
	TotalCount int
}

// SidecarUnavailableError는 사이드카 인덱스를 읽을 수 없을 때 반환됩니다 (REQ-SPC-004-013).
type SidecarUnavailableError struct {
	Cause error
}

// Error implements the error interface.
func (e *SidecarUnavailableError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("SidecarUnavailable: %v — '/moai mx --full' 를 실행하여 인덱스를 재빌드하세요", e.Cause)
	}
	return "SidecarUnavailable: 사이드카 인덱스를 읽을 수 없습니다 — '/moai mx --full' 를 실행하여 인덱스를 재빌드하세요"
}

// Unwrap returns the underlying error.
func (e *SidecarUnavailableError) Unwrap() error {
	return e.Cause
}

// LSPRequiredError는 MOAI_MX_QUERY_STRICT=1이고 LSP를 사용할 수 없을 때 반환됩니다 (REQ-SPC-004-030).
type LSPRequiredError struct {
	Language string
}

// Error implements the error interface.
func (e *LSPRequiredError) Error() string {
	return fmt.Sprintf("LSPRequired: 언어 '%s'에 대한 LSP 서버가 실행 중이지 않습니다 (MOAI_MX_QUERY_STRICT=1)", e.Language)
}

// InvalidQueryError는 쿼리 파라미터가 문법적으로 잘못되었을 때 반환됩니다 (REQ-SPC-004-041).
type InvalidQueryError struct {
	Field   string
	Value   string
	Message string
}

// Error implements the error interface.
func (e *InvalidQueryError) Error() string {
	return fmt.Sprintf("InvalidQuery: 필드 '%s'의 값 '%s'이 잘못되었습니다: %s", e.Field, e.Value, e.Message)
}

// Resolve는 Query에 기반하여 @MX TAG를 조회합니다.
// 사이드카 인덱스를 읽고 모든 필터를 AND 조합으로 적용합니다 (REQ-SPC-004-042).
//
// @MX:ANCHOR: [AUTO] Resolve — Query API 진입점의 불변 계약
// @MX:REASON: fan_in >= 3 — CLI mx_query.go, 코드맵 생성 도구, 평가자(evaluator-active)에서 호출
func (r *Resolver) Resolve(query Query) (QueryResult, error) {
	// RED 단계: 미구현 stub
	// GREEN 단계에서 실제 구현으로 교체됩니다
	return QueryResult{}, fmt.Errorf("not implemented")
}
