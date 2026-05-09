package template

// @MX:NOTE: [AUTO] Audit suite for SPEC-V3R2-WF-005 language-as-rules contract.
// Three tests: TestNoLangSkillDirectory (REQ-WF005-002), TestRelatedSkillsNoLangReference (REQ-WF005-013),
// TestLanguageNeutrality (REQ-WF005-014).

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// TestNoLangSkillDirectory walks the .claude/skills/ directory in the embedded FS
// and asserts that no directory matches the ^moai-lang-[a-z-]+$ pattern.
//
// Sentinel: LANG_AS_SKILL_FORBIDDEN
// Source: SPEC-V3R2-WF-005 REQ-WF005-002, REQ-WF005-007, REQ-WF005-009
//
// @MX:ANCHOR: [AUTO] SPEC-V3R2-WF-005 REQ-WF005-002,007,009 enforcer; guards 16 language rules
// against being migrated to skills. Touching this test signature affects the contract for all 16
// supported languages.
// @MX:REASON: fan_in=16 — this test is the single CI gate that prevents moai-lang-* skill creation
// across all 16 supported languages (cpp, csharp, elixir, flutter, go, java, javascript, kotlin,
// php, python, r, ruby, rust, scala, swift, typescript).
func TestNoLangSkillDirectory(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	langSkillPattern := regexp.MustCompile(`^moai-lang-[a-z-]+$`)

	walkErr := fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Ignore walk errors (e.g., missing optional directories)
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		// Check only the final path component (directory name)
		parts := strings.Split(path, "/")
		dirName := parts[len(parts)-1]
		if langSkillPattern.MatchString(dirName) {
			t.Errorf("LANG_AS_SKILL_FORBIDDEN: %s exists; language guidance must live in "+
				".claude/rules/moai/languages/, not as a skill", path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude/skills) error: %v", walkErr)
	}
}

// TestRelatedSkillsNoLangReference walks all .claude/skills/**/SKILL.md files in the
// embedded FS, parses the YAML frontmatter, and asserts that the metadata.related-skills
// field does not contain any moai-lang- substring.
//
// Sentinel: DEAD_LANG_FRONTMATTER_REFERENCE
// Source: SPEC-V3R2-WF-005 REQ-WF005-005, REQ-WF005-008; AC-WF005-06.
// Body-prose enforcement is delegated to TestSkillBodyNoLangReference
// (DEAD_LANG_SKILL_REFERENCE per REQ-WF005-013, AC-WF005-12).
func TestRelatedSkillsNoLangReference(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	langTokenPattern := regexp.MustCompile(`\bmoai-lang-[a-z]+\b`)

	var skillFiles []string
	walkErr := fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, "SKILL.md") {
			skillFiles = append(skillFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude/skills) error: %v", walkErr)
	}

	if len(skillFiles) == 0 {
		t.Fatal("no SKILL.md files found under .claude/skills/")
	}

	for _, path := range skillFiles {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			// Parse frontmatter: extract the metadata block
			relatedSkills := extractRelatedSkills(string(data))
			if relatedSkills == "" {
				return // No related-skills field — OK
			}

			// Check each token in the related-skills CSV
			tokens := strings.SplitSeq(relatedSkills, ",")
			for token := range tokens {
				token = strings.TrimSpace(strings.Trim(token, `"' `))
				if langTokenPattern.MatchString(token) {
					t.Errorf("DEAD_LANG_FRONTMATTER_REFERENCE: %s related-skills contains %q; "+
						"substitute with .claude/rules/moai/languages/<name>.md", path, token)
				}
			}
		})
	}

	t.Logf("audited %d SKILL.md files for related-skills lang references", len(skillFiles))
}

// TestSkillBodyNoLangReference walks all .claude/skills/**/SKILL.md files in the
// embedded FS, strips frontmatter and fenced code blocks via
// stripFrontmatterAndCodeBlocks, and asserts no body prose mentions a
// moai-lang-<name> identifier. Code-block examples and frontmatter are excluded
// from the scan to avoid false positives on documentation samples.
//
// Sentinel: DEAD_LANG_SKILL_REFERENCE
// Source: SPEC-V3R2-WF-005 REQ-WF005-013, AC-WF005-12.
// Frontmatter enforcement is delegated to TestRelatedSkillsNoLangReference
// (DEAD_LANG_FRONTMATTER_REFERENCE per REQ-WF005-005, REQ-WF005-008).
func TestSkillBodyNoLangReference(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	langTokenPattern := regexp.MustCompile(`\bmoai-lang-[a-z]+\b`)

	var skillFiles []string
	walkErr := fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, "SKILL.md") {
			skillFiles = append(skillFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(.claude/skills) error: %v", walkErr)
	}

	if len(skillFiles) == 0 {
		t.Fatal("no SKILL.md files found under .claude/skills/")
	}

	for _, path := range skillFiles {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			body := stripFrontmatterAndCodeBlocks(string(data))
			lines := strings.Split(body, "\n")
			for lineNum, line := range lines {
				if matches := langTokenPattern.FindAllString(line, -1); len(matches) > 0 {
					for _, match := range matches {
						t.Errorf("DEAD_LANG_SKILL_REFERENCE: %s body line %d mentions %q; "+
							"substitute with .claude/rules/moai/languages/<name>.md",
							path, lineNum+1, match)
					}
				}
			}
		})
	}

	t.Logf("audited %d SKILL.md bodies for moai-lang-* prose references", len(skillFiles))
}

// TestLanguageNeutrality walks all .claude/skills/**/SKILL.md and .claude/rules/**/*.md
// files in the embedded FS, scans body bytes (excluding fenced code blocks), and asserts
// that no language-primacy phrases appear.
//
// Sentinel: LANG_NEUTRALITY_VIOLATION
// Source: SPEC-V3R2-WF-005 REQ-WF005-014; CLAUDE.local.md §15 (16-language neutrality)
func TestLanguageNeutrality(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	// Language names matching the 16 supported languages per CLAUDE.local.md §15
	langNames := `(?:go|python|typescript|javascript|rust|java|kotlin|csharp|ruby|php|elixir|cpp|scala|r|flutter|swift)`

	// Primacy phrase patterns (per research.md §6.2)
	// These match phrases that declare a language as primary or only supported
	primacyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b` + langNames + `\s+is\s+(?:primary|the\s+primary|the\s+main|the\s+default)\b`),
		regexp.MustCompile(`(?i)\bonly\s+` + langNames + `\s+is\s+supported\b`),
		regexp.MustCompile(`(?i)\b` + langNames + `\s+is\s+the\s+primary\s+language\b`),
		regexp.MustCompile(`(?i)\b` + langNames + `\s+is\s+primary;\s+other\s+languages\b`),
	}

	var targetFiles []string

	// Collect SKILL.md files
	_ = fs.WalkDir(fsys, ".claude/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "SKILL.md") {
			targetFiles = append(targetFiles, path)
		}
		return nil
	})

	// Collect rule .md files
	_ = fs.WalkDir(fsys, ".claude/rules", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			targetFiles = append(targetFiles, path)
		}
		return nil
	})

	if len(targetFiles) == 0 {
		t.Fatal("no skill/rule files found to audit")
	}

	for _, path := range targetFiles {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			// Strip frontmatter and fenced code blocks before scanning body
			body := stripFrontmatterAndCodeBlocks(string(data))

			lines := strings.Split(body, "\n")
			for lineNum, line := range lines {
				for _, pattern := range primacyPatterns {
					if pattern.MatchString(line) {
						t.Errorf("LANG_NEUTRALITY_VIOLATION: %s body line %d contains primacy phrase %q; "+
							"per CLAUDE.local.md §15, all 16 languages must be treated equally",
							path, lineNum+1, line)
					}
				}
			}
		})
	}

	t.Logf("audited %d skill/rule files for language neutrality", len(targetFiles))
}

// extractRelatedSkills parses YAML frontmatter from a SKILL.md file and returns
// the value of the metadata.related-skills field, or empty string if not found.
// The frontmatter is delimited by --- lines.
func extractRelatedSkills(content string) string {
	lines := strings.Split(content, "\n")

	// Find opening ---
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}

	// Find closing ---
	end := -1
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return ""
	}

	// Scan frontmatter lines for related-skills under metadata:
	inMetadata := false
	for _, line := range lines[start+1 : end] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "metadata:" {
			inMetadata = true
			continue
		}
		if inMetadata {
			// Metadata block ends when we see a non-indented non-empty line
			// (except metadata sub-keys which are indented)
			if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				inMetadata = false
				continue
			}
			if strings.Contains(trimmed, "related-skills:") {
				_, after, _ := strings.Cut(trimmed, "related-skills:")
				val := strings.TrimSpace(after)
				// Strip surrounding quotes
				val = strings.Trim(val, `"'`)
				return val
			}
		}
	}
	return ""
}

// stripFrontmatterAndCodeBlocks removes YAML frontmatter (--- delimited) and
// fenced code blocks (``` delimited) from markdown content so that language
// primacy phrases inside code examples do not trigger false positives.
func stripFrontmatterAndCodeBlocks(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	// Skip frontmatter
	inFrontmatter := false
	frontmatterDone := false
	inCodeBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle frontmatter
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if trimmed == "---" {
				inFrontmatter = false
				frontmatterDone = true
			}
			continue
		}
		if !frontmatterDone && i == 0 {
			frontmatterDone = true
		}

		// Handle fenced code blocks (``` or ~~~)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
