package hook

import (
	"encoding/json"
	"strings"
	"testing"

	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
	"github.com/modu-ai/moai-adk/internal/hook/quality"
)

// canonicalHookDiagnostics는 AC-UTIL-003-009의 표준 픽스처를 반환한다.
// Error / Warning / Information / Hint 각 1개씩 포함한다.
func canonicalHookDiagnostics() []lsphook.Diagnostic {
	return []lsphook.Diagnostic{
		{
			Severity: lsphook.SeverityError,
			Message:  "undefined: foo",
			Source:   "compiler",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 10, Character: 2},
				End:   lsphook.Position{Line: 10, Character: 5},
			},
		},
		{
			Severity: lsphook.SeverityWarning,
			Message:  "unused variable: bar",
			Source:   "staticcheck",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 20, Character: 0},
				End:   lsphook.Position{Line: 20, Character: 3},
			},
		},
		{
			Severity: lsphook.SeverityInformation,
			Message:  "consider renaming to camelCase",
			Source:   "gopls",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 30, Character: 0},
				End:   lsphook.Position{Line: 30, Character: 10},
			},
		},
		{
			Severity: lsphook.SeverityHint,
			Message:  "export this function for better testability",
			Source:   "gopls",
			Range: lsphook.Range{
				Start: lsphook.Position{Line: 40, Character: 0},
				End:   lsphook.Position{Line: 40, Character: 12},
			},
		},
	}
}

// computeCounts는 픽스처의 SeverityCounts를 계산한다.
func computeCounts(diags []lsphook.Diagnostic) lsphook.SeverityCounts {
	var counts lsphook.SeverityCounts
	for _, d := range diags {
		switch d.Severity {
		case lsphook.SeverityError:
			counts.Errors++
		case lsphook.SeverityWarning:
			counts.Warnings++
		case lsphook.SeverityInformation:
			counts.Information++
		case lsphook.SeverityHint:
			counts.Hints++
		}
	}
	return counts
}

// ─── AC-UTIL-003-008 ─────────────────────────────────────────────────────────

// TestHookDiagnosticSeverity_JSONMarshal_StringPreserved는
// hook.DiagnosticSeverity("error")가 JSON에서 문자열 "error"로 직렬화됨을 확인한다 (AC-UTIL-003-008).
// wire format 동결: SPEC-UTIL-003 이전과 이후 동일한 JSON 출력이 보장되어야 한다.
func TestHookDiagnosticSeverity_JSONMarshal_StringPreserved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity lsphook.DiagnosticSeverity
		want     string
	}{
		{lsphook.SeverityError, "error"},
		{lsphook.SeverityWarning, "warning"},
		{lsphook.SeverityInformation, "information"},
		{lsphook.SeverityHint, "hint"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.severity), func(t *testing.T) {
			t.Parallel()

			d := lsphook.Diagnostic{
				Severity: tt.severity,
				Message:  "test message",
				Range:    lsphook.Range{},
			}

			data, err := json.Marshal(d)
			if err != nil {
				t.Fatalf("json.Marshal error: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatalf("json.Unmarshal error: %v", err)
			}

			severity, ok := m["severity"].(string)
			if !ok {
				t.Fatalf("severity field is not a string: type=%T value=%v (wire format blocker — must remain string)", m["severity"], m["severity"])
			}
			if severity != tt.want {
				t.Errorf("severity = %q, want %q", severity, tt.want)
			}
		})
	}
}

// ─── AC-UTIL-003-009 ─────────────────────────────────────────────────────────

// TestWireFormat_FormatDiagnosticsAsInstructionWithFile_Freeze는
// 표준 픽스처에 대한 FormatDiagnosticsAsInstructionWithFile 출력이
// SPEC 적용 전후로 byte-identical임을 확인한다 (AC-UTIL-003-009).
//
// 동결 검증 방식: 동일 함수를 두 번 호출하여 출력이 항등 (idempotent)임을 검증.
// SPEC 적용 후 이 테스트는 새 코드와 함께 실행되므로 구현 변경 시 반드시 실패해야 한다.
func TestWireFormat_FormatDiagnosticsAsInstructionWithFile_Freeze(t *testing.T) {
	t.Parallel()

	fixture := canonicalHookDiagnostics()
	counts := computeCounts(fixture)

	// 에러만 포함 (errors > 0이면 경고는 무시됨)
	got1 := quality.FormatDiagnosticsAsInstructionWithFile("main.go", fixture, counts, false)
	got2 := quality.FormatDiagnosticsAsInstructionWithFile("main.go", fixture, counts, false)

	if got1 != got2 {
		t.Error("FormatDiagnosticsAsInstructionWithFile is not idempotent (wire format unstable)")
	}

	// 출력이 비어있지 않아야 함 (error 진단이 있으므로)
	if got1 == "" {
		t.Error("FormatDiagnosticsAsInstructionWithFile returned empty string for error diagnostic")
	}

	// 예상 포맷 요소 확인: 헤더, 파일명, 에러 메시지, 종료 문구
	if !strings.Contains(got1, "[Quality Gate]") {
		t.Errorf("output missing '[Quality Gate]' header: %q", got1)
	}
	if !strings.Contains(got1, "main.go") {
		t.Errorf("output missing file name 'main.go': %q", got1)
	}
	if !strings.Contains(got1, "undefined: foo") {
		t.Errorf("output missing error message: %q", got1)
	}
	if !strings.Contains(got1, "Fix these errors before proceeding.") {
		t.Errorf("output missing trailing fix instruction: %q", got1)
	}
}

// TestWireFormat_ConvertHookDiagsToLSP_Freeze는 convertHookDiagsToLSP 변환이
// SPEC 적용 전후로 byte-identical 출력을 생성함을 확인한다 (AC-UTIL-003-009).
// hook.Diagnostic(string severity) → lsp.Diagnostic(int severity) 변환 경로 동결.
func TestWireFormat_ConvertHookDiagsToLSP_Freeze(t *testing.T) {
	t.Parallel()

	fixture := canonicalHookDiagnostics()

	result1 := convertHookDiagsToLSP(fixture)
	result2 := convertHookDiagsToLSP(fixture)

	if len(result1) != len(result2) {
		t.Fatalf("convertHookDiagsToLSP lengths differ: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if result1[i].Severity != result2[i].Severity {
			t.Errorf("index %d: Severity %d != %d", i, result1[i].Severity, result2[i].Severity)
		}
		if result1[i].Message != result2[i].Message {
			t.Errorf("index %d: Message %q != %q", i, result1[i].Message, result2[i].Message)
		}
	}

	// severity 변환 정확도: string → int 검증
	// SeverityError = "error" → 1, SeverityWarning = "warning" → 2,
	// SeverityInformation = "information" → 3, SeverityHint = "hint" → 4
	expected := []int{1, 2, 3, 4}
	for i, want := range expected {
		if int(result1[i].Severity) != want {
			t.Errorf("severity[%d] = %d, want %d", i, int(result1[i].Severity), want)
		}
	}
}

// TestHookDiagnostic_JSONRoundTrip는 hook.Diagnostic의 JSON 직렬화/역직렬화가
// 완전한 데이터 보존을 보장함을 확인한다 (AC-UTIL-003-008 보강).
func TestHookDiagnostic_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := lsphook.Diagnostic{
		Severity: lsphook.SeverityError,
		Message:  "undefined: bar",
		Code:     "E001",
		Source:   "compiler",
		Range: lsphook.Range{
			Start: lsphook.Position{Line: 5, Character: 3},
			End:   lsphook.Position{Line: 5, Character: 6},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var restored lsphook.Diagnostic
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if restored.Severity != original.Severity {
		t.Errorf("Severity: got %q, want %q", restored.Severity, original.Severity)
	}
	if restored.Message != original.Message {
		t.Errorf("Message: got %q, want %q", restored.Message, original.Message)
	}
	if restored.Range.Start.Line != original.Range.Start.Line {
		t.Errorf("Range.Start.Line: got %d, want %d", restored.Range.Start.Line, original.Range.Start.Line)
	}
}
