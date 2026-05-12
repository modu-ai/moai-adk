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
			return 0, "", &LSPRequiredError{Language: c.language()}
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
			return 0, "", &LSPRequiredError{Language: c.language()}
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

// language는 LSPRequiredError에 사용할 언어 식별자를 반환합니다.
// Language 필드가 비어있으면 "unknown"을 반환합니다.
func (c *LSPFanInCounter) language() string {
	if c.Language == "" {
		return "unknown"
	}
	return c.Language
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

// uriToPath는 LSP URI를 파일시스템 경로로 변환합니다.
//
// 변환 규칙:
//   - "file:///path/to/file.go"  → "/path/to/file.go"   (Unix)
//   - "file:///C:/path/file.go"  → "C:/path/file.go"    (Windows)
//   - "file://path"              → "path"                (scheme-only, passthrough)
//   - 기타 URI                   → uri 그대로 반환
func uriToPath(uri string) string {
	const fileTripleSlash = "file:///"
	const fileDoubleSlash = "file://"

	if strings.HasPrefix(uri, fileTripleSlash) {
		path := strings.TrimPrefix(uri, fileTripleSlash)
		// Windows: "C:/path" — 드라이브 문자 감지 (길이 >= 2, 두 번째 문자가 ':')
		if len(path) >= 2 && path[1] == ':' {
			return path
		}
		// Unix: "/path" 형식으로 복원
		return "/" + path
	}

	if strings.HasPrefix(uri, fileDoubleSlash) {
		return strings.TrimPrefix(uri, fileDoubleSlash)
	}

	return uri
}
