package lsp

import (
	"encoding/json"
	"errors"
)

// Sentinel errors for LSP operations.
var (
	// ErrServerNotRunning indicates a request was made to a server that is not running.
	ErrServerNotRunning = errors.New("lsp: server not running")

	// ErrServerStartFailed indicates a language server process failed to start.
	ErrServerStartFailed = errors.New("lsp: server failed to start")

	// ErrInitializeFailed indicates the LSP initialize handshake failed.
	ErrInitializeFailed = errors.New("lsp: initialization failed")

	// ErrConnectionClosed indicates the connection to the language server was closed.
	ErrConnectionClosed = errors.New("lsp: connection closed")
)

// DiagnosticSeverity represents the severity level of a diagnostic.
// Values match the LSP 3.17 specification.
type DiagnosticSeverity int

const (
	// SeverityError reports an error (severity 1).
	SeverityError DiagnosticSeverity = 1

	// SeverityWarning reports a warning (severity 2).
	SeverityWarning DiagnosticSeverity = 2

	// SeverityInfo reports an information message (severity 3).
	SeverityInfo DiagnosticSeverity = 3

	// SeverityHint reports a hint (severity 4).
	SeverityHint DiagnosticSeverity = 4
)

// String returns the human-readable name of the severity level.
func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityError:
		return "Error"
	case SeverityWarning:
		return "Warning"
	case SeverityInfo:
		return "Information"
	case SeverityHint:
		return "Hint"
	default:
		return "Unknown"
	}
}

// Position represents a zero-based position in a text document.
type Position struct {
	// Line is the zero-based line number.
	Line int `json:"line"`

	// Character is the zero-based character offset on the line.
	Character int `json:"character"`
}

// Range represents a range in a text document defined by start and end positions.
type Range struct {
	// Start is the range's start position (inclusive).
	Start Position `json:"start"`

	// End is the range's end position (exclusive).
	End Position `json:"end"`
}

// Contains reports whether the given position is within this range.
func (r Range) Contains(pos Position) bool {
	if r.Start.Line == r.End.Line {
		return pos.Line == r.Start.Line &&
			pos.Character >= r.Start.Character &&
			pos.Character <= r.End.Character
	}
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line {
		return pos.Character >= r.Start.Character
	}
	if pos.Line == r.End.Line {
		return pos.Character <= r.End.Character
	}
	return true
}

// Location represents a location inside a resource, such as a line in a text file.
type Location struct {
	// URI is the resource identifier (e.g., "file:///path/to/file.go").
	URI string `json:"uri"`

	// Range is the range within the resource.
	Range Range `json:"range"`
}

// Diagnostic represents a compiler error, warning, or informational message.
type Diagnostic struct {
	// Range is the range at which the message applies.
	Range Range `json:"range"`

	// Severity is the diagnostic's severity level.
	Severity DiagnosticSeverity `json:"severity"`

	// Code is the diagnostic's code (e.g., "E0001"). May be empty.
	Code string `json:"code,omitempty"`

	// Source identifies the tool that produced this diagnostic (e.g., "gopls").
	Source string `json:"source,omitempty"`

	// Message is the diagnostic's human-readable message.
	Message string `json:"message"`
}

// IsError reports whether this diagnostic is an error.
func (d Diagnostic) IsError() bool {
	return d.Severity == SeverityError
}

// IsWarning reports whether this diagnostic is a warning.
func (d Diagnostic) IsWarning() bool {
	return d.Severity == SeverityWarning
}

// HoverResult represents the result of a hover request.
type HoverResult struct {
	// Contents is the hover information content (may be markdown).
	Contents string `json:"contents"`

	// Range is the optional range for the symbol being hovered.
	Range *Range `json:"range,omitempty"`
}

// SymbolKind represents the kind of a document symbol.
// Values match the LSP 3.17 specification.
type SymbolKind int

// LSP symbol kind constants.
const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// DocumentSymbol represents a programming construct like a variable, class, or function
// that appears in a document. Symbols can be hierarchical via Children.
type DocumentSymbol struct {
	// Name is the symbol's name.
	Name string `json:"name"`

	// Kind is the symbol's kind (function, class, variable, etc.).
	Kind SymbolKind `json:"kind"`

	// Range is the range enclosing this symbol, not including leading/trailing whitespace.
	Range Range `json:"range"`

	// Children contains child symbols (e.g., struct fields, class methods).
	Children []DocumentSymbol `json:"children,omitempty"`
}

// InitializeResult represents the result of an LSP initialize request.
type InitializeResult struct {
	// Capabilities describes the server's capabilities.
	Capabilities json.RawMessage `json:"capabilities"`
}
