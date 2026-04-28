package mx

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Scanner scans source files for @MX tags.
type Scanner struct {
	ignorePatterns []string // .gitignore-style patterns to skip
	anchorIDs      map[string]string // AnchorID -> file:line (for duplicate detection)
	warnings      []string // Scanner warnings
	errors        []string // Scanner errors
}

// NewScanner creates a new tag scanner.
func NewScanner() *Scanner {
	return &Scanner{
		anchorIDs: make(map[string]string),
		warnings:  make([]string, 0),
		errors:    make([]string, 0),
	}
}

// SetIgnorePatterns sets the list of file patterns to ignore.
func (s *Scanner) SetIgnorePatterns(patterns []string) {
	s.ignorePatterns = patterns
}

// ScanFile scans a single file for @MX tags.
func (s *Scanner) ScanFile(filePath string) ([]Tag, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Get comment prefix for this file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	prefix := GetCommentPrefix(ext)
	if prefix == "" {
		// Unsupported language - skip
		return nil, nil
	}

	var tags []Tag
	scanner := bufio.NewScanner(file)
	lineNum := 0

	var pendingWarnTag *Tag // Track WARN tag that needs REASON

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check if line contains @MX tag
		if !strings.Contains(line, "@MX:") {
			if pendingWarnTag != nil && lineNum > pendingWarnTag.Line+3 {
				// No REASON found within 3 lines - emit warning
				s.warnings = append(s.warnings,
					fmt.Sprintf("MissingReasonForWarn: %s:%d - WARN tag without REASON within 3 lines",
						pendingWarnTag.File, pendingWarnTag.Line))
				pendingWarnTag = nil
			}
			continue
		}

		// Extract tag content after comment prefix
		tagContent, ok := extractTagContent(line, prefix)
		if !ok {
			continue
		}

		// Parse the tag
		tag, err := s.parseTag(filePath, lineNum, tagContent)
		if err != nil {
			s.errors = append(s.errors, fmt.Sprintf("parse error at %s:%d: %v", filePath, lineNum, err))
			continue
		}

		// Handle WARN tag REASON requirement
		if tag.Kind == MXWarn {
			pendingWarnTag = &tag
			// Continue scanning to find REASON in next 3 lines
		} else {
			pendingWarnTag = nil
		}

		// Check for REASON sub-line (for WARN and ANCHOR)
		if strings.Contains(line, "@MX:REASON") {
			reason := extractReason(line)
			if reason != "" {
				if pendingWarnTag != nil {
					pendingWarnTag.Reason = reason
					tags = append(tags, *pendingWarnTag)
					pendingWarnTag = nil
				} else if len(tags) > 0 {
					// Update the last tag if it's WARN or ANCHOR
					lastIdx := len(tags) - 1
					if tags[lastIdx].Kind == MXWarn || tags[lastIdx].Kind == MXAnchor {
						tags[lastIdx].Reason = reason
					}
				}
			}
		}

		// Check for duplicate AnchorID
		if tag.Kind == MXAnchor && tag.AnchorID != "" {
			if prevLoc, exists := s.anchorIDs[tag.AnchorID]; exists {
				s.errors = append(s.errors,
					fmt.Sprintf("DuplicateAnchorID: %s used at both %s and %s:%d",
						tag.AnchorID, prevLoc, filePath, lineNum))
			} else {
				s.anchorIDs[tag.AnchorID] = fmt.Sprintf("%s:%d", filePath, lineNum)
			}
		}

		// Check if we're beyond 3 lines for pending WARN tag
		if pendingWarnTag != nil && lineNum > pendingWarnTag.Line+3 {
			// No REASON found within 3 lines - emit warning
			s.warnings = append(s.warnings,
				fmt.Sprintf("MissingReasonForWarn: %s:%d - WARN tag without REASON within 3 lines",
					pendingWarnTag.File, pendingWarnTag.Line))
			pendingWarnTag = nil
		}

		// Add tag to results (WARN tags without reason are not added yet)
		if tag.Kind != MXWarn {
			tags = append(tags, tag)
		} else if pendingWarnTag == nil {
			// WARN tag with reason already added
			tags = append(tags, tag)
		}
	}

	// Check for unresolved pending WARN tag at end of file
	if pendingWarnTag != nil && pendingWarnTag.Reason == "" {
		s.warnings = append(s.warnings,
			fmt.Sprintf("MissingReasonForWarn: %s:%d - WARN tag without REASON",
				pendingWarnTag.File, pendingWarnTag.Line))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return tags, nil
}

// ScanDir recursively scans a directory for @MX tags.
func (s *Scanner) ScanDir(rootDir string) ([]Tag, error) {
	var allTags []Tag

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Check ignore patterns
			for _, pattern := range s.ignorePatterns {
				matched, err := filepath.Match(pattern, filepath.Base(path))
				if err == nil && matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Skip files matching ignore patterns
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}
		for _, pattern := range s.ignorePatterns {
			matched, err := filepath.Match(pattern, relPath)
			if err == nil && matched {
				return nil
			}
		}

		// Scan the file
		tags, err := s.ScanFile(path)
		if err != nil {
			// Log error but continue scanning
			s.errors = append(s.errors, fmt.Sprintf("scan %s: %v", path, err))
			return nil
		}

		allTags = append(allTags, tags...)
		return nil
	})

	return allTags, err
}

// GetWarnings returns all scanner warnings.
func (s *Scanner) GetWarnings() []string {
	return s.warnings
}

// GetErrors returns all scanner errors.
func (s *Scanner) GetErrors() []string {
	return s.errors
}

// HasErrors returns true if the scanner encountered any errors.
func (s *Scanner) HasErrors() bool {
	return len(s.errors) > 0
}

// extractTagContent extracts the @MX tag content from a line.
// Returns the content after "@MX:" and true if successful.
func extractTagContent(line, commentPrefix string) (string, bool) {
	// Find comment prefix
	_, rest, found := strings.Cut(line, commentPrefix)
	if !found {
		return "", false
	}

	// Find @MX: anywhere in the line (case-insensitive)
	mxIdx := strings.Index(strings.ToUpper(rest), "@MX:")
	if mxIdx == -1 {
		return "", false
	}

	// Extract tag content (everything after "@MX:")
	tagContent := strings.TrimSpace(rest[mxIdx+4:]) // "@MX:" is 4 chars
	return tagContent, true
}

// parseTag parses a tag content string into a Tag struct.
// Expected format: "KIND: body text" or "ANCHOR:anchor-id: body text"
func (s *Scanner) parseTag(filePath string, lineNum int, content string) (Tag, error) {
	// Split by first colon to get kind
	parts := strings.SplitN(content, ":", 2)
	if len(parts) < 2 {
		return Tag{}, fmt.Errorf("invalid tag format: %s", content)
	}

	kindStr := strings.TrimSpace(parts[0])
	kind := TagKind(strings.ToUpper(kindStr))

	// Validate kind
	switch kind {
	case MXNote, MXWarn, MXAnchor, MXTodo, MXLegacy:
		// Valid kind
	default:
		return Tag{}, fmt.Errorf("unknown tag kind: %s", kindStr)
	}

	// Extract body and optional AnchorID
	body := strings.TrimSpace(parts[1])
	var anchorID string

	if kind == MXAnchor {
		// ANCHOR format: "ANCHOR:anchor-id: body text"
		anchorParts := strings.SplitN(body, ":", 2)
		if len(anchorParts) >= 2 {
			anchorID = strings.TrimSpace(anchorParts[0])
			body = strings.TrimSpace(anchorParts[1])
		}
	}

	return Tag{
		Kind:       kind,
		File:       filePath,
		Line:       lineNum,
		Body:       body,
		AnchorID:   anchorID,
		CreatedBy:  "scanner", // TODO: Detect if human-created
		LastSeenAt: time.Now(),
	}, nil
}

// extractReason extracts the reason text from a @MX:REASON line.
func extractReason(line string) string {
	// Find @MX:REASON:
	re := regexp.MustCompile(`@MX:REASON:\s*(.+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

// DetectChanges compares new tags with existing sidecar tags and returns delta.
// Used by PostToolUse hook to emit only changed tags.
func DetectChanges(oldTags, newTags []Tag) (added, changed, removed []Tag) {
	oldIndex := make(map[string]Tag)
	for _, tag := range oldTags {
		oldIndex[tag.Key()] = tag
	}

	newIndex := make(map[string]Tag)
	for _, tag := range newTags {
		newIndex[tag.Key()] = tag
	}

	// Find added and changed
	for key, newTag := range newIndex {
		oldTag, exists := oldIndex[key]
		if !exists {
			added = append(added, newTag)
		} else if oldTag.Body != newTag.Body || oldTag.Reason != newTag.Reason {
			changed = append(changed, newTag)
		}
	}

	// Find removed
	for key, oldTag := range oldIndex {
		if _, exists := newIndex[key]; !exists {
			removed = append(removed, oldTag)
		}
	}

	return added, changed, removed
}
