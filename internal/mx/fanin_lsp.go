package mx

import (
	"context"
	"os"
	"strings"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// LSPReferencesClient는 LSPFanInCounter가 LSP 서버와 통신할 때 사용하는 인터페이스입니다.
// internal/lsp/core.Client의 서브셋으로, 테스트 가용성을 높이기 위해 별도 정의합니다 (REQ-SPC-004-003).
//
// @MX:NOTE: [AUTO] LSPReferencesClient — core.Client 의존 없이 mx 패키지 내부에서 LSP 참조 질의를 추상화.
// internal/lsp/core.Client의 FindReferences + IsAvailable subset만 노출.
type LSPReferencesClient interface {
	// FindReferences는 주어진 파일의 position에서 심볼의 모든 참조 위치를 반환합니다.
	// Returns ErrCapabilityUnsupported if server does not support references.
	FindReferences(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error)

	// IsAvailable은 LSP 서버가 현재 사용 가능한지 확인합니다.
	// nil 클라이언트 또는 서버 미시작 상태에서 false를 반환합니다.
	IsAvailable() bool
}

// LSPFanInCounter는 powernap LSP 클라이언트를 사용하여 fan-in을 계산하는 구현체입니다.
// LSP 사용 불가 시 TextualFanInCounter로 fallback합니다 (REQ-SPC-004-020).
//
// @MX:ANCHOR: [AUTO] LSPFanInCounter — FanInCounter 인터페이스의 LSP 구현체
// @MX:REASON: fan_in >= 3 — Resolver.Resolve(), CLI mx_query.go, M6 sweep test 모두 이 구현체를 사용
type LSPFanInCounter struct {
	// Client는 LSP textDocument/references 질의에 사용하는 클라이언트입니다.
	// nil이면 textual fallback으로 전환됩니다.
	Client LSPReferencesClient

	// ProjectRoot는 프로젝트 루트 디렉토리 경로입니다.
	// TextualFanInCounter fallback에도 사용됩니다.
	ProjectRoot string

	// Language는 LSPRequiredError 보고 시 사용하는 언어 식별자입니다.
	// 기본값은 "unknown"입니다.
	Language string
}

// Count는 주어진 Tag의 fan-in (caller count)을 계산합니다.
// LSP 사용 가능 시 textDocument/references를 통해 정확한 참조 수를 반환합니다.
// LSP 사용 불가 시 TextualFanInCounter로 fallback합니다.
//
// strictMode (MOAI_MX_QUERY_STRICT=1) 설정 시:
// - LSP 사용 불가이면 LSPRequiredError를 반환합니다 (REQ-SPC-004-030).
func (c *LSPFanInCounter) Count(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	// strictMode 감지 (REQ-SPC-004-030)
	strictMode := os.Getenv("MOAI_MX_QUERY_STRICT") == "1"

	// LSP 가용성 확인
	if !c.isLSPAvailable() {
		if strictMode {
			lang := c.Language
			if lang == "" {
				lang = "unknown"
			}
			return 0, "", &LSPRequiredError{Language: lang}
		}
		// LSP 사용 불가 → textual fallback
		return c.textualFallback(ctx, tag, projectRoot, excludeTests)
	}

	// LSP 경로: textDocument/references 호출
	pos := lsp.Position{
		Line:      tag.Line - 1, // LSP는 0-based, Tag.Line은 1-based
		Character: 0,
	}
	locations, err := c.Client.FindReferences(ctx, tag.File, pos)
	if err != nil {
		// LSP 오류 → textual fallback (non-strict 모드에서는 graceful)
		if strictMode {
			lang := c.Language
			if lang == "" {
				lang = "unknown"
			}
			return 0, "", &LSPRequiredError{Language: lang}
		}
		return c.textualFallback(ctx, tag, projectRoot, excludeTests)
	}

	// excludeTests 적용: _test.go 및 testdata 경로 제외 (REQ-SPC-004-040)
	count := 0
	for _, loc := range locations {
		filePath := uriToPath(loc.URI)
		if excludeTests && isTestFile(filePath) {
			continue
		}
		count++
	}

	return count, "lsp", nil
}

// isLSPAvailable은 LSP 클라이언트가 사용 가능한지 확인합니다.
func (c *LSPFanInCounter) isLSPAvailable() bool {
	if c.Client == nil {
		return false
	}
	return c.Client.IsAvailable()
}

// textualFallback은 TextualFanInCounter를 사용하여 fan-in을 계산합니다.
func (c *LSPFanInCounter) textualFallback(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	root := projectRoot
	if root == "" {
		root = c.ProjectRoot
	}
	fallback := &TextualFanInCounter{ProjectRoot: root}
	return fallback.Count(ctx, tag, root, excludeTests)
}

// uriToPath는 LSP URI (file:///path/to/file.go)를 파일시스템 경로로 변환합니다.
// Windows 경로 형식 (file:///C:/...)도 처리합니다.
func uriToPath(uri string) string {
	const fileScheme = "file://"
	if !strings.HasPrefix(uri, fileScheme) {
		return uri
	}
	path := strings.TrimPrefix(uri, fileScheme)
	// Unix: file:///path → /path (슬래시 하나 제거)
	// Windows: file:///C:/path → C:/path (슬래시 세 개 제거 후 드라이브 문자)
	if len(path) > 0 && path[0] == '/' {
		// 추가 슬래시 제거 (file:/// → /)
		// Unix: //path → /path
		if len(path) > 1 && path[1] != '/' {
			return path
		}
	}
	return path
}
