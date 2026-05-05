package template

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// @MX:NOTE: [AUTO] agents_frontmatter_test.go — REQ-CC2122-HOOK-002 정적 linter
// @MX:SPEC: SPEC-CC2122-HOOK-002 REQ-004/005/006
//
// 이 테스트는 .claude/agents/**/*.md 의 frontmatter 가 다음 규칙을 준수하는지
// 정적으로 검증한다:
//   1. tools 와 disallowedTools 는 mutually exclusive (claude-code v2.1.119+ 사양)
//   2. disallowedTools 가 정의된 경우 CSV 문자열 형식이어야 함 (공백 구분 금지)
//   3. tools 가 정의된 경우 CSV 문자열 형식이어야 함 (CLAUDE.local.md §12)
//
// 실제 `claude --print` 회귀 테스트는 GLM 통합 테스트 정책(CLAUDE.local.md §13)
// 으로 별도 manual 검증 분리.

// validateToolsCSVFormat는 tools 또는 disallowedTools 필드 값이 CSV 형식인지 검증한다.
// 빈 값은 허용 (필드 부재로 간주).
// YAML 배열 syntax([...]) 는 금지.
// 공백만으로 분리된 다중 토큰은 금지 (콤마 분리 필수).
//
// 반환: 위반 시 에러 메시지, 정상 시 빈 문자열.
func validateToolsCSVFormat(field, value string) string {
	if value == "" {
		return ""
	}

	trimmed := strings.TrimSpace(value)

	// YAML 배열 syntax 금지: "[A, B, C]" 형태는 frontmatter parser 가 그대로 문자열로 캡처함.
	if strings.HasPrefix(trimmed, "[") {
		return fmt.Sprintf("%s 값이 YAML 배열 syntax 로 시작함 (CSV 문자열 필수): %q", field, value)
	}

	// 콤마 없이 공백 구분된 다중 토큰은 위반 (예: "Read Write Edit").
	// 단일 토큰(예: "Read")은 허용.
	hasComma := strings.Contains(trimmed, ",")
	hasMultipleTokens := len(strings.Fields(trimmed)) > 1
	if hasMultipleTokens && !hasComma {
		return fmt.Sprintf("%s 값이 공백 구분으로 보임 (CSV 형식 필수, 콤마 분리): %q", field, value)
	}

	return ""
}

// collectAgentFiles는 embedded templates 에서 .claude/agents/ 하위 모든 .md 파일을 수집한다.
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

			// retired:true 에이전트는 tools: [] / skills: [] 빈 배열이 허용된다.
			// TestAgentFrontmatterAudit에서 별도로 검증하므로 여기서는 skip.
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

// TestValidateToolsCSVFormat_Cases는 validateToolsCSVFormat 헬퍼 자체가 정확히
// 동작하는지 단위 검증한다. 향후 frontmatter validator 가 강화될 때 회귀 방지 목적.
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
