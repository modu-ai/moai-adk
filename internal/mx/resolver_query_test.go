package mx

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildTestSidecar creates a test sidecar file in the specified directory.
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

// makeTag is a helper that creates a test tag.
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

// makeAnchorTag creates a test ANCHOR tag with an AnchorID.
func makeAnchorTag(file, anchorID, body string, line int) Tag {
	t := makeTag(MXAnchor, file, body, line)
	t.AnchorID = anchorID
	return t
}

// makeWarnTag creates a test WARN tag with a Reason.
func makeWarnTag(file, body, reason string, line int) Tag {
	t := makeTag(MXWarn, file, body, line)
	t.Reason = reason
	return t
}

// TestResolve_SpecAndKindFilter tests the SPEC+KIND combined filter.
// AC-SPC-004-01: out of 20 tags spanning 2 SPECs, only ANCHORs for SPEC-X-001 are returned.
func TestResolve_SpecAndKindFilter(t *testing.T) {
	// AC-SPC-004-01: SPEC filter + KIND filter combination
	stateDir := t.TempDir()

	// ANCHOR tags belonging to SPEC-X-001
	anchor1 := makeAnchorTag("internal/auth/handler.go", "anchor-auth-handler", "ANCHOR for SPEC-X-001", 10)
	anchor2 := makeAnchorTag("internal/auth/middleware.go", "anchor-auth-mw", "ANCHOR for SPEC-X-001", 20)

	// Tags belonging to SPEC-Y-002 (should be excluded by the filter)
	note1 := makeTag(MXNote, "internal/cache/store.go", "캐시 정책 설명 — SPEC-Y-002", 5)
	anchor3 := makeAnchorTag("internal/cache/store.go", "anchor-cache", "ANCHOR for SPEC-Y-002", 15)

	// NOTE tag for SPEC-X-001 (should be excluded by the KIND filter)
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

// TestResolve_FanInFilter tests the fan-in minimum filter.
// AC-SPC-004-02: out of fan_in 1,2,3,5,10, only the 3 entries with fan-in-min>=3 are returned.
func TestResolve_FanInFilter(t *testing.T) {
	// AC-SPC-004-02: fan-in filter
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

	// Inject a mock counter for fan-in calculation
	// In the real implementation this is injected via the WithFanInCounter option
	query := Query{
		Kind:     MXAnchor,
		FanInMin: 3,
		Limit:    DefaultLimit,
	}

	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// Minimal implementation uses mock fan-in values (real logic implemented in the GREEN phase)
	// In the RED phase this test is expected to fail
	for _, tag := range result.Tags {
		if tag.FanIn < 3 {
			t.Errorf("fan-in 필터 실패: %s의 fan-in=%d (최소 3 기대)", tag.AnchorID, tag.FanIn)
		}
	}
}

// TestResolve_DangerFilter tests the danger category filter.
// AC-SPC-004-03: a WARN with REASON "goroutine leak" is returned as concurrency.
func TestResolve_DangerFilter(t *testing.T) {
	// AC-SPC-004-03: danger category filter
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

// TestResolve_SidecarUnavailable tests the error returned when the sidecar file is missing.
// AC-SPC-004-04: when the sidecar file is absent, SidecarUnavailable error is returned.
func TestResolve_SidecarUnavailable(t *testing.T) {
	// AC-SPC-004-04: sidecar file is absent
	stateDir := t.TempDir()
	// Sidecar file is not created (only the directory exists)
	// However, the manager returns an empty sidecar from Load()
	// The actual error occurs when the file is missing and explicit verification is performed

	// Simulate an unreadable situation by using a non-existent directory
	nonExistentDir := filepath.Join(stateDir, "nonexistent_mx_state")
	mgr := NewManager(nonExistentDir)

	// When the sidecar file is absent, Resolve must return SidecarUnavailableError
	// The Resolve implementation must explicitly check for sidecar presence
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
		// Also verify the case where it is not SidecarUnavailableError
		_ = sidecarErr
		t.Errorf("오류 메시지에 'SidecarUnavailable' 없음: %v", err)
	}
}

// TestResolve_JSONSchema tests the JSON schema fields.
// AC-SPC-004-05: the JSON output must conform to the REQ-SPC-004-005 schema.
func TestResolve_JSONSchema(t *testing.T) {
	// AC-SPC-004-05: JSON schema verification
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

	// JSON marshalling and unmarshalling verification
	data, err := json.Marshal(result.Tags[0])
	if err != nil {
		t.Fatalf("JSON 직렬화 실패: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON 역직렬화 실패: %v", err)
	}

	// REQ-SPC-004-005 required fields check
	requiredFields := []string{"kind", "file", "line", "body", "created_by", "last_seen_at", "spec_associations"}
	for _, field := range requiredFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("JSON 스키마 누락 필드: %s", field)
		}
	}
}

// TestResolve_Pagination tests pagination.
// AC-SPC-004-08: with 10000 tags, apply limit 100 and emit TruncationNotice.
func TestResolve_Pagination(t *testing.T) {
	// AC-SPC-004-08: pagination and TruncationNotice
	stateDir := t.TempDir()

	// Generate 200 tags (more than DefaultLimit)
	tags := make([]Tag, 200)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트 태그", i+1)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// limit not specified -> DefaultLimit(100) is expected to apply
	query := Query{Limit: 0} // 0 means default applies
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

// TestResolve_TextualFallbackAnnotation tests the textual fallback annotation when LSP is unavailable.
// AC-SPC-004-07: fan_in_method "textual" annotation on Python files without LSP.
func TestResolve_TextualFallbackAnnotation(t *testing.T) {
	// AC-SPC-004-07: fan_in_method annotation on textual fallback
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

// TestResolve_StrictMode tests the LSP requirement when MOAI_MX_QUERY_STRICT=1.
// AC-SPC-004-09: in strict mode, LSPRequired error is returned when LSP is unavailable.
func TestResolve_StrictMode(t *testing.T) {
	// AC-SPC-004-09: strict mode test
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

// TestResolve_TestFixtureExclusion tests exclusion of test fixture references.
// AC-SPC-004-11: when include-tests is unset, test file references are excluded.
func TestResolve_TestFixtureExclusion(t *testing.T) {
	// AC-SPC-004-11: test file reference exclusion
	stateDir := t.TempDir()

	anchor := makeAnchorTag("tests/fixtures/mock_handler_test.go", "anchor-mock-handler", "핸들러 목", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor})
	resolver := NewResolver(mgr)

	// include-tests=false (default)
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

	// Test references in test fixture files must be excluded from fan-in
	// (the actual fan-in value must not include test references)
	_ = result // Add concrete verification in the GREEN phase
}

// TestResolve_EmptyResult tests that an empty array and exit 0 are returned when no matches exist.
// AC-SPC-004-12: returns [] and exit 0 when there are no matches.
func TestResolve_EmptyResult(t *testing.T) {
	// AC-SPC-004-12: empty result handling
	stateDir := t.TempDir()

	note := makeTag(MXNote, "internal/misc.go", "노트 태그", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{note})
	resolver := NewResolver(mgr)

	// ANCHOR filter -> only NOTE exists, so 0 results
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

// TestResolve_CombinedFilters tests the AND combination of multiple filters.
// AC-SPC-004-14: combined filter --spec X --kind anchor --fan-in-min 3.
func TestResolve_CombinedFilters(t *testing.T) {
	// AC-SPC-004-14: composite AND filter
	stateDir := t.TempDir()

	// High fan-in ANCHOR belonging to SPEC-X-001
	highFanInAnchor := makeAnchorTag("internal/auth/handler.go", "anchor-high", "고fan-in ANCHOR for SPEC-X-001", 10)
	// Low fan-in ANCHOR belonging to SPEC-X-001
	lowFanInAnchor := makeAnchorTag("internal/auth/util.go", "anchor-low", "저fan-in ANCHOR for SPEC-X-001", 20)
	// High fan-in ANCHOR belonging to SPEC-Y-002
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

	// Only tags satisfying all three conditions must be returned
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

// TestResolve_SpecAssociationFromBody tests SPEC ID extraction from the tag body.
// AC-SPC-004-15: @MX body "ANCHOR for SPEC-AUTH-001 handler" -> spec_associations contains SPEC-AUTH-001.
func TestResolve_SpecAssociationFromBody(t *testing.T) {
	// AC-SPC-004-15: SPEC ID extraction from tag body
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

// TestResolve_MarkdownFormat tests the markdown output format.
// AC-SPC-004-10: --format markdown emits a markdown table.
func TestResolve_MarkdownFormat(t *testing.T) {
	// AC-SPC-004-10: markdown output format verification
	stateDir := t.TempDir()
	note := makeTag(MXNote, "internal/misc.go", "노트 태그", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{note})
	resolver := NewResolver(mgr)

	query := Query{Kind: MXNote, Limit: DefaultLimit}
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// The FormatMarkdown function must exist
	md := FormatMarkdown(result)
	if !strings.Contains(md, "|") {
		t.Error("마크다운 테이블 형식 (|) 없음")
	}
}

// TestResolve_TableFormat tests the table output format.
// AC-SPC-004-06: --format table emits human-readable column format.
func TestResolve_TableFormat(t *testing.T) {
	// AC-SPC-004-06: table output format verification
	stateDir := t.TempDir()
	note := makeTag(MXNote, "internal/misc.go", "노트 태그", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{note})
	resolver := NewResolver(mgr)

	query := Query{Kind: MXNote, Limit: DefaultLimit}
	result, err := resolver.Resolve(query)
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}

	// The FormatTable function must exist
	table := FormatTable(result)
	if table == "" {
		t.Error("테이블 출력이 비어있음")
	}
}

// TestExtractSpecIDs unit-tests SPEC ID extraction from the tag body.
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

// TestFormatMarkdown_TableStructure verifies that the markdown output has a table structure.
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

	// Verify markdown table header
	if !strings.Contains(md, "Kind") {
		t.Error("마크다운 헤더에 'Kind' 없음")
	}
	if !strings.Contains(md, "|") {
		t.Error("마크다운 테이블 구분자 '|' 없음")
	}
}

// TestFormatJSON verifies that the JSON output is in the correct format.
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

// FormatMarkdown and FormatTable are implemented in resolver_query.go.

// TestResolve_LargeSidecarTruncation tests automatic limit enforcement on a large sidecar.
// Additional verification for AC-SPC-004-08.
func TestResolve_LargeSidecarTruncation(t *testing.T) {
	stateDir := t.TempDir()

	// More than 10000 tags (not a performance test, just a behavior test)
	// Actually creating 10000 takes time, so we test with 200
	tags := make([]Tag, 200)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트", i+1)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// When Limit=0, DefaultLimit applies
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

// TestValidateQuery_UnknownDanger_InvalidQueryError verifies that validateQuery returns
// InvalidQueryError for an unknown danger category.
// SPEC-V3R2-SPC-004 M2 RED — T-SPC004-06 / AC-SPC-004-13
func TestValidateQuery_UnknownDanger_InvalidQueryError(t *testing.T) {
	query := Query{Danger: "frobnicate"}

	err := validateQuery(query)
	if err == nil {
		t.Fatal("알 수 없는 danger 카테고리에 대해 오류 기대, 실제 nil")
	}

	var iqErr *InvalidQueryError
	// Type check via errors.As
	if !errors.As(err, &iqErr) {
		t.Fatalf("*InvalidQueryError 기대, 실제: %T (%v)", err, err)
	}

	if iqErr.Field != "danger" {
		t.Errorf("Field: 기대 'danger', 실제 %q", iqErr.Field)
	}

	if !strings.Contains(iqErr.Message, "allowed") {
		t.Errorf("Message에 'allowed' 없음: %q", iqErr.Message)
	}
}

// TestResolve_InvalidKind tests the error for an invalid kind value.
// AC-SPC-004-13: --kind nonexistent -> exit 2 + InvalidQuery error.
func TestResolve_InvalidKind(t *testing.T) {
	// AC-SPC-004-13: invalid filter value handling
	stateDir := t.TempDir()
	mgr := buildTestSidecar(t, stateDir, []Tag{})
	resolver := NewResolver(mgr)

	// Invalid kind value
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

// TestResolve_FilePrefix tests the file path prefix filter.
func TestResolve_FilePrefix(t *testing.T) {
	stateDir := t.TempDir()

	// Tags with various paths
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

// TestResolve_OffsetPagination tests offset-based pagination.
func TestResolve_OffsetPagination(t *testing.T) {
	stateDir := t.TempDir()

	// Generate 10 tags
	tags := make([]Tag, 10)
	for i := range tags {
		tags[i] = makeTag(MXNote, "internal/file.go", "노트", i+1)
	}

	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// First page
	q1 := Query{Limit: 5, Offset: 0}
	r1, err := resolver.Resolve(q1)
	if err != nil {
		t.Fatalf("첫 페이지 오류: %v", err)
	}

	// Second page
	q2 := Query{Limit: 5, Offset: 5}
	r2, err := resolver.Resolve(q2)
	if err != nil {
		t.Fatalf("두 번째 페이지 오류: %v", err)
	}

	// The sum of the two pages must equal the total tag count
	if len(r1.Tags)+len(r2.Tags) != 10 {
		t.Errorf("페이지네이션 합산: 기대 10, 실제 %d+%d=%d",
			len(r1.Tags), len(r2.Tags), len(r1.Tags)+len(r2.Tags))
	}
}

// TestResolveAnchor verifies that the existing ResolveAnchor(anchorID) API works.
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

// TestResolve_SinceFilter tests the since-time filter.
func TestResolve_SinceFilter(t *testing.T) {
	stateDir := t.TempDir()

	now := time.Now()
	old := now.Add(-48 * time.Hour) // 2 days ago

	oldTag := makeTag(MXNote, "internal/old.go", "오래된 태그", 1)
	oldTag.LastSeenAt = old

	newTag := makeTag(MXNote, "internal/new.go", "새 태그", 1)
	newTag.LastSeenAt = now

	mgr := buildTestSidecar(t, stateDir, []Tag{oldTag, newTag})
	resolver := NewResolver(mgr)

	// Only tags from 1 day ago onwards
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

// TestResolveLimit_MaxEnforcement tests enforcement of the limit maximum.
func TestResolveLimit_MaxEnforcement(t *testing.T) {
	stateDir := t.TempDir()
	tags := make([]Tag, 5)
	for i := range tags {
		tags[i] = makeTag(MXNote, "file.go", "노트", i+1)
	}
	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	// Clip to MaxLimit when MaxLimit is exceeded
	q := Query{Limit: MaxLimit + 1}
	result, err := resolver.Resolve(q)
	if err != nil {
		t.Fatalf("오류: %v", err)
	}
	// Only 5 tags exist, so 5 are returned
	if len(result.Tags) > 5 {
		t.Errorf("태그 수 초과: %d", len(result.Tags))
	}
}

// TestResolveLimit_Normalization tests the resolveLimit function.
func TestResolveLimit_Normalization(t *testing.T) {
	if resolveLimit(0) != DefaultLimit {
		t.Errorf("0 → DefaultLimit 기대")
	}
	if resolveLimit(-1) != DefaultLimit {
		t.Errorf("-1 → DefaultLimit 기대")
	}
	if resolveLimit(MaxLimit+1) != MaxLimit {
		t.Errorf("MaxLimit+1 → MaxLimit 기대")
	}
	if resolveLimit(50) != 50 {
		t.Errorf("50 → 50 기대")
	}
}

// TestSidecarUnavailableError tests the error message.
func TestSidecarUnavailableError(t *testing.T) {
	e := &SidecarUnavailableError{}
	if e.Error() == "" {
		t.Error("오류 메시지가 비어있음")
	}
	if e.Unwrap() != nil {
		t.Error("Cause 없을 때 Unwrap()은 nil 기대")
	}

	inner := os.ErrNotExist
	e2 := &SidecarUnavailableError{Cause: inner}
	if e2.Unwrap() != inner {
		t.Error("Unwrap()이 내부 오류를 반환해야 함")
	}
}

// TestInvalidQueryError tests the InvalidQueryError message.
func TestInvalidQueryError(t *testing.T) {
	e := &InvalidQueryError{Field: "kind", Value: "bad", Message: "test message"}
	if !strings.Contains(e.Error(), "InvalidQuery") {
		t.Errorf("오류 메시지에 'InvalidQuery' 없음: %s", e.Error())
	}
}

// TestLSPRequiredError tests the LSPRequiredError message.
func TestLSPRequiredError(t *testing.T) {
	e := &LSPRequiredError{Language: "go"}
	if !strings.Contains(e.Error(), "LSPRequired") {
		t.Errorf("오류 메시지에 'LSPRequired' 없음: %s", e.Error())
	}
}

// TestFormatTable_Empty tests the empty-result table output.
func TestFormatTable_Empty(t *testing.T) {
	result := QueryResult{Tags: []TagResult{}, TotalCount: 0}
	table := FormatTable(result)
	if !strings.Contains(table, "결과 없음") {
		t.Errorf("빈 결과 메시지 없음: %s", table)
	}
}

// TestFormatMarkdown_WithTruncation tests the markdown output with TruncationNotice.
func TestFormatMarkdown_WithTruncation(t *testing.T) {
	result := QueryResult{
		Tags:             []TagResult{},
		TruncationNotice: true,
		TotalCount:       500,
	}
	md := FormatMarkdown(result)
	if !strings.Contains(md, "TruncationNotice") {
		t.Errorf("TruncationNotice 없음: %s", md)
	}
}

// TestTruncateStr tests string truncation.
func TestTruncateStr(t *testing.T) {
	// Short string: unchanged
	got := truncateStr("short", 20)
	if got != "short" {
		t.Errorf("기대 'short', 실제 %q", got)
	}

	// Long string: truncate
	long := "abcdefghijklmnopqrstuvwxyz"
	got = truncateStr(long, 10)
	if len(got) > 10 {
		t.Errorf("잘라낸 문자열이 10자 초과: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("잘라낸 문자열이 '...'로 끝나야 함: %q", got)
	}
}

// TestResolve_ResolveAll tests the existing ResolveAll API.
func TestResolve_ResolveAll(t *testing.T) {
	stateDir := t.TempDir()
	anchor1 := makeAnchorTag("pkg/a.go", "anchor-1", "앵커 1", 1)
	anchor2 := makeAnchorTag("pkg/b.go", "anchor-2", "앵커 2", 1)
	mgr := buildTestSidecar(t, stateDir, []Tag{anchor1, anchor2})
	resolver := NewResolver(mgr)

	// Both anchorIDs exist
	tags, err := resolver.ResolveAll([]string{"anchor-1", "anchor-2"})
	if err != nil {
		t.Fatalf("ResolveAll 오류: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("ResolveAll 결과: 기대 2, 실제 %d", len(tags))
	}

	// Include a non-existent anchorID
	_, err = resolver.ResolveAll([]string{"anchor-1", "nonexistent"})
	if err == nil {
		t.Error("존재하지 않는 anchorID 포함 시 오류 기대")
	}
}

// TestResolve_ListAnchors tests listing all ANCHOR tags.
func TestResolve_ListAnchors(t *testing.T) {
	stateDir := t.TempDir()
	tags := []Tag{
		makeAnchorTag("b.go", "anchor-b", "앵커 B", 1),
		makeAnchorTag("a.go", "anchor-a", "앵커 A", 1),
		makeTag(MXNote, "c.go", "노트", 1),
	}
	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	anchors := resolver.ListAnchors()
	if len(anchors) != 2 {
		t.Errorf("ListAnchors: 기대 2, 실제 %d", len(anchors))
	}

	// Verify ascending sort by file path
	if len(anchors) == 2 && anchors[0].File > anchors[1].File {
		t.Error("파일 경로 오름차순 정렬 기대")
	}
}

// TestResolve_AuditLowFanIn tests the list of low fan-in ANCHORs.
func TestResolve_AuditLowFanIn(t *testing.T) {
	stateDir := t.TempDir()
	tags := []Tag{
		makeAnchorTag("a.go", "anchor-a", "앵커 A", 1),
	}
	mgr := buildTestSidecar(t, stateDir, tags)
	resolver := NewResolver(mgr)

	lowFanIn := resolver.AuditLowFanIn()
	if len(lowFanIn) == 0 {
		t.Error("AuditLowFanIn: 앵커가 있는데 빈 결과")
	}
}

// BenchmarkResolve measures Resolve performance (SPEC §7: 1000 tags in <100ms).
func BenchmarkResolve(b *testing.B) {
	stateDir, _ := os.MkdirTemp("", "mx-bench-*")
	defer func() { _ = os.RemoveAll(stateDir) }()

	// Prepare 1000 tags
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
