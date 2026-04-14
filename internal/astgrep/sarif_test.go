package astgrep_test

import (
	"encoding/json"
	"testing"

	"github.com/modu-ai/moai-adk/internal/astgrep"
)

// TestToSARIF_EmptyFindings: findings가 없을 때 유효한 SARIF 문서 생성 검증
func TestToSARIF_EmptyFindings(t *testing.T) {
	output, err := astgrep.ToSARIF(nil, "1.0.0-test")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}
	if len(output) == 0 {
		t.Fatal("ToSARIF() returned empty output")
	}

	var doc map[string]any
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("SARIF 출력이 유효한 JSON이 아님: %v", err)
	}

	// $schema 필드 검증
	schema, ok := doc["$schema"].(string)
	if !ok || schema == "" {
		t.Error("SARIF $schema 필드가 없거나 비어있음")
	}

	// version 필드 검증
	version, ok := doc["version"].(string)
	if !ok || version != "2.1.0" {
		t.Errorf("SARIF version = %q, want 2.1.0", version)
	}

	// runs 배열 검증
	runs, ok := doc["runs"].([]any)
	if !ok || len(runs) == 0 {
		t.Fatal("SARIF runs 배열이 없거나 비어있음")
	}
}

// TestToSARIF_ToolDriver: tool.driver 필드가 SPEC 요구사항을 만족하는지 검증 (AC5)
func TestToSARIF_ToolDriver(t *testing.T) {
	output, err := astgrep.ToSARIF(nil, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	tool := run["tool"].(map[string]any)
	driver := tool["driver"].(map[string]any)

	// tool.driver.name이 "moai-ast-grep"이어야 함 (AC5)
	if name := driver["name"].(string); name != "moai-ast-grep" {
		t.Errorf("tool.driver.name = %q, want moai-ast-grep", name)
	}

	// tool.driver.version이 전달된 버전을 반영해야 함 (AC5)
	if ver := driver["version"].(string); ver != "0.42.1" {
		t.Errorf("tool.driver.version = %q, want 0.42.1", ver)
	}
}

// TestToSARIF_FindingMapping: Finding이 SARIF result로 올바르게 매핑되는지 검증 (AC5)
func TestToSARIF_FindingMapping(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "go-no-raw-getenv",
			Severity: "warning",
			Message:  "환경변수를 직접 사용하지 마세요",
			File:     "internal/cli/main.go",
			Line:     10,
			Column:   5,
		},
		{
			RuleID:   "sec-sql-injection",
			Severity: "error",
			Message:  "SQL 인젝션 가능성",
			File:     "internal/db/query.go",
			Line:     42,
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	results, ok := run["results"].([]any)
	if !ok {
		t.Fatal("SARIF results 배열이 없음")
	}

	if len(results) != 2 {
		t.Fatalf("SARIF results len = %d, want 2", len(results))
	}
}

// TestToSARIF_SeverityMapping: severity가 SARIF level로 올바르게 매핑되는지 검증 (AC5)
func TestToSARIF_SeverityMapping(t *testing.T) {
	tests := []struct {
		severity  string
		wantLevel string
	}{
		{"error", "error"},
		{"warning", "warning"},
		{"info", "note"},
		{"", "note"},  // 빈 문자열 → note (SARIF 기본값)
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			findings := []astgrep.Finding{
				{
					RuleID:   "test-rule",
					Severity: tt.severity,
					Message:  "테스트",
					File:     "test.go",
					Line:     1,
				},
			}

			output, err := astgrep.ToSARIF(findings, "1.0.0")
			if err != nil {
				t.Fatalf("ToSARIF() error = %v", err)
			}

			var doc map[string]any
			_ = json.Unmarshal(output, &doc)

			runs := doc["runs"].([]any)
			run := runs[0].(map[string]any)
			results := run["results"].([]any)
			result := results[0].(map[string]any)

			level, ok := result["level"].(string)
			if !ok {
				t.Fatalf("SARIF result.level 필드가 없음")
			}
			if level != tt.wantLevel {
				t.Errorf("SARIF level = %q, want %q (severity=%q)", level, tt.wantLevel, tt.severity)
			}
		})
	}
}

// TestToSARIF_MetadataPreservation: Finding의 메타데이터(CWE/OWASP)가 SARIF에 전달되는지 검증 (AC5)
func TestToSARIF_MetadataPreservation(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "sec-sql-injection",
			Severity: "error",
			Message:  "SQL 인젝션",
			File:     "db.go",
			Line:     1,
			Metadata: map[string]string{
				"owasp": "A03:2021",
				"cwe":   "CWE-89",
			},
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	// JSON으로 파싱하여 properties 필드 존재 확인
	var doc map[string]any
	_ = json.Unmarshal(output, &doc)

	runs := doc["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	result := results[0].(map[string]any)

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("SARIF result.properties 필드가 없음")
	}

	if cwe, ok := props["cwe"].(string); !ok || cwe == "" {
		t.Error("SARIF result.properties.cwe 필드가 없거나 비어있음")
	}
}

// TestToSARIF_RulesDeterministicOrder: tool.driver.rules 배열이 rule ID 기준으로
// 오름차순 정렬되어 실행마다 동일한 순서로 출력되는지 검증.
// REGRESSION: Go map 반복 순서가 불확정적이어서 SARIF 출력이 실행마다 달라지는
// 문제(issue #644)를 방지한다. Snapshot 테스트 및 GitHub Code Scanning diff 안정성 확보용.
func TestToSARIF_RulesDeterministicOrder(t *testing.T) {
	// 의도적으로 알파벳 역순에 가까운 순서로 입력하여,
	// map 반복 순서에 의존하면 출력 순서가 불안정해지도록 구성한다.
	findings := []astgrep.Finding{
		{RuleID: "zeta-rule", Severity: "warning", Message: "zeta", File: "a.go", Line: 1},
		{RuleID: "alpha-rule", Severity: "error", Message: "alpha", File: "b.go", Line: 2},
		{RuleID: "mike-rule", Severity: "warning", Message: "mike", File: "c.go", Line: 3},
		{RuleID: "bravo-rule", Severity: "error", Message: "bravo", File: "d.go", Line: 4},
		{RuleID: "yankee-rule", Severity: "warning", Message: "yankee", File: "e.go", Line: 5},
		{RuleID: "charlie-rule", Severity: "error", Message: "charlie", File: "f.go", Line: 6},
	}

	wantOrder := []string{
		"alpha-rule",
		"bravo-rule",
		"charlie-rule",
		"mike-rule",
		"yankee-rule",
		"zeta-rule",
	}

	// 여러 번 호출해도 항상 동일한 순서여야 한다 (map 반복 비결정성 회피 검증).
	const iterations = 20
	var firstOutput []byte

	for i := 0; i < iterations; i++ {
		output, err := astgrep.ToSARIF(findings, "1.0.0-test")
		if err != nil {
			t.Fatalf("iteration %d: ToSARIF() error = %v", i, err)
		}

		if i == 0 {
			firstOutput = output
		} else if string(output) != string(firstOutput) {
			t.Fatalf("iteration %d: SARIF 출력이 실행마다 달라짐 (비결정적)\nfirst:\n%s\ngot:\n%s",
				i, string(firstOutput), string(output))
		}

		var doc map[string]any
		if err := json.Unmarshal(output, &doc); err != nil {
			t.Fatalf("iteration %d: JSON 파싱 실패: %v", i, err)
		}

		runs := doc["runs"].([]any)
		run := runs[0].(map[string]any)
		tool := run["tool"].(map[string]any)
		driver := tool["driver"].(map[string]any)
		rules, ok := driver["rules"].([]any)
		if !ok {
			t.Fatalf("iteration %d: tool.driver.rules 배열이 없음", i)
		}

		if len(rules) != len(wantOrder) {
			t.Fatalf("iteration %d: rules len = %d, want %d", i, len(rules), len(wantOrder))
		}

		for idx, rule := range rules {
			r := rule.(map[string]any)
			id := r["id"].(string)
			if id != wantOrder[idx] {
				t.Errorf("iteration %d: rules[%d].id = %q, want %q (전체 순서가 ID 기준 오름차순이어야 함)",
					i, idx, id, wantOrder[idx])
			}
		}
	}
}

// TestToSARIF_RoundTrip: ToSARIF 출력이 유효한 SARIF 2.1.0 스키마를 따르는지 검증
func TestToSARIF_RoundTrip(t *testing.T) {
	findings := []astgrep.Finding{
		{
			RuleID:   "test-rule-1",
			Severity: "error",
			Message:  "첫 번째 테스트 규칙",
			File:     "file1.go",
			Line:     10,
			Column:   5,
			EndLine:  10,
			EndColumn: 20,
		},
		{
			RuleID:   "test-rule-2",
			Severity: "warning",
			Message:  "두 번째 테스트 규칙",
			File:     "file2.go",
			Line:     42,
		},
	}

	output, err := astgrep.ToSARIF(findings, "0.42.1")
	if err != nil {
		t.Fatalf("ToSARIF() error = %v", err)
	}

	// JSON 파싱 및 재직렬화로 라운드트립 검증
	var doc any
	if err := json.Unmarshal(output, &doc); err != nil {
		t.Fatalf("ToSARIF() 출력이 유효한 JSON이 아님: %v", err)
	}

	reEncoded, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("재직렬화 실패: %v", err)
	}

	var redoc map[string]any
	if err := json.Unmarshal(reEncoded, &redoc); err != nil {
		t.Fatalf("재파싱 실패: %v", err)
	}

	// 핵심 필드 재확인
	if redoc["version"] != "2.1.0" {
		t.Errorf("라운드트립 후 version = %v, want 2.1.0", redoc["version"])
	}
}
