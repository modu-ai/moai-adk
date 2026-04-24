package gopls

import (
	"encoding/json"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// ─── JSON-RPC 2.0 봉투 타입 ────────────────────────────────────────────────
//
// REQ-GB-030: LSP 메시지는 Content-Length 헤더로 프레이밍된 JSON-RPC 2.0 형식이다.
// REQ-GB-033: id 필드로 요청-응답을 상관시킨다.
// REQ-GB-034: id가 없는 메시지는 알림으로 처리한다.

// Request는 JSON-RPC 2.0 요청 봉투다. 클라이언트가 서버에 보내는 메시지에 사용한다.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Notification은 id 없는 JSON-RPC 2.0 알림 봉투다.
// 클라이언트가 서버에, 또는 서버가 클라이언트에 단방향으로 보낸다.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response는 서버에서 클라이언트로 오는 JSON-RPC 2.0 응답 봉투다.
// 응답(id 있음)과 알림(id 없음, Method 있음) 양쪽을 표현한다.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// IsNotification은 이 메시지가 알림(id 없음)인지 판별한다.
// REQ-GB-034: id 필드가 없으면 알림으로 분류한다.
func (r *Response) IsNotification() bool {
	return len(r.ID) == 0
}

// ResponseError는 JSON-RPC 2.0 에러 객체다.
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ─── LSP 초기화 메시지 ─────────────────────────────────────────────────────
//
// REQ-GB-010: initialize 요청 params
// REQ-GB-011: initialized 알림
// REQ-GB-013: initializationOptions.staticcheck: true

// InitializeParams는 LSP `initialize` 요청의 파라미터다.
type InitializeParams struct {
	// RootURI는 프로젝트 루트 디렉토리의 파일 URI다.
	RootURI string `json:"rootUri"`
	// ClientCapabilities는 클라이언트가 지원하는 기능 목록이다.
	ClientCapabilities ClientCapabilities `json:"capabilities"`
	// InitializationOptions는 서버별 초기화 옵션이다.
	// REQ-GB-013: gopls에서 staticcheck를 활성화한다.
	InitializationOptions map[string]any `json:"initializationOptions,omitempty"`
}

// ClientCapabilities는 클라이언트가 지원하는 기능 집합이다.
type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

// TextDocumentClientCapabilities는 텍스트 문서 관련 클라이언트 기능이다.
type TextDocumentClientCapabilities struct {
	PublishDiagnostics PublishDiagnosticsClientCapabilities `json:"publishDiagnostics,omitempty"`
}

// PublishDiagnosticsClientCapabilities는 publishDiagnostics 알림 관련 클라이언트 기능이다.
// REQ-GB-010: relatedInformation: true를 설정해야 한다.
type PublishDiagnosticsClientCapabilities struct {
	RelatedInformation bool `json:"relatedInformation,omitempty"`
}

// InitializeResult는 LSP `initialize` 응답의 결과다.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities는 gopls 서버가 지원하는 기능 집합이다.
// 현재는 사용하지 않지만 구조체 역직렬화를 위해 정의한다.
type ServerCapabilities struct{}

// InitializedParams는 LSP `initialized` 알림의 파라미터다. 항상 빈 객체다.
// REQ-GB-011: initialize 응답 수신 후 전송한다.
type InitializedParams struct{}

// ─── 텍스트 문서 메시지 ────────────────────────────────────────────────────

// DidOpenTextDocumentParams는 LSP `textDocument/didOpen` 알림의 파라미터다.
// REQ-GB-020: 파일을 열어 diagnostics를 수집한다.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentItem은 LSP 텍스트 문서를 표현한다.
type TextDocumentItem struct {
	// URI는 문서의 파일 URI다. (예: "file:///workspace/main.go")
	URI string `json:"uri"`
	// LanguageID는 언어 식별자다. Go 파일은 "go"다.
	LanguageID string `json:"languageId"`
	// Version은 문서 버전 번호다. 1부터 시작한다.
	Version int `json:"version"`
	// Text는 문서의 전체 텍스트 내용이다.
	Text string `json:"text"`
}

// ─── 진단 메시지 ──────────────────────────────────────────────────────────
//
// REQ-GB-023: severity, source, code, message, range 필드를 포함해야 한다.
// REQ-UTIL-003-007: gopls.Diagnostic / Range / Position / DiagnosticSeverity는
// lsp 패키지 정의의 타입 별칭이다. 단일 출처(single source of truth) 보장.
// 기존 gopls 호출자는 타입 별칭의 동일성(identity) 의미론으로 수정 없이 컴파일된다.

// PublishDiagnosticsParams는 `textDocument/publishDiagnostics` 알림의 파라미터다.
type PublishDiagnosticsParams struct {
	// URI는 이 진단이 속하는 문서의 파일 URI다.
	URI string `json:"uri"`
	// Diagnostics는 이 문서에 대한 진단 목록이다. 빈 슬라이스면 문제 없음을 의미한다.
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostic은 lsp.Diagnostic의 타입 별칭이다 (REQ-UTIL-003-007).
// reflect.TypeOf(gopls.Diagnostic{}) == reflect.TypeOf(lsp.Diagnostic{})가 보장된다.
type Diagnostic = lsp.Diagnostic

// Range는 lsp.Range의 타입 별칭이다 (REQ-UTIL-003-007).
type Range = lsp.Range

// Position은 lsp.Position의 타입 별칭이다 (REQ-UTIL-003-007).
type Position = lsp.Position

// DiagnosticSeverity는 lsp.DiagnosticSeverity의 타입 별칭이다 (REQ-UTIL-003-007).
// LSP 3.17 사양의 DiagnosticSeverity 값과 일치한다 (int 기반, 1=Error, 2=Warning, 3=Info, 4=Hint).
type DiagnosticSeverity = lsp.DiagnosticSeverity

const (
	// SeverityError는 오류 진단이다 (값: 1).
	SeverityError DiagnosticSeverity = 1
	// SeverityWarning은 경고 진단이다 (값: 2).
	SeverityWarning DiagnosticSeverity = 2
	// SeverityInformation은 정보성 진단이다 (값: 3).
	// 참고: lsp 패키지는 동일 값을 SeverityInfo로 명명한다.
	SeverityInformation DiagnosticSeverity = 3
	// SeverityHint는 힌트 진단이다 (값: 4).
	SeverityHint DiagnosticSeverity = 4
)

// ─── 종료 메시지 ──────────────────────────────────────────────────────────
//
// REQ-GB-004: shutdown/exit 시퀀스로 gopls를 종료한다.

// ShutdownParams는 LSP `shutdown` 요청의 파라미터다. 항상 null이다.
type ShutdownParams struct{}

// ExitParams는 LSP `exit` 알림의 파라미터다. 항상 null이다.
type ExitParams struct{}
