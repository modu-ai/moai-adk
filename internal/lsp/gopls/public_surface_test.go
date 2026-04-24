package gopls_test

// TestPublicSurface_GoplsPackage는 SPEC-UTIL-003 이후 gopls 패키지의 공개 API 표면이
// 변경되지 않았음을 컴파일 타임에 확인한다 (AC-UTIL-003-012).
//
// 이 테스트 파일은 코드를 실행하는 대신 컴파일 가능성을 통해 표면을 잠근다.
// 아래의 변수/상수/타입 참조가 컴파일되면, 해당 식별자가 여전히 export되어 있다는 증거다.

import (
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// ─── 타입 surface ─────────────────────────────────────────────────────────────

// 컴파일 타임 표면 고정: 아래 타입들이 export되어 있어야 한다.
var (
	_ gopls.Diagnostic         // type alias → lsp.Diagnostic
	_ gopls.Range              // type alias → lsp.Range
	_ gopls.Position           // type alias → lsp.Position
	_ gopls.DiagnosticSeverity // type alias → lsp.DiagnosticSeverity
)

// ─── 상수 surface ─────────────────────────────────────────────────────────────

var (
	_ = gopls.SeverityError
	_ = gopls.SeverityWarning
	_ = gopls.SeverityInformation
	_ = gopls.SeverityHint
)

// ─── 기타 메시지 타입 surface ────────────────────────────────────────────────

var (
	_ gopls.Request
	_ gopls.Notification
	_ gopls.Response
	_ gopls.ResponseError
	_ gopls.InitializeParams
	_ gopls.ClientCapabilities
	_ gopls.TextDocumentClientCapabilities
	_ gopls.PublishDiagnosticsClientCapabilities
	_ gopls.InitializeResult
	_ gopls.ServerCapabilities
	_ gopls.InitializedParams
	_ gopls.DidOpenTextDocumentParams
	_ gopls.TextDocumentItem
	_ gopls.PublishDiagnosticsParams
	_ gopls.ShutdownParams
	_ gopls.ExitParams
)
