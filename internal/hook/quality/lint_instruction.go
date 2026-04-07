package quality

import (
	"fmt"
	"path/filepath"
	"strings"

	astgrep "github.com/modu-ai/moai-adk/internal/astgrep"
	lsphook "github.com/modu-ai/moai-adk/internal/lsp/hook"
)

// maxErrorsInMessage is the maximum number of diagnostic entries to include in a systemMessage.
const maxErrorsInMessage = 10

// FormatDiagnosticsAsInstruction formats LSP diagnostics into a systemMessage string
// suitable for injection into the AI's next prompt (REQ-LAI-001 through REQ-LAI-007).
//
// Returns an empty string when:
//   - counts.Errors == 0 and (counts.Warnings == 0 or warnAsInstruction == false)
//
// The returned message follows the format:
//
//	[Quality Gate] N error(s) detected in filename:
//	- file:line: message (source)
//	- file:line: message (source)
//	... and N more errors
//	Fix these errors before proceeding.
func FormatDiagnosticsAsInstruction(diagnostics []lsphook.Diagnostic, counts lsphook.SeverityCounts, warnAsInstruction bool) string {
	hasErrors := counts.Errors > 0
	hasActionableWarnings := warnAsInstruction && counts.Warnings > 0

	// REQ-LAI-007: no systemMessage when diagnostics are clean.
	if !hasErrors && !hasActionableWarnings {
		return ""
	}

	// Select which diagnostics to format.
	// Errors always take priority; warnings are included only when warnAsInstruction is true.
	var selected []lsphook.Diagnostic
	for _, d := range diagnostics {
		if d.Severity == lsphook.SeverityError {
			selected = append(selected, d)
		}
	}
	if warnAsInstruction && !hasErrors {
		// Only warnings, no errors — include warnings.
		for _, d := range diagnostics {
			if d.Severity == lsphook.SeverityWarning {
				selected = append(selected, d)
			}
		}
	}

	// Determine the filename for the header from the first diagnostic.
	// Fall back to an empty string if no diagnostics are present.
	filename := filenameFromDiagnostics(diagnostics)

	// Determine counts for the header.
	total := len(selected)
	kind := "error"
	if !hasErrors && hasActionableWarnings {
		kind = "warning"
	}

	var sb strings.Builder

	// Header line.
	fmt.Fprintf(&sb, "[Quality Gate] %d %s(s) detected", total, kind)
	if filename != "" {
		fmt.Fprintf(&sb, " in %s", filename)
	}
	sb.WriteString(":\n")

	// REQ-LAI-004: limit to maxErrorsInMessage entries.
	shown := selected
	remaining := 0
	if len(selected) > maxErrorsInMessage {
		shown = selected[:maxErrorsInMessage]
		remaining = len(selected) - maxErrorsInMessage
	}

	for _, d := range shown {
		line := d.Range.Start.Line + 1 // convert 0-based to 1-based
		entry := formatEntry(d, line)
		sb.WriteString("- ")
		sb.WriteString(entry)
		sb.WriteString("\n")
	}

	if remaining > 0 {
		fmt.Fprintf(&sb, "... and %d more errors\n", remaining)
	}

	sb.WriteString("Fix these errors before proceeding.")
	return sb.String()
}

// formatEntry formats a single diagnostic entry as "file:line: message (source)".
func formatEntry(d lsphook.Diagnostic, line int) string {
	// For now we cannot infer the file name from lsphook.Diagnostic directly,
	// so we use a generic placeholder. The caller enriches the message header
	// with the file name from the tool input.
	msg := d.Message
	if d.Source != "" {
		return fmt.Sprintf("line %d: %s (%s)", line, msg, d.Source)
	}
	return fmt.Sprintf("line %d: %s", line, msg)
}

// FormatDiagnosticsAsInstructionWithFile is the primary public API for systemMessage
// generation. It enriches diagnostic entries with the concrete file path so the
// AI can navigate directly to the error location (AC-LAI-002).
//
// filePath is the absolute or relative path to the modified file.
// If filePath is empty, file information is omitted from individual entries.
func FormatDiagnosticsAsInstructionWithFile(filePath string, diagnostics []lsphook.Diagnostic, counts lsphook.SeverityCounts, warnAsInstruction bool) string {
	hasErrors := counts.Errors > 0
	hasActionableWarnings := warnAsInstruction && counts.Warnings > 0

	// REQ-LAI-007: no systemMessage when diagnostics are clean.
	if !hasErrors && !hasActionableWarnings {
		return ""
	}

	// Select which diagnostics to format.
	var selected []lsphook.Diagnostic
	for _, d := range diagnostics {
		if d.Severity == lsphook.SeverityError {
			selected = append(selected, d)
		}
	}
	if warnAsInstruction && !hasErrors {
		for _, d := range diagnostics {
			if d.Severity == lsphook.SeverityWarning {
				selected = append(selected, d)
			}
		}
	}

	base := filepath.Base(filePath)
	if filePath == "" {
		base = ""
	}

	total := len(selected)
	kind := "error"
	if !hasErrors && hasActionableWarnings {
		kind = "warning"
	}

	var sb strings.Builder

	// Header: "[Quality Gate] N error(s) detected in filename:"
	fmt.Fprintf(&sb, "[Quality Gate] %d %s(s) detected", total, kind)
	if base != "" {
		fmt.Fprintf(&sb, " in %s", base)
	}
	sb.WriteString(":\n")

	// REQ-LAI-004: limit to maxErrorsInMessage entries.
	shown := selected
	remaining := 0
	if len(selected) > maxErrorsInMessage {
		shown = selected[:maxErrorsInMessage]
		remaining = len(selected) - maxErrorsInMessage
	}

	for _, d := range shown {
		line := d.Range.Start.Line + 1 // 0-based to 1-based
		var entry string
		if base != "" {
			if d.Source != "" {
				entry = fmt.Sprintf("%s:%d: %s (%s)", base, line, d.Message, d.Source)
			} else {
				entry = fmt.Sprintf("%s:%d: %s", base, line, d.Message)
			}
		} else {
			if d.Source != "" {
				entry = fmt.Sprintf("line %d: %s (%s)", line, d.Message, d.Source)
			} else {
				entry = fmt.Sprintf("line %d: %s", line, d.Message)
			}
		}
		sb.WriteString("- ")
		sb.WriteString(entry)
		sb.WriteString("\n")
	}

	// REQ-LAI-004: truncation notice.
	if remaining > 0 {
		fmt.Fprintf(&sb, "... and %d more errors\n", remaining)
	}

	sb.WriteString("Fix these errors before proceeding.")
	return sb.String()
}

// filenameFromDiagnostics returns the base file name from the first diagnostic's Source
// field, or an empty string when no diagnostics exist.
// This is a best-effort helper; FormatDiagnosticsAsInstructionWithFile is preferred.
func filenameFromDiagnostics(diagnostics []lsphook.Diagnostic) string {
	// Diagnostics do not carry a file path in the lsphook.Diagnostic struct;
	// the file is known only at the call site.  Return empty string here and
	// let the caller supply it via FormatDiagnosticsAsInstructionWithFile.
	if len(diagnostics) == 0 {
		return ""
	}
	return ""
}

// AppendAstSecurityFindings appends AST-grep security findings to an existing systemMessage.
// If there are no security matches, the original message is returned unchanged.
// If the existing message is empty and there are security matches, a standalone
// security section is returned (REQ-LAI-008).
//
// filePath is used as context for the header; pass empty string to omit.
func AppendAstSecurityFindings(existing string, filePath string, result *astgrep.ScanResult) string {
	if result == nil || len(result.Matches) == 0 {
		return existing
	}

	base := filepath.Base(filePath)
	if filePath == "" {
		base = ""
	}

	var sb strings.Builder

	// Prepend the existing LSP message when present.
	if existing != "" {
		// Remove the trailing "Fix these errors before proceeding." to merge cleanly.
		trimmed := strings.TrimSuffix(existing, "Fix these errors before proceeding.")
		trimmed = strings.TrimRight(trimmed, "\n")
		sb.WriteString(trimmed)
		sb.WriteString("\n")
	}

	// Security findings header.
	n := len(result.Matches)
	if base != "" {
		fmt.Fprintf(&sb, "[Security Gate] %d security issue(s) detected in %s:\n", n, base)
	} else {
		fmt.Fprintf(&sb, "[Security Gate] %d security issue(s) detected:\n", n)
	}

	// REQ-LAI-004 cap applies per category; use same limit for security findings.
	shown := result.Matches
	remaining := 0
	if len(result.Matches) > maxErrorsInMessage {
		shown = result.Matches[:maxErrorsInMessage]
		remaining = len(result.Matches) - maxErrorsInMessage
	}

	for _, m := range shown {
		entryFile := base
		if entryFile == "" && m.File != "" {
			entryFile = filepath.Base(m.File)
		}
		rule := m.Rule
		msg := m.Message
		if msg == "" {
			msg = m.Text
		}
		if entryFile != "" {
			if rule != "" {
				fmt.Fprintf(&sb, "- %s:%d: %s (%s)\n", entryFile, m.Line, msg, rule)
			} else {
				fmt.Fprintf(&sb, "- %s:%d: %s\n", entryFile, m.Line, msg)
			}
		} else {
			if rule != "" {
				fmt.Fprintf(&sb, "- line %d: %s (%s)\n", m.Line, msg, rule)
			} else {
				fmt.Fprintf(&sb, "- line %d: %s\n", m.Line, msg)
			}
		}
	}

	if remaining > 0 {
		fmt.Fprintf(&sb, "... and %d more security issues\n", remaining)
	}

	sb.WriteString("Fix these errors before proceeding.")
	return sb.String()
}
