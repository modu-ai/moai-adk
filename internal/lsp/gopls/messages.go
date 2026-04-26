package gopls

import (
	"encoding/json"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// ─── JSON-RPC 2.0 envelope types ─────────────────────────────────────────────
//
// REQ-GB-030: LSP messages use the JSON-RPC 2.0 format framed with Content-Length headers.
// REQ-GB-033: correlate requests and responses using the id field.
// REQ-GB-034: messages without an id are treated as notifications.

// Request is a JSON-RPC 2.0 request envelope. Used for messages sent from client to server.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Notification is a JSON-RPC 2.0 notification envelope without an id.
// Sent one-way from client to server, or from server to client.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is the JSON-RPC 2.0 response envelope received from server to client.
// It represents both responses (with id) and notifications (without id, with Method).
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// IsNotification returns true if this message is a notification (no id).
// REQ-GB-034: messages without an id field are classified as notifications.
func (r *Response) IsNotification() bool {
	return len(r.ID) == 0
}

// ResponseError is a JSON-RPC 2.0 error object.
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ─── LSP initialization messages ─────────────────────────────────────────────
//
// REQ-GB-010: initialize request params
// REQ-GB-011: initialized notification
// REQ-GB-013: initializationOptions.staticcheck: true

// InitializeParams holds the parameters for the LSP `initialize` request.
type InitializeParams struct {
	// RootURI is the file URI of the project root directory.
	RootURI string `json:"rootUri"`
	// ClientCapabilities lists the features supported by the client.
	ClientCapabilities ClientCapabilities `json:"capabilities"`
	// InitializationOptions holds server-specific initialization options.
	// REQ-GB-013: enables staticcheck in gopls.
	InitializationOptions map[string]any `json:"initializationOptions,omitempty"`
}

// ClientCapabilities is the set of features supported by the client.
type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

// TextDocumentClientCapabilities holds the client capabilities for text documents.
type TextDocumentClientCapabilities struct {
	PublishDiagnostics PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty"`
}

// PublishDiagnosticsClientCapabilities holds client capabilities for the publishDiagnostics notification.
// REQ-GB-010: relatedInformation must be set to true.
type PublishDiagnosticsClientCapabilities struct {
	RelatedInformation bool `json:"relatedInformation,omitempty"`
}

// InitializeResult holds the result of the LSP `initialize` response.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities is the set of features supported by the gopls server.
// Currently unused but defined for struct deserialization purposes.
type ServerCapabilities struct{}

// InitializedParams holds the parameters for the LSP `initialized` notification. Always an empty object.
// REQ-GB-011: sent after receiving the initialize response.
type InitializedParams struct{}

// ─── Text document messages ────────────────────────────────────────────────

// DidOpenTextDocumentParams holds the parameters for the LSP `textDocument/didOpen` notification.
// REQ-GB-020: opens a file to collect diagnostics.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentItem represents an LSP text document.
type TextDocumentItem struct {
	// URI is the file URI of the document (e.g., "file:///workspace/main.go").
	URI string `json:"uri"`
	// LanguageID is the language identifier. Go files use "go".
	LanguageID string `json:"languageId"`
	// Version is the document version number. Starts at 1.
	Version int `json:"version"`
	// Text is the full text content of the document.
	Text string `json:"text"`
}

// ─── Diagnostic messages ──────────────────────────────────────────────────
//
// REQ-GB-023: must include severity, source, code, message, and range fields.
// REQ-UTIL-003-007: gopls.Diagnostic / Range / Position / DiagnosticSeverity are
// type aliases for the types defined in the lsp package. This guarantees a single source of truth.
// Existing gopls callers compile without modification due to type alias identity semantics.

// PublishDiagnosticsParams holds the parameters for the `textDocument/publishDiagnostics` notification.
type PublishDiagnosticsParams struct {
	// URI is the file URI of the document to which these diagnostics belong.
	URI string `json:"uri"`
	// Diagnostics is the list of diagnostics for this document. An empty slice means no issues.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostic is a type alias for lsp.Diagnostic (REQ-UTIL-003-007).
// Guarantees reflect.TypeOf(gopls.Diagnostic{}) == reflect.TypeOf(lsp.Diagnostic{}).
type Diagnostic = lsp.Diagnostic

// Range is a type alias for lsp.Range (REQ-UTIL-003-007).
type Range = lsp.Range

// Position is a type alias for lsp.Position (REQ-UTIL-003-007).
type Position = lsp.Position

// DiagnosticSeverity is a type alias for lsp.DiagnosticSeverity (REQ-UTIL-003-007).
// Matches the DiagnosticSeverity values from LSP 3.17 spec (int-based: 1=Error, 2=Warning, 3=Info, 4=Hint).
type DiagnosticSeverity = lsp.DiagnosticSeverity

const (
	// SeverityError is an error diagnostic (value: 1).
	SeverityError DiagnosticSeverity = 1
	// SeverityWarning is a warning diagnostic (value: 2).
	SeverityWarning DiagnosticSeverity = 2
	// SeverityInformation is an informational diagnostic (value: 3).
	// Note: the lsp package names the same value SeverityInfo.
	SeverityInformation DiagnosticSeverity = 3
	// SeverityHint is a hint diagnostic (value: 4).
	SeverityHint DiagnosticSeverity = 4
)

// ─── Shutdown messages ──────────────────────────────────────────────────────
//
// REQ-GB-004: terminates gopls using the shutdown/exit sequence.

// ShutdownParams holds the parameters for the LSP `shutdown` request. Always null.
type ShutdownParams struct{}

// ExitParams holds the parameters for the LSP `exit` notification. Always null.
type ExitParams struct{}
