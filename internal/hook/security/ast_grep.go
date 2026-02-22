package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// reASTGrepFinding matches ast-grep text output lines.
var reASTGrepFinding = regexp.MustCompile(`([^:]+):(\d+):(\d+):\s*(error|warning|info|hint)\[([^\]]+)\]:\s*(.+)`)

// astGrepScanner implements ASTGrepScanner interface.
type astGrepScanner struct {
	mu        sync.RWMutex
	available bool
	checked   bool
	version   string
}

// NewASTGrepScanner creates a new ASTGrepScanner instance.
func NewASTGrepScanner() ASTGrepScanner {
	return &astGrepScanner{}
}

// IsAvailable checks if ast-grep (sg) binary is available in PATH.
// Implements REQ-HOOK-100.
func (s *astGrepScanner) IsAvailable() bool {
	s.mu.RLock()
	if s.checked {
		available := s.available
		s.mu.RUnlock()
		return available
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.checked {
		return s.available
	}

	_, err := exec.LookPath("sg")
	s.available = err == nil
	s.checked = true
	return s.available
}

// GetVersion returns the ast-grep version string.
func (s *astGrepScanner) GetVersion() string {
	if !s.IsAvailable() {
		return ""
	}

	s.mu.RLock()
	if s.version != "" {
		version := s.version
		s.mu.RUnlock()
		return version
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.version != "" {
		return s.version
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sg", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	s.version = strings.TrimSpace(string(output))
	return s.version
}

// Scan runs ast-grep scan on a single file.
// Implements REQ-HOOK-101, REQ-HOOK-102.
func (s *astGrepScanner) Scan(ctx context.Context, filePath string, configPath string) (*ScanResult, error) {
	start := time.Now()
	result := &ScanResult{
		Scanned:  false,
		Findings: []Finding{},
	}

	// Check if scanner is available (REQ-HOOK-100)
	if !s.IsAvailable() {
		return result, nil
	}

	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Check if file extension is supported (REQ-HOOK-141)
	ext := filepath.Ext(filePath)
	if !IsSupportedExtension(ext) {
		return result, nil
	}

	// Build command arguments
	args := []string{"scan", "--json"}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	args = append(args, filePath)

	// Execute ast-grep scan (REQ-HOOK-121)
	cmd := exec.CommandContext(ctx, "sg", args...)
	output, err := cmd.CombinedOutput()

	result.Duration = time.Since(start)
	result.Scanned = true

	// Check for context errors (timeout or cancellation)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Parse output
	if err != nil {
		// ast-grep returns non-zero exit code when findings exist
		// Try to parse the output anyway
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() > 1 {
			// Real error, not just findings
			result.Error = err.Error()
			// Try regex fallback (REQ-HOOK-122)
			result.Findings = parseASTGrepRegex(string(output))
		}
	}

	// Try JSON parsing first (REQ-HOOK-121)
	if len(output) > 0 && result.Error == "" {
		findings, parseErr := parseASTGrepJSON(output)
		if parseErr != nil {
			// Fall back to regex parsing (REQ-HOOK-122)
			result.Findings = parseASTGrepRegex(string(output))
		} else {
			result.Findings = findings
		}
	}

	// Update severity counts (REQ-HOOK-102)
	result.ErrorCount, result.WarningCount, result.InfoCount = result.CountBySeverity()

	return result, nil
}

// ScanMultiple runs ast-grep scan on multiple files.
// Implements REQ-HOOK-123.
func (s *astGrepScanner) ScanMultiple(ctx context.Context, filePaths []string, configPath string) ([]*ScanResult, error) {
	if len(filePaths) == 0 {
		return []*ScanResult{}, nil
	}

	results := make([]*ScanResult, len(filePaths))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, fp := range filePaths {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()

			result, err := s.Scan(ctx, path, configPath)
			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
			}
			if result != nil {
				results[idx] = result
			} else {
				// Create empty result on error
				results[idx] = &ScanResult{Scanned: false}
			}
		}(i, fp)
	}

	wg.Wait()

	if firstErr != nil {
		return results, firstErr
	}

	return results, nil
}

// parseASTGrepJSON parses JSON output from ast-grep.
// Implements REQ-HOOK-121.
func parseASTGrepJSON(data []byte) ([]Finding, error) {
	// ast-grep JSON output format
	type astGrepFinding struct {
		RuleID   string `json:"ruleId"`
		Severity string `json:"severity"`
		Message  string `json:"message"`
		File     string `json:"file"`
		Range    *struct {
			Start struct {
				Line   int `json:"line"`
				Column int `json:"column"`
			} `json:"start"`
			End struct {
				Line   int `json:"line"`
				Column int `json:"column"`
			} `json:"end"`
		} `json:"range"`
		// Alternative flat format
		Line      int    `json:"line"`
		Column    int    `json:"column"`
		EndLine   int    `json:"endLine"`
		EndColumn int    `json:"endColumn"`
		Code      string `json:"code"`
	}

	var rawFindings []astGrepFinding
	if err := json.Unmarshal(data, &rawFindings); err != nil {
		return nil, fmt.Errorf("failed to parse ast-grep JSON: %w", err)
	}

	findings := make([]Finding, 0, len(rawFindings))
	for _, rf := range rawFindings {
		f := Finding{
			RuleID:   rf.RuleID,
			Severity: parseSeverity(rf.Severity),
			Message:  rf.Message,
			File:     rf.File,
			Code:     rf.Code,
		}

		// Handle both range format and flat format
		if rf.Range != nil {
			f.Line = rf.Range.Start.Line
			f.Column = rf.Range.Start.Column
			f.EndLine = rf.Range.End.Line
			f.EndColumn = rf.Range.End.Column
		} else {
			f.Line = rf.Line
			f.Column = rf.Column
			f.EndLine = rf.EndLine
			f.EndColumn = rf.EndColumn
		}

		findings = append(findings, f)
	}

	return findings, nil
}

// parseASTGrepRegex parses text output using regex fallback.
// Implements REQ-HOOK-122.
func parseASTGrepRegex(output string) []Finding {
	findings := []Finding{}

	// Pattern: file:line:column: severity[rule]: message
	// Example: test.py:10:5: error[sql-injection]: Potential SQL injection

	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		matches := reASTGrepFinding.FindStringSubmatch(line)
		if len(matches) == 7 {
			// These conversions are safe because the regex ensures digits
			lineNum, err := strconv.Atoi(matches[2])
			if err != nil {
				continue // Skip malformed line
			}
			colNum, err := strconv.Atoi(matches[3])
			if err != nil {
				continue // Skip malformed line
			}

			findings = append(findings, Finding{
				File:     matches[1],
				Line:     lineNum,
				Column:   colNum,
				Severity: parseSeverity(matches[4]),
				RuleID:   matches[5],
				Message:  matches[6],
			})
		}
	}

	return findings
}

// parseSeverity converts a severity string to Severity type.
func parseSeverity(s string) Severity {
	switch strings.ToLower(s) {
	case "error":
		return SeverityError
	case "warning", "warn":
		return SeverityWarning
	case "info":
		return SeverityInfo
	case "hint":
		return SeverityHint
	default:
		return SeverityInfo
	}
}

// supportedLanguages defines the languages supported by ast-grep.
// Per SPEC 4.2.
var supportedLanguages = []SupportedLanguage{
	{Name: "python", Extensions: []string{".py", ".pyi"}, ASTGrepID: "python"},
	{Name: "javascript", Extensions: []string{".js", ".jsx", ".mjs", ".cjs"}, ASTGrepID: "javascript"},
	{Name: "typescript", Extensions: []string{".ts", ".tsx", ".mts", ".cts"}, ASTGrepID: "typescript"},
	{Name: "go", Extensions: []string{".go"}, ASTGrepID: "go"},
	{Name: "rust", Extensions: []string{".rs"}, ASTGrepID: "rust"},
	{Name: "java", Extensions: []string{".java"}, ASTGrepID: "java"},
	{Name: "kotlin", Extensions: []string{".kt", ".kts"}, ASTGrepID: "kotlin"},
	{Name: "c", Extensions: []string{".c", ".h"}, ASTGrepID: "c"},
	{Name: "cpp", Extensions: []string{".cpp", ".cc", ".hpp", ".cxx"}, ASTGrepID: "cpp"},
	{Name: "ruby", Extensions: []string{".rb"}, ASTGrepID: "ruby"},
	{Name: "php", Extensions: []string{".php"}, ASTGrepID: "php"},
	{Name: "swift", Extensions: []string{".swift"}, ASTGrepID: "swift"},
	{Name: "csharp", Extensions: []string{".cs"}, ASTGrepID: "csharp"},
	{Name: "elixir", Extensions: []string{".ex", ".exs"}, ASTGrepID: "elixir"},
	{Name: "scala", Extensions: []string{".scala"}, ASTGrepID: "scala"},
}

// extensionToLanguage maps file extensions to language names.
var extensionToLanguage = func() map[string]string {
	m := make(map[string]string)
	for _, lang := range supportedLanguages {
		for _, ext := range lang.Extensions {
			m[ext] = lang.Name
		}
	}
	return m
}()

// IsSupportedExtension checks if a file extension is supported for scanning.
// Per REQ-HOOK-141.
func IsSupportedExtension(ext string) bool {
	_, ok := extensionToLanguage[strings.ToLower(ext)]
	return ok
}

// GetLanguageForExtension returns the language name for a file extension.
func GetLanguageForExtension(ext string) string {
	return extensionToLanguage[strings.ToLower(ext)]
}

// GetSupportedLanguages returns all supported languages.
// Per REQ-HOOK-140.
func GetSupportedLanguages() []SupportedLanguage {
	return supportedLanguages
}
