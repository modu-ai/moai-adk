package template

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// @MX:NOTE: [AUTO] agents_frontmatter_test.go — REQ-CC2122-HOOK-002 static linter
// @MX:SPEC: SPEC-CC2122-HOOK-002 REQ-004/005/006
//
// This test statically verifies that the frontmatter of .claude/agents/**/*.md
// complies with the following rules:
//  1. tools and disallowedTools are mutually exclusive (claude-code v2.1.119+ spec)
//  2. When disallowedTools is defined, it MUST follow the CSV string format (no whitespace-only separators)
//  3. When tools is defined, it MUST follow the CSV string format (CLAUDE.local.md §12)
//
// Real `claude --print` regression coverage is split out into manual verification
// per the GLM integration test policy (CLAUDE.local.md §13).

// validateToolsCSVFormat verifies that the tools or disallowedTools field value follows the CSV format.
// An empty value is allowed (treated as field absence).
// YAML array syntax ([...]) is forbidden.
// Multi-token whitespace-only separation is forbidden (comma separator required).
//
// Returns: an error message on violation, an empty string when valid.
func validateToolsCSVFormat(field, value string) string {
	if value == "" {
		return ""
	}

	trimmed := strings.TrimSpace(value)

	// Forbid YAML array syntax: a "[A, B, C]" form is captured verbatim as a string by the frontmatter parser.
	if strings.HasPrefix(trimmed, "[") {
		return fmt.Sprintf("%s 값이 YAML 배열 syntax 로 시작함 (CSV 문자열 필수): %q", field, value)
	}

	// Multiple tokens separated by whitespace without commas are a violation (e.g. "Read Write Edit").
	// A single token (e.g. "Read") is allowed.
	hasComma := strings.Contains(trimmed, ",")
	hasMultipleTokens := len(strings.Fields(trimmed)) > 1
	if hasMultipleTokens && !hasComma {
		return fmt.Sprintf("%s 값이 공백 구분으로 보임 (CSV 형식 필수, 콤마 분리): %q", field, value)
	}

	return ""
}

// collectAgentFiles collects every .md file under .claude/agents/ from the embedded templates.
func collectAgentFiles(t *testing.T) []string {
	t.Helper()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	var agentFiles []string
	walkErr := fs.WalkDir(fsys, ".claude/agents", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".md.tmpl") {
			agentFiles = append(agentFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("WalkDir error: %v", walkErr)
	}

	if len(agentFiles) == 0 {
		t.Fatal("no agent files found under .claude/agents/")
	}
	return agentFiles
}

// TestAgentsFrontmatter_ToolsDisallowedMutualExclusion verifies that no agent
// definition declares both `tools:` and `disallowedTools:` simultaneously.
// REQ-CC2122-HOOK-002-004
func TestAgentsFrontmatter_ToolsDisallowedMutualExclusion(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	for _, path := range collectAgentFiles(t) {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				t.Fatalf("frontmatter 파싱 실패: %s", parseErr)
			}

			_, hasTools := fm["tools"]
			_, hasDisallowed := fm["disallowedTools"]

			if hasTools && hasDisallowed {
				t.Errorf("REQ-CC2122-HOOK-002-004 위반: tools 와 disallowedTools 가 동시 정의됨. claude-code v2.1.119+ 에서는 mutually exclusive 이어야 함.")
			}
		})
	}
}

// TestAgentsFrontmatter_ToolsCSVFormat verifies that `tools:` and `disallowedTools:`
// values follow CSV format (no whitespace-only separator, no array syntax).
// REQ-CC2122-HOOK-002-005
func TestAgentsFrontmatter_ToolsCSVFormat(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	for _, path := range collectAgentFiles(t) {
		t.Run(path, func(t *testing.T) {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) error: %v", path, readErr)
			}

			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				t.Fatalf("frontmatter 파싱 실패: %s", parseErr)
			}

			// retired:true agents are allowed to have empty arrays tools: [] / skills: [].
			// Skip here because TestAgentFrontmatterAudit verifies it separately.
			rf := parseRetiredFields(fm)
			if rf.retired {
				t.Skip("retired 에이전트: tools:[] 빈 배열은 TestAgentFrontmatterAudit에서 검증")
			}

			if msg := validateToolsCSVFormat("tools", fm["tools"]); msg != "" {
				t.Errorf("REQ-CC2122-HOOK-002-005 위반: %s", msg)
			}
			if msg := validateToolsCSVFormat("disallowedTools", fm["disallowedTools"]); msg != "" {
				t.Errorf("REQ-CC2122-HOOK-002-005 위반: %s", msg)
			}
		})
	}
}

// TestValidateToolsCSVFormat_Cases unit-tests the validateToolsCSVFormat helper
// itself. Protects against regressions as the frontmatter validator is hardened in the future.
// REQ-CC2122-HOOK-002-006
func TestValidateToolsCSVFormat_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		field     string
		value     string
		wantFails bool
	}{
		{name: "empty", field: "tools", value: "", wantFails: false},
		{name: "single_tool", field: "tools", value: "Read", wantFails: false},
		{name: "csv_with_space", field: "tools", value: "Read, Write, Edit", wantFails: false},
		{name: "csv_no_space", field: "disallowedTools", value: "Bash,Write", wantFails: false},
		{name: "long_csv", field: "tools", value: "Read, Write, Edit, Grep, Glob, Bash, TodoWrite, Skill, mcp__sequential-thinking__sequentialthinking", wantFails: false},
		{name: "yaml_array", field: "tools", value: "[Read, Write]", wantFails: true},
		{name: "yaml_array_with_space", field: "tools", value: "[ Read, Write ]", wantFails: true},
		{name: "space_separated", field: "tools", value: "Read Write Edit", wantFails: true},
		{name: "two_tokens_space", field: "disallowedTools", value: "Bash Write", wantFails: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			msg := validateToolsCSVFormat(tc.field, tc.value)
			gotFails := msg != ""
			if gotFails != tc.wantFails {
				t.Errorf("validateToolsCSVFormat(%q, %q) gotFails=%v (msg=%q), want %v",
					tc.field, tc.value, gotFails, msg, tc.wantFails)
			}
		})
	}
}
