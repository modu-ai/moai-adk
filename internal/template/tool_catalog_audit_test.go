package template

// Tool-catalog guard: validates that agent and skill tool declarations only
// reference tools that exist in the current Claude Code tool catalog.
//
// Allowlist source: the current Claude Code tool catalog (the set of tools
// exposed to agents and skills by the Claude Code runtime). Two tools that were
// previously valid are now disallowed in deployed agent/skill declarations:
//
//   - MultiEdit: removed from the Claude Code tool catalog. Only Edit (with
//     replace_all) and Write remain for file modification. An agent or skill
//     that declares MultiEdit declares a tool that no longer exists.
//   - TodoWrite: disabled by default in current Claude Code; the task-tracking
//     surface is the TaskCreate / TaskUpdate / TaskList / TaskGet family. A
//     retained agent that declares only TodoWrite for task tracking gets no
//     task-tracking tool unless an environment override is set. The migration
//     target is the Task* family.
//
// When the Claude Code tool catalog changes again, update the disallowed-token
// sets below and the rationale comment so a future maintainer can follow the
// reasoning.
//
// Scope:
//   - HARD-fail on any MultiEdit token in any agent `tools:` line or skill
//     `allowed-tools:` line (retired tool — never valid).
//   - HARD-fail on any TodoWrite token in a retained agent's `tools:` line
//     (default-disabled tool — migration target is the Task* family). Skill
//     `allowed-tools:` lines are NOT failed on TodoWrite, because skill
//     documentation may legitimately reference the historical tool name in a
//     teaching example; the agent frontmatter is the authoritative declaration
//     surface the runtime reads.
//   - Verbatim official-docs reproductions under the foundation-cc reference
//     directory are exempt: they document the historical catalog and are not
//     live tool declarations.

import (
	"bufio"
	"bytes"
	"io/fs"
	"strings"
	"testing"
)

// toolCatalogRetiredTokens are tool names that were removed from the Claude Code
// tool catalog. Any occurrence in an agent `tools:` line or skill
// `allowed-tools:` line is a hard failure.
var toolCatalogRetiredTokens = []string{"MultiEdit"}

// toolCatalogDisabledTokens are tool names that are default-disabled in the
// current Claude Code catalog. Any occurrence in a retained agent's `tools:`
// frontmatter line is a hard failure (the migration target is the Task* family).
var toolCatalogDisabledTokens = []string{"TodoWrite"}

// officialDocsExemptSubstrings marks verbatim official-docs reproductions that
// document the historical tool catalog rather than declaring live tools. Files
// whose path contains any of these substrings are exempt from the audit.
var officialDocsExemptSubstrings = []string{
	"claude-code-iam-official.md",
	"claude-code-sub-agents-official.md",
	"claude-code-plugins-official.md",
}

// toolListTokens extracts comma-separated tool tokens from a frontmatter line of
// the form `tools: A, B, C` or `allowed-tools: A, B, C`. The line is split on
// the first colon; the remainder is split on commas and trimmed.
func toolListTokens(line string) []string {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return nil
	}
	rest := line[idx+1:]
	parts := strings.Split(rest, ",")
	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

// containsToken reports whether the token list contains an exact match for name.
func containsToken(tokens []string, name string) bool {
	for _, t := range tokens {
		if t == name {
			return true
		}
	}
	return false
}

// isOfficialDocsExempt reports whether the path is a verbatim official-docs
// reproduction exempt from the audit.
func isOfficialDocsExempt(path string) bool {
	for _, sub := range officialDocsExemptSubstrings {
		if strings.Contains(path, sub) {
			return true
		}
	}
	return false
}

// TestToolCatalogNoMultiEdit walks the embedded agent and skill files and
// asserts that no `tools:` / `allowed-tools:` declaration contains the retired
// MultiEdit tool. Sentinel on violation: TOOL_CATALOG_RETIRED_TOOL.
func TestToolCatalogNoMultiEdit(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	for _, root := range []string{".claude/agents", ".claude/skills"} {
		if _, statErr := fs.Stat(fsys, root); statErr != nil {
			continue // directory may be absent in a partial embed; skip gracefully
		}
		walkErr := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			if isOfficialDocsExempt(path) {
				return nil
			}
			data, readErr := fs.ReadFile(fsys, path)
			if readErr != nil {
				return nil
			}
			scanner := bufio.NewScanner(bytes.NewReader(data))
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				line := strings.TrimSpace(scanner.Text())
				if !strings.HasPrefix(line, "tools:") && !strings.HasPrefix(line, "allowed-tools:") {
					continue
				}
				tokens := toolListTokens(line)
				for _, retired := range toolCatalogRetiredTokens {
					if containsToken(tokens, retired) {
						t.Errorf(
							"TOOL_CATALOG_RETIRED_TOOL: %s line %d declares retired tool %q; "+
								"it was removed from the Claude Code tool catalog (use Edit/Write).",
							path, lineNum, retired,
						)
					}
				}
			}
			return scanner.Err()
		})
		if walkErr != nil {
			t.Fatalf("WalkDir(%q) error: %v", root, walkErr)
		}
	}
}

// TestToolCatalogNoTodoWrite walks the embedded retained agent files and asserts
// that no `tools:` frontmatter declaration contains the default-disabled
// TodoWrite tool. The migration target is the Task* family
// (TaskCreate / TaskUpdate / TaskList / TaskGet). Sentinel on violation:
// TOOL_CATALOG_DISABLED_TOOL.
func TestToolCatalogNoTodoWrite(t *testing.T) {
	t.Parallel()

	fsys, err := EmbeddedTemplates()
	if err != nil {
		t.Fatalf("EmbeddedTemplates() error: %v", err)
	}

	const agentRoot = ".claude/agents/moai"
	if _, statErr := fs.Stat(fsys, agentRoot); statErr != nil {
		t.Skipf("%s not embedded; skipping", agentRoot)
	}

	walkErr := fs.WalkDir(fsys, agentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, readErr := fs.ReadFile(fsys, path)
		if readErr != nil {
			return nil
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "tools:") {
				continue
			}
			tokens := toolListTokens(line)
			for _, disabled := range toolCatalogDisabledTokens {
				if containsToken(tokens, disabled) {
					t.Errorf(
						"TOOL_CATALOG_DISABLED_TOOL: %s line %d declares default-disabled tool %q; "+
							"migrate to the Task* family (TaskCreate, TaskUpdate, TaskList, TaskGet).",
						path, lineNum, disabled,
					)
				}
			}
		}
		return scanner.Err()
	})
	if walkErr != nil {
		t.Fatalf("WalkDir(%q) error: %v", agentRoot, walkErr)
	}
}
