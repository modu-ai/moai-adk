package mx

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)

// FanInCounter is the interface for calculating the caller count of @MX:ANCHOR tags.
// Abstracts LSP-based implementation and text fallback implementation under the same interface (REQ-SPC-004-003).
type FanInCounter interface {
	// Count calculates the fan-in (caller count) of the given tag.
	// When excludeTests is true, excludes test file callers (REQ-SPC-004-040).
	// Returns: count (caller count), method ("lsp" or "textual"), err
	Count(ctx context.Context, tag Tag, projectRoot string, excludeTests bool) (count int, method string, err error)
}

// TextualFanInCounter is the implementation that calculates fan-in using text-based grep method.
// Used as a fallback for languages without LSP server (REQ-SPC-004-020).
//
// @MX:WARN: [AUTO] TextualFanInCounter — text search method has false positive risk on strings/comments
// @MX:REASON: grep-based fan-in counts string occurrences, not semantic callers; comments and dead code count
type TextualFanInCounter struct {
	// ProjectRoot is the project root directory path.
	ProjectRoot string
	// TestPaths is the list of user-defined glob patterns treated as test files (REQ-SPC-004-040).
	// When nil or empty, only the hard-coded patterns (_test.go, tests/, fixtures/, testdata/) are used.
	// Patterns use filepath.Match semantics + path-component matching (`**` is treated as path-segment matching).
	TestPaths []string
}

// isTestFile verifies if the file path is a test file (REQ-SPC-004-040).
// Test file detection based on: _test.go suffix or under tests/, fixtures/ directory.
// Backward-compat wrapper: calls isTestFileWithPatterns with nil user patterns.
func isTestFile(filePath string) bool {
	return isTestFileWithPatterns(filePath, nil)
}

// isTestFileWithPatterns verifies if the file path is a test file, considering user-supplied glob patterns.
// User patterns are checked first; if any match, returns true.
// After user patterns, hard-coded fallbacks are checked: _test.go suffix, /tests/, /fixtures/, /testdata/ path components.
//
// Pattern semantics: each pattern uses filepath.Match rules; '**' segments are matched against
// individual path components via path-prefix heuristic (contains the segment in the slash-normalised path).
func isTestFileWithPatterns(filePath string, userPatterns []string) bool {
	slashPath := filepath.ToSlash(filePath)

	// 1. Check user glob patterns first
	for _, pattern := range userPatterns {
		if matchesGlobPattern(slashPath, pattern) {
			return true
		}
	}

	// 2. Hard-coded fallback
	base := filepath.Base(filePath)
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	parts := strings.Split(slashPath, "/")
	for _, part := range parts {
		if part == "tests" || part == "fixtures" || part == "testdata" {
			return true
		}
	}
	return false
}

// matchesGlobPattern checks whether slashPath matches the glob pattern.
// `**` denotes an arbitrary path segment including path separators (doublestar
// semantics approximation).
func matchesGlobPattern(slashPath, pattern string) bool {
	// Try standard filepath.Match first
	if matched, _ := filepath.Match(pattern, slashPath); matched {
		return true
	}

	// Patterns with `**`: split on `**` and verify each segment exists
	// e.g.: "**/integration/**" → slashPath contains "/integration/"
	if strings.Contains(pattern, "**") {
		segments := strings.Split(pattern, "**")
		for _, seg := range segments {
			seg = strings.Trim(seg, "/")
			if seg == "" {
				continue
			}
			// The pattern segment must appear as a path component
			if !strings.Contains(slashPath, "/"+seg+"/") &&
				!strings.HasPrefix(slashPath, seg+"/") &&
				!strings.HasSuffix(slashPath, "/"+seg) &&
				slashPath != seg {
				return false
			}
		}
		return true
	}

	return false
}

// NewTextualFanInCounterWithTestPaths creates a TextualFanInCounter pre-configured with
// user-supplied test path glob patterns from mx.yaml (REQ-SPC-004-040).
// Callers that prefer a zero-value struct may still use &TextualFanInCounter{} directly.
//
// @MX:NOTE: [AUTO] NewTextualFanInCounterWithTestPaths — public constructor that injects mx.yaml test_paths.
// Used by the CLI wire-up (M6) when passing dangerCfg.TestPaths.
func NewTextualFanInCounterWithTestPaths(testPaths []string) *TextualFanInCounter {
	return &TextualFanInCounter{TestPaths: testPaths}
}

// Count calculates the caller count of AnchorID through text search.
// result's fan_in_method is always "textual".
func (c *TextualFanInCounter) Count(_ context.Context, tag Tag, projectRoot string, excludeTests bool) (int, string, error) {
	if tag.AnchorID == "" {
		return 0, "textual", nil
	}

	if projectRoot == "" {
		projectRoot = c.ProjectRoot
	}

	if projectRoot == "" {
		return 0, "textual", nil
	}

	count := 0

	// Recursively scan project root to search for AnchorID callerss
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore error and continue
		}

		if info.IsDir() {
			// exclude vendor, node_modules, etc.
			base := filepath.Base(path)
			if base == "vendor" || base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if path == tag.File {
			return nil
		}

		// exclude test files (when excludeTests=true)
		// If TestPaths is set, also evaluate the user glob patterns (REQ-SPC-004-040)
		if excludeTests && isTestFileWithPatterns(path, c.TestPaths) {
			return nil
		}

		// Search for AnchorID references in file
		refs := countReferencesInFile(path, tag.AnchorID)
		count += refs
		return nil
	})

	if err != nil {
		return 0, "textual", nil
	}

	return count, "textual", nil
}

func countReferencesInFile(filePath, symbol string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer func() { _ = f.Close() }()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, symbol) {
			count++
		}
	}
	return count
}
