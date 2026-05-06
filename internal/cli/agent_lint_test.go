package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper to create a temporary agent file
func createTempAgentFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "test-agent.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp agent file: %v", err)
	}

	return agentPath
}

// Test LR-01: Literal AskUserQuestion in body text
func TestCheckLiteralAskUserQuestion(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantCount   int
		wantMessage string
	}{
		{
			name: "AskUserQuestion in body",
			content: `---
name: test
---
Some text
AskUserQuestion should not be here
More text`,
			wantCount:   1,
			wantMessage: "Literal AskUserQuestion found",
		},
		{
			name: "AskUserQuestion in code block should be ignored",
			content: `---
name: test
---
Some text
` + "```" + `
AskUserQuestion is allowed here
` + "```" + `
More text`,
			wantCount:   0,
			wantMessage: "",
		},
		{
			name: "Multiple AskUserQuestion occurrences",
			content: `---
name: test
---
AskUserQuestion here
` + "```" + `
AskUserQuestion in code block
` + "```" + `
AskUserQuestion there`,
			wantCount:   2,
			wantMessage: "",
		},
		{
			name: "No AskUserQuestion",
			content: `---
name: test
---
Clean content without issues`,
			wantCount:   0,
			wantMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTempAgentFile(t, tt.content)
			content, _ := os.ReadFile(path)
			parts := strings.SplitN(string(content), "---", 3)
			if len(parts) < 3 {
				t.Fatal("invalid test content")
			}

			violations := checkLiteralAskUserQuestion(path, []byte(parts[2]))

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}

			if tt.wantCount > 0 && len(violations) > 0 {
				if !strings.Contains(violations[0].Message, tt.wantMessage) {
					t.Errorf("message = %q, want to contain %q", violations[0].Message, tt.wantMessage)
				}
			}
		})
	}
}

// Test LR-02: Agent in tools list
func TestCheckAgentInTools(t *testing.T) {
	tests := []struct {
		name      string
		frontmatter string
		wantCount int
	}{
		{
			name: `tools: "Agent" should fail`,
			frontmatter: `---
name: test
tools: Read, Write, Agent
---
`,
			wantCount: 1,
		},
		{
			name: `tools: "Read, Write" should pass`,
			frontmatter: `---
name: test
tools: Read, Write
---
`,
			wantCount: 0,
		},
		{
			name: "empty tools should pass",
			frontmatter: `---
name: test
---
`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTempAgentFile(t, tt.frontmatter)
			content, _ := os.ReadFile(path)
			parts := strings.SplitN(string(content), "---", 3)

			var fm AgentFrontmatter
			_ = parseYAMLFrontmatter([]byte(parts[1]), &fm)
			fm.Tools = extractToolsFromFrontmatter(tt.frontmatter)

			violations := checkAgentInTools(path, fm)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}
		})
	}
}

// Helper to extract tools from frontmatter string
func extractToolsFromFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "tools:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// Test LR-03: Missing effort field
func TestCheckMissingEffort(t *testing.T) {
	tests := []struct {
		name      string
		frontmatter string
		wantCount int
		wantSev   LintSeverity
	}{
		{
			name: "missing effort",
			frontmatter: `---
name: test
---
`,
			wantCount: 1,
			wantSev:   SeverityError,
		},
		{
			name: "effort present",
			frontmatter: `---
name: test
effort: high
---
`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTempAgentFile(t, tt.frontmatter)
			content, _ := os.ReadFile(path)
			parts := strings.SplitN(string(content), "---", 3)

			var fm AgentFrontmatter
			_ = parseYAMLFrontmatter([]byte(parts[1]), &fm)
			fm.Effort = extractEffortFromFrontmatter(tt.frontmatter)

			violations := checkMissingEffort(path, fm)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}

			if tt.wantCount > 0 && violations[0].Severity != tt.wantSev {
				t.Errorf("severity = %v, want %v", violations[0].Severity, tt.wantSev)
			}
		})
	}
}

// Helper to extract effort from frontmatter string
func extractEffortFromFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "effort:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// Test LR-05: Missing isolation for write-heavy agents
func TestCheckMissingIsolation(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		isolation string
		wantCount int
	}{
		{
			name:      "implementer without worktree",
			agentName: "my-implementer",
			isolation: "",
			wantCount: 1,
		},
		{
			name:      "implementer with worktree",
			agentName: "my-implementer",
			isolation: "worktree",
			wantCount: 0,
		},
		{
			name:      "tester without worktree",
			agentName: "test-specialist",
			isolation: "",
			wantCount: 1,
		},
		{
			name:      "designer without worktree",
			agentName: "ui-designer",
			isolation: "",
			wantCount: 1,
		},
		{
			name:      "non-role agent",
			agentName: "manager-spec",
			isolation: "",
			wantCount: 0,
		},
		{
			name:      "write-heavy expert-backend without worktree",
			agentName: "expert-backend",
			isolation: "",
			wantCount: 1,
		},
		{
			name:      "write-heavy expert-backend with worktree",
			agentName: "expert-backend",
			isolation: "worktree",
			wantCount: 0,
		},
		{
			name:      "write-heavy researcher without worktree",
			agentName: "researcher",
			isolation: "",
			wantCount: 1,
		},
		{
			name:      "write-heavy manager-cycle without worktree",
			agentName: "manager-cycle",
			isolation: "",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := AgentFrontmatter{
				Name:      tt.agentName,
				Isolation: tt.isolation,
			}

			violations := checkMissingIsolation("test.md", fm)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}
			if len(violations) > 0 && violations[0].Severity != SeverityError {
				t.Errorf("expected error severity, got %s", violations[0].Severity)
			}
		})
	}
}

// Test LR-09: isolation: worktree on read-only agents
func TestCheckReadOnlyIsolation(t *testing.T) {
	tests := []struct {
		name           string
		permissionMode string
		isolation      string
		wantCount      int
	}{
		{
			name:           "plan mode with worktree",
			permissionMode: "plan",
			isolation:      "worktree",
			wantCount:      1,
		},
		{
			name:           "plan mode without worktree",
			permissionMode: "plan",
			isolation:      "",
			wantCount:      0,
		},
		{
			name:           "acceptEdits mode with worktree",
			permissionMode: "acceptEdits",
			isolation:      "worktree",
			wantCount:      0,
		},
		{
			name:           "default mode with worktree",
			permissionMode: "",
			isolation:      "worktree",
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := AgentFrontmatter{
				PermissionMode: tt.permissionMode,
				Isolation:      tt.isolation,
			}

			violations := checkReadOnlyIsolation("test.md", fm)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}
			if len(violations) > 0 && violations[0].Rule != "LR-09" {
				t.Errorf("expected LR-09 rule, got %s", violations[0].Rule)
			}
		})
	}
}

// Test LR-06: Deepthink boilerplate
func TestCheckDeepthinkBoilerplate(t *testing.T) {
	tests := []struct {
		name      string
		frontmatter string
		strict    bool
		wantCount int
	}{
		{
			name: "deepthink boilerplate present",
			frontmatter: `---
name: test
description: |
  Some text
  --deepthink flag: Activate when needed
---
`,
			strict:    false,
			wantCount: 1,
		},
		{
			name: "no deepthink boilerplate",
			frontmatter: `---
name: test
description: |
  Some text
---
`,
			strict:    false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTempAgentFile(t, tt.frontmatter)
			content, _ := os.ReadFile(path)
			parts := strings.SplitN(string(content), "---", 3)

			violations := checkDeepthinkBoilerplate(path, []byte(parts[1]), tt.strict)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}
		})
	}
}

// Test LR-07: Duplicate mandate blocks
func TestCheckDuplicateMandateBlocks(t *testing.T) {
	t.Run("single mandate block should pass", func(t *testing.T) {
		content := `---
name: test
---
## Evaluation
- Evaluate code quality
- Score readability
- Assess performance`

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "agent1.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write agent file: %v", err)
		}

		violations := checkDuplicateMandateBlocks([]string{path})

		if len(violations) != 0 {
			t.Errorf("got %d violations, want 0", len(violations))
		}
	})

	t.Run("duplicate mandate blocks should fail", func(t *testing.T) {
		content1 := `---
name: test1
---
## Evaluation
- Evaluate code quality
- Score readability
- Assess performance`

		content2 := `---
name: test2
---
## Evaluation
- Evaluate security
- Score robustness
- Assess scalability`

		tmpDir := t.TempDir()
		path1 := filepath.Join(tmpDir, "agent1.md")
		path2 := filepath.Join(tmpDir, "agent2.md")
		if err := os.WriteFile(path1, []byte(content1), 0644); err != nil {
			t.Fatalf("write agent file 1: %v", err)
		}
		if err := os.WriteFile(path2, []byte(content2), 0644); err != nil {
			t.Fatalf("write agent file 2: %v", err)
		}

		violations := checkDuplicateMandateBlocks([]string{path1, path2})

		if len(violations) != 1 {
			t.Errorf("got %d violations, want 1", len(violations))
		}

		if violations[0].Rule != "LR-07" {
			t.Errorf("rule = %s, want LR-07", violations[0].Rule)
		}
	})
}

// Test LR-08: Skill preload drift
func TestCheckSkillPreloadDrift(t *testing.T) {
	t.Run("consistent skill preloads should pass", func(t *testing.T) {
		content1 := `---
name: manager-spec
skills:
  - moai-foundation-core
---
`

		content2 := `---
name: manager-ddd
skills:
  - moai-foundation-core
---
`

		tmpDir := t.TempDir()
		path1 := filepath.Join(tmpDir, "agent1.md")
		path2 := filepath.Join(tmpDir, "agent2.md")
		if err := os.WriteFile(path1, []byte(content1), 0644); err != nil {
			t.Fatalf("write agent file 1: %v", err)
		}
		if err := os.WriteFile(path2, []byte(content2), 0644); err != nil {
			t.Fatalf("write agent file 2: %v", err)
		}

		violations := checkSkillPreloadDrift([]string{path1, path2})

		if len(violations) != 0 {
			t.Errorf("got %d violations, want 0", len(violations))
		}
	})

	t.Run("skill drift should warn", func(t *testing.T) {
		content1 := `---
name: manager-spec
skills:
  - moai-foundation-core
  - moai-workflow-spec
---
`

		content2 := `---
name: manager-ddd
skills:
  - moai-foundation-core
---
`

		tmpDir := t.TempDir()
		path1 := filepath.Join(tmpDir, "agent1.md")
		path2 := filepath.Join(tmpDir, "agent2.md")
		if err := os.WriteFile(path1, []byte(content1), 0644); err != nil {
			t.Fatalf("write agent file 1: %v", err)
		}
		if err := os.WriteFile(path2, []byte(content2), 0644); err != nil {
			t.Fatalf("write agent file 2: %v", err)
		}

		violations := checkSkillPreloadDrift([]string{path1, path2})

		if len(violations) != 1 {
			t.Errorf("got %d violations, want 1", len(violations))
		}

		if violations[0].Rule != "LR-08" {
			t.Errorf("rule = %s, want LR-08", violations[0].Rule)
		}
	})
}

// Test JSON output format
func TestAgentLintJSONOutput(t *testing.T) {
	content := `---
name: test
tools: Read, Write, Agent
effort: high
---
Some text
AskUserQuestion here`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// We can't easily run the full cobra command here,
	// but we can test the output structure
	violations, _ := lintAgentFile(path, false)

	output := LintOutput{
		Version: "1.0.0",
		Summary: LintSummary{
			Total:    len(violations),
			Errors:   2,
			Warnings: 0,
		},
		Violations: violations,
	}

	if output.Version != "1.0.0" {
		t.Errorf("version = %s, want 1.0.0", output.Version)
	}

	if len(output.Violations) != 2 {
		t.Errorf("got %d violations, want 2 (LR-01 and LR-02)", len(output.Violations))
	}
}

// Test malformed frontmatter handling
func TestAgentLintMalformedFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing closing delimiter",
			content: `---
name: test
Some content without closing delimiter`,
		},
		{
			name: "only one delimiter",
			content: `---name: test`,
		},
		{
			name: "empty file",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTempAgentFile(t, tt.content)
			_, err := lintAgentFile(path, false)

			if err == nil {
				t.Error("expected error for malformed frontmatter, got nil")
			}
		})
	}
}

// Test exit codes
func TestAgentLintExitCodes(t *testing.T) {
	t.Run("clean file should exit 0", func(t *testing.T) {
		content := `---
name: test
effort: high
tools: Read, Write, Edit
---
Clean content`

		path := createTempAgentFile(t, content)
		violations, _ := lintAgentFile(path, false)

		if len(violations) != 0 {
			t.Errorf("expected 0 violations, got %d", len(violations))
		}
	})

	t.Run("violations should be reported", func(t *testing.T) {
		content := `---
name: test
tools: Read, Write, Agent
---
AskUserQuestion here`

		path := createTempAgentFile(t, content)
		violations, _ := lintAgentFile(path, false)

		if len(violations) == 0 {
			t.Error("expected violations, got 0")
		}
	})
}

// Test fenced code block exemption for LR-01
func TestFencedCodeBlockExemption(t *testing.T) {
	content := `---
name: test
---
Text before
` + "```" + `
AskUserQuestion in code block
should be ignored
` + "```" + `
Text after
AskUserQuestion here should fail`

	path := createTempAgentFile(t, content)
	fileContent, _ := os.ReadFile(path)
	parts := strings.SplitN(string(fileContent), "---", 3)

	violations := checkLiteralAskUserQuestion(path, []byte(parts[2]))

	if len(violations) != 1 {
		t.Errorf("expected 1 violation (only outside code block), got %d", len(violations))
	}

	// Line 10 in the full file is "AskUserQuestion here should fail"
	if violations[0].Line != 10 {
		t.Errorf("line = %d, want 10", violations[0].Line)
	}
}

// Test LR-10: Static team-* agent file detection (SPEC-V3R2-ORC-005)
func TestCheckStaticTeamAgent(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		wantCount int
		wantRule  string
	}{
		{
			name:      "team-custom.md should be rejected",
			fileName:  "team-custom.md",
			wantCount: 1,
			wantRule:  "LR-10",
		},
		{
			name:      "team-implementer.md should be rejected",
			fileName:  "team-implementer.md",
			wantCount: 1,
			wantRule:  "LR-10",
		},
		{
			name:      "teamlead.md should NOT be rejected (no dash after team)",
			fileName:  "teamlead.md",
			wantCount: 0,
			wantRule:  "",
		},
		{
			name:      "expert-team.md should NOT be rejected (not starting with team-)",
			fileName:  "expert-team.md",
			wantCount: 0,
			wantRule:  "",
		},
		{
			name:      "manager-spec.md should NOT be rejected",
			fileName:  "manager-spec.md",
			wantCount: 0,
			wantRule:  "",
		},
		{
			name:      "team-designer-v2.md should be rejected",
			fileName:  "team-designer-v2.md",
			wantCount: 1,
			wantRule:  "LR-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := "---\nname: " + strings.TrimSuffix(tt.fileName, ".md") + "\n---\nbody"
			tmpDir := t.TempDir()
			agentPath := filepath.Join(tmpDir, tt.fileName)

			if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
				t.Fatalf("failed to create temp agent file: %v", err)
			}

			violations := checkStaticTeamAgent(agentPath)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
				return
			}

			if tt.wantCount > 0 {
				if violations[0].Rule != tt.wantRule {
					t.Errorf("rule = %s, want %s", violations[0].Rule, tt.wantRule)
				}
				if violations[0].Severity != SeverityError {
					t.Errorf("severity = %s, want %s", violations[0].Severity, SeverityError)
				}
				if !strings.Contains(violations[0].Message, "ORC_STATIC_TEAM_AGENT_PROHIBITED") {
					t.Errorf("message should contain ORC_STATIC_TEAM_AGENT_PROHIBITED, got: %s", violations[0].Message)
				}
				if violations[0].Line != 1 {
					t.Errorf("line = %d, want 1", violations[0].Line)
				}
			}
		})
	}
}

// Test LR-10 integration: lintAgentFile should detect static team-* agent
func TestLintAgentFile_StaticTeamAgent(t *testing.T) {
	content := `---
name: team-custom
tools: "Read,Grep,Glob"
effort: low
---
Custom team agent body`

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "team-custom.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	violations, err := lintAgentFile(agentPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, v := range violations {
		if v.Rule == "LR-10" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected LR-10 violation for team-custom.md, not found")
	}
}
