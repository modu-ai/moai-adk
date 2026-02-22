package hook

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/lsp"
)

// diagnosticsCollector implements LSPDiagnosticsCollector interface.
// It provides LSP diagnostics with CLI fallback per REQ-HOOK-150 through REQ-HOOK-153.
type diagnosticsCollector struct {
	lspClient lsp.DiagnosticsProvider
	fallback  FallbackDiagnostics
}

// NewDiagnosticsCollector creates a new diagnostics collector.
// lspClient may be nil if LSP is not available.
// fallback may be nil if no fallback tools are configured.
func NewDiagnosticsCollector(lspClient lsp.DiagnosticsProvider, fallback FallbackDiagnostics) *diagnosticsCollector {
	return &diagnosticsCollector{
		lspClient: lspClient,
		fallback:  fallback,
	}
}

// GetDiagnostics retrieves diagnostics for the given file path.
// Per REQ-HOOK-151: LSP server takes priority when available.
// Per REQ-HOOK-152: Falls back to CLI tools when LSP unavailable.
// Per REQ-HOOK-153: MUST NOT block file writes on diagnostic failure (observation only).
func (c *diagnosticsCollector) GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Try LSP first per REQ-HOOK-151
	if c.lspClient != nil {
		diagnostics, err := c.tryLSP(ctx, filePath)
		if err == nil {
			return diagnostics, nil
		}
		// LSP failed, fall through to fallback
	}

	// Try fallback per REQ-HOOK-152
	if c.fallback != nil {
		return c.tryFallback(ctx, filePath)
	}

	// No diagnostics available - return empty slice per REQ-HOOK-153
	// This is observation-only, so we don't return an error
	return []Diagnostic{}, nil
}

// tryLSP attempts to get diagnostics from the LSP server.
func (c *diagnosticsCollector) tryLSP(ctx context.Context, filePath string) ([]Diagnostic, error) {
	uri := filePathToURI(filePath)
	lspDiags, err := c.lspClient.Diagnostics(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("LSP diagnostics failed: %w", err)
	}

	// Convert LSP diagnostics to our format
	result := make([]Diagnostic, 0, len(lspDiags))
	for _, d := range lspDiags {
		result = append(result, convertLSPDiagnostic(d))
	}

	return result, nil
}

// tryFallback attempts to get diagnostics from CLI fallback tools.
func (c *diagnosticsCollector) tryFallback(ctx context.Context, filePath string) ([]Diagnostic, error) {
	language := c.fallback.GetLanguage(filePath)
	if !c.fallback.IsAvailable(language) {
		return nil, &ErrDiagnosticsUnavailable{
			Language: language,
			Reason:   "no fallback tool available",
		}
	}

	return c.fallback.RunFallback(ctx, filePath)
}

// GetSeverityCounts calculates severity counts from diagnostics.
func (c *diagnosticsCollector) GetSeverityCounts(diagnostics []Diagnostic) SeverityCounts {
	counts := SeverityCounts{}

	for _, d := range diagnostics {
		switch d.Severity {
		case SeverityError:
			counts.Errors++
		case SeverityWarning:
			counts.Warnings++
		case SeverityInformation:
			counts.Information++
		case SeverityHint:
			counts.Hints++
		}
	}

	return counts
}

// convertLSPDiagnostic converts an LSP diagnostic to our format.
func convertLSPDiagnostic(d lsp.Diagnostic) Diagnostic {
	return Diagnostic{
		Range: Range{
			Start: Position{
				Line:      d.Range.Start.Line,
				Character: d.Range.Start.Character,
			},
			End: Position{
				Line:      d.Range.End.Line,
				Character: d.Range.End.Character,
			},
		},
		Severity: convertLSPSeverity(d.Severity),
		Code:     d.Code,
		Source:   d.Source,
		Message:  d.Message,
	}
}

// convertLSPSeverity converts LSP severity to our severity type.
func convertLSPSeverity(s lsp.DiagnosticSeverity) DiagnosticSeverity {
	switch s {
	case lsp.SeverityError:
		return SeverityError
	case lsp.SeverityWarning:
		return SeverityWarning
	case lsp.SeverityInfo:
		return SeverityInformation
	case lsp.SeverityHint:
		return SeverityHint
	default:
		return SeverityInformation
	}
}

// filePathToURI converts a file path to a file:// URI.
func filePathToURI(filePath string) string {
	// Ensure absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	// Convert to URI format
	absPath = filepath.ToSlash(absPath)
	if !strings.HasPrefix(absPath, "/") {
		absPath = "/" + absPath
	}

	return "file://" + absPath
}

// String returns a string representation of the severity.
func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityError:
		return "Error"
	case SeverityWarning:
		return "Warning"
	case SeverityInformation:
		return "Information"
	case SeverityHint:
		return "Hint"
	default:
		return "Unknown"
	}
}

// FormatDiagnostics formats diagnostics as a human-readable string.
// This is used for hook output per the format specified in SPEC-HOOK-004.
func FormatDiagnostics(filePath string, diagnostics []Diagnostic) string {
	if len(diagnostics) == 0 {
		return fmt.Sprintf("LSP: No diagnostics in %s", filepath.Base(filePath))
	}

	counts := (&diagnosticsCollector{}).GetSeverityCounts(diagnostics)
	var sb strings.Builder

	fmt.Fprintf(&sb, "LSP: %d error(s), %d warning(s) in %s\n",
		counts.Errors, counts.Warnings, filepath.Base(filePath))

	for _, d := range diagnostics {
		fmt.Fprintf(&sb, "  - [%s] Line %d: %s",
			strings.ToUpper(d.Severity.String()), d.Range.Start.Line+1, d.Message)
		if d.Source != "" {
			fmt.Fprintf(&sb, " [%s]", d.Source)
		}
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// Compile-time interface compliance check.
var _ LSPDiagnosticsCollector = (*diagnosticsCollector)(nil)
