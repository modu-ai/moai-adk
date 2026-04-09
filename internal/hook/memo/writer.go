package memo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// memoFileName is the relative path within the project directory.
const memoFileName = ".moai/state/session-memo.md"

// Write serializes the given sections to the session memo file at
// {projectDir}/.moai/state/session-memo.md.
// Sections are ordered by priority (P1 first). Parent directories are
// created as needed. Any existing file is overwritten.
func Write(projectDir string, sections []Section) error {
	if projectDir == "" {
		return fmt.Errorf("memo: projectDir must not be empty")
	}

	// Sort a copy so the caller's slice is not mutated.
	sorted := make([]Section, len(sections))
	copy(sorted, sections)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority < sorted[j].Priority
		}
		return sorted[i].Title < sorted[j].Title
	})

	var sb strings.Builder
	sb.WriteString("# Session Memo\n\n")

	for _, s := range sorted {
		if s.Content == "" {
			continue
		}
		label := priorityLabel(s.Priority)
		fmt.Fprintf(&sb, "## %s: %s\n\n%s\n\n", label, s.Title, strings.TrimSpace(s.Content))
	}

	dest := filepath.Join(projectDir, memoFileName)

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("memo: create state dir: %w", err)
	}

	if err := os.WriteFile(dest, []byte(sb.String()), 0o644); err != nil {
		return fmt.Errorf("memo: write file: %w", err)
	}

	return nil
}
