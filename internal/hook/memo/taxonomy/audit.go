package taxonomy

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// AuditCode identifies the type of audit finding.
type AuditCode string

// Audit warning codes (SPEC-V3R2-EXT-001).
const (
	WarnMissingType          AuditCode = "MEMORY_MISSING_TYPE"
	WarnMissingFrontmatter   AuditCode = "MEMORY_MISSING_FRONTMATTER"
	WarnIndexOverflow        AuditCode = "MEMORY_INDEX_OVERFLOW"
	WarnBodyStructureMissing AuditCode = "MEMORY_BODY_STRUCTURE_MISSING"
	WarnExcludedCategory     AuditCode = "MEMORY_EXCLUDED_CATEGORY"
	WarnDuplicate            AuditCode = "MEMORY_DUPLICATE"
)

// AuditFinding represents a single audit warning for a memory file.
type AuditFinding struct {
	// Code is the machine-readable warning code.
	Code AuditCode
	// Path is the file or directory that triggered the finding.
	Path string
	// Detail contains human-readable context (category name, line count, missing keys, etc.).
	Detail string
}

// feedbackBodyRegex matches a feedback/project body that has both
// **Why:** and **How to apply:** markers in order (REQ-EXT001-010).
// The (?s) flag makes . match newlines.
var feedbackBodyRegex = regexp.MustCompile(`(?s).+?\*\*Why:\*\*.+?\*\*How to apply:\*\*`)

// excludedCategoryKeywords is the static v1 keyword map for MEMORY_EXCLUDED_CATEGORY
// detection (OPEN-2 RESOLVED: static-keyword v1, plan.md §8).
//
// Each entry maps a category name to a list of case-insensitive substrings.
// When any keyword matches the file body, the category is flagged.
var excludedCategoryKeywords = map[string][]string{
	"code_pattern":   {"```go", "```python", "```typescript", "```js", "```javascript", "```rust", "func ", "def ", "class "},
	"git_history":    {"commit ", "git log", "git blame", "HEAD~", "git diff"},
	"debug_recipe":   {"fix recipe", "debug solution", "root cause:", "fixed in commit"},
	"claude_md_mirror": {"CLAUDE.md", "TRUST 5", "MoAI-ADK"},
	"ephemeral_state": {"currently working on", "in progress:", "TODO for this session", "current branch"},
}

// claudeMdMirrorThreshold is the minimum number of claude_md_mirror keywords
// that must match before flagging a file (avoids single-mention false positives).
const claudeMdMirrorThreshold = 3

// AuditFile audits a single memory file and returns findings for any violations.
//
// Checks performed:
//   - MEMORY_MISSING_TYPE: frontmatter has no type or empty type
//   - MEMORY_MISSING_FRONTMATTER: frontmatter is present but name or description is absent
//   - MEMORY_BODY_STRUCTURE_MISSING: feedback/project files missing **Why:** / **How to apply:**
//   - MEMORY_EXCLUDED_CATEGORY: body matches excluded-category keywords
func AuditFile(path string) ([]AuditFinding, error) {
	fm, body, err := ParseFile(path)
	if err != nil {
		// File without frontmatter → missing type.
		if err == ErrNoFrontmatter {
			return []AuditFinding{{
				Code:   WarnMissingType,
				Path:   path,
				Detail: "file has no YAML frontmatter",
			}}, nil
		}
		// Symlink or read error → skip silently (security).
		return nil, nil //nolint:nilerr
	}

	var findings []AuditFinding

	// Check type presence and validity.
	if fm.Type == "" {
		findings = append(findings, AuditFinding{
			Code:   WarnMissingType,
			Path:   path,
			Detail: "frontmatter 'type' key is missing",
		})
	}

	// Check required frontmatter keys (REQ-EXT001-002).
	missingKeys := []string{}
	if fm.Name == "" {
		missingKeys = append(missingKeys, "name")
	}
	if fm.Description == "" {
		missingKeys = append(missingKeys, "description")
	}
	if len(missingKeys) > 0 {
		findings = append(findings, AuditFinding{
			Code:   WarnMissingFrontmatter,
			Path:   path,
			Detail: fmt.Sprintf("missing required frontmatter keys: %s", strings.Join(missingKeys, ", ")),
		})
	}

	// Check body structure for feedback and project types (REQ-EXT001-010/011).
	if fm.Type == TypeFeedback || fm.Type == TypeProject {
		if !feedbackBodyRegex.MatchString(body) {
			findings = append(findings, AuditFinding{
				Code:   WarnBodyStructureMissing,
				Path:   path,
				Detail: fmt.Sprintf("type %q body must contain rule text then **Why:** then **How to apply:** markers", fm.Type),
			})
		}
	}

	// Check for excluded categories (REQ-EXT001-015, OPEN-2 static-keyword v1).
	if cat := detectExcludedCategory(body); cat != "" {
		findings = append(findings, AuditFinding{
			Code:   WarnExcludedCategory,
			Path:   path,
			Detail: fmt.Sprintf("content matches excluded category %q — do not store this in persistent memory", cat),
		})
	}

	return findings, nil
}

// detectExcludedCategory checks body against the static keyword map.
// Returns the matching category name, or empty string if no match.
// For "claude_md_mirror", requires claudeMdMirrorThreshold distinct keyword matches.
func detectExcludedCategory(body string) string {
	lower := strings.ToLower(body)
	for cat, keywords := range excludedCategoryKeywords {
		if cat == "claude_md_mirror" {
			matches := 0
			for _, kw := range keywords {
				if strings.Contains(body, kw) { // case-sensitive for CLAUDE.md/TRUST 5/MoAI-ADK
					matches++
				}
			}
			if matches >= claudeMdMirrorThreshold {
				return cat
			}
			continue
		}
		for _, kw := range keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				return cat
			}
		}
	}
	return ""
}

// AuditIndex checks a MEMORY.md index file for line-count overflow.
// Returns MEMORY_INDEX_OVERFLOW when the file exceeds lineCap lines
// (REQ-EXT001-003/008, AC-EXT001-03).
func AuditIndex(indexPath string, lineCap int) ([]AuditFinding, error) {
	if lineCap <= 0 {
		lineCap = config.DefaultMemoryIndexLineCap
	}

	f, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("taxonomy: AuditIndex open %s: %w", indexPath, err)
	}
	defer func() { _ = f.Close() }()

	lines := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("taxonomy: AuditIndex scan %s: %w", indexPath, err)
	}

	if lines > lineCap {
		return []AuditFinding{{
			Code:   WarnIndexOverflow,
			Path:   indexPath,
			Detail: fmt.Sprintf("MEMORY.md has %d lines; cap is %d — archive older entries", lines, lineCap),
		}}, nil
	}

	return nil, nil
}

// AuditDuplicates scans dir for memory files (.md, excluding MEMORY.md) and
// reports any pair that shares the same frontmatter description value
// (REQ-EXT001-016, AC-EXT001-11).
//
// Detection is description-based and type-agnostic per acceptance.md edge cases.
func AuditDuplicates(dir string) ([]AuditFinding, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("taxonomy: AuditDuplicates read dir %s: %w", dir, err)
	}

	// Map description → list of file paths.
	descToFiles := make(map[string][]string)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if e.Name() == "MEMORY.md" {
			continue
		}
		// Skip symlinks.
		if e.Type()&os.ModeSymlink != 0 {
			continue
		}

		fullPath := filepath.Join(dir, e.Name())
		fm, _, err := ParseFile(fullPath)
		if err != nil {
			continue
		}
		if fm.Description == "" {
			continue
		}

		descToFiles[fm.Description] = append(descToFiles[fm.Description], fullPath)
	}

	var findings []AuditFinding
	for desc, paths := range descToFiles {
		if len(paths) < 2 {
			continue
		}
		findings = append(findings, AuditFinding{
			Code:   WarnDuplicate,
			Path:   dir,
			Detail: fmt.Sprintf("duplicate description %q found in: %s — consider merging", desc, strings.Join(paths, ", ")),
		})
	}

	return findings, nil
}
