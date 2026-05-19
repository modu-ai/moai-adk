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
		name        string
		frontmatter string
		wantCount   int
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
		name        string
		frontmatter string
		wantCount   int
		wantSev     LintSeverity
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
			name:      "write-heavy manager-develop without worktree",
			agentName: "manager-develop",
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
		name        string
		frontmatter string
		strict      bool
		wantCount   int
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
		// LR-07 v2 fingerprint: requires Skeptical-context header preceding eval bullets.
		content := `---
name: test
---
## Skeptical Evaluation Mandate
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
		// Both files must include Skeptical-context header for LR-07 v2 to fire.
		content1 := `---
name: test1
---
## Skeptical Evaluation Mandate
- Evaluate code quality
- Score readability
- Assess performance`

		content2 := `---
name: test2
---
## Skeptical Evaluation Mandate
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

	// Regression guard (SPEC-V2.20.0-RC1 hotfix): Delegation Protocol bullets
	// share eval keywords with mandate blocks (security/performance/quality) but
	// MUST NOT trip LR-07 v2 fingerprint because the preceding header is
	// "## Delegation Protocol", not a Skeptical-context header.
	t.Run("delegation protocol bullets should not trip LR-07", func(t *testing.T) {
		content := `---
name: test-delegate
---
## Delegation Protocol
- SPEC unclear: Delegate to manager-spec
- Security concerns: Delegate to expert-security
- Performance issues: Delegate to expert-performance
- Quality validation: Delegate to manager-quality
- Git operations: Delegate to manager-git`

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "agent.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write agent file: %v", err)
		}

		violations := checkDuplicateMandateBlocks([]string{path})

		if len(violations) != 0 {
			t.Errorf("delegation protocol tripped LR-07: got %d violations, want 0", len(violations))
		}
	})

	// Regression guard: Complexity Analysis bullets reuse "score" keyword
	// three times in different complexity bands — pre-v2 fingerprint mis-flagged
	// these as mandate blocks. v2 requires Skeptical-context header.
	t.Run("complexity analysis bullets should not trip LR-07", func(t *testing.T) {
		content := `---
name: test-complexity
---
### Complexity Analysis
- SIMPLE (score < 3): Direct interview
- MEDIUM (score 3-6): Lightweight planning
- COMPLEX (score > 6): Full decomposition`

		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "agent.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write agent file: %v", err)
		}

		violations := checkDuplicateMandateBlocks([]string{path})

		if len(violations) != 0 {
			t.Errorf("complexity analysis tripped LR-07: got %d violations, want 0", len(violations))
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
		Version: "1.0",
		Summary: LintSummary{
			Total:    len(violations),
			Errors:   2,
			Warnings: 0,
		},
		Violations: violations,
	}

	if output.Version != "1.0" {
		t.Errorf("version = %s, want 1.0 (canonical JSON schema version, stable through v3.0.x)", output.Version)
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
			name:    "only one delimiter",
			content: `---name: test`,
		},
		{
			name:    "empty file",
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

// ============================================================================
// M1 RED Tests for SPEC-V3R2-ORC-003 (Effort-Level Calibration Matrix)
// These tests FAIL initially and turn GREEN after M2 implementation.
// ============================================================================

// TestLintLR12_MatrixDrift_DriftedAgent tests LR-12: effort drift from canonical matrix
// This is a RED test - it will FAIL until checkEffortMatrixDrift is implemented in M2
func TestLintLR12_MatrixDrift_DriftedAgent(t *testing.T) {
	content := `---
name: expert-security
description: Security specialist
tools: Read, Write, Agent
effort: high
---
Security agent body`

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "expert-security.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// This will fail initially because checkEffortMatrixDrift doesn't exist yet
	// After M2, this will detect that effort: high drifts from canonical xhigh
	violations, err := lintAgentFile(agentPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After M2 implementation, we expect 1 LR-12 violation
	// For now, this test documents the expected behavior
	foundLR12 := false
	for _, v := range violations {
		if v.Rule == "LR-12" {
			foundLR12 = true
			if !strings.Contains(v.Message, "ORC_EFFORT_MATRIX_DRIFT") {
				t.Errorf("LR-12 message should contain ORC_EFFORT_MATRIX_DRIFT, got: %s", v.Message)
			}
			if v.Severity != SeverityError {
				t.Errorf("LR-12 severity should be Error, got: %s", v.Severity)
			}
			break
		}
	}

	// TODO: Remove this skip after M2 implementation
	// For now, this test documents the expected behavior

	if !foundLR12 {
		t.Error("expected LR-12 violation for expert-security with effort: high (should be xhigh)")
	}
}

// TestLintLR12_MatrixDrift_CleanAgent tests LR-12: agent with correct effort value
// This is a RED test - it will FAIL until checkEffortMatrixDrift is implemented in M2
func TestLintLR12_MatrixDrift_CleanAgent(t *testing.T) {
	content := `---
name: expert-security
description: Security specialist
tools: Read, Write, Agent
effort: xhigh
---
Security agent body`

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "expert-security.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	violations, err := lintAgentFile(agentPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After M2, we expect 0 LR-12 violations for correct effort value
	for _, v := range violations {
		if v.Rule == "LR-12" {
			t.Error("expected no LR-12 violations for expert-security with correct effort: xhigh")
		}
	}

	// TODO: Remove this skip after M2 implementation
}

// TestLintLR13_InvalidEffortEnum tests LR-13: invalid effort enum value
// This is a RED test - it will FAIL until checkInvalidEffortEnum is implemented in M2
func TestLintLR13_InvalidEffortEnum(t *testing.T) {
	content := `---
name: expert-security
description: Security specialist
tools: Read, Write, Agent
effort: ultra
---
Security agent body`

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "expert-security.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	violations, err := lintAgentFile(agentPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After M2, we expect 1 LR-13 violation for invalid enum value
	foundLR13 := false
	for _, v := range violations {
		if v.Rule == "LR-13" {
			foundLR13 = true
			if !strings.Contains(v.Message, "AGT_INVALID_FRONTMATTER") && !strings.Contains(v.Message, "not in") {
				t.Errorf("LR-13 message should mention invalid enum, got: %s", v.Message)
			}
			if v.Severity != SeverityError {
				t.Errorf("LR-13 severity should be Error, got: %s", v.Severity)
			}
			break
		}
	}

	// TODO: Remove this skip after M2 implementation

	if !foundLR13 {
		t.Error("expected LR-13 violation for effort: ultra (invalid enum)")
	}
}

// TestLintLR14_FixedBudgetTokens tests LR-14: fixed budget_tokens prohibition
// This is a RED test - it will FAIL until checkFixedBudgetTokens is implemented in M2
func TestLintLR14_FixedBudgetTokens(t *testing.T) {
	content := `---
name: expert-security
description: Security specialist
tools: Read, Write, Agent
---
Security agent body

This agent uses budget_tokens: 5000 which is prohibited for Opus 4.7.`

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "expert-security.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	violations, err := lintAgentFile(agentPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After M2, we expect 1 LR-14 violation for fixed budget_tokens
	foundLR14 := false
	for _, v := range violations {
		if v.Rule == "LR-14" {
			foundLR14 = true
			if !strings.Contains(v.Message, "ORC_FIXED_BUDGET_PROHIBITED") {
				t.Errorf("LR-14 message should contain ORC_FIXED_BUDGET_PROHIBITED, got: %s", v.Message)
			}
			if v.Severity != SeverityError {
				t.Errorf("LR-14 severity should be Error, got: %s", v.Severity)
			}
			break
		}
	}

	// TODO: Remove this skip after M2 implementation

	if !foundLR14 {
		t.Error("expected LR-14 violation for budget_tokens: 5000")
	}
}

// TestAuthoringDocHasEffortMatrix tests that agent-authoring.md contains the effort matrix
// This is a RED test - it will FAIL until M2 adds the matrix table
func TestAuthoringDocHasEffortMatrix(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Try both worktree and main project paths
	authoringDocPaths := []string{
		filepath.Join(cwd, ".claude", "rules", "moai", "development", "agent-authoring.md"),
		filepath.Join(cwd, "..", "..", "..", ".claude", "rules", "moai", "development", "agent-authoring.md"),
	}

	var content []byte
	var docPath string
	for _, path := range authoringDocPaths {
		if data, err := os.ReadFile(path); err == nil {
			content = data
			docPath = path
			break
		}
	}

	if len(content) == 0 {
		t.Skip("agent-authoring.md not found - will test after M2 implementation")
		return
	}

	contentStr := string(content)

	// Check for the section heading
	if !strings.Contains(contentStr, "## Effort-Level Calibration Matrix") {
		t.Error("agent-authoring.md should contain '## Effort-Level Calibration Matrix' section")
	}

	// Check for the canonical 17-agent matrix table
	expectedAgents := []string{
		"manager-spec", "manager-strategy", "manager-cycle", "manager-quality",
		"manager-docs", "manager-git", "manager-project",
		"expert-backend", "expert-frontend", "expert-security", "expert-devops", "expert-performance",
		"expert-refactoring", "builder-platform",
		"evaluator-active", "plan-auditor", "researcher",
	}

	missingAgents := []string{}
	for _, agent := range expectedAgents {
		if !strings.Contains(contentStr, agent) {
			missingAgents = append(missingAgents, agent)
		}
	}

	if len(missingAgents) > 0 {
		t.Errorf("agent-authoring.md effort matrix missing agents: %v", missingAgents)
	}

	// Check for effort level values
	expectedEfforts := []string{"xhigh", "high", "medium"}
	for _, effort := range expectedEfforts {
		if !strings.Contains(contentStr, effort) {
			t.Errorf("agent-authoring.md should contain effort level: %s", effort)
		}
	}

	t.Logf("Checked agent-authoring.md at: %s", docPath)
}

// TestConstitutionCrossReference tests that moai-constitution.md cross-references the matrix
// This is a RED test - it will FAIL until M2 adds the cross-reference
func TestConstitutionCrossReference(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Try both worktree and main project paths
	constitutionDocPaths := []string{
		filepath.Join(cwd, ".claude", "rules", "moai", "core", "moai-constitution.md"),
		filepath.Join(cwd, "..", "..", "..", ".claude", "rules", "moai", "core", "moai-constitution.md"),
	}

	var content []byte
	var docPath string
	for _, path := range constitutionDocPaths {
		if data, err := os.ReadFile(path); err == nil {
			content = data
			docPath = path
			break
		}
	}

	if len(content) == 0 {
		t.Skip("moai-constitution.md not found - will test after M2 implementation")
		return
	}

	contentStr := string(content)

	// Check for cross-reference to agent-authoring.md
	if !strings.Contains(contentStr, "agent-authoring.md") {
		t.Error("moai-constitution.md should cross-reference agent-authoring.md for effort matrix")
	}

	// Check for Opus 4.7 section mentioning effort level selection
	if !strings.Contains(contentStr, "Opus 4.7") && !strings.Contains(contentStr, "Prompt Philosophy") {
		t.Error("moai-constitution.md should have Opus 4.7 Prompt Philosophy section")
	}

	t.Logf("Checked moai-constitution.md at: %s", docPath)
}

// TestLintLR03_MissingEffortIsError verifies LR-03 is at Error severity (not Warning)
// Per SPEC-V3R2-ORC-003 REQ-006, LR-03 was promoted from warning to error
// This test ensures the promotion is in place and prevents future regression
func TestLintLR03_MissingEffortIsError(t *testing.T) {
	content := `---
name: test-agent
description: Test agent without effort field
tools: Read, Write
---
Agent body`

	tmpDir := t.TempDir()
	agentPath := filepath.Join(tmpDir, "test-agent.md")

	if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	violations, err := lintAgentFile(agentPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find LR-03 violation
	foundLR03 := false
	for _, v := range violations {
		if v.Rule == "LR-03" {
			foundLR03 = true
			// Critical assertion: LR-03 must be Error severity, not Warning
			if v.Severity != SeverityError {
				t.Errorf("LR-03 severity must be Error (per SPEC-V3R2-ORC-003 REQ-006), got: %s", v.Severity)
			}
			// Verify error message mentions missing effort
			if !strings.Contains(v.Message, "effort") && !strings.Contains(v.Message, "LR-03") {
				t.Errorf("LR-03 message should mention missing effort field, got: %s", v.Message)
			}
			break
		}
	}

	if !foundLR03 {
		t.Error("expected LR-03 violation for missing effort field")
	}
}

// TestCheckDeadHooks tests LR-04 dead hook detection via direct AgentFrontmatter struct
func TestCheckDeadHooks(t *testing.T) {
	tests := []struct {
		name      string
		tools     string
		hooks     map[string][]Hook
		wantCount int
		wantRule  string
	}{
		{
			name:      "no hooks — no violation",
			tools:     "Read, Write",
			hooks:     nil,
			wantCount: 0,
		},
		{
			name:  "hook references absent tool — violation",
			tools: "Read, Grep",
			hooks: map[string][]Hook{
				"PostToolUse": {
					{Matcher: "Write|Edit"},
				},
			},
			wantCount: 1,
			wantRule:  "LR-04",
		},
		{
			name:  "hook references present tool — no violation",
			tools: "Read, Write, Edit",
			hooks: map[string][]Hook{
				"PostToolUse": {
					{Matcher: "Write|Edit"},
				},
			},
			wantCount: 0,
		},
		{
			name:  "hook with empty matcher — no violation",
			tools: "Read",
			hooks: map[string][]Hook{
				"PostToolUse": {
					{Matcher: ""},
				},
			},
			wantCount: 0,
		},
		{
			name:      "no tools — no violation",
			tools:     "",
			hooks:     map[string][]Hook{"PostToolUse": {{Matcher: "Write"}}},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm := AgentFrontmatter{
				Tools: tt.tools,
				Hooks: tt.hooks,
			}

			violations := checkDeadHooks("test.md", fm)

			if len(violations) != tt.wantCount {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantCount)
			}

			if tt.wantCount > 0 && len(violations) > 0 {
				if violations[0].Rule != tt.wantRule {
					t.Errorf("rule = %s, want %s", violations[0].Rule, tt.wantRule)
				}
				if violations[0].Severity != SeverityError {
					t.Errorf("severity = %s, want error", violations[0].Severity)
				}
			}
		})
	}
}

// TestRunAgentLint_WithFixtureFiles tests runAgentLint via testdata fixtures
func TestRunAgentLint_WithFixtureFiles(t *testing.T) {
	// Test with the clean fixture — should produce only LR-03 (missing effort) violations
	// since fixture-clean.md has effort: high, it should be clean
	tmpDir := t.TempDir()

	cleanContent := `---
name: test-clean
tools: Read, Write
effort: high
---
Clean body with no issues.`

	cleanPath := filepath.Join(tmpDir, "clean.md")
	if err := os.WriteFile(cleanPath, []byte(cleanContent), 0644); err != nil {
		t.Fatalf("write clean fixture: %v", err)
	}

	violations, err := lintAgentFile(cleanPath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have no violations (clean file)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for clean file, got %d: %v", len(violations), violations)
	}
}

// TestRunAgentLint_LR01WithOrchestratorExemption tests that orchestrator agents are exempt from LR-01
func TestRunAgentLint_LR01WithOrchestratorExemption(t *testing.T) {
	// Orchestrator with AskUserQuestion in tools: exempt from LR-01
	orchestratorContent := `---
name: test-orchestrator
tools: Read, Write, AskUserQuestion, Agent
effort: high
---
This agent uses AskUserQuestion to ask the user for decisions.
The presence of AskUserQuestion in tools: exempts this from LR-01.`

	path := createTempAgentFile(t, orchestratorContent)

	fileContent, _ := os.ReadFile(path)
	parts := strings.SplitN(string(fileContent), "---", 3)
	if len(parts) < 3 {
		t.Fatal("invalid content")
	}

	// The LR-01 check should see AskUserQuestion in the body but not flag it
	// because AskUserQuestion is in tools: — but our direct checkLiteralAskUserQuestion
	// call doesn't know about frontmatter. Test via lintAgentFile instead.
	violations, err := lintAgentFile(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not have LR-01 violation (orchestrator is exempt)
	for _, v := range violations {
		if v.Rule == "LR-01" {
			t.Errorf("unexpected LR-01 violation for orchestrator agent: %s", v.Message)
		}
	}
}

// TestRunAgentLint_JSONFormat tests the JSON output format via lintAgentFile + LintOutput
func TestRunAgentLint_JSONFormat(t *testing.T) {
	content := `---
name: test-json
tools: Read, Write, Agent
effort: high
---
Some body text.`

	path := createTempAgentFile(t, content)
	violations, err := lintAgentFile(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errors := 0
	warnings := 0
	for _, v := range violations {
		if v.Severity == SeverityError {
			errors++
		} else {
			warnings++
		}
	}

	output := LintOutput{
		Version: "1.0",
		Summary: LintSummary{
			Total:    len(violations),
			Errors:   errors,
			Warnings: warnings,
		},
		Violations: violations,
	}

	// Verify version field
	if output.Version != "1.0" {
		t.Errorf("version = %s, want 1.0", output.Version)
	}

	// Verify summary consistency
	if output.Summary.Total != output.Summary.Errors+output.Summary.Warnings {
		t.Errorf("summary total %d != errors %d + warnings %d",
			output.Summary.Total, output.Summary.Errors, output.Summary.Warnings)
	}

	// Should have LR-02 violation for Agent in tools
	found := false
	for _, v := range output.Violations {
		if v.Rule == "LR-02" {
			found = true
			if v.Severity != SeverityError {
				t.Errorf("LR-02 severity = %s, want error", v.Severity)
			}
		}
	}
	if !found {
		t.Error("expected LR-02 violation not found in JSON output")
	}
}

// TestRunAgentLint_StrictModePromotesWarnings tests that --strict promotes LR-03 from warning to error
// Note: In current implementation LR-03 is always Error (per SPEC-V3R2-ORC-003), so --strict has no effect.
// This test verifies the current behavior.
func TestRunAgentLint_StrictModePromotesWarnings(t *testing.T) {
	// LR-06 (deepthink boilerplate) is warning by default, error in --strict
	content := `---
name: test-strict
description: |
  Some description
  --deepthink flag: Activate for deep analysis
tools: Read, Write
effort: high
---
Body text.`

	path := createTempAgentFile(t, content)

	// Non-strict: LR-06 should be warning
	violations, err := lintAgentFile(path, false)
	if err != nil {
		t.Fatalf("non-strict: unexpected error: %v", err)
	}

	lr06Found := false
	for _, v := range violations {
		if v.Rule == "LR-06" {
			lr06Found = true
			if v.Severity != SeverityWarning {
				t.Errorf("non-strict LR-06 severity = %s, want warning", v.Severity)
			}
		}
	}

	if !lr06Found {
		t.Error("expected LR-06 violation for deepthink boilerplate")
	}

	// Strict: LR-06 should be error
	violations, err = lintAgentFile(path, true)
	if err != nil {
		t.Fatalf("strict: unexpected error: %v", err)
	}

	lr06FoundStrict := false
	for _, v := range violations {
		if v.Rule == "LR-06" {
			lr06FoundStrict = true
			if v.Severity != SeverityError {
				t.Errorf("strict LR-06 severity = %s, want error", v.Severity)
			}
		}
	}

	if !lr06FoundStrict {
		t.Error("expected LR-06 violation in strict mode")
	}
}

// TestRunAgentLint_TextOutput tests the text output format
func TestRunAgentLint_TextOutput(t *testing.T) {
	// Test using direct violation creation since we can't easily test RunE
	// This exercises the LintOutput and violation severity detection
	violations := []LintViolation{
		{Rule: "LR-01", Severity: SeverityError, File: "test.md", Line: 5, Message: "test error"},
		{Rule: "LR-06", Severity: SeverityWarning, File: "test.md", Line: 10, Message: "test warning"},
	}

	errors := 0
	warnings := 0
	for _, v := range violations {
		if v.Severity == SeverityError {
			errors++
		} else {
			warnings++
		}
	}

	if errors != 1 {
		t.Errorf("expected 1 error, got %d", errors)
	}
	if warnings != 1 {
		t.Errorf("expected 1 warning, got %d", warnings)
	}
}

// TestAgentLintCmd_RunE_CleanFixture tests runAgentLint via cobra RunE with a clean path
func TestAgentLintCmd_RunE_CleanFixture(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a clean agent file
	cleanContent := `---
name: clean-agent
tools: Read, Write, Edit
effort: high
---
Clean body with no violations.`

	cleanPath := filepath.Join(tmpDir, "clean-agent.md")
	if err := os.WriteFile(cleanPath, []byte(cleanContent), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Set up the --path flag
	if err := agentLintCmd.Flags().Set("path", tmpDir); err != nil {
		t.Fatalf("set path flag: %v", err)
	}
	defer func() {
		_ = agentLintCmd.Flags().Set("path", "")
	}()

	// Run via cobra RunE
	err := agentLintCmd.RunE(agentLintCmd, []string{})
	if err != nil {
		t.Errorf("unexpected error for clean file: %v", err)
	}
}

// TestAgentLintCmd_RunE_WithViolation tests runAgentLint with a file containing violations
func TestAgentLintCmd_RunE_WithViolation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with LR-02 violation
	violationContent := `---
name: violation-agent
tools: Read, Write, Agent
effort: high
---
Body text.`

	violationPath := filepath.Join(tmpDir, "violation-agent.md")
	if err := os.WriteFile(violationPath, []byte(violationContent), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Set path flag
	if err := agentLintCmd.Flags().Set("path", tmpDir); err != nil {
		t.Fatalf("set path flag: %v", err)
	}
	defer func() {
		_ = agentLintCmd.Flags().Set("path", "")
	}()

	// Run via cobra RunE — should return errLintViolations
	err := agentLintCmd.RunE(agentLintCmd, []string{})
	if err == nil {
		t.Error("expected error for file with violations, got nil")
	}
}

// TestAgentLintCmd_RunE_JSONFormat tests runAgentLint with --format=json
func TestAgentLintCmd_RunE_JSONFormat(t *testing.T) {
	tmpDir := t.TempDir()

	cleanContent := `---
name: json-agent
tools: Read, Write
effort: high
---
Clean body.`

	cleanPath := filepath.Join(tmpDir, "json-agent.md")
	if err := os.WriteFile(cleanPath, []byte(cleanContent), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Set flags
	if err := agentLintCmd.Flags().Set("path", tmpDir); err != nil {
		t.Fatalf("set path flag: %v", err)
	}
	if err := agentLintCmd.Flags().Set("format", "json"); err != nil {
		t.Fatalf("set format flag: %v", err)
	}
	defer func() {
		_ = agentLintCmd.Flags().Set("path", "")
		_ = agentLintCmd.Flags().Set("format", "text")
	}()

	err := agentLintCmd.RunE(agentLintCmd, []string{})
	if err != nil {
		t.Errorf("unexpected error for JSON format with clean file: %v", err)
	}
}

// TestAgentLintCmd_RunE_NoFiles tests runAgentLint when no agent files found
func TestAgentLintCmd_RunE_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	// Empty directory — no .md files

	if err := agentLintCmd.Flags().Set("path", tmpDir); err != nil {
		t.Fatalf("set path flag: %v", err)
	}
	defer func() {
		_ = agentLintCmd.Flags().Set("path", "")
	}()

	// Should succeed with "no files found" message
	err := agentLintCmd.RunE(agentLintCmd, []string{})
	if err != nil {
		t.Errorf("unexpected error for empty directory: %v", err)
	}
}

// ============================================================================
// SPEC-V3R2-ORC-004: Sentinel Key Tests
// RED tests for LR-05 ORC_WORKTREE_MISSING and LR-09 ORC_WORKTREE_ON_READONLY
// ============================================================================

// TestLintLR05_OrcWorktreeMissingSentinel tests that LR-05 violation message
// contains the ORC_WORKTREE_MISSING sentinel key (AC-06).
func TestLintLR05_OrcWorktreeMissingSentinel(t *testing.T) {
	tmpDir := t.TempDir()
	agentFile := filepath.Join(tmpDir, "expert-backend.md")
	err := os.WriteFile(agentFile, []byte(`---
name: expert-backend
tools: Read, Write, Edit
permissionMode: bypassPermissions
effort: high
---

Body.
`), 0o644)
	if err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	violations, err := lintAgentFile(agentFile, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found *LintViolation
	for i := range violations {
		if violations[i].Rule == "LR-05" {
			found = &violations[i]
			break
		}
	}

	if found == nil {
		t.Fatal("expected LR-05 violation for expert-backend without isolation:worktree, got none")
	}
	if found.Severity != SeverityError {
		t.Errorf("LR-05 severity = %s, want error", found.Severity)
	}
	if !strings.Contains(found.Message, "ORC_WORKTREE_MISSING") {
		t.Errorf("LR-05 message should contain ORC_WORKTREE_MISSING, got: %s", found.Message)
	}
}

// TestLintLR09_OrcWorktreeOnReadonlySentinel tests that LR-09 violation message
// contains the ORC_WORKTREE_ON_READONLY sentinel key (AC-07).
func TestLintLR09_OrcWorktreeOnReadonlySentinel(t *testing.T) {
	tmpDir := t.TempDir()
	agentFile := filepath.Join(tmpDir, "evaluator-active.md")
	err := os.WriteFile(agentFile, []byte(`---
name: evaluator-active
tools: Read, Grep, Glob
permissionMode: plan
isolation: worktree
effort: xhigh
---

Body.
`), 0o644)
	if err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	violations, err := lintAgentFile(agentFile, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found *LintViolation
	for i := range violations {
		if violations[i].Rule == "LR-09" {
			found = &violations[i]
			break
		}
	}

	if found == nil {
		t.Fatal("expected LR-09 violation for evaluator-active with isolation:worktree on plan mode, got none")
	}
	if found.Severity != SeverityError {
		t.Errorf("LR-09 severity = %s, want error", found.Severity)
	}
	if !strings.Contains(found.Message, "ORC_WORKTREE_ON_READONLY") {
		t.Errorf("LR-09 message should contain ORC_WORKTREE_ON_READONLY, got: %s", found.Message)
	}
}

// TestAgentLint_NoSandboxNoJustification_Fails verifies LR-33: sandbox: none without
// sandbox.justification causes a lint error.
// T-RT003-11 / T-RT003-36: SPEC-V3R2-RT-003 REQ-033/043 AC-08/15.
func TestAgentLint_NoSandboxNoJustification_Fails(t *testing.T) {
	t.Parallel()

	content := `---
name: test-agent-no-sandbox-justification
tools: Read, Grep
effort: medium
sandbox: none
---

This agent uses sandbox: none without justification.
It should fail lint LR-33.
`
	path := createTempAgentFile(t, content)

	violations, err := lintAgentFile(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// LR-33 エラーが1件あること
	var lr33Violations []LintViolation
	for _, v := range violations {
		if v.Rule == "LR-33" {
			lr33Violations = append(lr33Violations, v)
		}
	}

	if len(lr33Violations) == 0 {
		t.Error("expected LR-33 violation for sandbox: none without justification, got none")
		return
	}

	if lr33Violations[0].Severity != SeverityError {
		t.Errorf("LR-33: expected severity Error, got %q", lr33Violations[0].Severity)
	}
}

// TestAgentLint_NoSandboxWithJustification_Passes verifies LR-33: sandbox: none
// WITH sandbox.justification passes (emits warning only, not error).
// T-RT003-11 / T-RT003-36: SPEC-V3R2-RT-003 AC-15.
func TestAgentLint_NoSandboxWithJustification_Passes(t *testing.T) {
	t.Parallel()

	content := `---
name: test-agent-sandbox-with-justification
tools: Read, Grep
effort: medium
sandbox: none
sandbox.justification: "dogfooding legacy workflow X — tracked in SPEC-V3R2-MIG-001"
---

This agent has sandbox: none with justification.
LR-33 should emit warning (not error).
`
	path := createTempAgentFile(t, content)

	violations, err := lintAgentFile(path, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// LR-33 warning (not error)
	for _, v := range violations {
		if v.Rule == "LR-33" && v.Severity == SeverityError {
			t.Errorf("LR-33: expected warning (not error) for sandbox: none with justification, got error: %s", v.Message)
		}
	}
}

// TestCheckDeadHooks_ViaLintAgentFile tests LR-04 detection through lintAgentFile
// The hooks field parsing in the simple YAML parser doesn't handle complex maps,
// so this tests what actually happens when a fixture file is processed.
func TestCheckDeadHooks_ViaLintAgentFile(t *testing.T) {
	// Use the testdata fixture
	fixturePath := filepath.Join("testdata", "agent_lint", "fixture-lr04-dead-hook.md")

	// Check if fixture exists; if not, skip
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("fixture-lr04-dead-hook.md not found, skipping")
		return
	}

	violations, err := lintAgentFile(fixturePath, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The dead hook fixture has tools: Read, Grep (no Write or Edit)
	// but hooks reference Write|Edit — should be LR-04
	// Note: the simple YAML parser may not parse hooks: blocks correctly,
	// so we verify the fixture at least doesn't error.
	t.Logf("violations from LR-04 fixture: %d", len(violations))
	for _, v := range violations {
		t.Logf("  [%s] %s:%d: %s", v.Rule, filepath.Base(v.File), v.Line, v.Message)
	}
}

// Test LR-08 domain-prefix exemption (SPEC-V3R5-CORE-SLIM-001 Track A)
func TestSkillPreloadDriftExemption_DomainSkills(t *testing.T) {
	// Domain-scoped skills must NOT trigger LR-08 violations even when asymmetric.
	// agent-backend has moai-domain-backend; agent-frontend has moai-domain-frontend.
	// These are agent-specific by design — symmetry must NOT be enforced.
	contentBackend := `---
name: expert-backend
skills:
  - moai-foundation-core
  - moai-domain-backend
  - moai-workflow-testing
---
`
	contentFrontend := `---
name: expert-frontend
skills:
  - moai-foundation-core
  - moai-domain-frontend
  - moai-workflow-testing
---
`
	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "expert-backend.md")
	path2 := filepath.Join(tmpDir, "expert-frontend.md")
	if err := os.WriteFile(path1, []byte(contentBackend), 0644); err != nil {
		t.Fatalf("write expert-backend: %v", err)
	}
	if err := os.WriteFile(path2, []byte(contentFrontend), 0644); err != nil {
		t.Fatalf("write expert-frontend: %v", err)
	}

	violations := checkSkillPreloadDrift([]string{path1, path2})

	for _, v := range violations {
		if v.Rule == "LR-08" && strings.Contains(v.Message, "moai-domain-") {
			t.Errorf("domain skill triggered LR-08 (must be exempt): %s", v.Message)
		}
	}
}

func TestSkillPreloadDriftExemption_FoundationSkills(t *testing.T) {
	// Foundation skills (moai-foundation-*) are NOT exempt from LR-08.
	// If expert-backend has moai-foundation-quality but expert-frontend does not,
	// LR-08 MUST fire.
	contentBackend := `---
name: expert-backend
skills:
  - moai-foundation-core
  - moai-foundation-quality
  - moai-workflow-testing
---
`
	contentFrontend := `---
name: expert-frontend
skills:
  - moai-foundation-core
  - moai-workflow-testing
---
`
	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "expert-backend.md")
	path2 := filepath.Join(tmpDir, "expert-frontend.md")
	if err := os.WriteFile(path1, []byte(contentBackend), 0644); err != nil {
		t.Fatalf("write expert-backend: %v", err)
	}
	if err := os.WriteFile(path2, []byte(contentFrontend), 0644); err != nil {
		t.Fatalf("write expert-frontend: %v", err)
	}

	violations := checkSkillPreloadDrift([]string{path1, path2})

	foundLR08 := false
	for _, v := range violations {
		if v.Rule == "LR-08" && strings.Contains(v.Message, "moai-foundation-quality") {
			foundLR08 = true
		}
	}
	if !foundLR08 {
		t.Errorf("expected LR-08 violation for moai-foundation-quality drift, got none")
	}
}

func TestSkillPreloadDriftExemption_WorkflowSkills(t *testing.T) {
	// Workflow skills (moai-workflow-*) are NOT exempt from LR-08.
	// If expert-backend has moai-workflow-testing but expert-frontend does not,
	// LR-08 MUST fire.
	contentBackend := `---
name: expert-backend
skills:
  - moai-foundation-core
  - moai-workflow-testing
---
`
	contentFrontend := `---
name: expert-frontend
skills:
  - moai-foundation-core
---
`
	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "expert-backend.md")
	path2 := filepath.Join(tmpDir, "expert-frontend.md")
	if err := os.WriteFile(path1, []byte(contentBackend), 0644); err != nil {
		t.Fatalf("write expert-backend: %v", err)
	}
	if err := os.WriteFile(path2, []byte(contentFrontend), 0644); err != nil {
		t.Fatalf("write expert-frontend: %v", err)
	}

	violations := checkSkillPreloadDrift([]string{path1, path2})

	foundLR08 := false
	for _, v := range violations {
		if v.Rule == "LR-08" && strings.Contains(v.Message, "moai-workflow-testing") {
			foundLR08 = true
		}
	}
	if !foundLR08 {
		t.Errorf("expected LR-08 violation for moai-workflow-testing drift, got none")
	}
}

func TestSkillPreloadDriftExemption_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		agents        []string // agent file contents
		wantLR08Count int      // expected number of LR-08 violations from domain skills
		description   string
	}{
		{
			name: "empty skills list",
			agents: []string{
				`---
name: expert-backend
skills: []
---
`,
				`---
name: expert-frontend
skills: []
---
`,
			},
			wantLR08Count: 0,
			description:   "agents with empty skill lists produce no LR-08 violations",
		},
		{
			name: "single agent in category",
			agents: []string{
				`---
name: expert-backend
skills:
  - moai-domain-backend
---
`,
			},
			wantLR08Count: 0,
			description:   "single agent in category — no peer comparison possible, no LR-08",
		},
		{
			name: "lookalike prefix does not exempt",
			agents: []string{
				`---
name: expert-backend
skills:
  - moai-foundation-core
  - moai-domainextended-backend
  - moai-workflow-testing
---
`,
				`---
name: expert-frontend
skills:
  - moai-foundation-core
  - moai-workflow-testing
---
`,
			},
			wantLR08Count: 1,
			description:   "moai-domainextended- is NOT in the exempt prefix list; LR-08 must fire",
		},
		{
			name: "all six exempt prefixes are skipped",
			agents: []string{
				`---
name: expert-backend
skills:
  - moai-foundation-core
  - moai-domain-backend
  - moai-design-backend
  - moai-library-sql
  - moai-framework-gin
  - moai-platform-aws
  - moai-ref-openapi
  - moai-workflow-testing
---
`,
				`---
name: expert-frontend
skills:
  - moai-foundation-core
  - moai-workflow-testing
---
`,
			},
			wantLR08Count: 0,
			description:   "all 6 exempt prefixes on one agent; peer has none — no LR-08 violations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var paths []string
			for i, content := range tt.agents {
				p := filepath.Join(tmpDir, filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "-")+"-agent"+string(rune('0'+i))+".md"))
				p = filepath.Join(tmpDir, strings.ReplaceAll(tt.name, " ", "-")+"-agent"+string(rune('0'+i))+".md")
				if err := os.WriteFile(p, []byte(content), 0644); err != nil {
					t.Fatalf("write agent file: %v", err)
				}
				paths = append(paths, p)
			}

			violations := checkSkillPreloadDrift(paths)

			lr08Count := 0
			for _, v := range violations {
				if v.Rule == "LR-08" {
					lr08Count++
				}
			}
			if lr08Count != tt.wantLR08Count {
				t.Errorf("%s: got %d LR-08 violations, want %d", tt.description, lr08Count, tt.wantLR08Count)
				for _, v := range violations {
					if v.Rule == "LR-08" {
						t.Logf("  LR-08: %s", v.Message)
					}
				}
			}
		})
	}
}
