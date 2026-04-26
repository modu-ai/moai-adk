package gopls_test

import (
	"reflect"
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// TestDiagnosticAlias_TypeIdentity verifies via reflect.TypeOf that
// gopls.Diagnostic and lsp.Diagnostic are the same type (AC-UTIL-003-007).
func TestDiagnosticAlias_TypeIdentity(t *testing.T) {
	t.Parallel()

	goplsType := reflect.TypeOf(gopls.Diagnostic{})
	lspType := reflect.TypeOf(lsp.Diagnostic{})

	if goplsType != lspType {
		t.Errorf("reflect.TypeOf(gopls.Diagnostic{}) = %v, reflect.TypeOf(lsp.Diagnostic{}) = %v — types are not identical (type alias expected)",
			goplsType, lspType)
	}
}

// TestRangeAlias_TypeIdentity verifies that gopls.Range and lsp.Range are
// the same type (AC-UTIL-003-007).
func TestRangeAlias_TypeIdentity(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(gopls.Range{}) != reflect.TypeOf(lsp.Range{}) {
		t.Error("gopls.Range and lsp.Range are not the same type (type alias expected)")
	}
}

// TestPositionAlias_TypeIdentity verifies that gopls.Position and lsp.Position
// are the same type (AC-UTIL-003-007).
func TestPositionAlias_TypeIdentity(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(gopls.Position{}) != reflect.TypeOf(lsp.Position{}) {
		t.Error("gopls.Position and lsp.Position are not the same type (type alias expected)")
	}
}

// TestSeverityError_Equality verifies gopls.SeverityError == lsp.SeverityError (AC-UTIL-003-007).
func TestSeverityError_Equality(t *testing.T) {
	t.Parallel()

	if gopls.SeverityError != gopls.DiagnosticSeverity(lsp.SeverityError) {
		t.Errorf("gopls.SeverityError (%d) != lsp.SeverityError (%d)",
			int(gopls.SeverityError), int(lsp.SeverityError))
	}
}

// TestDiagnosticAlias_Interoperability verifies at compile time that an
// lsp.Diagnostic value can be assigned to a gopls.Diagnostic without a type
// conversion (AC-UTIL-003-007).
func TestDiagnosticAlias_Interoperability(t *testing.T) {
	t.Parallel()

	// If aliased, the assignment compiles and the value must be the same.
	lspDiag := lsp.Diagnostic{
		Severity: lsp.SeverityError,
		Message:  "undefined: foo",
	}

	// Type alias: assignment without an explicit conversion compiles.
	// ST1023 nolint: the type annotation is intentional — it documents the test intent
	// (proving alias identity through the assignment).
	//nolint:staticcheck // ST1023
	var goplsDiag gopls.Diagnostic = lspDiag
	if goplsDiag.Message != lspDiag.Message {
		t.Errorf("goplsDiag.Message = %q, want %q", goplsDiag.Message, lspDiag.Message)
	}
	if int(goplsDiag.Severity) != int(lspDiag.Severity) {
		t.Errorf("goplsDiag.Severity = %d, want %d", int(goplsDiag.Severity), int(lspDiag.Severity))
	}
}
