// Package complexity measures per-function cyclomatic complexity for MX validation.
// It supports 5 seeded languages (Go, Python, TypeScript, JavaScript, Rust) via
// tree-sitter AST queries. Unsupported languages return Result{Supported: false}.
//
// Build constraints:
//   - CGO enabled: full tree-sitter implementation (measure_cgo.go)
//   - CGO disabled: stub returning Supported: false (measure_nocgo.go)
package complexity

// Result holds complexity metrics for a single function.
type Result struct {
	// Cyclomatic is the McCabe cyclomatic complexity (decision nodes + 1).
	Cyclomatic int
	// IfBranches is the count of direct if / else-if nodes within the function.
	IfBranches int
	// Supported is false when the language has no seeded tree-sitter query.
	Supported bool
}

// maxFileSizeBytes is the file-size cap (1 MiB) to prevent tree-sitter memory exhaustion.
const maxFileSizeBytes = 1 << 20 // 1 MiB

// Measure computes cyclomatic complexity for funcName in the given source content.
//
// Parameters:
//   - lang: language identifier ("go", "python", "typescript", "javascript", "rust",
//     or any of the 11 scaffolded language names).
//   - content: full file source bytes (UTF-8).
//   - funcName: exported function or method name to locate in the AST.
//   - startLine: 1-indexed hint for disambiguation when the same name appears multiple times.
//     Pass 0 to pick the first match.
//
// Returns Result{Supported: false} for:
//   - unsupported/scaffolded languages
//   - content exceeding maxFileSizeBytes (1 MiB)
//   - any parse or query error (logged, never propagated)
//
// Never panics; always returns a nil error (errors are swallowed gracefully
// to avoid blocking the validation pipeline).
//
// The implementation is in measure_cgo.go (CGO build) or measure_nocgo.go (non-CGO build).
func Measure(lang string, content []byte, funcName string, startLine int) (Result, error) {
	return measure(lang, content, funcName, startLine)
}
