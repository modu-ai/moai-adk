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
	// TestPaths는 테스트 파일로 간주할 사용자 정의 glob 패턴 목록입니다 (REQ-SPC-004-040).
	// nil 또는 빈 슬라이스이면 하드코딩된 패턴(_test.go, tests/, fixtures/, testdata/)만 사용합니다.
	// 패턴은 filepath.Match 의미론 + 경로 구성요소 매칭을 사용합니다 (**는 경로 세그먼트 매칭으로 처리).
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

	// 1. 사용자 glob 패턴 우선 확인
	for _, pattern := range userPatterns {
		if matchesGlobPattern(slashPath, pattern) {
			return true
		}
	}

	// 2. 하드코딩 폴백
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

// matchesGlobPattern은 slashPath가 glob 패턴과 일치하는지 확인합니다.
// **는 경로 구분자를 포함한 임의의 경로 세그먼트를 의미합니다 (더블스타 의미론 근사).
func matchesGlobPattern(slashPath, pattern string) bool {
	// 표준 filepath.Match를 먼저 시도
	if matched, _ := filepath.Match(pattern, slashPath); matched {
		return true
	}

	// ** 포함 패턴: **를 기준으로 분할하여 각 세그먼트 존재 여부 확인
	// 예: "**/integration/**" → slashPath에 "/integration/" 포함 여부
	if strings.Contains(pattern, "**") {
		segments := strings.Split(pattern, "**")
		for _, seg := range segments {
			seg = strings.Trim(seg, "/")
			if seg == "" {
				continue
			}
			// 패턴 세그먼트가 경로 구성요소로 포함되어야 함
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
// @MX:NOTE: [AUTO] NewTextualFanInCounterWithTestPaths — mx.yaml test_paths를 주입하는 공개 생성자.
// CLI wire-up(M6)에서 dangerCfg.TestPaths를 전달할 때 사용.
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
		// TestPaths 필드가 있으면 사용자 glob 패턴과 함께 검사 (REQ-SPC-004-040)
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
