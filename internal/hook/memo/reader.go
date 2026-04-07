package memo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// charsPerToken is the approximate number of UTF-8 characters per token
// used for budget estimation.
const charsPerToken = 4

// Read reads the session memo from {projectDir}/.moai/state/session-memo.md
// and returns its content trimmed to fit within maxTokens.
//
// If the file does not exist, Read returns ("", nil) — not an error.
// Content is trimmed from the lowest-priority sections (P4 first) until
// the estimated token count is within budget.
func Read(projectDir string, maxTokens int) (string, error) {
	if projectDir == "" {
		return "", fmt.Errorf("memo: projectDir must not be empty")
	}

	path := filepath.Join(projectDir, memoFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("memo: read file: %w", err)
	}

	content := string(data)
	if maxTokens <= 0 {
		return content, nil
	}

	return trimToTokenBudget(content, maxTokens), nil
}

// trimToTokenBudget trims sections from the end of the memo (lowest priority
// first) until the estimated token count is within budget.
// Sections are identified by "## Px:" headings.
func trimToTokenBudget(content string, maxTokens int) string {
	if estimateTokens(content) <= maxTokens {
		return content
	}

	// Split into header + sections. The first element is anything before
	// the first "## " heading (the "# Session Memo\n\n" part).
	parts := splitSections(content)
	if len(parts) <= 1 {
		// No section headings found; truncate by characters.
		limit := maxTokens * charsPerToken
		if len(content) > limit {
			return content[:limit]
		}
		return content
	}

	// Remove sections from the end until we fit.
	for len(parts) > 1 && estimateTokens(strings.Join(parts, "")) > maxTokens {
		parts = parts[:len(parts)-1]
	}

	return strings.Join(parts, "")
}

// splitSections splits the memo content on "## " headings, keeping the
// heading attached to its body in each element (except the first element
// which is the preamble before any heading).
func splitSections(content string) []string {
	var parts []string
	lines := strings.Split(content, "\n")

	var current strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && current.Len() > 0 {
			parts = append(parts, current.String())
			current.Reset()
		}
		current.WriteString(line)
		current.WriteByte('\n')
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// estimateTokens returns an approximate token count for the given string.
func estimateTokens(s string) int {
	return (len(s) + charsPerToken - 1) / charsPerToken
}
