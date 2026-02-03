package merge

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// FormatConflictMarkers produces file content with Git-style conflict markers
// inserted at the conflict positions.
func FormatConflictMarkers(content []byte, conflicts []Conflict) []byte {
	if len(conflicts) == 0 {
		return content
	}

	lines := splitLines(string(content))

	// Build a set of conflict lines for quick lookup.
	conflictByLine := make(map[int]Conflict)
	for _, c := range conflicts {
		conflictByLine[c.StartLine] = c
	}

	var result []string
	for i, line := range lines {
		lineNum := i + 1 // 1-based
		if c, ok := conflictByLine[lineNum]; ok {
			result = append(result,
				"<<<<<<< current",
				c.Current,
				"=======",
				c.Updated,
				">>>>>>> updated",
			)
		} else {
			result = append(result, line)
		}
	}

	// Handle conflicts that reference lines beyond current content length.
	for _, c := range conflicts {
		if c.StartLine > len(lines) {
			result = append(result,
				"<<<<<<< current",
				c.Current,
				"=======",
				c.Updated,
				">>>>>>> updated",
			)
		}
	}

	return []byte(strings.Join(result, "\n"))
}

// WriteConflictFile creates a .conflict file with Git-style conflict markers.
// The original file at originalPath is NOT modified.
// Returns the path of the created conflict file.
func WriteConflictFile(originalPath string, mergedContent []byte, conflicts []Conflict) (string, error) {
	if len(conflicts) == 0 {
		return "", errors.New("conflict: no conflicts to write")
	}

	conflictContent := FormatConflictMarkers(mergedContent, conflicts)
	conflictPath := originalPath + ".conflict"

	if err := os.WriteFile(conflictPath, conflictContent, 0o644); err != nil {
		return "", fmt.Errorf("conflict: write file: %w", err)
	}

	return conflictPath, nil
}
