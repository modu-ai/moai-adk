package gopls_test

// TestPublicSurface_GoplsPackage verifies at compile time that the public API
// surface of the gopls package has not changed since SPEC-UTIL-003 (AC-UTIL-003-012).
//
// Instead of executing code, this file locks the surface through compilability:
// if the type/constant/variable references below compile, the corresponding
// identifiers are still exported.

import (
	"github.com/modu-ai/moai-adk/internal/lsp/gopls"
)

// ─── Type surface ─────────────────────────────────────────────────────────────

// Compile-time surface lock: the following types must remain exported.
var (
	_ gopls.Diagnostic         // type alias → lsp.Diagnostic
	_ gopls.Range              // type alias → lsp.Range
	_ gopls.Position           // type alias → lsp.Position
	_ gopls.DiagnosticSeverity // type alias → lsp.DiagnosticSeverity
)

// ─── Constant surface ─────────────────────────────────────────────────────────

var (
	_ = gopls.SeverityError
	_ = gopls.SeverityWarning
	_ = gopls.SeverityInformation
	_ = gopls.SeverityHint
)

// ─── Other message-type surface ───────────────────────────────────────────────

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
