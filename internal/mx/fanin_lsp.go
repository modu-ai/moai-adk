package mx

import (
	"context"
	"os"
	"strings"

	lsp "github.com/modu-ai/moai-adk/internal/lsp"
)

// LSPReferencesClient is the interface that LSPFanInCounter uses to talk to
// the LSP server. It is a subset of internal/lsp/core.Client, separately
// defined to improve testability (REQ-SPC-004-003).
//
// @MX:NOTE: [AUTO] LSPReferencesClient — abstracts LSP reference queries inside the mx package without depending on core.Client.
// Exposes only the FindReferences + IsAvailable subset of internal/lsp/core.Client.
type LSPReferencesClient interface {
	// FindReferences returns all reference locations for the symbol at the given file position.
	// Returns ErrCapabilityUnsupported if server does not support references.
	FindReferences(ctx context.Context, path string, pos lsp.Position) ([]lsp.Location, error)

	// IsAvailable reports whether the LSP server is currently available.
	// Returns false when the client is nil or the server has not started.
	IsAvailable() bool
}

// LSPFanInCounter is an implementation that uses the powernap LSP client to
// calculate fan-in. When LSP is unavailable it falls back to
// TextualFanInCounter (REQ-SPC-004-020).
//
// @MX:ANCHOR: [AUTO] LSPFanInCounter — LSP implementation of the FanInCounter interface
// @MX:REASON: fan_in >= 3 — used by Resolver.Resolve(), CLI mx_query.go, and the M6 sweep test
type LSPFanInCounter struct {
	// Client is the client used for LSP textDocument/references queries.
	// When nil, behaviour switches to textual fallback.
	Client LSPReferencesClient

	// ProjectRoot is the project root directory path.
	// Also used by the TextualFanInCounter fallback.
	ProjectRoot string

	// Language is the language identifier used when reporting LSPRequiredError.
	// Defaults to "unknown".
	Language string
}

// Count calculates the fan-in (caller count) of the given Tag.
// When LSP is available it returns an accurate reference count via
// textDocument/references. When LSP is unavailable it falls back to
// TextualFanInCounter.
//
// In strictMode (MOAI_MX_QUERY_STRICT=1):
// - If LSP is unavailable, returns LSPRequiredError (REQ-SPC-004-030).
func (c *LSPFanInCounter) Count(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	// strictMode detection (REQ-SPC-004-030)
	strictMode := os.Getenv("MOAI_MX_QUERY_STRICT") == "1"

	// LSP availability check
	if !c.isLSPAvailable() {
		if strictMode {
			return 0, "", &LSPRequiredError{Language: c.language()}
		}
		// LSP unavailable → textual fallback
		return c.textualFallback(ctx, tag, projectRoot, excludeTests)
	}

	// LSP path: invoke textDocument/references
	pos := lsp.Position{
		Line:      tag.Line - 1, // LSP is 0-based, Tag.Line is 1-based
		Character: 0,
	}
	locations, err := c.Client.FindReferences(ctx, tag.File, pos)
	if err != nil {
		// LSP error → textual fallback (graceful in non-strict mode)
		if strictMode {
			return 0, "", &LSPRequiredError{Language: c.language()}
		}
		return c.textualFallback(ctx, tag, projectRoot, excludeTests)
	}

	// Apply excludeTests: exclude _test.go and testdata paths (REQ-SPC-004-040)
	count := 0
	for _, loc := range locations {
		filePath := uriToPath(loc.URI)
		if excludeTests && isTestFile(filePath) {
			continue
		}
		count++
	}

	return count, "lsp", nil
}

// isLSPAvailable reports whether the LSP client is usable.
func (c *LSPFanInCounter) isLSPAvailable() bool {
	if c.Client == nil {
		return false
	}
	return c.Client.IsAvailable()
}

// language returns the language identifier used for LSPRequiredError.
// Returns "unknown" when the Language field is empty.
func (c *LSPFanInCounter) language() string {
	if c.Language == "" {
		return "unknown"
	}
	return c.Language
}

// textualFallback uses TextualFanInCounter to calculate fan-in.
func (c *LSPFanInCounter) textualFallback(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	root := projectRoot
	if root == "" {
		root = c.ProjectRoot
	}
	fallback := &TextualFanInCounter{ProjectRoot: root}
	return fallback.Count(ctx, tag, root, excludeTests)
}

// uriToPath converts an LSP URI to a filesystem path.
//
// Conversion rules:
//   - "file:///path/to/file.go"  → "/path/to/file.go"   (Unix)
//   - "file:///C:/path/file.go"  → "C:/path/file.go"    (Windows)
//   - "file://path"              → "path"                (scheme-only, passthrough)
//   - other URIs                 → URI returned as-is
func uriToPath(uri string) string {
	const fileTripleSlash = "file:///"
	const fileDoubleSlash = "file://"

	if strings.HasPrefix(uri, fileTripleSlash) {
		path := strings.TrimPrefix(uri, fileTripleSlash)
		// Windows: "C:/path" — detect drive letter (length >= 2, second char is ':')
		if len(path) >= 2 && path[1] == ':' {
			return path
		}
		// Unix: restore "/path" form
		return "/" + path
	}

	if strings.HasPrefix(uri, fileDoubleSlash) {
		return strings.TrimPrefix(uri, fileDoubleSlash)
	}

	return uri
}
