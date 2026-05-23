package mx

import (
	"context"
	"errors"
	"os"
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// mockLSPReferencesClient is a test mock implementation of the LSPReferencesClient interface.
type mockLSPReferencesClient struct {
	// locations are the locations returned by FindReferences calls.
	locations []lsp.Location
	// err is the error returned by FindReferences calls.
	err error
	// available indicates whether the LSP server is available.
	available bool
}

// FindReferences returns the preconfigured location list or error.
func (m *mockLSPReferencesClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.locations, nil
}

// IsAvailable returns the LSP server availability.
func (m *mockLSPReferencesClient) IsAvailable() bool {
	return m.available
}

// TestLSPFanInCounter_BasicCount verifies that when the LSP mock returns 3 references,
// fan_in=3 and method="lsp" are returned.
// AC-SPC-004-02: includes the case where fan_in_method is "lsp".
func TestLSPFanInCounter_BasicCount(t *testing.T) {
	// Arrange
	locations := []lsp.Location{
		{URI: "file:///project/internal/a.go", Range: lsp.Range{}},
		{URI: "file:///project/internal/b.go", Range: lsp.Range{}},
		{URI: "file:///project/internal/c.go", Range: lsp.Range{}},
	}
	mock := &mockLSPReferencesClient{
		locations: locations,
		available: true,
	}
	counter := &LSPFanInCounter{
		Client:      mock,
		ProjectRoot: "/project",
	}
	tag := Tag{
		Kind:     MXAnchor,
		File:     "/project/internal/auth.go",
		Line:     10,
		AnchorID: "anchor-auth",
	}

	// Act
	count, method, err := counter.Count(context.Background(), tag, "/project", false)

	// Assert
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}
	if count != 3 {
		t.Errorf("fan_in: 기대 3, 실제 %d", count)
	}
	if method != "lsp" {
		t.Errorf("fan_in_method: 기대 'lsp', 실제 '%s'", method)
	}
}

// TestLSPFanInCounter_TextualFallback verifies that when the LSP client is nil
// it falls back to TextualFanInCounter.
// AC-SPC-004-07: when LSP is absent, fall back to the textual method and return method="textual".
func TestLSPFanInCounter_TextualFallback(t *testing.T) {
	// Arrange: scenario with no LSP (nil client)
	counter := &LSPFanInCounter{
		Client:      nil, // LSP unavailable
		ProjectRoot: "/project",
	}
	tag := Tag{
		Kind:     MXAnchor,
		File:     "/project/internal/auth.go",
		Line:     10,
		AnchorID: "anchor-auth",
	}

	// Act
	_, method, err := counter.Count(context.Background(), tag, "/project", false)

	// Assert
	if err != nil {
		t.Fatalf("textual fallback 중 예기치 않은 오류: %v", err)
	}
	if method != "textual" {
		t.Errorf("fallback fan_in_method: 기대 'textual', 실제 '%s'", method)
	}
}

// TestLSPFanInCounter_StrictMode_LSPRequiredError verifies that when MOAI_MX_QUERY_STRICT=1
// and LSP is unavailable, LSPRequiredError is returned.
// AC-SPC-004-09: strictMode + LSP unavailable -> exit non-zero + LSPRequired.
func TestLSPFanInCounter_StrictMode_LSPRequiredError(t *testing.T) {
	// Arrange
	t.Setenv("MOAI_MX_QUERY_STRICT", "1")
	counter := &LSPFanInCounter{
		Client:      nil, // LSP unavailable
		ProjectRoot: "/project",
		Language:    "go",
	}
	tag := Tag{
		Kind:     MXAnchor,
		File:     "/project/internal/auth.go",
		Line:     10,
		AnchorID: "anchor-auth",
	}

	// Act
	_, _, err := counter.Count(context.Background(), tag, "/project", false)

	// Assert
	if err == nil {
		t.Fatal("strictMode + LSP 불가: 오류 기대했으나 nil 반환됨")
	}
	var lspErr *LSPRequiredError
	if !errors.As(err, &lspErr) {
		t.Errorf("LSPRequiredError 기대, 실제: %T: %v", err, err)
	}
}

// TestLSPFanInCounter_ExcludeTests verifies that when excludeTests=true
// references in _test.go files are excluded.
// AC-SPC-004-11: test file reference exclusion.
func TestLSPFanInCounter_ExcludeTests(t *testing.T) {
	// Arrange: out of 3 references, 1 is in a _test.go file
	locations := []lsp.Location{
		{URI: "file:///project/internal/a.go", Range: lsp.Range{}},
		{URI: "file:///project/internal/b_test.go", Range: lsp.Range{}}, // test file
		{URI: "file:///project/internal/c.go", Range: lsp.Range{}},
	}
	mock := &mockLSPReferencesClient{
		locations: locations,
		available: true,
	}
	counter := &LSPFanInCounter{
		Client:      mock,
		ProjectRoot: "/project",
	}
	tag := Tag{
		Kind:     MXAnchor,
		File:     "/project/internal/auth.go",
		Line:     10,
		AnchorID: "anchor-auth",
	}

	// Act: excludeTests=true
	countExcluded, _, err := counter.Count(context.Background(), tag, "/project", true)
	if err != nil {
		t.Fatalf("excludeTests=true 오류: %v", err)
	}

	// Act: excludeTests=false
	countAll, _, err := counter.Count(context.Background(), tag, "/project", false)
	if err != nil {
		t.Fatalf("excludeTests=false 오류: %v", err)
	}

	// Assert: when tests are excluded, the count must be 1 less
	if countExcluded != 2 {
		t.Errorf("excludeTests=true: fan_in 2 기대, 실제 %d", countExcluded)
	}
	if countAll != 3 {
		t.Errorf("excludeTests=false: fan_in 3 기대, 실제 %d", countAll)
	}
}

// TestLSPFanInCounter_InterfaceCompliance verifies that LSPFanInCounter implements
// the FanInCounter interface at compile time.
func TestLSPFanInCounter_InterfaceCompliance(t *testing.T) {
	var _ FanInCounter = &LSPFanInCounter{}
}

// TestLSPFanInCounter_LSPErrorFallback verifies that when the LSP client returns an error,
// textual fallback works.
func TestLSPFanInCounter_LSPErrorFallback(t *testing.T) {
	// Arrange: LSP client returns an error
	mock := &mockLSPReferencesClient{
		err:       errors.New("lsp: connection refused"),
		available: true,
	}
	counter := &LSPFanInCounter{
		Client:      mock,
		ProjectRoot: t.TempDir(),
	}
	tag := Tag{
		Kind:     MXAnchor,
		File:     "/project/internal/auth.go",
		Line:     10,
		AnchorID: "anchor-auth",
	}

	// Act
	_, method, err := counter.Count(context.Background(), tag, t.TempDir(), false)

	// Assert: on LSP error, switch to textual fallback; no error must be returned
	if err != nil {
		t.Fatalf("LSP 오류 시 textual fallback 중 예기치 않은 오류: %v", err)
	}
	if method != "textual" {
		t.Errorf("LSP 오류 후 fallback method: 기대 'textual', 실제 '%s'", method)
	}
}

// TestLSPFanInCounter_StrictMode_Env verifies that when MOAI_MX_QUERY_STRICT is unset,
// LSP unavailability triggers textual fallback without an error.
func TestLSPFanInCounter_StrictMode_Env(t *testing.T) {
	// Case where MOAI_MX_QUERY_STRICT env var is unset
	os.Unsetenv("MOAI_MX_QUERY_STRICT") //nolint:errcheck

	counter := &LSPFanInCounter{
		Client:      nil, // LSP unavailable
		ProjectRoot: "/project",
	}
	tag := Tag{
		Kind:     MXAnchor,
		File:     "/project/internal/auth.go",
		Line:     10,
		AnchorID: "anchor-auth",
	}

	// Act
	_, method, err := counter.Count(context.Background(), tag, "/project", false)

	// Assert: in non-strict mode, textual fallback without error
	if err != nil {
		t.Fatalf("non-strict fallback 중 예기치 않은 오류: %v", err)
	}
	if method != "textual" {
		t.Errorf("non-strict fallback method: 기대 'textual', 실제 '%s'", method)
	}
}
