package template

// Korean Leftover Audit — issue #1055
//
// Test A: TestNoKoreanRecommendedMarkerInEnglishInstructionDocs
//   Asserts that the English instruction docs no longer contain the hardcoded
//   Korean "(권장)" marker. The canonical SSOT definition in askuser-protocol.md
//   is intentionally kept; only the two English instruction files are checked here.
//
// Test B: TestNoKoreanInConfigComments
//   Walks all *.yaml / *.yaml.tmpl files under .moai/config/sections/ in the
//   embedded FS and asserts that no YAML comment line contains Hangul syllables.
//   ALLOWLIST: language.yaml.tmpl "# Examples:" line (multilingual example).

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// hangulRangeRE matches any Hangul syllable (U+AC00 – U+D7A3).
var hangulRangeRE = regexp.MustCompile(`[\x{AC00}-\x{D7A3}]`)

// TestNoKoreanRecommendedMarkerInEnglishInstructionDocs asserts that the two
// main English instruction documents no longer contain the literal "(권장)".
//
// Sentinel: KOREAN_RECOMMENDED_MARKER_IN_ENGLISH_DOCS
// Source: GitHub issue #1055 — English-default templates ship hardcoded Korean.
func TestNoKoreanRecommendedMarkerInEnglishInstructionDocs(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	targets := []string{
		"CLAUDE.md",
		".claude/rules/moai/core/moai-constitution.md",
	}

	const marker = "(권장)"

	for _, path := range targets {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			lines := strings.Split(string(data), "\n")
			for lineNum, line := range lines {
				if strings.Contains(line, marker) {
					t.Errorf(
						"KOREAN_RECOMMENDED_MARKER_IN_ENGLISH_DOCS: %s line %d contains %q; "+
							"replace with \"(Recommended)\" only",
						path, lineNum+1, marker,
					)
				}
			}
		})
	}
}

// TestNoKoreanInConfigComments walks every *.yaml and *.yaml.tmpl file under
// .moai/config/sections/ in the embedded FS.  For each line, the substring
// after the first '#' that is the comment opener (i.e., '#' is either the
// first non-whitespace character, or is preceded only by whitespace) is
// inspected for Hangul syllables.  Any match is a test failure.
//
// ALLOWLIST:
//   - language.yaml.tmpl line containing "# Examples:" — intentional
//     multilingual example showing "Korean (한국어)" / "Japanese (日本語)".
//
// Sentinel: KOREAN_IN_CONFIG_YAML_COMMENT
// Source: GitHub issue #1055 — Englishize ALL hardcoded Korean in English context.
func TestNoKoreanInConfigComments(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	configDir := ".moai/config/sections"

	var yamlFiles []string
	walkErr := fs.WalkDir(fsys, configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yaml.tmpl") {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(%q) error: %v", configDir, walkErr)
	}

	if len(yamlFiles) == 0 {
		t.Fatalf("no *.yaml or *.yaml.tmpl files found under %q", configDir)
	}

	for _, path := range yamlFiles {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			lines := strings.Split(string(data), "\n")
			for lineNum, line := range lines {
				commentPortion := extractYAMLCommentPortion(line)
				if commentPortion == "" {
					continue
				}

				// ALLOWLIST: language.yaml.tmpl "# Examples:" multilingual line.
				if strings.HasSuffix(path, "language.yaml.tmpl") &&
					strings.Contains(commentPortion, "Examples:") {
					continue
				}

				if hangulRangeRE.MatchString(commentPortion) {
					t.Errorf(
						"KOREAN_IN_CONFIG_YAML_COMMENT: %s line %d has Korean in comment: %q",
						path, lineNum+1, line,
					)
				}
			}
		})
	}

	t.Logf("audited %d yaml files for Korean in comments", len(yamlFiles))
}

// extractYAMLCommentPortion returns the comment text (everything after the '#')
// when the line's first non-whitespace character is '#', or when '#' is
// preceded only by whitespace (i.e., it is a standalone comment line, not a
// '#' character embedded inside a quoted string value).
//
// This simple heuristic is sufficient for the config/sections/*.yaml files
// because those files use '#' only as YAML comment openers.
func extractYAMLCommentPortion(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(trimmed, "#") {
		return ""
	}
	// Return everything after the '#'
	return trimmed[1:]
}
