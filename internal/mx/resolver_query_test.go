package mx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildTestSidecar는 테스트용 사이드카 파일을 지정된 디렉토리에 생성합니다.
func buildTestSidecar(t *testing.T, stateDir string, tags []Tag) *Manager {
	t.Helper()
	mgr := NewManager(stateDir)
	sidecar := &Sidecar{
		SchemaVersion: SchemaVersion,
		Tags:          tags,
		ScannedAt:     time.Now(),
	}
	if err := mgr.Write(sidecar); err != nil {
		t.Fatalf("사이드카 쓰기 실패: %v", err)
	}
	return mgr
}

// makeTag는 테스트용 태그를 생성하는 헬퍼입니다.
func makeTag(kind TagKind, file, body string, line int) Tag {
	return Tag{
		Kind:       kind,
		File:       file,
		Line:       line,
		Body:       body,
		CreatedBy:  "test",
		LastSeenAt: time.Now(),
	}
}

// makeAnchorTag는 AnchorID가 있는 테스트용 ANCHOR 태그를 생성합니다.
func makeAnchorTag(file, anchorID, body string, line int) Tag {
	t := makeTag(MXAnchor, file, body, line)
	t.AnchorID = anchorID
	return t
}

// makeWarnTag는 Reason이 있는 테스트용 WARN 태그를 생성합니다.
func makeWarnTag(file, body, reason string, line int) Tag {
	t := makeTag(MXWarn, file, body, line)
	t.Reason = reason
	return t
}

// TestResolve_SpecAndKindFilter는 SPEC+KIND 복합 필터링을 테스트합니다.
// AC-SPC-004-01: 20개 태그, 2개 SPEC 중 SPEC-X-001의 ANCHOR만 반환
func TestResolve_SpecAndKindFilter(t *testing.T) {
	// AC-SPC-004-01: SPEC 필터 + KIND 필터 조합
	stateDir := t.TempDir()

	// SPEC-X-001에 속하는 ANCHOR 태그
	anchor1 := makeAnchorTag("internal/auth/handler.go", "anchor-auth-handler", "ANCHOR for SPEC-X-001", 10)
	anchor2 := makeAnchorTag("internal/auth/middleware.go", "anchor-auth-mw", "ANCHOR for SPEC-X-001", 20)

	// SPEC-Y-002에 속하는 태그들 (필터에서 제외되어야 함)
	note1 := makeTag(MXNote, "internal/cache/store.go", "캐시 정책 설명 — SPEC-Y-002", 5)
	anchor3 := makeAnchorTag("internal/cache/store.go", "anchor-cache", "ANCHOR for SPEC-Y-002", 15)

	// SPEC-X-001의 NOTE 태그 (KIND 필터로 제외되어야 함)
	note2 := makeTag(MXNote, "internal/auth/handler.go", "인증 흐름 설명 — SPEC-X-001", 8)

	mgr := buildTestSidecar(t, stateDir, []Tag{anchor1, anchor2, note1, anchor3, note2})
	resolver := NewResolver(mgr)

	query := Query{
		SpecID: "SPEC-X-001",
		Kind:   MXAnchor,
		Limit:  DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) != 2 {
		t.Errorf("태그 수: 기대 2, 실제 %d", len(result.Tags))
	}

	for _, tag := range result.Tags {
		if tag.Kind != MXAnchor {
			t.Errorf("KIND 필터 실패: %s (ANCHOR 기대)", tag.Kind)
		}
		found := false
		for _, spec := range tag.SpecAssociations {
			if spec == "SPEC-X-001" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SPEC 연결 누락: %v에 SPEC-X-001 없음", tag.SpecAssociations)
		}
	}
}

// TestResolve_FanInFilter는 fan-in 최소값 필터를 테스트합니다.
// AC-SPC-004-02: fan_in 1,2,3,5,10 중 fan-in-min=3 이상인 3개만 반환
func TestResolve_FanInFilter(t *testing.T) {
	// AC-SPC-004-02: fan-in 필터
	stateDir := t.TempDir()

	anchors := []Tag{
		makeAnchorTag("pkg/a.go", "anchor-a", "low fan-in 1", 1),
		makeAnchorTag("pkg/b.go", "anchor-b", "low fan-in 2", 1),
		makeAnchorTag("pkg/c.go", "anchor-c", "medium fan-in 3", 1),
		makeAnchorTag("pkg/d.go", "anchor-d", "high fan-in 5", 1),
		makeAnchorTag("pkg/e.go", "anchor-e", "high fan-in 10", 1),
	}

	mgr := buildTestSidecar(t, stateDir, anchors)
	resolver := NewResolver(mgr)

	// fan-in 계산을 위한 mock counter 주입
	// 실제 구현에서는 WithFanInCounter 옵션으로 주입
	query := Query{
		Kind:     MXAnchor,
		FanInMin: 3,
		Limit:    DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// 최소 구현에서는 mock fan-in 값 사용 (GREEN 단계에서 실제 로직 구현)
	// RED 단계에서는 이 테스트가 실패해야 합니다
	for _, tag := range result.Tags {
		if tag.FanIn < 3 {
			t.Errorf("fan-in 필터 실패: %s의 fan-in=%d (최소 3 기대)", tag.AnchorID, tag.FanIn)
		}
	}
}

// TestResolve_DangerFilter는 위험 카테고리 필터를 테스트합니다.
// AC-SPC-004-03: "goroutine leak" REASON을 가진 WARN이 concurrency로 반환
func TestResolve_DangerFilter(t *testing.T) {
	// AC-SPC-004-03: 위험 카테고리 필터
	stateDir := t.TempDir()

	warnConcurrency := makeWarnTag("internal/worker.go", "무제한 고루틴 생성", "goroutine leak on panic", 15)
	warnOther := makeWarnTag("internal/db.go", "DB 연결 미반환", "missing Close on error path", 30)
	note1 := makeTag(MXNote, "internal/misc.go", "일반 노트", 5)

	mgr := buildTestSidecar(t, stateDir, []Tag{warnConcurrency, warnOther, note1})
	resolver := NewResolver(mgr)

	query := Query{
		Danger: "concurrency",
		Limit:  DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) != 1 {
		t.Errorf("태그 수: 기대 1 (concurrency만), 실제 %d", len(result.Tags))
	}

	if len(result.Tags) > 0 {
		if result.Tags[0].DangerCategory != "concurrency" {
			t.Errorf("위험 카테고리: 기대 concurrency, 실제 %s", result.Tags[0].DangerCategory)
		}
	}
}

// TestResolve_SidecarUnavailable은 사이드카 파일 없을 때 오류를 테스트합니다.
// AC-SPC-004-04: 사이드카 파일 없을 때 SidecarUnavailable 오류 반환
func TestResolve_SidecarUnavailable(t *testing.T) {
	// AC-SPC-004-04: 사이드카 파일 없을 때
	stateDir := t.TempDir()
	// 사이드카 파일 생성하지 않음 (디렉토리만 존재)
	// 그러나 매니저는 Load()에서 빈 사이드카를 반환함
	// 실제 오류는 파일이 없고 명시적 확인이 있을 때

	// 존재하지 않는 디렉토리 사용하여 읽기 불가 상황 시뮬레이션
	nonExistentDir := filepath.Join(stateDir, "nonexistent_mx_state")
	mgr := NewManager(nonExistentDir)

	// 사이드카 파일이 없을 때 Resolve가 SidecarUnavailableError를 반환해야 함
	// Resolve 구현이 사이드카 존재 여부를 명시적으로 확인해야 함
	resolver := NewResolver(mgr)
	query := Query{Limit: DefaultLimit}

	_, err := resolver.Resolve(query)
	if err == nil {
		t.Error("사이드카 없을 때 오류 기대, 실제 nil 반환")
		return
	}

	var sidecarErr *SidecarUnavailableError
	errStr := err.Error()
	if !strings.Contains(errStr, "SidecarUnavailable") {
		// SidecarUnavailableError가 아닌 경우도 확인
		_ = sidecarErr
		t.Errorf("오류 메시지에 'SidecarUnavailable' 없음: %v", err)
	}
}

// TestResolve_JSONSchema는 JSON 스키마 필드를 테스트합니다.
// AC-SPC-004-05: JSON 출력이 REQ-SPC-004-005 스키마를 준수해야 함
func TestResolve_JSONSchema(t *testing.T) {
	// AC-SPC-004-05: JSON 스키마 검증
	stateDir := t.TempDir()

	anchor := makeAnchorTag("internal/auth.go", "anchor-test", "테스트 앵커", 10)
	anchor.Reason = "fan_in >= 3 이유"

	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	query := Query{Kind: MXAnchor, Limit: DefaultLimit}
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) == 0 {
		t.Fatal("결과 태그 없음")
	}

	// JSON 직렬화 및 역직렬화 검증
	data, err := json.Marshal(result.Tags[0])
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON 역직렬화 실패: %v", err)
	}

	// REQ-SPC-004-005 필수 필드 확인
	requiredFields := []string{"kind", "file", "line", "body", "created_by", "last_seen_at", "spec_associations"}
	for _, field := range requiredFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("JSON 스키마 누락 필드: %s", field)
		}
	}
}

// TestResolve_Pagination은 페이지네이션을 테스트합니다.
// AC-SPC-004-08: 10000개 태그에서 limit 100 적용 및 TruncationNotice
func TestResolve_Pagination(t *testing.T) {
	// AC-SPC-004-08: 페이지네이션 및 TruncationNotice
	stateDir := t.TempDir()

	// 200개 태그 생성 (DefaultLimit보다 많음)
	tags := make([]Tag, 200)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트 태그", i+1)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// limit 명시하지 않음 → DefaultLimit(100) 적용 기대
	query := Query{Limit: 0} // 0이면 기본값 적용
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) > DefaultLimit {
		t.Errorf("Limit 초과: %d개 반환 (최대 %d 기대)", len(result.Tags), DefaultLimit)
	}

	if result.TotalCount <= DefaultLimit && result.TruncationNotice {
		t.Error("TruncationNotice가 필요없는데 true")
	}
}

// TestResolve_TextualFallbackAnnotation은 LSP 없을 때 textual 폴백 어노테이션을 테스트합니다.
// AC-SPC-004-07: LSP 없는 Python 파일에서 fan_in_method: "textual" 어노테이션
func TestResolve_TextualFallbackAnnotation(t *testing.T) {
	// AC-SPC-004-07: textual 폴백 시 fan_in_method 어노테이션
	stateDir := t.TempDir()

	anchor := makeAnchorTag("internal/py/worker.py", "anchor-py-worker", "파이썬 워커", 5)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	query := Query{
		Kind:       MXAnchor,
		FanInMin:   2,
		FilePrefix: "internal/py/",
		Limit:      DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	for _, tag := range result.Tags {
		if tag.FanInMethod != "textual" && tag.FanInMethod != "lsp" {
			t.Errorf("fan_in_method 값 오류: %q (lsp 또는 textual 기대)", tag.FanInMethod)
		}
	}
}

// TestResolve_StrictMode는 MOAI_MX_QUERY_STRICT=1일 때 LSP 요구를 테스트합니다.
// AC-SPC-004-09: strict 모드에서 LSP 없으면 LSPRequired 오류
func TestResolve_StrictMode(t *testing.T) {
	// AC-SPC-004-09: strict 모드 테스트
	t.Setenv("MOAI_MX_QUERY_STRICT", "1")

	stateDir := t.TempDir()
	anchor := makeAnchorTag("internal/auth.go", "anchor-auth", "인증 핸들러", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	query := Query{
		Kind:     MXAnchor,
		FanInMin: 3,
		Limit:    DefaultLimit,
	}

	_, err := resolver.Resolve(query)
	if err == nil {
		t.Error("strict 모드에서 LSP 없을 때 오류 기대, 실제 nil")
		return
	}

	if !strings.Contains(err.Error(), "LSPRequired") {
		t.Errorf("오류 메시지에 'LSPRequired' 없음: %v", err)
	}
}

// TestResolve_TestFixtureExclusion은 테스트 픽스처 참조 제외를 테스트합니다.
// AC-SPC-004-11: include-tests 없을 때 테스트 파일 참조 제외
func TestResolve_TestFixtureExclusion(t *testing.T) {
	// AC-SPC-004-11: 테스트 파일 참조 제외
	stateDir := t.TempDir()

	anchor := makeAnchorTag("tests/fixtures/mock_handler_test.go", "anchor-mock-handler", "핸들러 목", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	// include-tests=false (기본값)
	query := Query{
		Kind:         MXAnchor,
		FanInMin:     1,
		IncludeTests: false,
		Limit:        DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// 테스트 픽스처 파일의 fan-in에서 테스트 참조가 제외되어야 함
	// (실제 fan-in 값이 테스트 참조를 포함하지 않음)
	_ = result // GREEN 단계에서 구체적인 검증 추가
}

// TestResolve_EmptyResult는 매칭 없을 때 빈 배열과 exit 0을 테스트합니다.
// AC-SPC-004-12: 매칭 없을 때 [] 반환 및 exit 0
func TestResolve_EmptyResult(t *testing.T) {
	// AC-SPC-004-12: 빈 결과 처리
	stateDir := t.TempDir()

	note := makeTag(MXNote, "internal/misc.go", "노트 태그", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{note})
	resolver := NewResolver(mgr)

	// ANCHOR 필터 → NOTE만 있으므로 0개 결과
	query := Query{
		Kind:  MXAnchor,
		Limit: DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("빈 결과 시 오류 없어야 함: %v", err)
	}

	if result.Tags == nil {
		t.Error("빈 결과 시 nil이 아닌 빈 슬라이스 기대")
	}

	if len(result.Tags) != 0 {
		t.Errorf("빈 결과 기대, 실제 %d개", len(result.Tags))
	}
}

// TestResolve_CombinedFilters는 여러 필터의 AND 조합을 테스트합니다.
// AC-SPC-004-14: --spec X --kind anchor --fan-in-min 3 조합 필터
func TestResolve_CombinedFilters(t *testing.T) {
	// AC-SPC-004-14: 복합 AND 필터
	stateDir := t.TempDir()

	// SPEC-X-001에 속하는 높은 fan-in ANCHOR
	highFanInAnchor := makeAnchorTag("internal/auth/handler.go", "anchor-high", "고fan-in ANCHOR for SPEC-X-001", 10)
	// SPEC-X-001에 속하는 낮은 fan-in ANCHOR
	lowFanInAnchor := makeAnchorTag("internal/auth/util.go", "anchor-low", "저fan-in ANCHOR for SPEC-X-001", 20)
	// SPEC-Y-002에 속하는 높은 fan-in ANCHOR
	otherSpecAnchor := makeAnchorTag("internal/cache/store.go", "anchor-other", "고fan-in ANCHOR for SPEC-Y-002", 5)

	mgr := buildTestSidecar(t, stateDir, []Tag{highFanInAnchor, lowFanInAnchor, otherSpecAnchor})
	resolver := NewResolver(mgr)

	query := Query{
		SpecID:   "SPEC-X-001",
		Kind:     MXAnchor,
		FanInMin: 3,
		Limit:    DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// 세 조건을 모두 만족하는 태그만 반환되어야 함
	for _, tag := range result.Tags {
		if tag.Kind != MXAnchor {
			t.Errorf("KIND 조건 미충족: %s", tag.Kind)
		}
		if tag.FanIn < 3 {
			t.Errorf("fan-in 조건 미충족: %d < 3", tag.FanIn)
		}
		found := false
		for _, spec := range tag.SpecAssociations {
			if spec == "SPEC-X-001" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SPEC 연결 조건 미충족: SPEC-X-001 없음")
		}
	}
}

// TestResolve_SpecAssociationFromBody는 태그 본문에서 SPEC ID 추출을 테스트합니다.
// AC-SPC-004-15: @MX body "ANCHOR for SPEC-AUTH-001 handler" → spec_associations에 SPEC-AUTH-001
func TestResolve_SpecAssociationFromBody(t *testing.T) {
	// AC-SPC-004-15: 태그 본문 SPEC ID 추출
	stateDir := t.TempDir()

	anchor := makeAnchorTag("internal/auth/handler.go", "anchor-auth-handler", "ANCHOR for SPEC-AUTH-001 handler", 10)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	query := Query{
		Kind:  MXAnchor,
		Limit: DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) == 0 {
		t.Fatal("결과 태그 없음")
	}

	found := false
	for _, spec := range result.Tags[0].SpecAssociations {
		if spec == "SPEC-AUTH-001" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("spec_associations에 SPEC-AUTH-001 없음: %v", result.Tags[0].SpecAssociations)
	}
}

// TestResolve_MarkdownFormat는 마크다운 출력 형식을 테스트합니다.
// AC-SPC-004-10: --format markdown 마크다운 테이블 출력
func TestResolve_MarkdownFormat(t *testing.T) {
	// AC-SPC-004-10: 마크다운 출력 형식 확인
	stateDir := t.TempDir()
	note := makeTag(MXNote, "internal/misc.go", "노트 태그", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{note})
	resolver := NewResolver(mgr)

	query := Query{Kind: MXNote, Limit: DefaultLimit}
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// FormatMarkdown 함수가 있어야 함
	md := FormatMarkdown(result)
	if !strings.Contains(md, "|") {
		t.Error("마크다운 테이블 형식 (|) 없음")
	}
}

// TestResolve_TableFormat는 테이블 출력 형식을 테스트합니다.
// AC-SPC-004-06: --format table 사람이 읽을 수 있는 컬럼 형식 출력
func TestResolve_TableFormat(t *testing.T) {
	// AC-SPC-004-06: 테이블 출력 형식 확인
	stateDir := t.TempDir()
	note := makeTag(MXNote, "internal/misc.go", "노트 태그", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{note})
	resolver := NewResolver(mgr)

	query := Query{Kind: MXNote, Limit: DefaultLimit}
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// FormatTable 함수가 있어야 함
	table := FormatTable(result)
	if table == "" {
		t.Error("테이블 출력이 비어있음")
	}
}

// TestExtractSpecIDs는 태그 본문에서 SPEC ID 추출을 단위 테스트합니다.
func TestExtractSpecIDs(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []string
	}{
		{
			name:     "단순 SPEC ID",
			body:     "ANCHOR for SPEC-AUTH-001 handler",
			expected: []string{"SPEC-AUTH-001"},
		},
		{
			name:     "여러 SPEC ID",
			body:     "SPEC-AUTH-001과 SPEC-DB-002를 위한 앵커",
			expected: []string{"SPEC-AUTH-001", "SPEC-DB-002"},
		},
		{
			name:     "SPEC ID 없음",
			body:     "일반 설명 텍스트",
			expected: []string{},
		},
		{
			name:     "중복 SPEC ID",
			body:     "SPEC-AUTH-001 참조: SPEC-AUTH-001 처리",
			expected: []string{"SPEC-AUTH-001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractSpecIDs(tt.body)
			if len(got) != len(tt.expected) {
				t.Errorf("SPEC ID 수: 기대 %d, 실제 %d (got=%v)", len(tt.expected), len(got), got)
				return
			}
			seen := make(map[string]bool)
			for _, id := range got {
				seen[id] = true
			}
			for _, id := range tt.expected {
				if !seen[id] {
					t.Errorf("누락된 SPEC ID: %s (got=%v)", id, got)
				}
			}
		})
	}
}

// TestFormatMarkdown_TableStructure는 마크다운 출력이 테이블 구조를 갖는지 확인합니다.
func TestFormatMarkdown_TableStructure(t *testing.T) {
	result := QueryResult{
		Tags: []TagResult{
			{
				Kind:             MXAnchor,
				File:             "internal/auth.go",
				Line:             10,
				Body:             "인증 핸들러",
				AnchorID:         "anchor-auth",
				CreatedBy:        "agent",
				LastSeenAt:       time.Now(),
				FanIn:            3,
				FanInMethod:      "lsp",
				SpecAssociations: []string{"SPEC-AUTH-001"},
			},
		},
		TruncationNotice: false,
		TotalCount:       1,
	}

	md := FormatMarkdown(result)

	// 마크다운 테이블 헤더 확인
	if !strings.Contains(md, "Kind") {
		t.Error("마크다운 헤더에 'Kind' 없음")
	}
	if !strings.Contains(md, "|") {
		t.Error("마크다운 테이블 구분자 '|' 없음")
	}
}

// TestFormatJSON은 JSON 출력이 올바른 형식인지 확인합니다.
func TestFormatJSON(t *testing.T) {
	result := QueryResult{
		Tags: []TagResult{
			{
				Kind:             MXNote,
				File:             "internal/misc.go",
				Line:             1,
				Body:             "노트",
				CreatedBy:        "agent",
				LastSeenAt:       time.Now(),
				SpecAssociations: []string{},
			},
		},
	}

	data, err := json.Marshal(result.Tags)
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}

	var decoded []map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON 역직렬화 실패: %v", err)
	}

	if len(decoded) != 1 {
		t.Errorf("JSON 배열 크기: 기대 1, 실제 %d", len(decoded))
	}
}

// FormatMarkdown은 QueryResult를 마크다운 테이블로 변환합니다 (REQ-SPC-004-031).
// AC-SPC-004-10 검증용 — 실제 구현은 resolver_query.go에 있어야 함.
func FormatMarkdown(result QueryResult) string {
	// RED 단계: 미구현 stub
	// GREEN 단계에서 실제 구현으로 교체됩니다
	return ""
}

// FormatTable은 QueryResult를 사람이 읽을 수 있는 테이블 형식으로 변환합니다 (REQ-SPC-004-004).
// AC-SPC-004-06 검증용 — 실제 구현은 resolver_query.go에 있어야 함.
func FormatTable(result QueryResult) string {
	// RED 단계: 미구현 stub
	// GREEN 단계에서 실제 구현으로 교체됩니다
	return ""
}

// TestResolve_LargeSidecarTruncation은 대규모 사이드카에서 자동 limit 적용을 테스트합니다.
// AC-SPC-004-08의 추가 검증
func TestResolve_LargeSidecarTruncation(t *testing.T) {
	stateDir := t.TempDir()

	// 10000개 이상 태그 (성능 테스트는 아님, 동작 테스트)
	// 실제 10000개는 시간이 걸리므로 200개로 테스트
	tags := make([]Tag, 200)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트", i+1)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// Limit=0이면 DefaultLimit 적용
	query := Query{Limit: 0}
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) > DefaultLimit {
		t.Errorf("자동 limit 미적용: %d개 반환 (최대 %d 기대)", len(result.Tags), DefaultLimit)
	}

	if result.TotalCount > DefaultLimit && !result.TruncationNotice {
		t.Error("TruncationNotice가 true여야 함")
	}
}

// TestResolve_InvalidKind는 잘못된 kind 값에 대한 오류를 테스트합니다.
// AC-SPC-004-13: --kind nonexistent → exit 2 + InvalidQuery 오류
func TestResolve_InvalidKind(t *testing.T) {
	// AC-SPC-004-13: 잘못된 필터 값 처리
	stateDir := t.TempDir()
	mgr := buildTestSidecar(t, stateDir, []Tag{})
	resolver := NewResolver(mgr)

	// 잘못된 kind 값
	query := Query{
		Kind:  TagKind("nonexistent"),
		Limit: DefaultLimit,
	}

	_, err := resolver.Resolve(query)
	if err == nil {
		t.Error("잘못된 kind에 대해 오류 기대, 실제 nil")
		return
	}

	if !strings.Contains(err.Error(), "InvalidQuery") {
		t.Errorf("오류 메시지에 'InvalidQuery' 없음: %v", err)
	}
}

// TestResolve_FilePrefix는 파일 경로 접두사 필터를 테스트합니다.
func TestResolve_FilePrefix(t *testing.T) {
	stateDir := t.TempDir()

	// 다양한 경로의 태그들
	tags := []Tag{
		makeTag(MXNote, "internal/auth/handler.go", "인증 핸들러", 1),
		makeTag(MXNote, "internal/cache/store.go", "캐시 저장소", 1),
		makeTag(MXNote, "pkg/utils/helper.go", "유틸리티", 1),
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	query := Query{
		FilePrefix: "internal/auth/",
		Limit:      DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	for _, tag := range result.Tags {
		if !strings.HasPrefix(tag.File, "internal/auth/") {
			t.Errorf("파일 접두사 필터 실패: %s", tag.File)
		}
	}
}

// TestResolve_OffsetPagination은 offset 기반 페이지네이션을 테스트합니다.
func TestResolve_OffsetPagination(t *testing.T) {
	stateDir := t.TempDir()

	// 10개 태그 생성
	tags := make([]Tag, 10)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트", i+1)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// 첫 페이지
	q1 := Query{Limit: 5, Offset: 0}
	r1, err := resolver.Resolve(q1)
	if err != nil {
		t.Fatalf("첫 페이지 오류: %v", err)
	}

	// 두 번째 페이지
	q2 := Query{Limit: 5, Offset: 5}
	r2, err := resolver.Resolve(q2)
	if err != nil {
		t.Fatalf("두 번째 페이지 오류: %v", err)
	}

	// 두 페이지 합산이 전체 태그 수와 같아야 함
	if len(r1.Tags)+len(r2.Tags) != 10 {
		t.Errorf("페이지네이션 합산: 기대 10, 실제 %d+%d=%d",
			len(r1.Tags), len(r2.Tags), len(r1.Tags)+len(r2.Tags))
	}
}

// TestResolveAnchor는 기존 ResolveAnchor(anchorID) API가 동작하는지 확인합니다.
func TestResolveAnchor(t *testing.T) {
	stateDir := t.TempDir()

	anchor := makeAnchorTag("internal/auth.go", "anchor-auth-001", "인증 핸들러", 10)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	tag, err := resolver.ResolveAnchor("anchor-auth-001")
	if err != nil {
		t.Fatalf("ResolveAnchor 오류: %v", err)
	}

	if tag.AnchorID != "anchor-auth-001" {
		t.Errorf("AnchorID: 기대 anchor-auth-001, 실제 %s", tag.AnchorID)
	}
}

// TestResolve_SinceFilter는 since 시간 필터를 테스트합니다.
func TestResolve_SinceFilter(t *testing.T) {
	stateDir := t.TempDir()

	now := time.Now()
	old := now.Add(-48 * time.Hour) // 2일 전

	oldTag := makeTag(MXNote, "internal/old.go", "오래된 태그", 1)
	oldTag.LastSeenAt = old

	newTag := makeTag(MXNote, "internal/new.go", "새 태그", 1)
	newTag.LastSeenAt = now

	mgr := buildTestSidecar(t, stateDir, []Tag{oldTag, newTag})
	resolver := NewResolver(mgr)

	// 1일 전 이후의 태그만
	query := Query{
		Since: now.Add(-24 * time.Hour),
		Limit: DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	if len(result.Tags) != 1 {
		t.Errorf("since 필터 결과: 기대 1, 실제 %d", len(result.Tags))
	}
}

// BenchmarkResolve는 Resolve 성능을 측정합니다 (SPEC §7: 1000개 태그 < 100ms).
func BenchmarkResolve(b *testing.B) {
	stateDir, _ := os.MkdirTemp("", "mx-bench-*")
	defer func() { _ = os.RemoveAll(stateDir) }()

	// 1000개 태그 준비
	tags := make([]Tag, 1000)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트", i+1)
	}

	mgr := NewManager(stateDir)
	sidecar := &Sidecar{SchemaVersion: SchemaVersion, Tags: tags, ScannedAt: time.Now()}
	_ = mgr.Write(sidecar)

	resolver := NewResolver(mgr)
	query := Query{Kind: MXNote, Limit: DefaultLimit}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = resolver.Resolve(query)
	}
}
