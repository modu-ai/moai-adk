// agent_frontmatter_audit_test.go: standardization audit for retired agent frontmatter.
// Maps to REQ-RA-002, REQ-RA-013, REQ-RA-016.
//
// M1 RED phase: the tests in this file are intentionally designed to fail
// against the not-yet-implemented state.
//
// Expected RED state:
//   - TestAgentFrontmatterAudit: FAIL because manager-tdd.md lacks retired:true (GREEN in M2)
//   - TestRetirementCompletenessAssertion: FAIL because manager-develop.md is missing from the embedded FS (GREEN in M2)
//   - TestNoOrphanedManagerTDDReference: FAIL because several files still reference manager-tdd (GREEN in M5)
package template

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// retiredFrontmatter holds the retired-related fields parsed from an agent file.
type retiredFrontmatter struct {
	retired            bool
	retiredReplacement string
	retiredParamHint   string
	tools              string // raw value (empty array parses as "[]")
	skills             string // raw value (empty array parses as "[]")
	hasStatusRetired   bool   // whether the legacy status: retired field is present
}

// parseRetiredFields extracts retired-related fields from a frontmatter map.
func parseRetiredFields(fm map[string]string) retiredFrontmatter {
	result := retiredFrontmatter{}

	if val, ok := fm["retired"]; ok {
		// YAML boolean: parsed as the string "true" (parseFrontmatterAndBody strips quotes and returns the raw value)
		result.retired = strings.TrimSpace(val) == "true"
	}
	if val, ok := fm["retired_replacement"]; ok {
		result.retiredReplacement = strings.TrimSpace(val)
	}
	if val, ok := fm["retired_param_hint"]; ok {
		result.retiredParamHint = strings.TrimSpace(val)
	}
	if val, ok := fm["tools"]; ok {
		result.tools = strings.TrimSpace(val)
	}
	if val, ok := fm["skills"]; ok {
		result.skills = strings.TrimSpace(val)
	}
	// Detect the legacy status: retired field
	if val, ok := fm["status"]; ok && strings.TrimSpace(val) == "retired" {
		result.hasStatusRetired = true
	}

	return result
}

// TestAgentFrontmatterAudit walks the retained-agent subfolders (.claude/agents/{core,meta}/*.md)
// and verifies that the five standard retired:true frontmatter fields are present
// when an agent declares retired:true.
//
// REQ-RA-002: standard retired frontmatter field validation
// REQ-TST-011: walk path updated to current retained catalog reality (post SPEC-V3R6-AGENT-TEAM-REBUILD-001).
// Post-ATR-001: no retired stubs remain in the embedded template (12 archived agents
// physically moved to .moai/backups/agent-archive-2026-05-25/). The audit therefore
// validates retained-agent frontmatter cleanliness (no orphan retired:true keys, no
// legacy status: retired field) across the {core, meta} subfolders.
func TestAgentFrontmatterAudit(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	var agentFiles []string
	for _, domain := range []string{"core", "meta"} {
		agentDir := ".claude/agents/" + domain
		walkErr := fs.WalkDir(fsys, agentDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".md") {
				agentFiles = append(agentFiles, path)
			}
			return nil
		})
		if walkErr != nil {
			t.Fatalf("WalkDir(%q) 오류: %v", agentDir, walkErr)
		}
	}
	if len(agentFiles) == 0 {
		t.Fatal(".claude/agents/{core,meta}/ 하위 에이전트 파일이 없음")
	}

	// Validation rules:
	// 1. retired:true agents: all five standard fields are required
	// 2. non-retired agents: the retired: key itself MUST be absent
	// 3. legacy status:retired field is forbidden
	for _, path := range agentFiles {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) 오류: %v", path, readErr)
			}

			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				t.Fatalf("frontmatter 파싱 실패: %s", parseErr)
			}

			rf := parseRetiredFields(fm)

			// Reject the legacy status:retired field (REQ-RA-002 requires the 'retired: true' boolean)
			if rf.hasStatusRetired {
				t.Errorf("RETIREMENT_INCOMPLETE: legacy 'status: retired' 필드 감지. 'retired: true' boolean 필드로 교체 필요 (REQ-RA-002)")
			}

			if rf.retired {
				// retired:true agents: validate the five standard fields
				if rf.retiredReplacement == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'retired_replacement' 필드 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
				if rf.retiredParamHint == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'retired_param_hint' 필드 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
				// tools: [] empty array MUST be explicit
				if rf.tools == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'tools: []' 명시적 빈 배열 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
				// skills: [] empty array MUST be explicit
				if rf.skills == "" {
					t.Errorf("RETIREMENT_INCOMPLETE_%s: retired:true 에이전트에 'skills: []' 명시적 빈 배열 없음 (REQ-RA-002)", agentNameFromPath(path))
				}
			}
		})
	}

	// Workflow audit 2026-05-16 Bundle C / F-003: the two subtests (manager-tdd must be retired,
	// manager-ddd must be retired) were removed. The policy shifted from stub-preservation to
	// full purge for those zombie agents. The new contract is verified by TestPurgedZombieAgentsAbsent.
}

// TestRetirementCompletenessAssertion verifies that, for every retired:true agent,
// the replacement agent file exists in the embedded FS.
//
// REQ-RA-016: CI must perform the RETIREMENT_INCOMPLETE_<agent> check
// REQ-TST-010: path drift fix — .claude/agents/moai/ → .claude/agents/core/
// per SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (pre-existing per ATR-001 §F.2.8).
func TestRetirementCompletenessAssertion(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	// Explicit assertion of the manager-tdd → manager-develop replacement.
	// When manager-tdd.md has retired:true, manager-develop.md must exist in the embedded FS.
	// Post AGENT-FOLDER-SPLIT-001: manager-develop.md lives under core/ subfolder.
	t.Run("manager-tdd replacement manager-develop must exist", func(t *testing.T) {
		t.Parallel()

		const replacementPath = ".claude/agents/core/manager-develop.md"
		_, statErr := fs.Stat(fsys, replacementPath)
		if statErr != nil {
			t.Errorf("RETIREMENT_INCOMPLETE_manager-tdd: 교체 에이전트 '%s'가 embedded FS에 없음. "+
				"SPEC-V3R3-RETIRED-AGENT-001 M2에서 manager-develop.md 추가 필요 (REQ-RA-016)", replacementPath)
		}
	})

	// Explicit assertion of the manager-ddd → manager-develop replacement.
	// When manager-ddd.md has retired:true, manager-develop.md must exist in the embedded FS.
	// Post AGENT-FOLDER-SPLIT-001: manager-develop.md lives under core/ subfolder.
	t.Run("manager-ddd replacement manager-develop must exist", func(t *testing.T) {
		t.Parallel()

		const replacementPath = ".claude/agents/core/manager-develop.md"
		_, statErr := fs.Stat(fsys, replacementPath)
		if statErr != nil {
			t.Errorf("RETIREMENT_INCOMPLETE_manager-ddd: 교체 에이전트 '%s'가 embedded FS에 없음. "+
				"SPEC-V3R3-RETIRED-DDD-001 M2에서 manager-develop.md 추가 필요 (REQ-RD-012)", replacementPath)
		}
	})

	// Generic check: for every retired:true agent in the embedded FS, verify the replacement file exists.
	// Post AGENT-FOLDER-SPLIT-001: walk retained subfolders {core, meta} and resolve
	// retired_replacement to whichever subfolder contains the file.
	t.Run("all retired agents have replacement in embedded FS", func(t *testing.T) {
		t.Parallel()

		var agentFiles []string
		for _, domain := range []string{"core", "meta"} {
			_ = fs.WalkDir(fsys, ".claude/agents/"+domain, func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				if strings.HasSuffix(path, ".md") {
					agentFiles = append(agentFiles, path)
				}
				return nil
			})
		}

		for _, path := range agentFiles {
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				continue
			}
			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				continue
			}
			rf := parseRetiredFields(fm)
			if !rf.retired || rf.retiredReplacement == "" {
				continue
			}

			// Resolve the replacement file path: search both retained subfolders.
			found := false
			var attempted []string
			for _, domain := range []string{"core", "meta"} {
				replacementPath := fmt.Sprintf(".claude/agents/%s/%s.md", domain, rf.retiredReplacement)
				attempted = append(attempted, replacementPath)
				if _, statErr := fs.Stat(fsys, replacementPath); statErr == nil {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("RETIREMENT_INCOMPLETE_%s: retired_replacement '%s' 파일이 embedded FS에 없음 (시도: %v)",
					agentNameFromPath(path), rf.retiredReplacement, attempted)
			}
		}
	})
}

// TestNoOrphanedManagerTDDReference verifies that the specified core files no
// longer reference manager-tdd.
//
// REQ-RA-013: when manager-develop is the active unified agent, every documentation reference must be updated
// Expected RED: FAIL because several files still reference manager-tdd
func TestNoOrphanedManagerTDDReference(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	// Files that MUST NOT reference manager-tdd (REQ-RA-013 §M5 substitution scope)
	// Each file fails if the string "manager-tdd" is found
	// Exceptions: the manager-tdd.md file itself (frontmatter name: line, migration notes)
	checkFiles := []struct {
		path        string
		description string
	}{
		{
			path:        "CLAUDE.md",
			description: "CLAUDE.md §4 Manager Agents 및 §5 Agent Chain",
		},
		{
			path:        ".claude/rules/moai/development/agent-authoring.md",
			description: "agent-authoring.md Manager Agents 목록",
		},
		{
			path:        ".claude/rules/moai/core/agent-hooks.md",
			description: "agent-hooks.md Agent Hook Actions 테이블",
		},
		{
			path:        ".claude/rules/moai/workflow/spec-workflow.md",
			description: "spec-workflow.md Phase Overview 테이블",
		},
		{
			path:        ".claude/agents/moai/manager-strategy.md",
			description: "manager-strategy.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/manager-ddd.md",
			description: "manager-ddd.md 인라인 참조 (2개 라인)",
		},
	}

	for _, cf := range checkFiles {
		t.Run(cf.path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, cf.path)
			if readErr != nil {
				// Skip the test when the file is missing (file presence is verified in M5)
				t.Skipf("파일 %q 읽기 실패 (make build 필요): %v", cf.path, readErr)
				return
			}

			content := string(data)
			// Search for manager-tdd references (case-sensitive)
			// Plain substring scan, augmented by a common-pattern search to handle word boundaries
			orphanedRefs := findManagerTDDReferences(content)
			if len(orphanedRefs) > 0 {
				t.Errorf("ORPHANED_MANAGER_TDD_REFERENCE in %s (%s): %d개 참조 발견. "+
					"SPEC-V3R3-RETIRED-AGENT-001 M5에서 'manager-develop'로 교체 필요 (REQ-RA-013):\n%s",
					cf.path, cf.description, len(orphanedRefs), strings.Join(orphanedRefs, "\n"))
			}
		})
	}
}

// TestNoOrphanedManagerDDDReference verifies that the specified core files no
// longer reference manager-ddd.
//
// REQ-RD-010: when manager-develop is the active unified agent, every documentation reference must be updated
// Expected RED: FAIL because 30 Cat A files still reference manager-ddd
func TestNoOrphanedManagerDDDReference(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	// Files that MUST NOT reference manager-ddd (REQ-RD-010 §M3 Cat A substitution scope)
	// Each file fails if the string "manager-ddd" is found
	// Exceptions: the manager-ddd.md file itself (frontmatter name: line, migration notes)
	checkFiles := []struct {
		path        string
		description string
	}{
		// Cat A1 — Rule files (3 files)
		{
			path:        ".claude/rules/moai/development/agent-authoring.md",
			description: "agent-authoring.md Manager Agents 목록",
		},
		{
			path:        ".claude/rules/moai/workflow/spec-workflow.md",
			description: "spec-workflow.md Phase Overview 테이블",
		},
		{
			path:        ".claude/rules/moai/workflow/worktree-integration.md",
			description: "worktree-integration.md Worktree Isolation Rules",
		},
		// Cat A2 — Agent definition files (11 files)
		{
			path:        ".claude/agents/moai/manager-strategy.md",
			description: "manager-strategy.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/manager-quality.md",
			description: "manager-quality.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/manager-spec.md",
			description: "manager-spec.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-backend.md",
			description: "expert-backend.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-frontend.md",
			description: "expert-frontend.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-testing.md",
			description: "expert-testing.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-debug.md",
			description: "expert-debug.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-devops.md",
			description: "expert-devops.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-mobile.md",
			description: "expert-mobile.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/expert-refactoring.md",
			description: "expert-refactoring.md 에이전트 위임 참조",
		},
		{
			path:        ".claude/agents/moai/evaluator-active.md",
			description: "evaluator-active.md 에이전트 위임 참조",
		},
		// Cat A3 — Output-style file (1 file)
		{
			path:        ".claude/output-styles/moai/moai.md",
			description: "moai.md Command Reference 테이블",
		},
		// Cat A4 — Skill files (15 files)
		{
			path:        ".claude/skills/moai/SKILL.md",
			description: "moai SKILL.md Agent Catalog",
		},
		{
			path:        ".claude/skills/moai/references/mx-tag.md",
			description: "mx-tag.md @MX TAG Protocol",
		},
		{
			path:        ".claude/skills/moai/references/reference.md",
			description: "reference.md Reference Architecture",
		},
		{
			path:        ".claude/skills/moai/workflows/moai.md",
			description: "moai.md MoAI Unified Workflow",
		},
		{
			path:        ".claude/skills/moai/workflows/run.md",
			description: "run.md Run Phase Workflow",
		},
		{
			path:        ".claude/skills/moai-foundation-cc/SKILL.md",
			description: "foundation-cc SKILL.md",
		},
		{
			path:        ".claude/skills/moai-foundation-core/SKILL.md",
			description: "foundation-core SKILL.md",
		},
		{
			path:        ".claude/skills/moai-foundation-quality/SKILL.md",
			description: "foundation-quality SKILL.md",
		},
		{
			path:        ".claude/skills/moai-meta-harness/SKILL.md",
			description: "meta-harness SKILL.md",
		},
		{
			path:        ".claude/skills/moai-workflow-ddd/SKILL.md",
			description: "workflow-ddd SKILL.md",
		},
		{
			path:        ".claude/skills/moai-workflow-loop/SKILL.md",
			description: "workflow-loop SKILL.md",
		},
		{
			path:        ".claude/skills/moai-workflow-spec/SKILL.md",
			description: "workflow-spec SKILL.md",
		},
		{
			path:        ".claude/skills/moai-workflow-spec/references/reference.md",
			description: "workflow-spec reference.md",
		},
		{
			path:        ".claude/skills/moai-workflow-spec/references/examples.md",
			description: "workflow-spec examples.md",
		},
		{
			path:        ".claude/skills/moai-workflow-testing/SKILL.md",
			description: "workflow-testing SKILL.md",
		},
	}

	for _, cf := range checkFiles {
		t.Run(cf.path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, cf.path)
			if readErr != nil {
				// Skip the test when the file is missing (file presence is verified in M3)
				t.Skipf("파일 %q 읽기 실패 (make build 필요): %v", cf.path, readErr)
				return
			}

			content := string(data)
			// Search for manager-ddd references (case-sensitive)
			// Plain substring scan, augmented by a common-pattern search to handle word boundaries
			orphanedRefs := findManagerDDDReferences(content)
			if len(orphanedRefs) > 0 {
				t.Errorf("ORPHANED_MANAGER_DDD_REFERENCE in %s (%s): %d개 참조 발견. "+
					"SPEC-V3R3-RETIRED-DDD-001 M3에서 'manager-develop'로 교체 필요 (REQ-RD-010):\n%s",
					cf.path, cf.description, len(orphanedRefs), strings.Join(orphanedRefs, "\n"))
			}
		})
	}
}

// TestAgentFrontmatter_PermissionModeStrictEnum verifies that the permissionMode
// key in .claude/agents/**/*.md is one of the five allowed values.
//
// AC-09: a value such as permissionMode: ultra-bypass MUST fail the build.
// Sentinel: PERMISSION_MODE_UNKNOWN_VALUE: <file> declares permissionMode: <value>;
//
//	allowed: default|acceptEdits|bypassPermissions|plan|bubble.
//
// T-RT002-12.
func TestAgentFrontmatter_PermissionModeStrictEnum(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() 오류: %v", err)
	}

	// The five allowed permissionMode values.
	allowedModes := map[string]bool{
		"default":            true,
		"acceptEdits":        true,
		"bypassPermissions":  true,
		"plan":               true,
		"bubble":             true,
	}

	var agentFiles []string
	_ = fs.WalkDir(fsys, ".claude/agents", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			agentFiles = append(agentFiles, path)
		}
		return nil
	})

	for _, path := range agentFiles {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				t.Fatalf("ReadFile(%q) 오류: %v", path, readErr)
			}

			fm, _, parseErr := parseFrontmatterAndBody(string(data))
			if parseErr != "" {
				// Skip files without frontmatter.
				return
			}

			val, ok := fm["permissionMode"]
			if !ok {
				// Skip when the permissionMode key is absent (optional field).
				return
			}

			val = strings.TrimSpace(val)
			if !allowedModes[val] {
				t.Errorf("PERMISSION_MODE_UNKNOWN_VALUE: %s declares permissionMode: %s; "+
					"allowed: default|acceptEdits|bypassPermissions|plan|bubble",
					path, val)
			}
		})
	}
}

// agentNameFromPath extracts the agent name from a file path.
// ".claude/agents/moai/manager-tdd.md" → "manager-tdd"
func agentNameFromPath(path string) string {
	base := path
	// Segment after the last /
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		base = path[idx+1:]
	}
	// Strip .md
	return strings.TrimSuffix(base, ".md")
}

// findManagerTDDReferences searches the content for manager-tdd references and returns them.
// The frontmatter name: line and migration notes are allowed exceptions.
func findManagerTDDReferences(content string) []string {
	var refs []string
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Search for manager-tdd references
		if !strings.Contains(line, "manager-tdd") {
			continue
		}
		// Allowed exception: the frontmatter name field (manager-tdd.md's own name: manager-tdd)
		// Allowed exception: migration notes (# deprecated, # previous name, <!-- , [DEPRECATED, …)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name: manager-tdd") {
			continue
		}
		if strings.HasPrefix(trimmed, "#") && strings.Contains(strings.ToLower(trimmed), "deprecated") {
			continue
		}
		if strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		refs = append(refs, fmt.Sprintf("  L%d: %s", i+1, trimmed))
	}
	return refs
}

// findManagerDDDReferences searches the content for manager-ddd references and returns them.
// The frontmatter name: line and migration notes are allowed exceptions.
func findManagerDDDReferences(content string) []string {
	var refs []string
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Search for manager-ddd references
		if !strings.Contains(line, "manager-ddd") {
			continue
		}
		// Allowed exception: the frontmatter name field (manager-ddd.md's own name: manager-ddd)
		// Allowed exception: migration notes (# deprecated, # previous name, <!-- , [DEPRECATED, …)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "name: manager-ddd") {
			continue
		}
		if strings.HasPrefix(trimmed, "#") && strings.Contains(strings.ToLower(trimmed), "deprecated") {
			continue
		}
		if strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		refs = append(refs, fmt.Sprintf("  L%d: %s", i+1, trimmed))
	}
	return refs
}
