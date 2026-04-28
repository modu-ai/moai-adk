package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// ValidStatuses defines all allowed status values
var ValidStatuses = []string{
	"draft",
	"planned",
	"in-progress",
	"implemented",
	"completed",
	"superseded",
}

// IsValidStatus checks if a status value is valid
func IsValidStatus(status string) bool {
	return slices.Contains(ValidStatuses, status)
}

// UpdateStatus updates the status field in a SPEC document
// specDir is the directory containing spec.md (e.g., ".moai/specs/SPEC-XXX")
func UpdateStatus(specDir, newStatus string) error {
	// Validate status
	if !IsValidStatus(newStatus) {
		return fmt.Errorf("invalid status: %q", newStatus)
	}

	// Locate spec.md
	specPath := filepath.Join(specDir, "spec.md")
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return fmt.Errorf("spec.md not found in %s", specDir)
	}

	// Read the file
	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec.md: %w", err)
	}

	// Detect format and update status
	updated, err := updateStatusInContent(string(content), newStatus)
	if err != nil {
		return err
	}

	// Write back
	if err := os.WriteFile(specPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write spec.md: %w", err)
	}

	return nil
}

// ParseStatus extracts the current status from a SPEC document
func ParseStatus(specDir string) (string, error) {
	specPath := filepath.Join(specDir, "spec.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to read spec.md: %w", err)
	}

	return parseStatusFromContent(string(content))
}

// updateStatusInContent detects format and updates status field
func updateStatusInContent(content, newStatus string) (string, error) {
	lines := strings.Split(content, "\n")

	// Read first 30 lines for format detection
	sampleLines := lines
	if len(sampleLines) > 30 {
		sampleLines = sampleLines[:30]
	}
	sample := strings.Join(sampleLines, "\n")

	// Format detection (order matters: check more specific patterns first)
	if strings.Contains(sample, "| 상태 |") || strings.Contains(sample, "| Status |") {
		// Table format (Format E) - check before YAML since tables may have ---
		return updateStatusInTable(lines, newStatus)
	}

	if strings.Contains(sample, "- **Status**:") || strings.Contains(sample, "- **상태**:") {
		// Markdown list format (Format D)
		return updateStatusInMarkdownList(lines, newStatus)
	}

	if strings.Contains(sample, "---") {
		// YAML frontmatter (Format A/B)
		return updateStatusInYAML(lines, newStatus)
	}

	// No frontmatter - add YAML frontmatter (Format F)
	return addYAMLFrontmatter(content, newStatus)
}

// updateStatusInYAML updates status in YAML frontmatter
func updateStatusInYAML(lines []string, newStatus string) (string, error) {
	inFrontmatter := false
	statusLineIdx := -1
	lastFrontmatterFieldIdx := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
			} else {
				// End of frontmatter
				break
			}
			continue
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "status:") {
				statusLineIdx = i
			}
			// Track the last field line (before the closing ---)
			if strings.Contains(line, ":") && !strings.HasPrefix(line, "#") {
				lastFrontmatterFieldIdx = i
			}
		}
	}

	// Update existing status line
	if statusLineIdx >= 0 {
		lines[statusLineIdx] = fmt.Sprintf("status: %s", newStatus)
		return strings.Join(lines, "\n"), nil
	}

	// Fallback: return original content on error
	originalContent := strings.Join(lines, "\n")

	// Add status field after last frontmatter field
	if lastFrontmatterFieldIdx >= 0 {
		// Insert after last field
		newLines := make([]string, len(lines)+1)
		copy(newLines, lines[:lastFrontmatterFieldIdx+1])
		newLines[lastFrontmatterFieldIdx+1] = fmt.Sprintf("status: %s", newStatus)
		copy(newLines[lastFrontmatterFieldIdx+2:], lines[lastFrontmatterFieldIdx+1:])
		return strings.Join(newLines, "\n"), nil
	}

	// No fields found - add status at start of frontmatter
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			newLines := make([]string, len(lines)+1)
			copy(newLines, lines[:i+1])
			newLines[i+1] = fmt.Sprintf("status: %s", newStatus)
			copy(newLines[i+2:], lines[i+1:])
			return strings.Join(newLines, "\n"), nil
		}
	}

	return originalContent, fmt.Errorf("could not find YAML frontmatter")
}

// updateStatusInTable updates status in table format
func updateStatusInTable(lines []string, newStatus string) (string, error) {
	originalContent := strings.Join(lines, "\n")

	for i, line := range lines {
		// Match Korean table: | 상태 | value |
		if strings.Contains(line, "| 상태 |") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				// Replace the value after "상태"
				for j, part := range parts {
					if strings.TrimSpace(part) == "상태" && j+1 < len(parts) {
						parts[j+1] = " " + newStatus + " "
						lines[i] = strings.Join(parts, "|")
						return strings.Join(lines, "\n"), nil
					}
				}
			}
		}

		// Match English table: | Status | value |
		if strings.Contains(line, "| Status |") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				for j, part := range parts {
					if strings.TrimSpace(part) == "Status" && j+1 < len(parts) {
						parts[j+1] = " " + newStatus + " "
						lines[i] = strings.Join(parts, "|")
						return strings.Join(lines, "\n"), nil
					}
				}
			}
		}
	}

	return originalContent, fmt.Errorf("could not find status in table")
}

// updateStatusInMarkdownList updates status in Markdown list format
func updateStatusInMarkdownList(lines []string, newStatus string) (string, error) {
	originalContent := strings.Join(lines, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "- **Status**:") || strings.HasPrefix(line, "- **상태**:") {
			lines[i] = strings.TrimRight(line, "\r")
			// Replace the value after the colon
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				lines[i] = fmt.Sprintf("%s: %s", parts[0], newStatus)
			}
			return strings.Join(lines, "\n"), nil
		}
	}

	return originalContent, fmt.Errorf("could not find status in Markdown list")
}

// addYAMLFrontmatter adds YAML frontmatter with status field
func addYAMLFrontmatter(content, newStatus string) (string, error) {
	frontmatter := fmt.Sprintf("---\nstatus: %s\n---\n", newStatus)
	return frontmatter + content, nil
}

// parseStatusFromContent extracts status from content
func parseStatusFromContent(content string) (string, error) {
	lines := strings.Split(content, "\n")

	// Read first 30 lines
	sampleLines := lines
	if len(sampleLines) > 30 {
		sampleLines = sampleLines[:30]
	}
	sample := strings.Join(sampleLines, "\n")

	// Try table format first (most specific pattern)
	if strings.Contains(sample, "| 상태 |") || strings.Contains(sample, "| Status |") {
		if status, found := parseStatusFromTable(lines); found {
			return status, nil
		}
	}

	// Try Markdown list format
	if strings.Contains(sample, "- **Status**:") || strings.Contains(sample, "- **상태**:") {
		if status, found := parseStatusFromMarkdownList(lines); found {
			return status, nil
		}
	}

	// Try YAML frontmatter
	if strings.Contains(sample, "---") {
		if status, found := parseStatusFromYAML(lines); found {
			return status, nil
		}
	}

	return "", fmt.Errorf("could not parse status from spec.md")
}

// parseStatusFromYAML extracts status from YAML frontmatter
func parseStatusFromYAML(lines []string) (string, bool) {
	inFrontmatter := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if !inFrontmatter {
				inFrontmatter = true
			} else {
				break
			}
			continue
		}

		if inFrontmatter && strings.HasPrefix(line, "status:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), true
			}
		}
	}

	return "", false
}

// parseStatusFromTable extracts status from table format
func parseStatusFromTable(lines []string) (string, bool) {
	reStatus := regexp.MustCompile(`\|\s*Status\s*\|\s*([^\||]+)\s*\|`)
	reSangtae := regexp.MustCompile(`\|\s*상태\s*\|\s*([^\||]+)\s*\|`)

	for _, line := range lines {
		// Try English first
		if matches := reStatus.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1]), true
		}

		// Try Korean
		if matches := reSangtae.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1]), true
		}
	}

	return "", false
}

// parseStatusFromMarkdownList extracts status from Markdown list format
func parseStatusFromMarkdownList(lines []string) (string, bool) {
	reStatus := regexp.MustCompile(`-\s*\*\*Status\*\*:\s*(.+)`)
	reSangtae := regexp.MustCompile(`-\s*\*\*상태\*\*:\s*(.+)`)

	for _, line := range lines {
		if matches := reStatus.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1]), true
		}

		if matches := reSangtae.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1]), true
		}
	}

	return "", false
}
