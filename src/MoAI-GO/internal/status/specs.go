package status

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SpecInfo represents information about a SPEC document
type SpecInfo struct {
	ID     string
	Title  string
	Status string
}

// GetActiveSpecs retrieves all active SPEC documents from .moai/specs/
func GetActiveSpecs(projectDir string) ([]SpecInfo, error) {
	specsDir := filepath.Join(projectDir, ".moai", "specs")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return []SpecInfo{}, nil
	}

	// Read all subdirectories
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading specs directory: %w", err)
	}

	specs := make([]SpecInfo, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specID := entry.Name()
		specPath := filepath.Join(specsDir, specID, "spec.md")

		// Try to read spec.md
		content, err := os.ReadFile(specPath)
		if err != nil {
			// Skip if spec.md doesn't exist
			continue
		}

		// Extract title and status
		title, status := parseSpecInfo(string(content))

		specs = append(specs, SpecInfo{
			ID:     specID,
			Title:  title,
			Status: status,
		})
	}

	return specs, nil
}

// parseSpecInfo extracts title and status from SPEC markdown content
func parseSpecInfo(content string) (title, status string) {
	// Extract title from first heading
	reTitle := regexp.MustCompile(`^#\s+SPEC[^\n]*\n+([^\n]+)`)
	titleMatches := reTitle.FindStringSubmatch(content)
	if len(titleMatches) > 1 {
		title = strings.TrimSpace(titleMatches[1])
	} else {
		title = "Unknown"
	}

	// Extract status from metadata
	reStatus := regexp.MustCompile(`(?i)\|\s*\*\*Status\s*\*\*\s*\|\s*([^\||]+)`)
	statusMatches := reStatus.FindStringSubmatch(content)
	if len(statusMatches) > 1 {
		status = strings.TrimSpace(statusMatches[1])
	} else {
		status = "Unknown"
	}

	return title, status
}

// FormatStatusIcon returns an icon for a given status
func FormatStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "completed", "done", "approved":
		return "✓"
	case "in progress", "active", "pending":
		return "◐"
	case "draft", "planned":
		return "○"
	default:
		return "?"
	}
}
