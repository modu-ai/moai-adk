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
type TextualFanInCounter struct {
	// ProjectRoot is the project root directory path.
	ProjectRoot string
}

// isTestFile verifies if the file path is a test file (REQ-SPC-004-040).
// Test file detection based on: _test.go suffix or under tests/, fixtures/ directory.
func isTestFile(filePath string) bool {
	base := filepath.Base(filePath)
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	// If path includes tests/ or fixtures/ directory, treat as test file
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	for _, part := range parts {
		if part == "tests" || part == "fixtures" || part == "testdata" {
			return true
		}
	}
	return false
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
		if excludeTests && isTestFile(path) {
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
