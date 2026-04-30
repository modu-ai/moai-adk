package mx

import (
	"context"
	"fmt"
	"os"
	"strings"
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

	// fanInCounter는 fan-in 계산에 사용할 구현체입니다. nil이면 TextualFanInCounter 사용.
	fanInCounter FanInCounter

	// dangerMatcher는 위험 카테고리 매칭에 사용할 구현체입니다.
	dangerMatcher *DangerCategoryMatcher

	// specAssociator는 SPEC 연결에 사용할 구현체입니다.
	specAssociator *SpecAssociator

	// projectRoot는 fan-in 계산용 프로젝트 루트 경로입니다.
	projectRoot string
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

// validKinds는 허용된 TagKind 값 목록입니다.
var validKinds = map[TagKind]bool{
	MXNote:   true,
	MXWarn:   true,
	MXAnchor: true,
	MXTodo:   true,
	MXLegacy: true,
	"":       true, // 빈 문자열은 모든 종류를 의미
}

// resolveLimit은 쿼리의 Limit 값을 정규화합니다.
// 0 또는 음수이면 DefaultLimit, MaxLimit 초과이면 MaxLimit을 반환합니다.
func resolveLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

// Resolve는 Query에 기반하여 @MX TAG를 조회합니다.
// 사이드카 인덱스를 읽고 모든 필터를 AND 조합으로 적용합니다 (REQ-SPC-004-042).
//
// @MX:ANCHOR: [AUTO] Resolve — Query API 진입점의 불변 계약
// @MX:REASON: fan_in >= 3 — CLI mx_query.go, 코드맵 생성 도구, 평가자(evaluator-active)에서 호출
func (r *Resolver) Resolve(query Query) (QueryResult, error) {
	// 1. 쿼리 유효성 검증 (REQ-SPC-004-041)
	if err := validateQuery(query); err != nil {
		return QueryResult{}, err
	}

	// 2. 사이드카 인덱스 읽기 (REQ-SPC-004-013)
	sidecarPath := r.manager.sidecarPath
	if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
		return QueryResult{}, &SidecarUnavailableError{Cause: err}
	}

	sidecar, err := r.manager.Load()
	if err != nil {
		return QueryResult{}, &SidecarUnavailableError{Cause: err}
	}

	// 3. 위험 카테고리 매처 및 SPEC 연결자 설정
	dangerMatcher := query.dangerMatcher
	if dangerMatcher == nil {
		dangerMatcher = NewDangerCategoryMatcher(DangerCategoryConfig{})
	}

	specAssociator := query.specAssociator
	if specAssociator == nil {
		specAssociator = NewSpecAssociator(map[string][]string{})
	}

	fanInCounter := query.fanInCounter
	if fanInCounter == nil {
		fanInCounter = &TextualFanInCounter{}
	}

	// 4. MOAI_MX_QUERY_STRICT 확인 (REQ-SPC-004-030)
	strictMode := os.Getenv("MOAI_MX_QUERY_STRICT") == "1"
	needsFanIn := query.FanInMin > 0
	if strictMode && needsFanIn {
		// strict 모드에서는 LSP가 필요하지만 항상 textual fallback이 제공됨
		// 실제 LSP 가용성 확인은 fanInCounter 구현에 위임
		// 여기서는 LSPRequired 오류 반환 (LSP 클라이언트 없음)
		return QueryResult{}, &LSPRequiredError{Language: "unknown"}
	}

	// 5. 모든 태그에 필터 적용 (AND 조합, REQ-SPC-004-042)
	limit := resolveLimit(query.Limit)

	var matched []TagResult
	for _, tag := range sidecar.Tags {
		tagResult, ok := applyFilters(tag, query, dangerMatcher, specAssociator)
		if !ok {
			continue
		}

		// fan-in 계산 (ANCHOR이고 FanInMin이 설정된 경우)
		if tag.Kind == MXAnchor && needsFanIn {
			count, method, err := fanInCounter.Count(context.Background(), tag, query.projectRoot, !query.IncludeTests)
			if err != nil {
				// fan-in 계산 실패 시 0으로 처리
				count = 0
				method = "textual"
			}
			tagResult.FanIn = count
			tagResult.FanInMethod = method

			// fan-in 최소값 필터
			if tagResult.FanIn < query.FanInMin {
				continue
			}
		}

		matched = append(matched, tagResult)
	}

	totalCount := len(matched)

	// 6. 페이지네이션 적용 (REQ-SPC-004-007)
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var paginated []TagResult
	if offset >= len(matched) {
		paginated = []TagResult{}
	} else {
		end := offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		paginated = matched[offset:end]
	}

	// 7. 빈 결과 처리 (REQ-SPC-004-041: nil이 아닌 빈 슬라이스)
	if paginated == nil {
		paginated = []TagResult{}
	}

	truncationNotice := totalCount > limit+offset

	return QueryResult{
		Tags:             paginated,
		TruncationNotice: truncationNotice,
		TotalCount:       totalCount,
	}, nil
}

// validateQuery는 쿼리 파라미터의 유효성을 검증합니다 (REQ-SPC-004-041).
func validateQuery(query Query) error {
	if query.Kind != "" && !validKinds[query.Kind] {
		return &InvalidQueryError{
			Field:   "kind",
			Value:   string(query.Kind),
			Message: fmt.Sprintf("허용 값: note, warn, anchor, todo, legacy (실제: %s)", query.Kind),
		}
	}
	return nil
}

// applyFilters는 단일 태그에 모든 필터를 AND 조합으로 적용합니다.
// 태그가 모든 필터를 통과하면 TagResult와 true를 반환합니다.
func applyFilters(tag Tag, query Query, dangerMatcher *DangerCategoryMatcher, specAssociator *SpecAssociator) (TagResult, bool) {
	// KIND 필터 (REQ-SPC-004-001)
	if query.Kind != "" && tag.Kind != query.Kind {
		return TagResult{}, false
	}

	// 파일 접두사 필터 (REQ-SPC-004-001)
	if query.FilePrefix != "" && !strings.HasPrefix(tag.File, query.FilePrefix) {
		return TagResult{}, false
	}

	// Since 시간 필터 (REQ-SPC-004-001)
	if !query.Since.IsZero() && tag.LastSeenAt.Before(query.Since) {
		return TagResult{}, false
	}

	// SPEC 연결 및 필터 (REQ-SPC-004-006, REQ-SPC-004-010)
	specAssociations := specAssociator.Associate(tag)
	if query.SpecID != "" {
		found := false
		for _, spec := range specAssociations {
			if spec == query.SpecID {
				found = true
				break
			}
		}
		if !found {
			return TagResult{}, false
		}
	}

	// Danger 카테고리 필터 (REQ-SPC-004-012, WARN 전용)
	var dangerCategory string
	if query.Danger != "" {
		if tag.Kind != MXWarn {
			return TagResult{}, false
		}
		if !dangerMatcher.Match(tag.Reason, query.Danger) {
			return TagResult{}, false
		}
		dangerCategory = query.Danger
	} else if tag.Kind == MXWarn && tag.Reason != "" {
		// Danger 필터 없어도 카테고리 자동 설정
		dangerCategory = dangerMatcher.CategoryOf(tag.Reason)
	}

	// TagResult 구성 (REQ-SPC-004-005)
	if specAssociations == nil {
		specAssociations = []string{}
	}

	return TagResult{
		Kind:             tag.Kind,
		File:             tag.File,
		Line:             tag.Line,
		Body:             tag.Body,
		Reason:           tag.Reason,
		AnchorID:         tag.AnchorID,
		CreatedBy:        tag.CreatedBy,
		LastSeenAt:       tag.LastSeenAt,
		DangerCategory:   dangerCategory,
		SpecAssociations: specAssociations,
	}, true
}

// FormatMarkdown은 QueryResult를 마크다운 테이블로 변환합니다 (REQ-SPC-004-031).
func FormatMarkdown(result QueryResult) string {
	var sb strings.Builder

	if result.TruncationNotice {
		sb.WriteString(fmt.Sprintf("> **TruncationNotice**: 전체 %d개 중 %d개만 표시됩니다.\n\n", result.TotalCount, len(result.Tags)))
	}

	sb.WriteString("| Kind | File | Line | Body | FanIn | Danger | SPECs |\n")
	sb.WriteString("|------|------|------|------|-------|--------|-------|\n")

	for _, tag := range result.Tags {
		specs := strings.Join(tag.SpecAssociations, ", ")
		fanIn := ""
		if tag.FanIn > 0 || tag.FanInMethod != "" {
			fanIn = fmt.Sprintf("%d (%s)", tag.FanIn, tag.FanInMethod)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s | %s | %s |\n",
			tag.Kind, tag.File, tag.Line,
			truncateStr(tag.Body, 40),
			fanIn, tag.DangerCategory, specs))
	}

	return sb.String()
}

// FormatTable은 QueryResult를 사람이 읽을 수 있는 텍스트 테이블로 변환합니다 (REQ-SPC-004-004).
func FormatTable(result QueryResult) string {
	if len(result.Tags) == 0 {
		return "(결과 없음)\n"
	}

	var sb strings.Builder

	if result.TruncationNotice {
		sb.WriteString(fmt.Sprintf("TruncationNotice: 전체 %d개 중 %d개만 표시됩니다.\n\n", result.TotalCount, len(result.Tags)))
	}

	// 컬럼 너비 계산
	header := fmt.Sprintf("%-8s %-50s %5s %-40s\n", "KIND", "FILE", "LINE", "BODY")
	separator := strings.Repeat("-", len(header)) + "\n"

	sb.WriteString(header)
	sb.WriteString(separator)

	for _, tag := range result.Tags {
		sb.WriteString(fmt.Sprintf("%-8s %-50s %5d %-40s\n",
			tag.Kind,
			truncateStr(tag.File, 50),
			tag.Line,
			truncateStr(tag.Body, 40)))
	}

	return sb.String()
}

// truncateStr은 문자열이 maxLen을 초과하면 잘라서 "..."를 붙입니다.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
