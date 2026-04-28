//go:build !cgo

package complexity

// measure is the stub implementation for non-CGO builds.
// All languages return Result{Supported: false} because tree-sitter requires CGO.
func measure(_ string, _ []byte, _ string, _ int) (Result, error) {
	return Result{Supported: false}, nil
}
