package gopls_test

import (
	"reflect"
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// TestDiagnosticAlias_TypeIdentity는 gopls.Diagnostic이 lsp.Diagnostic과
// 동일한 타입임을 reflect.TypeOf로 확인한다 (AC-UTIL-003-007).
func TestDiagnosticAlias_TypeIdentity(t *testing.T) {
	t.Parallel()

	goplsType := reflect.TypeOf(gopls.Diagnostic{})
	lspType := reflect.TypeOf(lsp.Diagnostic{})

	if goplsType != lspType {
		t.Errorf("reflect.TypeOf(gopls.Diagnostic{}) = %v, reflect.TypeOf(lsp.Diagnostic{}) = %v — types are not identical (type alias expected)",
			goplsType, lspType)
	}
}

// TestRangeAlias_TypeIdentity는 gopls.Range가 lsp.Range와
// 동일한 타입임을 확인한다 (AC-UTIL-003-007).
func TestRangeAlias_TypeIdentity(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(gopls.Range{}) != reflect.TypeOf(lsp.Range{}) {
		t.Error("gopls.Range and lsp.Range are not the same type (type alias expected)")
	}
}

// TestPositionAlias_TypeIdentity는 gopls.Position이 lsp.Position과
// 동일한 타입임을 확인한다 (AC-UTIL-003-007).
func TestPositionAlias_TypeIdentity(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(gopls.Position{}) != reflect.TypeOf(lsp.Position{}) {
		t.Error("gopls.Position and lsp.Position are not the same type (type alias expected)")
	}
}

// TestSeverityError_Equality는 gopls.SeverityError == lsp.SeverityError를 확인한다 (AC-UTIL-003-007).
func TestSeverityError_Equality(t *testing.T) {
	t.Parallel()

	if gopls.SeverityError != gopls.DiagnosticSeverity(lsp.SeverityError) {
		t.Errorf("gopls.SeverityError (%d) != lsp.SeverityError (%d)",
			int(gopls.SeverityError), int(lsp.SeverityError))
	}
}

// TestDiagnosticAlias_Interoperability는 lsp.Diagnostic 값을 gopls.Diagnostic으로
// 타입 변환 없이 대입할 수 있음을 컴파일 시점에 검증한다 (AC-UTIL-003-007).
func TestDiagnosticAlias_Interoperability(t *testing.T) {
	t.Parallel()

	// 타입 별칭이면 대입이 컴파일되고 값이 동일해야 함
	lspDiag := lsp.Diagnostic{
		Severity: lsp.SeverityError,
		Message:  "undefined: foo",
	}

	// 타입 별칭: 추가 변환 없이 대입 가능
	var goplsDiag gopls.Diagnostic = lspDiag
	if goplsDiag.Message != lspDiag.Message {
		t.Errorf("goplsDiag.Message = %q, want %q", goplsDiag.Message, lspDiag.Message)
	}
	if int(goplsDiag.Severity) != int(lspDiag.Severity) {
		t.Errorf("goplsDiag.Severity = %d, want %d", int(goplsDiag.Severity), int(lspDiag.Severity))
	}
}
