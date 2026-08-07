//go:build !cgo

package astx

// extractImpl is the non-CGO stub: every language returns Supported: false
// because tree-sitter requires CGO. The build still compiles and Extract
// never panics (REQ-NT-015).
func extractImpl(_ string, _ string) (SymbolSet, error) {
	return SymbolSet{Supported: false}, nil
}
