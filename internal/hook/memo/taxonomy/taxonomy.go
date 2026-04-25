// Package taxonomy provides 4-type memory taxonomy enforcement for MoAI agent
// persistent memory files stored under .claude/agent-memory/<agent-name>/.
//
// It implements REQ-EXT001-001..004 from SPEC-V3R2-EXT-001.
package taxonomy

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// MemoryType is the enum for agent memory file types (REQ-EXT001-001/004).
// Exactly 4 values are allowed in v3.0.0; adding a type requires a new SPEC.
type MemoryType string

// The four canonical memory types (REQ-EXT001-004: fixed at 4 values in v3.0.0).
const (
	TypeUser      MemoryType = "user"
	TypeFeedback  MemoryType = "feedback"
	TypeProject   MemoryType = "project"
	TypeReference MemoryType = "reference"
)

// ValidTypes is the immutable set of all valid MemoryType values.
// Invariant: len(ValidTypes) == 4 (REQ-EXT001-004, AC-EXT001-09).
var ValidTypes = []MemoryType{TypeUser, TypeFeedback, TypeProject, TypeReference}

// ErrNoFrontmatter is returned when a markdown file contains no YAML frontmatter block.
var ErrNoFrontmatter = errors.New("taxonomy: no YAML frontmatter found")

// ErrSymlink is returned when ParseFile encounters a symbolic link.
var ErrSymlink = errors.New("taxonomy: symlink files must not be followed")

// Frontmatter holds the parsed fields from a memory file's YAML frontmatter.
// Raw contains all parsed key-value pairs as a fallback for unrecognised keys.
type Frontmatter struct {
	Name        string
	Description string
	Type        MemoryType
	Raw         map[string]string
}

// ValidateType returns nil if t is one of the 4 canonical MemoryType values,
// or an error otherwise. Implements REQ-EXT001-001.
func ValidateType(t MemoryType) error {
	for _, v := range ValidTypes {
		if t == v {
			return nil
		}
	}
	return fmt.Errorf("taxonomy: unknown memory type %q; valid types are user, feedback, project, reference (REQ-EXT001-004)", t)
}

// ParseFile reads the markdown file at path, extracts its YAML frontmatter, and
// returns the parsed Frontmatter plus the body text after the closing "---" line.
//
// Rules:
//   - Symlinks are rejected with ErrSymlink (security: no path traversal via symlinks).
//   - Files with no frontmatter return ErrNoFrontmatter.
//   - Missing type/name/description keys result in empty string fields; callers should
//     use AuditFile to detect and report these conditions (REQ-EXT001-007).
func ParseFile(path string) (Frontmatter, string, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return Frontmatter{}, "", fmt.Errorf("taxonomy: lstat %s: %w", path, err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return Frontmatter{}, "", ErrSymlink
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Frontmatter{}, "", fmt.Errorf("taxonomy: read %s: %w", path, err)
	}

	return parseFrontmatter(string(data))
}

// parseFrontmatter parses raw markdown content and extracts the YAML frontmatter block.
// It returns ErrNoFrontmatter when the content does not start with "---".
func parseFrontmatter(content string) (Frontmatter, string, error) {
	if len(content) == 0 {
		return Frontmatter{}, "", ErrNoFrontmatter
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Frontmatter{}, "", ErrNoFrontmatter
	}

	// Find closing "---"
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx < 0 {
		return Frontmatter{}, "", ErrNoFrontmatter
	}

	raw := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(strings.Join(lines[1:closingIdx], "\n")))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		idx := strings.IndexByte(line, ':')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		raw[key] = val
	}

	fm := Frontmatter{
		Name:        raw["name"],
		Description: raw["description"],
		Type:        MemoryType(raw["type"]),
		Raw:         raw,
	}

	body := ""
	if closingIdx+1 < len(lines) {
		body = strings.Join(lines[closingIdx+1:], "\n")
	}

	return fm, body, nil
}
