package hook

import (
	"context"
	"encoding/json"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Package-level compiled regexps to avoid repeated compilation.
var (
	reGoVetPattern      = regexp.MustCompile(`\.?/?([^:]+):(\d+):(\d+): (.+)`)
	reTypeScriptPattern = regexp.MustCompile(`([^(]+)\((\d+),(\d+)\): (error|warning) (TS\d+): (.+)`)
)

// FallbackTool represents a CLI tool configuration for fallback diagnostics.
type FallbackTool struct {
	Name       string
	Command    string
	Args       []string
	JSONOutput bool
	Parser     func([]byte) ([]Diagnostic, error)
}

// fallbackDiagnostics implements FallbackDiagnostics interface.
// It provides CLI tool fallback when LSP is unavailable per REQ-HOOK-160 through REQ-HOOK-162.
type fallbackDiagnostics struct {
	mu        sync.RWMutex
	tools     map[string][]FallbackTool
	available map[string]bool
}

// NewFallbackDiagnostics creates a new fallback diagnostics provider.
func NewFallbackDiagnostics() *fallbackDiagnostics {
	fb := &fallbackDiagnostics{
		tools:     make(map[string][]FallbackTool),
		available: make(map[string]bool),
	}
	fb.registerDefaultTools()
	return fb
}

// registerDefaultTools registers fallback tools for supported languages per REQ-HOOK-160.
func (f *fallbackDiagnostics) registerDefaultTools() {
	// Python: ruff check --output-format=json
	f.tools["python"] = []FallbackTool{
		{
			Name:       "ruff",
			Command:    "ruff",
			Args:       []string{"check", "--output-format=json", "{file}"},
			JSONOutput: true,
			Parser:     parseRuffOutput,
		},
	}

	// TypeScript: tsc --pretty false
	f.tools["typescript"] = []FallbackTool{
		{
			Name:       "tsc",
			Command:    "tsc",
			Args:       []string{"--noEmit", "--pretty", "false", "{file}"},
			JSONOutput: false,
			Parser:     func(b []byte) ([]Diagnostic, error) { return parseTypeScriptOutput(string(b)), nil },
		},
	}

	// JavaScript: eslint --format json
	f.tools["javascript"] = []FallbackTool{
		{
			Name:       "eslint",
			Command:    "eslint",
			Args:       []string{"--format", "json", "{file}"},
			JSONOutput: true,
			Parser:     parseESLintOutput,
		},
	}

	// Go: go vet
	f.tools["go"] = []FallbackTool{
		{
			Name:       "go vet",
			Command:    "go",
			Args:       []string{"vet", "{file}"},
			JSONOutput: false,
			Parser:     nil, // Uses custom parsing in RunFallback
		},
	}

	// Rust: cargo clippy --message-format=json
	f.tools["rust"] = []FallbackTool{
		{
			Name:       "clippy",
			Command:    "cargo",
			Args:       []string{"clippy", "--message-format=json", "--", "{file}"},
			JSONOutput: true,
			Parser:     parseClippyOutput,
		},
	}
}

// GetLanguage returns the detected language for a file path.
func (f *fallbackDiagnostics) GetLanguage(filePath string) string {
	ext := strings.ToLower(getExtension(filePath))

	switch ext {
	case ".py", ".pyi":
		return "python"
	case ".ts", ".tsx", ".mts", ".cts":
		return "typescript"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	default:
		return "unknown"
	}
}

// IsAvailable checks if a fallback tool is available for the language.
func (f *fallbackDiagnostics) IsAvailable(language string) bool {
	f.mu.RLock()
	if cached, ok := f.available[language]; ok {
		f.mu.RUnlock()
		return cached
	}
	f.mu.RUnlock()

	tools := f.tools[language]
	if len(tools) == 0 {
		f.cacheAvailability(language, false)
		return false
	}

	// Check if any tool is available
	for _, tool := range tools {
		if _, err := exec.LookPath(tool.Command); err == nil {
			f.cacheAvailability(language, true)
			return true
		}
	}

	f.cacheAvailability(language, false)
	return false
}

// cacheAvailability caches the tool availability result.
func (f *fallbackDiagnostics) cacheAvailability(language string, available bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.available[language] = available
}

// GetToolsForLanguage returns configured tools for a language.
func (f *fallbackDiagnostics) GetToolsForLanguage(language string) []FallbackTool {
	return f.tools[language]
}

// RunFallback executes fallback CLI tool for the given file.
// Returns "diagnostics unavailable" error if no tool is available per REQ-HOOK-162.
func (f *fallbackDiagnostics) RunFallback(ctx context.Context, filePath string) ([]Diagnostic, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	language := f.GetLanguage(filePath)
	tools := f.tools[language]

	if len(tools) == 0 {
		return nil, &ErrDiagnosticsUnavailable{
			Language: language,
			Reason:   "no tools configured",
		}
	}

	// Try each tool in order
	for _, tool := range tools {
		if _, err := exec.LookPath(tool.Command); err != nil {
			continue
		}

		diagnostics, err := f.runTool(ctx, tool, filePath)
		if err == nil {
			return diagnostics, nil
		}
		// Tool failed, try next one
	}

	return nil, &ErrDiagnosticsUnavailable{
		Language: language,
		Reason:   "no working tool found",
	}
}

// runTool executes a single tool and parses its output.
func (f *fallbackDiagnostics) runTool(ctx context.Context, tool FallbackTool, filePath string) ([]Diagnostic, error) {
	args := make([]string, len(tool.Args))
	for i, arg := range tool.Args {
		if arg == "{file}" {
			args[i] = filePath
		} else {
			args[i] = arg
		}
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tool.Command, args...)
	// Linters return non-zero exit code when issues are found, which is expected.
	// We intentionally ignore the error and parse the output regardless.
	output, _ := cmd.CombinedOutput() //nolint:errcheck // linters return non-zero on findings

	// For linters, non-zero exit code often means "found issues" not "error"
	// So we try to parse the output regardless

	if tool.Parser != nil {
		return tool.Parser(output)
	}

	// Special handling for go vet
	if tool.Command == "go" && len(args) > 0 && args[0] == "vet" {
		return parseGoVetOutput(string(output), filePath), nil
	}

	return []Diagnostic{}, nil
}

// parseRuffOutput parses ruff JSON output per REQ-HOOK-161.
func parseRuffOutput(data []byte) ([]Diagnostic, error) {
	if len(data) == 0 {
		return []Diagnostic{}, nil
	}

	var ruffDiags []struct {
		Code     string `json:"code"`
		Message  string `json:"message"`
		Location struct {
			Row    int `json:"row"`
			Column int `json:"column"`
		} `json:"location"`
		EndLocation struct {
			Row    int `json:"row"`
			Column int `json:"column"`
		} `json:"end_location"`
		Filename string `json:"filename"`
	}

	if err := json.Unmarshal(data, &ruffDiags); err != nil {
		return []Diagnostic{}, nil // Return empty on parse error
	}

	result := make([]Diagnostic, 0, len(ruffDiags))
	for _, d := range ruffDiags {
		severity := classifyRuffSeverity(d.Code)
		result = append(result, Diagnostic{
			Range: Range{
				Start: Position{Line: d.Location.Row - 1, Character: d.Location.Column - 1},
				End:   Position{Line: d.EndLocation.Row - 1, Character: d.EndLocation.Column - 1},
			},
			Severity: severity,
			Code:     d.Code,
			Source:   "ruff",
			Message:  d.Message,
		})
	}

	return result, nil
}

// classifyRuffSeverity maps ruff codes to severity levels.
func classifyRuffSeverity(code string) DiagnosticSeverity {
	// E and F codes are typically errors
	if strings.HasPrefix(code, "E") || strings.HasPrefix(code, "F") {
		return SeverityError
	}
	// W codes are warnings
	if strings.HasPrefix(code, "W") {
		return SeverityWarning
	}
	// Default to warning
	return SeverityWarning
}

// parseGoVetOutput parses go vet output per REQ-HOOK-161.
func parseGoVetOutput(output string, basePath string) []Diagnostic {
	result := make([]Diagnostic, 0)

	// Pattern: ./file.go:line:column: message

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := reGoVetPattern.FindStringSubmatch(line)
		if len(matches) != 5 {
			continue
		}

		lineNum, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}
		charNum, err := strconv.Atoi(matches[3])
		if err != nil {
			continue
		}
		message := matches[4]

		result = append(result, Diagnostic{
			Range: Range{
				Start: Position{Line: lineNum - 1, Character: charNum - 1},
				End:   Position{Line: lineNum - 1, Character: charNum - 1},
			},
			Severity: SeverityWarning,
			Source:   "go vet",
			Message:  message,
		})
	}

	return result
}

// parseTypeScriptOutput parses tsc output per REQ-HOOK-161.
func parseTypeScriptOutput(output string) []Diagnostic {
	var result []Diagnostic

	// Pattern: file.ts(line,column): error TS1234: message

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := reTypeScriptPattern.FindStringSubmatch(line)
		if len(matches) != 7 {
			continue
		}

		lineNum, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}
		charNum, err := strconv.Atoi(matches[3])
		if err != nil {
			continue
		}
		severityStr := matches[4]
		code := matches[5]
		message := matches[6]

		severity := SeverityError
		if severityStr == "warning" {
			severity = SeverityWarning
		}

		result = append(result, Diagnostic{
			Range: Range{
				Start: Position{Line: lineNum - 1, Character: charNum - 1},
				End:   Position{Line: lineNum - 1, Character: charNum - 1},
			},
			Severity: severity,
			Code:     code,
			Source:   "tsc",
			Message:  message,
		})
	}

	return result
}

// parseESLintOutput parses eslint JSON output per REQ-HOOK-161.
func parseESLintOutput(data []byte) ([]Diagnostic, error) {
	if len(data) == 0 {
		return []Diagnostic{}, nil
	}

	var eslintResults []struct {
		FilePath string `json:"filePath"`
		Messages []struct {
			RuleID   string `json:"ruleId"`
			Severity int    `json:"severity"`
			Message  string `json:"message"`
			Line     int    `json:"line"`
			Column   int    `json:"column"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(data, &eslintResults); err != nil {
		return []Diagnostic{}, nil
	}

	var result []Diagnostic
	for _, file := range eslintResults {
		for _, msg := range file.Messages {
			severity := SeverityWarning
			if msg.Severity == 2 {
				severity = SeverityError
			}

			result = append(result, Diagnostic{
				Range: Range{
					Start: Position{Line: msg.Line - 1, Character: msg.Column - 1},
					End:   Position{Line: msg.Line - 1, Character: msg.Column - 1},
				},
				Severity: severity,
				Code:     msg.RuleID,
				Source:   "eslint",
				Message:  msg.Message,
			})
		}
	}

	return result, nil
}

// parseClippyOutput parses cargo clippy JSON output per REQ-HOOK-161.
func parseClippyOutput(data []byte) ([]Diagnostic, error) {
	if len(data) == 0 {
		return []Diagnostic{}, nil
	}

	var result []Diagnostic

	// Clippy outputs one JSON object per line
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var msg struct {
			Reason  string `json:"reason"`
			Message *struct {
				Code *struct {
					Code string `json:"code"`
				} `json:"code"`
				Level   string `json:"level"`
				Message string `json:"message"`
				Spans   []struct {
					LineStart   int `json:"line_start"`
					LineEnd     int `json:"line_end"`
					ColumnStart int `json:"column_start"`
					ColumnEnd   int `json:"column_end"`
				} `json:"spans"`
			} `json:"message"`
		}

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		if msg.Reason != "compiler-message" || msg.Message == nil {
			continue
		}

		severity := SeverityWarning
		if msg.Message.Level == "error" {
			severity = SeverityError
		}

		code := ""
		if msg.Message.Code != nil {
			code = msg.Message.Code.Code
		}

		for _, span := range msg.Message.Spans {
			result = append(result, Diagnostic{
				Range: Range{
					Start: Position{Line: span.LineStart - 1, Character: span.ColumnStart - 1},
					End:   Position{Line: span.LineEnd - 1, Character: span.ColumnEnd - 1},
				},
				Severity: severity,
				Code:     code,
				Source:   "clippy",
				Message:  msg.Message.Message,
			})
			break // Only use first span
		}
	}

	return result, nil
}

// getExtension extracts file extension from path.
func getExtension(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx == -1 {
		return ""
	}
	return path[idx:]
}

// Compile-time interface compliance check.
var _ FallbackDiagnostics = (*fallbackDiagnostics)(nil)
