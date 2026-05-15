//go:build ignore
// +build ignore

// lint-skip-cleanup.go — 원클릭 bulk cleanup for SPEC-V3R4-LINT-SKIP-CLEANUP-001
//
// 동작:
//   1. affected-list.txt 에서 55개 대상 SPEC 경로를 읽는다.
//   2. 각 SPEC spec.md 에서 frontmatter의 lint.skip StatusGitConsistency 엔트리를 제거한다.
//   3. version 필드를 patch +1 bump 한다.
//   4. updated 필드를 2026-05-16 으로 갱신한다.
//   5. ## HISTORY 표에 새 row 1줄을 추가한다 (없으면 섹션 생성).
//   6. 이미 정리된 SPEC은 건너뛴다 (idempotent).
//
// 실행:
//   cd /path/to/moai-adk-go
//   go run .moai/scripts/lint-skip-cleanup.go
//
// 주의: gopkg.in/yaml.v3 가 go.mod에 있어야 한다.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	affectedListPath = ".moai/specs/SPEC-V3R4-LINT-SKIP-CLEANUP-001/affected-list.txt"
	updatedDate      = "2026-05-16"
	historyRowAuthor = "manager-develop (run-phase)"
	historyRowDesc   = "lint.skip StatusGitConsistency 회피책 제거 — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001 walker filter 머지로 불필요해짐."
)

func main() {
	// Read affected list
	listData, err := os.ReadFile(affectedListPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read affected-list.txt: %v\n", err)
		os.Exit(1)
	}

	paths := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(listData)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}

	fmt.Printf("Processing %d SPECs...\n\n", len(paths))

	cleaned := 0
	skipped := 0
	errors := 0

	for _, specPath := range paths {
		specID := filepath.Base(filepath.Dir(specPath))
		action, err := processSpec(specPath, specID)
		if err != nil {
			fmt.Printf("ERROR: %s: %v\n", specID, err)
			errors++
			continue
		}
		fmt.Printf("%s: %s\n", specID, action)
		if action == "already cleaned: skipped" {
			skipped++
		} else {
			cleaned++
		}
	}

	fmt.Printf("\nSummary: cleaned=%d skipped=%d errors=%d\n", cleaned, skipped, errors)
	if errors > 0 {
		os.Exit(1)
	}
}

func processSpec(specPath, specID string) (string, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("read: %w", err)
	}

	content := string(data)

	// Split frontmatter from body
	// Format: starts with "---\n", ends with "\n---\n" or "\n---" at end
	if !strings.HasPrefix(content, "---\n") {
		return "", fmt.Errorf("no frontmatter start")
	}

	// Find the closing "---" of frontmatter (second occurrence)
	rest := content[4:] // skip leading "---\n"
	fmEnd := strings.Index(rest, "\n---\n")
	var fmText, bodyText string
	if fmEnd >= 0 {
		fmText = rest[:fmEnd]
		bodyText = rest[fmEnd+5:] // skip "\n---\n"
	} else {
		// Check for trailing "---" at end of file
		fmEnd2 := strings.Index(rest, "\n---")
		if fmEnd2 >= 0 && fmEnd2 == len(rest)-4 {
			fmText = rest[:fmEnd2]
			bodyText = ""
		} else {
			return "", fmt.Errorf("no frontmatter end")
		}
	}

	// Idempotency check: if no lint: key, already cleaned
	if !strings.Contains(fmText, "\nlint:") && !strings.HasPrefix(fmText, "lint:") {
		return "already cleaned: skipped", nil
	}

	// Also check that it actually has StatusGitConsistency in frontmatter
	if !strings.Contains(fmText, "StatusGitConsistency") {
		return "already cleaned: skipped", nil
	}

	// Parse YAML with Node API to preserve key ordering
	var docNode yaml.Node
	if err := yaml.Unmarshal([]byte(fmText), &docNode); err != nil {
		return "", fmt.Errorf("yaml parse: %w", err)
	}

	if docNode.Kind != yaml.DocumentNode || len(docNode.Content) == 0 {
		return "", fmt.Errorf("yaml: unexpected document structure")
	}

	mapping := docNode.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return "", fmt.Errorf("yaml: root is not mapping")
	}

	// Find and process lint key
	lintIdx := findKey(mapping.Content, "lint")
	if lintIdx < 0 {
		return "already cleaned: skipped", nil
	}

	// Get current version before modifying
	versionIdx := findKey(mapping.Content, "version")
	if versionIdx < 0 {
		return "", fmt.Errorf("no version field")
	}
	currentVersion := strings.Trim(mapping.Content[versionIdx+1].Value, `"'`)
	// Preserve original quoting style
	originalVersionValue := mapping.Content[versionIdx+1].Value
	wasQuoted := strings.HasPrefix(originalVersionValue, `"`) || strings.HasPrefix(originalVersionValue, `'`)
	quoteChar := ""
	if wasQuoted {
		quoteChar = string(originalVersionValue[0])
	}

	newVersion := patchBump(currentVersion)
	if wasQuoted {
		mapping.Content[versionIdx+1].Value = quoteChar + newVersion + quoteChar
	} else {
		mapping.Content[versionIdx+1].Value = newVersion
	}
	mapping.Content[versionIdx+1].Style = mapping.Content[versionIdx+1].Style

	// Update `updated` field (not `updated_at`)
	updatedIdx := findKey(mapping.Content, "updated")
	if updatedIdx >= 0 {
		mapping.Content[updatedIdx+1].Value = updatedDate
		mapping.Content[updatedIdx+1].Style = 0 // unquoted
	}

	// Remove lint key+value pair (lintIdx and lintIdx+1 in flat list)
	newContent := make([]*yaml.Node, 0, len(mapping.Content)-2)
	for i, node := range mapping.Content {
		if i == lintIdx || i == lintIdx+1 {
			continue
		}
		newContent = append(newContent, node)
	}
	mapping.Content = newContent

	// Serialize frontmatter back
	newFMBytes, err := yaml.Marshal(&docNode)
	if err != nil {
		return "", fmt.Errorf("yaml marshal: %w", err)
	}
	newFM := strings.TrimSuffix(string(newFMBytes), "\n")

	// Add HISTORY row to body
	newBody := insertHistoryRow(bodyText, newVersion, updatedDate)

	// Reconstruct file
	result := "---\n" + newFM + "\n---\n" + newBody

	if err := os.WriteFile(specPath, []byte(result), 0644); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}

	return fmt.Sprintf("cleaned (version %s → %s, updated → %s)", currentVersion, newVersion, updatedDate), nil
}

// findKey finds the index of a key in a flat yaml mapping node content slice.
// Returns -1 if not found.
func findKey(content []*yaml.Node, key string) int {
	for i := 0; i < len(content)-1; i += 2 {
		if content[i].Value == key {
			return i
		}
	}
	return -1
}

// patchBump increments the patch version (Z in X.Y.Z).
// Handles "X.Y" (rare) by appending ".1".
func patchBump(version string) string {
	// Strip surrounding quotes if any
	v := strings.Trim(version, `"'`)
	parts := strings.Split(v, ".")
	if len(parts) < 3 {
		// "X.Y" → "X.Y.1"
		parts = append(parts, "0")
	}
	patch, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		// Non-numeric patch: append .1
		return v + ".1"
	}
	parts[len(parts)-1] = strconv.Itoa(patch + 1)
	return strings.Join(parts, ".")
}

// insertHistoryRow adds a new row to the ## HISTORY section in the body.
// If no ## HISTORY section exists, creates one after the first H1 (# title).
// Handles both table-format and bullet-format HISTORY sections.
// If the section has a table, determines top-newest vs bottom-newest ordering.
func insertHistoryRow(body, newVersion, date string) string {
	newRow := fmt.Sprintf("| %-7s | %-10s | %s | %s |",
		newVersion, date, historyRowAuthor, historyRowDesc)

	lines := strings.Split(body, "\n")

	// Find ## HISTORY section
	historyIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## HISTORY" {
			historyIdx = i
			break
		}
	}

	if historyIdx < 0 {
		// No HISTORY section: create one after the first H1 heading or at start of body
		return insertHistorySection(body, newRow)
	}

	// Find the table header separator line after ## HISTORY
	// Also detect bullet lists
	tableHeaderIdx := -1
	tableSepIdx := -1
	firstDataRowIdx := -1
	lastDataRowIdx := -1
	bulletStartIdx := -1

	for i := historyIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Stop at next section
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "---") {
			break
		}

		if strings.HasPrefix(trimmed, "| ") || strings.HasPrefix(trimmed, "|---") || strings.HasPrefix(trimmed, "|:--") {
			if tableHeaderIdx < 0 && strings.Contains(trimmed, "|") && !strings.Contains(trimmed, "---") {
				tableHeaderIdx = i
			} else if tableHeaderIdx >= 0 && tableSepIdx < 0 && strings.Contains(trimmed, "---") {
				tableSepIdx = i
			} else if tableSepIdx >= 0 && strings.HasPrefix(trimmed, "|") {
				if firstDataRowIdx < 0 {
					firstDataRowIdx = i
				}
				lastDataRowIdx = i
			}
		} else if strings.HasPrefix(trimmed, "- ") {
			if bulletStartIdx < 0 {
				bulletStartIdx = i
			}
		}
	}

	if bulletStartIdx >= 0 && tableHeaderIdx < 0 {
		// Bullet-list format: append new bullet at the end of the list
		bulletRow := fmt.Sprintf("- %s v%s: %s", date, newVersion, historyRowDesc)
		return insertAtBulletEnd(lines, historyIdx, bulletRow)
	}

	if tableHeaderIdx < 0 {
		// No table or bullet found after ## HISTORY: just append a table
		return insertTableAfterHistoryHeader(lines, historyIdx, newRow)
	}

	if firstDataRowIdx < 0 {
		// Table header exists but no data rows yet
		insertPos := tableSepIdx + 1
		if tableSepIdx < 0 {
			insertPos = tableHeaderIdx + 1
		}
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertPos]...)
		newLines = append(newLines, newRow)
		newLines = append(newLines, lines[insertPos:]...)
		return strings.Join(newLines, "\n")
	}

	// Determine ordering: compare first data row version vs last data row version
	// If first row version > last row version → top-newest → insert after separator
	// If first row version < last row version → bottom-newest → append after last data row
	firstRowVersion := extractVersionFromTableRow(lines[firstDataRowIdx])
	lastRowVersion := extractVersionFromTableRow(lines[lastDataRowIdx])

	isTopNewest := compareVersions(firstRowVersion, lastRowVersion) > 0

	var insertPos int
	if isTopNewest || firstDataRowIdx == lastDataRowIdx {
		// Insert after separator (top-newest)
		if tableSepIdx >= 0 {
			insertPos = tableSepIdx + 1
		} else {
			insertPos = tableHeaderIdx + 1
		}
	} else {
		// Append after last data row (bottom-newest)
		insertPos = lastDataRowIdx + 1
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, newRow)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

// insertHistorySection creates a ## HISTORY section with a table.
// Places it after the first H1 heading if found, otherwise at the beginning of body.
func insertHistorySection(body, newRow string) string {
	lines := strings.Split(body, "\n")

	// Find first H1 heading
	h1Idx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			h1Idx = i
			break
		}
	}

	tableSection := []string{
		"",
		"## HISTORY",
		"",
		"| Version | Date       | Author                     | Description |",
		"|---------|------------|----------------------------|-------------|",
		newRow,
		"",
	}

	var insertPos int
	if h1Idx >= 0 {
		// Insert after H1 line, then skip blank lines
		insertPos = h1Idx + 1
		for insertPos < len(lines) && strings.TrimSpace(lines[insertPos]) == "" {
			insertPos++
		}
	} else {
		insertPos = 0
	}

	newLines := make([]string, 0, len(lines)+len(tableSection))
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, tableSection...)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

// insertAtBulletEnd appends a bullet item at the end of the bullet list in HISTORY.
func insertAtBulletEnd(lines []string, historyIdx int, bulletRow string) string {
	// Find the last bullet line in HISTORY section
	lastBulletIdx := -1
	for i := historyIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "---") {
			break
		}
		if strings.HasPrefix(trimmed, "- ") {
			lastBulletIdx = i
		}
	}

	insertPos := lastBulletIdx + 1
	if lastBulletIdx < 0 {
		insertPos = historyIdx + 1
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, bulletRow)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

// insertTableAfterHistoryHeader inserts a full table after ## HISTORY when there's no table.
func insertTableAfterHistoryHeader(lines []string, historyIdx int, newRow string) string {
	insertPos := historyIdx + 1
	// Skip blank lines after ## HISTORY heading
	for insertPos < len(lines) && strings.TrimSpace(lines[insertPos]) == "" {
		insertPos++
	}

	tableLines := []string{
		"| Version | Date       | Author                     | Description |",
		"|---------|------------|----------------------------|-------------|",
		newRow,
	}

	newLines := make([]string, 0, len(lines)+len(tableLines))
	newLines = append(newLines, lines[:insertPos]...)
	newLines = append(newLines, tableLines...)
	newLines = append(newLines, lines[insertPos:]...)
	return strings.Join(newLines, "\n")
}

// extractVersionFromTableRow extracts the version string from the first column of a markdown table row.
func extractVersionFromTableRow(row string) string {
	// Format: "| version | date | ..."
	parts := strings.SplitN(row, "|", 3)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// compareVersions compares two version strings (e.g. "0.3.1" vs "0.1.0").
// Returns positive if a > b, negative if a < b, 0 if equal.
func compareVersions(a, b string) int {
	aParts := strings.Split(strings.Trim(a, `"' `), ".")
	bParts := strings.Split(strings.Trim(b, `"' `), ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aVal, bVal int
		if i < len(aParts) {
			aVal, _ = strconv.Atoi(strings.TrimSpace(aParts[i]))
		}
		if i < len(bParts) {
			bVal, _ = strconv.Atoi(strings.TrimSpace(bParts[i]))
		}
		if aVal != bVal {
			return aVal - bVal
		}
	}
	return 0
}
