package astgrep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/foundation"
)

// Timeouts for sg CLI operations.
const (
	SGTimeout      = 60 * time.Second
	SearchTimeout  = 120 * time.Second
	VersionTimeout = 5 * time.Second
)

// Analyzer defines the interface for AST-based code analysis.
type Analyzer interface {
	// Scan performs AST-based code scanning using patterns or rules.
	Scan(ctx context.Context, patterns []string, paths []string) (*ScanResult, error)

	// FindPattern searches for a single AST pattern in the specified language.
	FindPattern(ctx context.Context, pattern string, lang string) ([]Match, error)

	// Replace replaces matching AST patterns with the specified replacement.
	Replace(ctx context.Context, pattern, replacement, lang string, paths []string) ([]FileChange, error)
}

// CommandExecutor abstracts command execution for testability.
type CommandExecutor interface {
	Execute(ctx context.Context, workDir string, name string, args ...string) ([]byte, error)
}

// execCommandExecutor is the production implementation using os/exec.
type execCommandExecutor struct{}

func (e *execCommandExecutor) Execute(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// sg may return non-zero exit code when no matches found.
			// Return whatever output we have and let the caller parse it.
			return output, nil
		}
		return nil, fmt.Errorf("execute %s: %w", name, err)
	}
	return output, nil
}

// Option configures the SGAnalyzer.
type Option func(*SGAnalyzer)

// WithCommandExecutor sets a custom command executor for testing.
func WithCommandExecutor(executor CommandExecutor) Option {
	return func(a *SGAnalyzer) {
		a.executor = executor
	}
}

// SGAnalyzer implements Analyzer using the ast-grep (sg) CLI.
type SGAnalyzer struct {
	executor    CommandExecutor
	workDir     string
	sgAvailable bool
	sgChecked   bool
	mu          sync.Mutex
}

// Compile-time interface check.
var _ Analyzer = (*SGAnalyzer)(nil)

// NewAnalyzer creates a new SGAnalyzer with the given work directory.
func NewAnalyzer(workDir string, opts ...Option) *SGAnalyzer {
	a := &SGAnalyzer{
		executor: &execCommandExecutor{},
		workDir:  workDir,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// sgMatch is the internal representation of an ast-grep CLI JSON output match.
type sgMatch struct {
	Text     string  `json:"text"`
	Range    sgRange `json:"range"`
	File     string  `json:"file"`
	Lines    string  `json:"lines"`
	RuleID   string  `json:"ruleId,omitempty"`
	Severity string  `json:"severity,omitempty"`
	Message  string  `json:"message,omitempty"`
}

// sgRange represents the range of a match in the sg JSON output.
type sgRange struct {
	Start sgPosition `json:"start"`
	End   sgPosition `json:"end"`
}

// sgPosition represents a line/column position in the sg JSON output.
type sgPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// toMatch converts an internal sgMatch to the public Match type.
func (m *sgMatch) toMatch() Match {
	return Match{
		File:     m.File,
		Line:     m.Range.Start.Line + 1, // sg uses 0-indexed lines
		Column:   m.Range.Start.Column,
		EndLine:  m.Range.End.Line + 1,
		EndCol:   m.Range.End.Column,
		Text:     m.Text,
		Rule:     m.RuleID,
		Severity: m.Severity,
		Message:  m.Message,
	}
}

// IsSGAvailable checks whether the sg CLI is installed and accessible.
// The result is cached for the lifetime of the analyzer instance.
func (a *SGAnalyzer) IsSGAvailable(ctx context.Context) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sgChecked {
		return a.sgAvailable
	}

	ctx, cancel := context.WithTimeout(ctx, VersionTimeout)
	defer cancel()

	_, err := a.executor.Execute(ctx, a.workDir, "sg", "--version")
	a.sgAvailable = err == nil
	a.sgChecked = true
	return a.sgAvailable
}

// DetectLanguage returns the programming language for a file based on its extension.
// Uses foundation.DefaultRegistry for language lookup and returns the ast-grep
// compatible language name. Returns "text" for unrecognized extensions.
func DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		return "text"
	}

	info, err := foundation.DefaultRegistry.ByExtension(ext)
	if err != nil {
		return "text"
	}

	// Return the ast-grep specific language name if defined,
	// otherwise return the language ID.
	return info.AstGrepLanguageName(ext)
}

// ShouldIncludeFile checks if a file should be included in scanning
// based on include/exclude patterns.
func ShouldIncludeFile(filePath string, config *ScanConfig) bool {
	if config == nil {
		return true
	}

	base := filepath.Base(filePath)

	// Check exclude patterns: if any segment of the path matches, exclude it.
	for _, pattern := range config.ExcludePatterns {
		if containsSegment(filePath, pattern) {
			return false
		}
		matched, matchErr := filepath.Match(pattern, base)
		if matchErr != nil {
			continue // skip invalid patterns
		}
		if matched {
			return false
		}
	}

	// If no include patterns specified, include everything.
	if len(config.IncludePatterns) == 0 {
		return true
	}

	// Check include patterns: file must match at least one.
	for _, pattern := range config.IncludePatterns {
		matched, matchErr := filepath.Match(pattern, base)
		if matchErr != nil {
			continue // skip invalid patterns
		}
		if matched {
			return true
		}
	}

	return false
}

// containsSegment checks if the path contains a directory segment matching the pattern.
func containsSegment(filePath, segment string) bool {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for _, part := range parts {
		if part == segment {
			return true
		}
	}
	return false
}

// Scan performs an AST-based scan using the given patterns on the specified paths.
// If sg CLI is not available, returns an empty ScanResult without error.
func (a *SGAnalyzer) Scan(ctx context.Context, patterns []string, paths []string) (*ScanResult, error) {
	start := time.Now()

	if !a.IsSGAvailable(ctx) {
		return &ScanResult{Duration: time.Since(start)}, nil
	}

	if len(paths) == 0 {
		paths = []string{a.workDir}
	}

	var allMatches []Match
	filesScanned := 0

	for _, pattern := range patterns {
		for _, path := range paths {
			ctx, cancel := context.WithTimeout(ctx, SGTimeout)
			args := []string{"run", "--pattern", pattern, "--json", path}
			output, err := a.executor.Execute(ctx, a.workDir, "sg", args...)
			cancel()
			if err != nil {
				return nil, fmt.Errorf("scan pattern %q on %s: %w", pattern, path, err)
			}

			matches, err := parseSGOutput(output)
			if err != nil {
				continue // non-parseable output treated as no matches
			}
			allMatches = append(allMatches, matches...)
			filesScanned++
		}
	}

	return &ScanResult{
		Matches:  allMatches,
		Duration: time.Since(start),
		Files:    filesScanned,
	}, nil
}

// FindPattern searches for a single AST pattern in the specified language.
// If sg CLI is not available, returns an empty slice without error.
func (a *SGAnalyzer) FindPattern(ctx context.Context, pattern string, lang string) ([]Match, error) {
	if !a.IsSGAvailable(ctx) {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, SearchTimeout)
	defer cancel()

	args := []string{"run", "--pattern", pattern, "--lang", lang, "--json", a.workDir}
	output, err := a.executor.Execute(ctx, a.workDir, "sg", args...)
	if err != nil {
		return nil, fmt.Errorf("find pattern %q: %w", pattern, err)
	}

	return parseSGOutput(output)
}

// Replace replaces matching AST patterns with the specified replacement.
// If sg CLI is not available, returns an empty slice without error.
func (a *SGAnalyzer) Replace(ctx context.Context, pattern, replacement, lang string, paths []string) ([]FileChange, error) {
	if !a.IsSGAvailable(ctx) {
		return nil, nil
	}

	if len(paths) == 0 {
		paths = []string{a.workDir}
	}

	var allChanges []FileChange
	for _, path := range paths {
		ctx, cancel := context.WithTimeout(ctx, SearchTimeout)
		args := []string{"run", "--pattern", pattern, "--rewrite", replacement, "--lang", lang, "--json", path}
		output, err := a.executor.Execute(ctx, a.workDir, "sg", args...)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("replace pattern %q: %w", pattern, err)
		}

		matches, err := parseSGOutput(output)
		if err != nil {
			continue
		}
		for _, m := range matches {
			allChanges = append(allChanges, FileChange{
				FilePath: m.File,
				OldCode:  m.Text,
				NewCode:  replacement,
				Line:     m.Line,
				Column:   m.Column,
			})
		}
	}

	return allChanges, nil
}

// ScanFile scans a single file using ast-grep rules.
// Returns an error if the file does not exist.
// If sg CLI is not available, returns an empty ScanResult without error.
func (a *SGAnalyzer) ScanFile(ctx context.Context, filePath string, config *ScanConfig) (*ScanResult, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filePath)
		}
		return nil, fmt.Errorf("stat file %s: %w", filePath, err)
	}

	start := time.Now()
	lang := DetectLanguage(filePath)

	if !a.IsSGAvailable(ctx) {
		return &ScanResult{
			Matches:  []Match{},
			Duration: time.Since(start),
			Files:    1,
			Language: lang,
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, SGTimeout)
	defer cancel()

	args := buildScanArgs(filePath, config)
	output, err := a.executor.Execute(ctx, a.workDir, "sg", args...)
	if err != nil {
		return nil, fmt.Errorf("scan file %s: %w", filePath, err)
	}

	matches, err := parseSGOutput(output)
	if err != nil {
		matches = nil
	}

	return &ScanResult{
		Matches:  matches,
		Duration: time.Since(start),
		Files:    1,
		Language: lang,
	}, nil
}

// ScanProject recursively scans all supported files in a project directory.
// Returns an error if the directory does not exist.
func (a *SGAnalyzer) ScanProject(ctx context.Context, projectPath string, config *ScanConfig) (*ProjectScanResult, error) {
	info, err := os.Stat(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project path not found: %s", projectPath)
		}
		return nil, fmt.Errorf("stat project path %s: %w", projectPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path is not a directory: %s", projectPath)
	}

	if config == nil {
		dc := DefaultScanConfig()
		config = &dc
	}

	start := time.Now()
	result := &ProjectScanResult{
		BySeverity: make(map[string]int),
	}

	err = filepath.Walk(projectPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // skip inaccessible files
		}
		if info.IsDir() {
			// Check if this directory should be excluded.
			base := filepath.Base(path)
			for _, pattern := range config.ExcludePatterns {
				if base == pattern {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check file extension is supported.
		lang := DetectLanguage(path)
		if lang == "text" {
			return nil
		}

		if !ShouldIncludeFile(path, config) {
			return nil
		}

		scanResult, scanErr := a.ScanFile(ctx, path, config)
		if scanErr != nil {
			return nil // skip files that fail to scan
		}

		result.TotalFiles++
		result.TotalMatches += len(scanResult.Matches)
		for _, m := range scanResult.Matches {
			sev := m.Severity
			if sev == "" {
				sev = "info"
			}
			result.BySeverity[sev]++
		}
		result.FileResults = append(result.FileResults, *scanResult)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk project %s: %w", projectPath, err)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// PatternSearch searches for a custom AST pattern in the specified language and path.
// The rule_id for each match is set to "pattern:<first 30 chars>".
func (a *SGAnalyzer) PatternSearch(ctx context.Context, pattern, lang, path string) ([]Match, error) {
	if !a.IsSGAvailable(ctx) {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, SearchTimeout)
	defer cancel()

	args := []string{"run", "--pattern", pattern, "--lang", lang, "--json", path}
	output, err := a.executor.Execute(ctx, a.workDir, "sg", args...)
	if err != nil {
		return nil, fmt.Errorf("pattern search %q: %w", pattern, err)
	}

	matches, err := parseSGOutput(output)
	if err != nil {
		return nil, nil
	}

	// Set rule_id for pattern-based matches.
	ruleID := "pattern:" + truncate(pattern, 30)
	for i := range matches {
		matches[i].Rule = ruleID
	}

	return matches, nil
}

// PatternReplace replaces matching code with the specified replacement pattern.
// If dryRun is true, no actual file modifications are made.
func (a *SGAnalyzer) PatternReplace(ctx context.Context, pattern, replacement, lang, path string, dryRun bool) (*ReplaceResult, error) {
	result := &ReplaceResult{DryRun: dryRun}

	if !a.IsSGAvailable(ctx) {
		return result, nil
	}

	// First, find matches.
	ctx, cancel := context.WithTimeout(ctx, SearchTimeout)
	defer cancel()

	findArgs := []string{"run", "--pattern", pattern, "--lang", lang, "--json", path}
	output, err := a.executor.Execute(ctx, a.workDir, "sg", findArgs...)
	if err != nil {
		return nil, fmt.Errorf("pattern replace search %q: %w", pattern, err)
	}

	matches, err := parseSGOutput(output)
	if err != nil {
		return result, nil
	}

	result.MatchesFound = len(matches)

	// Build change list.
	fileSet := make(map[string]struct{})
	for _, m := range matches {
		result.Changes = append(result.Changes, FileChange{
			FilePath: m.File,
			OldCode:  m.Text,
			NewCode:  replacement,
			Line:     m.Line,
			Column:   m.Column,
		})
		fileSet[m.File] = struct{}{}
	}
	result.FilesModified = len(fileSet)

	// If not dry_run, execute the actual replacement.
	if !dryRun && result.MatchesFound > 0 {
		replaceArgs := []string{"run", "--pattern", pattern, "--rewrite", replacement, "--lang", lang, path}
		_, err := a.executor.Execute(ctx, a.workDir, "sg", replaceArgs...)
		if err != nil {
			return nil, fmt.Errorf("pattern replace execute %q: %w", pattern, err)
		}
	}

	return result, nil
}

// parseSGOutput parses the JSON output from the sg CLI into Match objects.
func parseSGOutput(output []byte) ([]Match, error) {
	if len(output) == 0 {
		return nil, nil
	}

	var sgMatches []sgMatch
	if err := json.Unmarshal(output, &sgMatches); err != nil {
		return nil, fmt.Errorf("parse sg output: %w", err)
	}

	matches := make([]Match, 0, len(sgMatches))
	for i := range sgMatches {
		matches = append(matches, sgMatches[i].toMatch())
	}
	return matches, nil
}

// buildScanArgs constructs the sg CLI arguments for a file scan.
func buildScanArgs(filePath string, config *ScanConfig) []string {
	args := []string{"scan", "--json"}
	if config != nil && config.RulesPath != "" {
		args = append(args, "--config", config.RulesPath)
	}
	args = append(args, filePath)
	return args
}

// truncate returns the first n characters of s, or s itself if shorter.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
