package project

// initializer_expansion.go — Phase 1 yaml write helpers for SPEC-V3R5-INIT-WIZARD-EXPANSION-001.
//
// Each function writes a single section yaml file when StandardMode is active.
// Defaults for coverage_exemptions sibling fields are sourced here rather than
// hardcoded, satisfying plan.md R-IWE-003 mitigation (no hardcoded sibling values).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// defaultMaxExemptPercentage matches internal/config/defaults.go DefaultMaxExemptPercentage.
const defaultMaxExemptPercentage = 15

// defaultRequireJustification matches internal/config/defaults.go default for CoverageExemptions.
const defaultRequireJustification = true

// WritePhase1Configs writes the Phase 1 section yaml files when StandardMode is active.
// When opts.StandardMode is false this function is a no-op (Quick mode backward-compat).
func WritePhase1Configs(opts InitOptions, result *InitResult) error {
	if !opts.StandardMode {
		return nil
	}

	sectionsDir := filepath.Clean(filepath.Join(opts.ProjectRoot, defs.MoAIDir, defs.SectionsSubdir))

	if err := writeProjectModeYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeHarnessProfileYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeLSPYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeQualityExpansionYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	if err := writeDesignYAML(sectionsDir, opts, result); err != nil {
		return err
	}
	return nil
}

// writeProjectModeYAML writes project.mode to project.yaml (B1, REQ-IWE-001).
// It reads the existing project.yaml and updates only the mode key.
func writeProjectModeYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	projectYAMLPath := filepath.Join(sectionsDir, defs.ProjectYAML)
	mode := opts.ProjectMode
	if mode == "" {
		mode = "personal"
	}

	// Read existing file or create a fresh block
	var content string
	existing, readErr := os.ReadFile(projectYAMLPath) //nolint:govet
	if readErr == nil {
		// Replace or append mode key
		content = patchYAMLKey(string(existing), "project", "mode", mode)
	} else {
		// Fresh project.yaml with mode key only (other keys written by generateConfigsFallback)
		content = fmt.Sprintf("project:\n  mode: %s\n", mode)
	}

	if err := os.WriteFile(projectYAMLPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write project.yaml mode: %w", err)
	}
	if readErr != nil {
		// Only append to CreatedFiles if newly created
		result.CreatedFiles = append(result.CreatedFiles,
			filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.ProjectYAML))
	}
	return nil
}

// writeHarnessProfileYAML writes harness.default_profile to harness.yaml (B2, REQ-IWE-002).
func writeHarnessProfileYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	profile := opts.HarnessProfile
	if profile == "" {
		profile = "default"
	}
	content := fmt.Sprintf("harness:\n  default_profile: %s\n", profile)
	harnessPath := filepath.Join(sectionsDir, defs.HarnessYAML)
	if err := os.WriteFile(harnessPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write harness.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles,
		filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.HarnessYAML))
	return nil
}

// writeLSPYAML writes lsp.enabled to lsp.yaml (B3, REQ-IWE-003).
func writeLSPYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	content := fmt.Sprintf("lsp:\n  enabled: %t\n", opts.LSPEnabled)
	lspPath := filepath.Join(sectionsDir, defs.LSPYAML)
	if err := os.WriteFile(lspPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write lsp.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles,
		filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.LSPYAML))
	return nil
}

// writeQualityExpansionYAML extends quality.yaml with coverage_exemptions block (B5, REQ-IWE-004).
// The existing quality.yaml (written by generateConfigsFallback) is read and the
// coverage_exemptions block is appended under the constitution: section.
func writeQualityExpansionYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	qualityPath := filepath.Join(sectionsDir, defs.QualityYAML)

	// Read existing content (may or may not exist)
	var existing string
	if data, err := os.ReadFile(qualityPath); err == nil {
		existing = string(data)
	}

	// Build the expansion block
	exemptBlock := fmt.Sprintf(`  coverage_exemptions:
    enabled: %t
    require_justification: %t
    max_exempt_percentage: %d
`,
		opts.CoverageExemptionsEnabled,
		defaultRequireJustification,
		defaultMaxExemptPercentage,
	)

	// Replace enforce_quality line and append exemptions block
	var content string
	if existing == "" {
		// Fallback: write the whole constitution block
		content = fmt.Sprintf(`constitution:
  development_mode: tdd
  enforce_quality: %t
  test_coverage_target: 85
%s`, opts.EnforceQuality, exemptBlock)
	} else {
		// Patch enforce_quality value and add exemptions block
		content = patchYAMLKey(existing, "constitution", "enforce_quality", fmt.Sprintf("%t", opts.EnforceQuality))
		// Append coverage_exemptions block if not already present
		if !yamlContains(content, "coverage_exemptions:") {
			content += exemptBlock
		}
	}

	if err := os.WriteFile(qualityPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write quality.yaml expansion: %w", err)
	}
	return nil
}

// writeDesignYAML writes design.enabled and design.claude_design.enabled to design.yaml (B8, REQ-IWE-005).
func writeDesignYAML(sectionsDir string, opts InitOptions, result *InitResult) error {
	content := fmt.Sprintf(`design:
  enabled: %t
  claude_design:
    enabled: %t
`,
		opts.DesignEnabled,
		opts.ClaudeDesignEnabled,
	)
	designPath := filepath.Join(sectionsDir, defs.DesignYAML)
	if err := os.WriteFile(designPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write design.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles,
		filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.DesignYAML))
	return nil
}

// patchYAMLKey is a simple line-by-line YAML key patcher.
// It locates "  key: <value>" under the given section and replaces the value.
// If the key is not found, it does not add it (the caller handles that).
func patchYAMLKey(content, section, key, newValue string) string {
	lines := splitLines(content)
	inSection := false
	for i, line := range lines {
		stripped := trimLeadingSpaces(line)
		if stripped == section+":" {
			inSection = true
			continue
		}
		// Leave section when we encounter another top-level key
		if inSection && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inSection = false
		}
		if inSection && trimLeadingSpaces(stripped) != "" {
			// Check for "key: ..."
			if len(stripped) > len(key)+2 && stripped[:len(key)+2] == key+": " {
				lines[i] = "  " + key + ": " + newValue
			} else if stripped == key+":" {
				lines[i] = "  " + key + ": " + newValue
			}
		}
	}
	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	// Remove trailing extra newline
	if len(result) > 0 && result[len(result)-1] == '\n' {
		// Keep single trailing newline
		for len(result) > 1 && result[len(result)-2] == '\n' {
			result = result[:len(result)-1]
		}
	}
	return result
}

// yamlFrame is one ancestor mapping key held in scope while walking a YAML
// document line by line, so patchYAMLPathValue can reconstruct the full dotted
// path of every key it visits.
type yamlFrame struct {
	indent int
	key    string
}

// patchYAMLPathValue rewrites the value of the key identified by its full
// dotted path (for example "design.claude_design.enabled"), preserving that
// line's original indentation and leaving every other byte of the document
// untouched. It returns the content and whether the path was found.
//
// It exists because patchYAMLKey is depth-blind: that helper matches on
// whitespace-stripped lines and rewrites at a hardcoded 2-space indent, so it
// rewrites EVERY key of a given name inside the section and flattens each to
// depth 2. Against the deployed lsp.yaml (two `enabled:` keys) and design.yaml
// (five) it would trade a visible clobber for silent structural corruption.
// patchYAMLKey is deliberately left untouched — its two existing callers target
// keys that are unique within their files.
//
// Only the first match is rewritten. When the path is absent the content is
// returned byte-identical with ok=false, so a caller can decide whether that is
// a no-op or an error rather than appending a duplicate mapping key.
func patchYAMLPathValue(content, path, newValue string) (string, bool) {
	if content == "" || path == "" {
		return content, false
	}
	segments := strings.Split(path, ".")
	lines := splitLines(content)
	var stack []yamlFrame

	for i, line := range lines {
		trimmed := trimLeadingSpaces(line)
		// Blank lines, comments and sequence entries carry no mapping key.
		if trimmed == "" || trimmed[0] == '#' || trimmed[0] == '-' {
			continue
		}
		colon := strings.IndexByte(trimmed, ':')
		if colon <= 0 {
			continue
		}
		key := trimmed[:colon]
		indent := len(line) - len(trimmed)

		// Leave every scope that is no longer an ancestor of this line.
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}

		if yamlPathMatches(stack, key, segments) {
			lines[i] = line[:indent] + key + ": " + newValue
			return joinYAMLLines(lines, strings.HasSuffix(content, "\n")), true
		}

		// A key with no inline value opens a nested mapping.
		if strings.TrimSpace(trimmed[colon+1:]) == "" {
			stack = append(stack, yamlFrame{indent: indent, key: key})
		}
	}
	return content, false
}

// yamlPathMatches reports whether the ancestor stack plus key spells out
// exactly the target segment list.
func yamlPathMatches(stack []yamlFrame, key string, segments []string) bool {
	if len(stack)+1 != len(segments) {
		return false
	}
	for i := range stack {
		if stack[i].key != segments[i] {
			return false
		}
	}
	return key == segments[len(segments)-1]
}

// joinYAMLLines reassembles lines produced by splitLines, restoring the
// document's original trailing-newline state.
func joinYAMLLines(lines []string, trailingNewline bool) string {
	var b strings.Builder
	for i, l := range lines {
		b.WriteString(l)
		if i < len(lines)-1 || trailingNewline {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimLeadingSpaces(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[i:]
}

func yamlContains(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
