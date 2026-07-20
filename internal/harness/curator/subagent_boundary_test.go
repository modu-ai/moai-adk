package curator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriter_NoAskUserQuestion enforces the C-HRA-008 subagent boundary
// (REQ-HEV2-035 / AC-HEV2-044): the curator package MUST NOT call
// AskUserQuestion or reference mcp__askuser. L5 approval is routed through
// the orchestrator; the Curator returns proposal artifacts, never prompts.
//
// This static guard scans every .go file in this package directory
// (excluding _test.go and comment lines) for forbidden references.
func TestWriter_NoAskUserQuestion(t *testing.T) {
	dir := "."
	forbidden := []string{"AskUserQuestion", "mcp__askuser"}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		// Only scan Go source files; skip test files.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for lineNo, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			// Skip comment-only lines.
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			for _, token := range forbidden {
				if strings.Contains(line, token) {
					t.Errorf("%s:%d: forbidden %q reference in curator package (C-HRA-008): %s",
						path, lineNo+1, token, strings.TrimSpace(line))
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
