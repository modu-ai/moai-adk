package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Initializer handles LSP server lifecycle management.
// Use this interface when you only need to start or stop the server.
type Initializer interface {
	// Initialize sends the LSP initialize request and initialized notification.
	Initialize(ctx context.Context, rootURI string) error

	// Shutdown sends the LSP shutdown request followed by an exit notification.
	Shutdown(ctx context.Context) error
}

// DiagnosticsProvider provides diagnostic information from the language server.
// Use this interface when you only need to retrieve diagnostics.
type DiagnosticsProvider interface {
	// Diagnostics retrieves diagnostics for the given document URI.
	Diagnostics(ctx context.Context, uri string) ([]Diagnostic, error)
}

// NavigationProvider provides code navigation features.
// Use this interface when you only need references or go-to-definition.
type NavigationProvider interface {
	// References returns all reference locations for the symbol at the given position.
	References(ctx context.Context, uri string, pos Position) ([]Location, error)

	// Definition returns the definition location(s) for the symbol at the given position.
	Definition(ctx context.Context, uri string, pos Position) ([]Location, error)
}

// HoverProvider provides hover information for symbols.
// Use this interface when you only need hover documentation.
type HoverProvider interface {
	// Hover returns hover information for the symbol at the given position.
	Hover(ctx context.Context, uri string, pos Position) (*HoverResult, error)
}

// SymbolsProvider provides document symbol information.
// Use this interface when you only need to query document symbols.
type SymbolsProvider interface {
	// Symbols returns the document symbols for the given document URI.
	Symbols(ctx context.Context, uri string) ([]DocumentSymbol, error)
}

// Client composes all LSP capabilities and communicates with a single
// Language Server over JSON-RPC 2.0. All methods accept a context.Context
// for cancellation and timeout control.
//
// This interface composes the following focused interfaces for consumers
// that only need a subset of functionality:
//   - Initializer: Server lifecycle (Initialize, Shutdown)
//   - DiagnosticsProvider: Diagnostic retrieval
//   - NavigationProvider: References and Definition
//   - HoverProvider: Hover information
//   - SymbolsProvider: Document symbols
//
// Example usage:
//
//	client := lsp.NewClient(conn)
//	if err := client.Initialize(ctx, "file:///project"); err != nil { ... }
//	diagnostics, err := client.Diagnostics(ctx, "file:///project/main.go")
//	if err := client.Shutdown(ctx); err != nil { ... }
type Client interface {
	Initializer
	DiagnosticsProvider
	NavigationProvider
	HoverProvider
	SymbolsProvider
}

// lspClient implements the Client interface using a Conn for JSON-RPC communication.
type lspClient struct {
	conn         Conn
	initialized  bool
	capabilities json.RawMessage
}

// Compile-time interface compliance checks.
// lspClient implements all segregated interfaces.
var (
	_ Initializer         = (*lspClient)(nil)
	_ DiagnosticsProvider = (*lspClient)(nil)
	_ NavigationProvider  = (*lspClient)(nil)
	_ HoverProvider       = (*lspClient)(nil)
	_ SymbolsProvider     = (*lspClient)(nil)
	_ Client              = (*lspClient)(nil)
)

// NewClient creates a new LSP Client that communicates over the given connection.
func NewClient(conn Conn) Client {
	return &lspClient{conn: conn}
}

// --- LSP parameter types (internal, not exported) ---

type initializeParams struct {
	ProcessID    int            `json:"processId"`
	RootURI      string         `json:"rootUri"`
	Capabilities map[string]any `json:"capabilities"`
}

type textDocumentIdentifier struct {
	URI string `json:"uri"`
}

type textDocumentPositionParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type referenceParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      referenceContext       `json:"context"`
}

type referenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

type documentParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
}

// --- LSP response types (internal, not exported) ---

type initializeResponse struct {
	Capabilities json.RawMessage `json:"capabilities"`
}

type diagnosticReport struct {
	Kind  string       `json:"kind"`
	Items []Diagnostic `json:"items"`
}

type hoverResponse struct {
	Contents json.RawMessage `json:"contents"`
	Range    *Range          `json:"range,omitempty"`
}

// --- Client method implementations ---

// Initialize performs the LSP initialize handshake.
func (c *lspClient) Initialize(ctx context.Context, rootURI string) error {
	params := initializeParams{
		ProcessID:    os.Getpid(),
		RootURI:      rootURI,
		Capabilities: map[string]any{},
	}

	var result initializeResponse
	if err := c.conn.Call(ctx, "initialize", params, &result); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	c.capabilities = result.Capabilities
	c.initialized = true

	// Send initialized notification.
	if err := c.conn.Notify(ctx, "initialized", map[string]any{}); err != nil {
		return fmt.Errorf("initialized notification: %w", err)
	}

	return nil
}

// Diagnostics retrieves diagnostics for the given document URI using pull diagnostics.
func (c *lspClient) Diagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
	params := documentParams{
		TextDocument: textDocumentIdentifier{URI: uri},
	}

	var report diagnosticReport
	if err := c.conn.Call(ctx, "textDocument/diagnostic", params, &report); err != nil {
		return nil, fmt.Errorf("textDocument/diagnostic: %w", err)
	}

	if report.Items == nil {
		return []Diagnostic{}, nil
	}
	return report.Items, nil
}

// References returns all reference locations for the symbol at the given position.
func (c *lspClient) References(ctx context.Context, uri string, pos Position) ([]Location, error) {
	params := referenceParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     pos,
		Context:      referenceContext{IncludeDeclaration: true},
	}

	var locations []Location
	if err := c.conn.Call(ctx, "textDocument/references", params, &locations); err != nil {
		return nil, fmt.Errorf("textDocument/references: %w", err)
	}

	if locations == nil {
		return []Location{}, nil
	}
	return locations, nil
}

// Hover returns hover information for the symbol at the given position.
func (c *lspClient) Hover(ctx context.Context, uri string, pos Position) (*HoverResult, error) {
	params := textDocumentPositionParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     pos,
	}

	var resp hoverResponse
	if err := c.conn.Call(ctx, "textDocument/hover", params, &resp); err != nil {
		return nil, fmt.Errorf("textDocument/hover: %w", err)
	}

	if len(resp.Contents) == 0 {
		return nil, nil
	}

	contents := parseHoverContents(resp.Contents)
	return &HoverResult{Contents: contents, Range: resp.Range}, nil
}

// Definition returns the definition location(s) for the symbol at the given position.
func (c *lspClient) Definition(ctx context.Context, uri string, pos Position) ([]Location, error) {
	params := textDocumentPositionParams{
		TextDocument: textDocumentIdentifier{URI: uri},
		Position:     pos,
	}

	// LSP definition can return Location, Location[], or LocationLink[].
	// We handle Location[] and single Location.
	var raw json.RawMessage
	if err := c.conn.Call(ctx, "textDocument/definition", params, &raw); err != nil {
		return nil, fmt.Errorf("textDocument/definition: %w", err)
	}

	if len(raw) == 0 || string(raw) == "null" {
		return []Location{}, nil
	}

	// Try as array first.
	var locations []Location
	if err := json.Unmarshal(raw, &locations); err == nil {
		return locations, nil
	}

	// Try as single Location.
	var loc Location
	if err := json.Unmarshal(raw, &loc); err == nil {
		return []Location{loc}, nil
	}

	return []Location{}, nil
}

// Symbols returns the document symbols for the given document URI.
func (c *lspClient) Symbols(ctx context.Context, uri string) ([]DocumentSymbol, error) {
	params := documentParams{
		TextDocument: textDocumentIdentifier{URI: uri},
	}

	var symbols []DocumentSymbol
	if err := c.conn.Call(ctx, "textDocument/documentSymbol", params, &symbols); err != nil {
		return nil, fmt.Errorf("textDocument/documentSymbol: %w", err)
	}

	if symbols == nil {
		return []DocumentSymbol{}, nil
	}
	return symbols, nil
}

// Shutdown sends the LSP shutdown request followed by an exit notification.
func (c *lspClient) Shutdown(ctx context.Context) error {
	if err := c.conn.Call(ctx, "shutdown", nil, nil); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := c.conn.Notify(ctx, "exit", nil); err != nil {
		return fmt.Errorf("exit notification: %w", err)
	}

	return c.conn.Close()
}

// parseHoverContents extracts a string from various LSP hover content formats.
func parseHoverContents(raw json.RawMessage) string {
	// Try MarkupContent: {"kind":"markdown","value":"..."}
	var mc struct {
		Kind  string `json:"kind"`
		Value string `json:"value"`
	}
	if json.Unmarshal(raw, &mc) == nil && mc.Value != "" {
		return mc.Value
	}

	// Try plain string.
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}

	// Fallback: return raw JSON as string.
	return string(raw)
}
