package core

import (
	"encoding/json"
	"fmt"
)

// TextDocumentClientCapabilities holds text document capabilities declared to the server.
// These match the LSP 3.17 specification minimum for the three query operations.
type TextDocumentClientCapabilities struct {
	// PublishDiagnostics indicates the client supports diagnostic push notifications.
	PublishDiagnostics *DiagnosticClientCapability `json:"publishDiagnostics,omitempty"`

	// References indicates the client supports find-references requests.
	References *ReferenceClientCapability `json:"references,omitempty"`

	// Definition indicates the client supports go-to-definition requests.
	Definition *DefinitionClientCapability `json:"definition,omitempty"`
}

// DiagnosticClientCapability signals that the client accepts publishDiagnostics notifications.
type DiagnosticClientCapability struct {
	// RelatedInformation indicates support for related diagnostic information.
	RelatedInformation bool `json:"relatedInformation,omitempty"`
}

// ReferenceClientCapability signals that the client can send textDocument/references requests.
type ReferenceClientCapability struct {
	// DynamicRegistration indicates support for dynamic capability registration.
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// DefinitionClientCapability signals that the client can send textDocument/definition requests.
type DefinitionClientCapability struct {
	// DynamicRegistration indicates support for dynamic capability registration.
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
}

// ClientCapabilities declares the client capabilities sent in the LSP initialize request
// (REQ-LC-032). Only the minimum capabilities required for our three query operations
// are declared.
type ClientCapabilities struct {
	// TextDocument groups text document capabilities.
	TextDocument TextDocumentClientCapabilities `json:"textDocument"`
}

// DefaultClientCapabilities returns the standard capabilities declared by this client.
// Enables diagnostics push, references, and definition queries.
func DefaultClientCapabilities() ClientCapabilities {
	return ClientCapabilities{
		TextDocument: TextDocumentClientCapabilities{
			PublishDiagnostics: &DiagnosticClientCapability{
				RelatedInformation: true,
			},
			References: &ReferenceClientCapability{},
			Definition: &DefinitionClientCapability{},
		},
	}
}

// serverCapabilitiesRaw는 initialize 응답에서 파싱하는 내부 타입.
// 최소한의 필드만 파싱하며 알 수 없는 필드는 무시합니다.
type serverCapabilitiesRaw struct {
	TextDocumentSync   any  `json:"textDocumentSync"`
	ReferencesProvider bool `json:"referencesProvider"`
	DefinitionProvider bool `json:"definitionProvider"`
}

// ServerCapabilities captures the capabilities declared by the language server
// in its initialize response (REQ-LC-033).
type ServerCapabilities struct {
	// TextDocumentSync indicates the server supports text document synchronization.
	// When non-zero, diagnostic push is assumed enabled (v1 simplification).
	TextDocumentSync int

	// ReferencesProvider indicates the server supports textDocument/references.
	ReferencesProvider bool

	// DefinitionProvider indicates the server supports textDocument/definition.
	DefinitionProvider bool

	// DiagnosticProvider is true if the server declared any text document sync,
	// indicating it will push textDocument/publishDiagnostics notifications.
	DiagnosticProvider bool
}

// ParseServerCapabilities parses the capabilities field from a LSP initialize response.
// Unknown fields in raw are silently ignored (tolerant parsing).
//
// Returns (ServerCapabilities{}, nil) for empty raw input.
func ParseServerCapabilities(raw json.RawMessage) (ServerCapabilities, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return ServerCapabilities{}, nil
	}

	var rawCaps serverCapabilitiesRaw
	if err := json.Unmarshal(raw, &rawCaps); err != nil {
		return ServerCapabilities{}, fmt.Errorf("parse server capabilities: %w", err)
	}

	syncKind := parseTextDocumentSync(rawCaps.TextDocumentSync)

	return ServerCapabilities{
		TextDocumentSync:   syncKind,
		ReferencesProvider: rawCaps.ReferencesProvider,
		DefinitionProvider: rawCaps.DefinitionProvider,
		DiagnosticProvider: syncKind > 0,
	}, nil
}

// parseTextDocumentSync extracts the sync kind integer from the potentially
// polymorphic textDocumentSync field (can be int or object with {change: int}).
func parseTextDocumentSync(v any) int {
	if v == nil {
		return 0
	}
	// JSON 숫자는 float64로 디코딩됨
	switch vt := v.(type) {
	case float64:
		return int(vt)
	case map[string]any:
		if change, ok := vt["change"]; ok {
			if f, ok := change.(float64); ok {
				return int(f)
			}
		}
	}
	return 0
}

// Supports reports whether the server advertises support for the given LSP method.
// Returns false for unknown methods.
//
// Supported methods: "textDocument/publishDiagnostics", "textDocument/references",
// "textDocument/definition".
func (sc ServerCapabilities) Supports(method string) bool {
	switch method {
	case "textDocument/publishDiagnostics":
		return sc.DiagnosticProvider
	case "textDocument/references":
		return sc.ReferencesProvider
	case "textDocument/definition":
		return sc.DefinitionProvider
	default:
		return false
	}
}
