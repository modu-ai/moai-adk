package mx

// Callsite represents a single location where an @MX:ANCHOR symbol is referenced.
// Method indicates whether the location was resolved via LSP ("lsp") or
// text-based search ("textual").
//
// @MX:NOTE: [AUTO] Callsite — G-04 위치 정보 타입; ResolveAnchorCallsites의 반환 요소
type Callsite struct {
	// File is the absolute path to the file containing the reference.
	File string `json:"file"`

	// Line is the 1-based line number of the reference.
	Line int `json:"line"`

	// Column is the 1-based column number of the reference (0 when unknown, e.g. textual method).
	Column int `json:"column,omitempty"`

	// Method indicates how the callsite was discovered: "lsp" or "textual".
	Method string `json:"method"`
}
