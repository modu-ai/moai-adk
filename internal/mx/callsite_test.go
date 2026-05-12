package mx

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// TestResolver_ResolveAnchorCallsites_LSP는 LSP mock이 3개의 Location을 반환할 때
// ResolveAnchorCallsites가 Method="lsp"인 3개 Callsite를 반환하는지 확인합니다.
// AC-SPC-004-02: LSP 기반 callsite 위치 목록 반환.
func TestResolver_ResolveAnchorCallsites_LSP(t *testing.T) {
	// Arrange: Manager with a single ANCHOR tag
	manager := newTestManager(t, []Tag{
		{
			Kind:     MXAnchor,
			AnchorID: "auth-handler",
			File:     "/project/internal/auth.go",
			Line:     10,
		},
	})
	resolver := NewResolver(manager)

	// LSP mock returning 3 references
	locations := []lsp.Location{
		{URI: "file:///project/internal/a.go", Range: lsp.Range{Start: lsp.Position{Line: 4, Character: 12}}},
		{URI: "file:///project/internal/b.go", Range: lsp.Range{Start: lsp.Position{Line: 8, Character: 0}}},
		{URI: "file:///project/internal/c.go", Range: lsp.Range{Start: lsp.Position{Line: 20, Character: 3}}},
	}
	lspClient := &mockLSPReferencesClient{
		locations: locations,
		available: true,
	}

	// Act
	callsites, err := resolver.ResolveAnchorCallsites(context.Background(), "auth-handler", "/project", false, lspClient)

	// Assert
	if err != nil {
		t.Fatalf("예기치 않은 오류: %v", err)
	}
	if len(callsites) != 3 {
		t.Fatalf("callsite 수: 기대 3, 실제 %d", len(callsites))
	}
	for i, cs := range callsites {
		if cs.Method != "lsp" {
			t.Errorf("callsite[%d].Method: 기대 'lsp', 실제 '%s'", i, cs.Method)
		}
	}
	// File and Line spot-check for first callsite
	if callsites[0].File != "/project/internal/a.go" {
		t.Errorf("callsite[0].File: 기대 '/project/internal/a.go', 실제 '%s'", callsites[0].File)
	}
	if callsites[0].Line != 5 { // lsp 0-based → 1-based
		t.Errorf("callsite[0].Line: 기대 5, 실제 %d", callsites[0].Line)
	}
}

// TestResolver_ResolveAnchorCallsites_TextualFallback는 LSP 사용 불가 시
// walk-based textual fallback으로 callsite 목록을 반환하는지 확인합니다.
// AC-SPC-004-07: LSP unavailable → textual fallback.
func TestResolver_ResolveAnchorCallsites_TextualFallback(t *testing.T) {
	// Arrange: tmp directory with fixture files referencing the anchor
	projectRoot := t.TempDir()

	// Create two fixture files containing the anchor symbol
	writeFixture(t, filepath.Join(projectRoot, "caller_a.go"), `package main
// calls auth-handler here
func doA() { _ = "auth-handler" }
`)
	writeFixture(t, filepath.Join(projectRoot, "caller_b.go"), `package main
// another reference to auth-handler
func doB() { _ = "auth-handler" }
`)
	// Create a file that does NOT reference the anchor
	writeFixture(t, filepath.Join(projectRoot, "unrelated.go"), `package main
func unrelated() {}
`)

	manager := newTestManager(t, []Tag{
		{
			Kind:     MXAnchor,
			AnchorID: "auth-handler",
			File:     filepath.Join(projectRoot, "auth.go"),
			Line:     5,
		},
	})
	resolver := NewResolver(manager)

	// LSP client is nil → unavailable
	callsites, err := resolver.ResolveAnchorCallsites(context.Background(), "auth-handler", projectRoot, false, nil)

	// Assert
	if err != nil {
		t.Fatalf("textual fallback 중 예기치 않은 오류: %v", err)
	}
	if len(callsites) < 2 {
		t.Fatalf("textual callsite 수: 최소 2 기대, 실제 %d", len(callsites))
	}
	for i, cs := range callsites {
		if cs.Method != "textual" {
			t.Errorf("callsite[%d].Method: 기대 'textual', 실제 '%s'", i, cs.Method)
		}
		if cs.Column != 0 {
			t.Errorf("callsite[%d].Column: textual fallback은 0 기대, 실제 %d", i, cs.Column)
		}
	}
}

// TestResolver_ResolveAnchor_BackwardCompat는 ResolveAnchor의 시그니처가
// 변경되지 않았음을 확인합니다 (PR #746 baseline 호환).
// G-04: 기존 API 시그니처 불변 보장.
func TestResolver_ResolveAnchor_BackwardCompat(t *testing.T) {
	// Arrange
	manager := newTestManager(t, []Tag{
		{
			Kind:     MXAnchor,
			AnchorID: "my-anchor",
			File:     "/project/internal/x.go",
			Line:     1,
		},
	})
	resolver := NewResolver(manager)

	// Act: invoke with original signature (anchorID string) → (Tag, error)
	tag, err := resolver.ResolveAnchor("my-anchor")

	// Assert
	if err != nil {
		t.Fatalf("ResolveAnchor 예기치 않은 오류: %v", err)
	}
	if tag.AnchorID != "my-anchor" {
		t.Errorf("tag.AnchorID: 기대 'my-anchor', 실제 '%s'", tag.AnchorID)
	}
	if tag.Kind != MXAnchor {
		t.Errorf("tag.Kind: 기대 MXAnchor, 실제 '%s'", tag.Kind)
	}

	// Assert type (compile-time verified by usage above)
	_ = tag
}

// TestResolver_ResolveAnchorCallsites_IncludeTests는 includeTests=false 시
// _test.go 파일의 callsite를 제외하는지 확인합니다.
func TestResolver_ResolveAnchorCallsites_IncludeTests(t *testing.T) {
	// Arrange
	manager := newTestManager(t, []Tag{
		{
			Kind:     MXAnchor,
			AnchorID: "some-anchor",
			File:     "/project/internal/handler.go",
			Line:     5,
		},
	})
	resolver := NewResolver(manager)

	// 3 references: 2 normal + 1 test file
	locations := []lsp.Location{
		{URI: "file:///project/internal/a.go", Range: lsp.Range{Start: lsp.Position{Line: 1, Character: 0}}},
		{URI: "file:///project/internal/b_test.go", Range: lsp.Range{Start: lsp.Position{Line: 3, Character: 0}}},
		{URI: "file:///project/internal/c.go", Range: lsp.Range{Start: lsp.Position{Line: 7, Character: 0}}},
	}
	lspClient := &mockLSPReferencesClient{
		locations: locations,
		available: true,
	}

	// Act: includeTests=false → excludes test files
	csExclude, err := resolver.ResolveAnchorCallsites(context.Background(), "some-anchor", "/project", false, lspClient)
	if err != nil {
		t.Fatalf("includeTests=false 오류: %v", err)
	}

	// Act: includeTests=true → includes test files
	csInclude, err := resolver.ResolveAnchorCallsites(context.Background(), "some-anchor", "/project", true, lspClient)
	if err != nil {
		t.Fatalf("includeTests=true 오류: %v", err)
	}

	if len(csExclude) != 2 {
		t.Errorf("includeTests=false: callsite 2 기대, 실제 %d", len(csExclude))
	}
	if len(csInclude) != 3 {
		t.Errorf("includeTests=true: callsite 3 기대, 실제 %d", len(csInclude))
	}
}

// TestResolver_ResolveAnchorCallsites_AnchorNotFound는 존재하지 않는 AnchorID에 대해
// error를 반환하는지 확인합니다.
func TestResolver_ResolveAnchorCallsites_AnchorNotFound(t *testing.T) {
	manager := newTestManager(t, nil)
	resolver := NewResolver(manager)

	_, err := resolver.ResolveAnchorCallsites(context.Background(), "nonexistent", "/project", false, nil)
	if err == nil {
		t.Fatal("존재하지 않는 AnchorID: 오류 기대했으나 nil 반환됨")
	}
}

// --- helpers ---

// newTestManager는 주어진 Tag 목록을 가진 임시 Manager를 생성합니다.
func newTestManager(t *testing.T, tags []Tag) *Manager {
	t.Helper()
	tmpDir := t.TempDir()

	mgr := NewManager(tmpDir)
	// Inject tags directly via the manager's internal state for testing
	mgr.mu.Lock()
	for _, tag := range tags {
		mgr.currentTags[tag.Key()] = tag
	}
	mgr.mu.Unlock()
	return mgr
}

// writeFixture는 주어진 경로에 파일을 생성하고 내용을 씁니다.
func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("fixture 디렉토리 생성 실패: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("fixture 파일 쓰기 실패: %v", err)
	}
}
