package mx

import (
	"context"
	"errors"
	"os"
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// mockLSPReferencesClient는 LSPReferencesClient 인터페이스의 테스트용 mock 구현체입니다.
//
// @MX:TODO: [AUTO] mockLSPReferencesClient — resolved in GREEN phase once LSPReferencesClient interface is defined
type mockLSPReferencesClient struct {
	// locations는 FindReferences 호출 시 반환할 위치 목록입니다.
	locations []lsp.Location
	// err는 FindReferences 호출 시 반환할 오류입니다.
	err error
	// available은 LSP 서버가 사용 가능한지 여부입니다.
	available bool
}

// FindReferences는 미리 설정된 location 목록 또는 오류를 반환합니다.
func (m *mockLSPReferencesClient) FindReferences(_ context.Context, _ string, _ lsp.Position) ([]lsp.Location, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.locations, nil
}

// IsAvailable은 LSP 서버 가용성을 반환합니다.
func (m *mockLSPReferencesClient) IsAvailable() bool {
	return m.available
}

// TestLSPFanInCounter_BasicCount는 LSP mock이 3개의 references를 반환할 때
// fan_in=3, method="lsp"를 반환하는지 확인합니다.
// AC-SPC-004-02: fan_in_method가 "lsp"인 경우 포함.
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

// TestLSPFanInCounter_TextualFallback는 LSP 클라이언트가 nil일 때
// TextualFanInCounter로 fallback하는지 확인합니다.
// AC-SPC-004-07: LSP 없으면 textual 방식으로 fallback, method="textual" 반환.
func TestLSPFanInCounter_TextualFallback(t *testing.T) {
	// Arrange: nil 클라이언트로 LSP 미사용 시나리오
	counter := &LSPFanInCounter{
		Client:      nil, // LSP 사용 불가
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

// TestLSPFanInCounter_StrictMode_LSPRequiredError는 MOAI_MX_QUERY_STRICT=1 설정 시
// LSP 사용 불가이면 LSPRequiredError를 반환하는지 확인합니다.
// AC-SPC-004-09: strictMode + LSP unavailable → exit non-zero + LSPRequired.
func TestLSPFanInCounter_StrictMode_LSPRequiredError(t *testing.T) {
	// Arrange
	t.Setenv("MOAI_MX_QUERY_STRICT", "1")
	counter := &LSPFanInCounter{
		Client:      nil, // LSP 사용 불가
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

// TestLSPFanInCounter_ExcludeTests는 excludeTests=true일 때
// _test.go 파일에서의 references를 제외하는지 확인합니다.
// AC-SPC-004-11: 테스트 파일 참조 제외.
func TestLSPFanInCounter_ExcludeTests(t *testing.T) {
	// Arrange: 3개 references 중 1개는 _test.go 파일
	locations := []lsp.Location{
		{URI: "file:///project/internal/a.go", Range: lsp.Range{}},
		{URI: "file:///project/internal/b_test.go", Range: lsp.Range{}}, // 테스트 파일
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

	// Assert: 테스트 제외 시 count가 1 적어야 함
	if countExcluded != 2 {
		t.Errorf("excludeTests=true: fan_in 2 기대, 실제 %d", countExcluded)
	}
	if countAll != 3 {
		t.Errorf("excludeTests=false: fan_in 3 기대, 실제 %d", countAll)
	}
}

// TestLSPFanInCounter_InterfaceCompliance는 LSPFanInCounter가
// FanInCounter 인터페이스를 구현하는지 컴파일 타임에 확인합니다.
func TestLSPFanInCounter_InterfaceCompliance(t *testing.T) {
	var _ FanInCounter = &LSPFanInCounter{}
}

// TestLSPFanInCounter_LSPErrorFallback은 LSP 클라이언트가 오류를 반환할 때
// textual fallback이 작동하는지 확인합니다.
func TestLSPFanInCounter_LSPErrorFallback(t *testing.T) {
	// Arrange: LSP 클라이언트가 오류 반환
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

	// Assert: LSP 오류 시 textual fallback으로 전환, 오류 없어야 함
	if err != nil {
		t.Fatalf("LSP 오류 시 textual fallback 중 예기치 않은 오류: %v", err)
	}
	if method != "textual" {
		t.Errorf("LSP 오류 후 fallback method: 기대 'textual', 실제 '%s'", method)
	}
}

// TestLSPFanInCounter_StrictMode_Env 는 MOAI_MX_QUERY_STRICT 환경변수가
// 설정되지 않은 경우 LSP 불가 시 오류 없이 textual fallback으로 전환됨을 확인합니다.
func TestLSPFanInCounter_StrictMode_Env(t *testing.T) {
	// MOAI_MX_QUERY_STRICT 환경변수가 없는 경우
	os.Unsetenv("MOAI_MX_QUERY_STRICT") //nolint:errcheck

	counter := &LSPFanInCounter{
		Client:      nil, // LSP 사용 불가
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

	// Assert: non-strict 모드에서는 오류 없이 textual fallback
	if err != nil {
		t.Fatalf("non-strict fallback 중 예기치 않은 오류: %v", err)
	}
	if method != "textual" {
		t.Errorf("non-strict fallback method: 기대 'textual', 실제 '%s'", method)
	}
}
