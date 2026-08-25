//go:build !cgo

package astx

// extractImpl is the non-CGO stub: every language returns Supported: false
// because tree-sitter requires CGO. The build still compiles and Extract
// never panics (REQ-NT-015).
func extractImpl(_ string, _ string) (SymbolSet, error) {
	return SymbolSet{Supported: false}, nil
}

// extractCallsImpl mirrors the stub contract for call extraction (SPEC-
// V3R6-GRAPH-FRESHNESS-001): no CGO, no tree-sitter, Supported: false.
func extractCallsImpl(_ string, _ string) (CallSet, error) {
	return CallSet{Supported: false}, nil
}
