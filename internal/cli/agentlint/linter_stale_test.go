package agentlint

// SPEC-CLIFIX-LINTER-STALE-001 regression suite.
//
// This file holds the REQ-007 regression tests proving each previously-dead
// check now fires AND has a non-firing negative case:
//   - LR-04 (dead-hook detection) — TestParseFrontmatter_PopulatesHooksSkills,
//     TestLR04_DeadHookFiringAndNonFiring
//   - LR-05 (writeHeavyAgents roster) — TestWriteHeavyAgents_RosterClean,
//     TestWriteHeavyAgents_NoArchivedNames (drift guard)
//   - LR-07 (live/mirror dedupe) — TestLR07_LiveMirrorPairNoFinding,
//     TestLR07_GenuineDuplicateStillFires
//
// The four checks live across two packages; the ClaimTask pending-validation
// regression lives in internal/cli/taskledger.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseFrontmatter_PopulatesHooksSkills proves REQ-LINT-001-001: yaml.v3
// parsing populates the Hooks and Skills fields that the former hand-rolled
// parser left as zero values (TODO-stub setField cases), un-blocking LR-04.
func TestParseFrontmatter_PopulatesHooksSkills(t *testing.T) {
	fm := `name: test-agent
tools: Read, Write
skills:
  - moai-workflow-tdd
  - moai-ref-testing-pyramid
hooks:
  PostToolUse:
    - matcher: Write|Edit
      hooks:
        - command: ./hook.sh
`
	var got AgentFrontmatter
	if err := parseYAMLFrontmatter([]byte(fm), &got); err != nil {
		t.Fatalf("parseYAMLFrontmatter: %v", err)
	}

	if got.Name != "test-agent" {
		t.Errorf("Name = %q, want test-agent", got.Name)
	}
	if len(got.Skills) != 2 {
		t.Errorf("Skills len = %d, want 2 (yaml.v3 must populate the list the stub dropped); got %+v", len(got.Skills), got.Skills)
	}
	postHooks, ok := got.Hooks["PostToolUse"]
	if !ok {
		t.Fatalf("Hooks[PostToolUse] missing — yaml.v3 must populate the map the stub dropped; got %+v", got.Hooks)
	}
	if len(postHooks) != 1 {
		t.Fatalf("PostToolUse hook list len = %d, want 1", len(postHooks))
	}
	if postHooks[0].Matcher != "Write|Edit" {
		t.Errorf("matcher = %q, want Write|Edit", postHooks[0].Matcher)
	}
}

// TestParseFrontmatter_SandboxScalarResilience proves the yaml.v3 parser does
// not abort when frontmatter carries `sandbox: none` (scalar into the struct).
func TestParseFrontmatter_SandboxScalarResilience(t *testing.T) {
	fm := `name: sandboxed-agent
sandbox: none
`
	var got AgentFrontmatter
	if err := parseYAMLFrontmatter([]byte(fm), &got); err != nil {
		t.Fatalf("scalar sandbox must not abort parsing: %v", err)
	}
	if got.Sandbox.Backend != "none" {
		t.Errorf("Sandbox.Backend = %q, want none", got.Sandbox.Backend)
	}
}

// TestLR04_DeadHookFiringAndNonFiring proves REQ-LINT-001-001 un-blocks LR-04
// end-to-end via the parser: a fixture declaring a hook whose matcher tool is
// absent from tools: yields an LR-04 finding (firing), while a clean fixture
// yields zero (non-firing).
func TestLR04_DeadHookFiringAndNonFiring(t *testing.T) {
	dir := t.TempDir()

	deadHook := `---
name: dead-hook-agent
tools: Read, Grep
hooks:
  PostToolUse:
    - matcher: Write|Edit
      hooks:
        - command: ./hook.sh
---
body
`
	cleanHook := `---
name: clean-hook-agent
tools: Read, Write, Edit
hooks:
  PostToolUse:
    - matcher: Write|Edit
      hooks:
        - command: ./hook.sh
---
body
`
	deadPath := filepath.Join(dir, "dead.md")
	cleanPath := filepath.Join(dir, "clean.md")
	if err := os.WriteFile(deadPath, []byte(deadHook), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cleanPath, []byte(cleanHook), 0o644); err != nil {
		t.Fatal(err)
	}

	deadVios, err := lintAgentFile(deadPath, false)
	if err != nil {
		t.Fatalf("lint dead: %v", err)
	}
	if !hasRule(deadVios, "LR-04") {
		t.Errorf("dead-hook fixture must yield an LR-04 finding; got %+v", deadVios)
	}

	cleanVios, err := lintAgentFile(cleanPath, false)
	if err != nil {
		t.Fatalf("lint clean: %v", err)
	}
	if hasRule(cleanVios, "LR-04") {
		t.Errorf("clean-hook fixture must yield zero LR-04 findings; got %+v", cleanVios)
	}
}

// TestWriteHeavyAgents_RosterClean proves REQ-LINT-001-002: a retained
// write-heavy agent (manager-develop) without isolation: worktree fires LR-05,
// while an archived name (expert-backend) does NOT.
func TestWriteHeavyAgents_RosterClean(t *testing.T) {
	// Firing: retained write-heavy agent missing isolation.
	devVios := checkMissingIsolation("test.md", AgentFrontmatter{Name: "manager-develop", Isolation: ""})
	if !hasRule(devVios, "LR-05") {
		t.Errorf("manager-develop without worktree must fire LR-05; got %+v", devVios)
	}

	// Non-firing: archived name must NOT be treated as write-heavy.
	archivedVios := checkMissingIsolation("test.md", AgentFrontmatter{Name: "expert-backend", Isolation: ""})
	if hasRule(archivedVios, "LR-05") {
		t.Errorf("archived name expert-backend must NOT fire LR-05; got %+v", archivedVios)
	}
}

// TestWriteHeavyAgents_NoArchivedNames is the drift guard (plan §G
// anti-pattern): the writeHeavyAgents slice must contain no archived agent
// name, so the roster cannot silently drift back to dead names. Scoped to the
// slice literal only (not comments elsewhere in the file).
func TestWriteHeavyAgents_NoArchivedNames(t *testing.T) {
	src, err := os.ReadFile("agent_lint.go")
	if err != nil {
		t.Fatal(err)
	}
	block := extractWriteHeavyAgentsBlock(string(src))
	if block == "" {
		t.Fatal("could not locate writeHeavyAgents slice literal in agent_lint.go")
	}
	archived := []string{"expert-backend", "expert-frontend", "expert-refactoring", "researcher",
		"expert-security", "expert-devops", "expert-performance", "manager-strategy", "manager-quality",
		"manager-brain", "manager-project", "claude-code-guide"}
	for _, name := range archived {
		needle := `"` + name + `"`
		if strings.Contains(block, needle) {
			t.Errorf("writeHeavyAgents must not name archived agent %q (CLAUDE.md §4 SSOT)", name)
		}
	}
}

// extractWriteHeavyAgentsBlock returns the source text of the
// `writeHeavyAgents := []string{ ... }` literal, or "" if not found.
func extractWriteHeavyAgentsBlock(src string) string {
	marker := "writeHeavyAgents := []string{"
	idx := strings.Index(src, marker)
	if idx < 0 {
		return ""
	}
	end := strings.Index(src[idx:], "}")
	if end < 0 {
		return ""
	}
	return src[idx : idx+end]
}

// TestLR07_LiveMirrorPairNoFinding proves REQ-LINT-001-003: a live agent file
// and its template mirror both carrying a Skeptical-Evaluator Mandate block
// produce ZERO LR-07 findings (the pair is deduped by path mapping).
func TestLR07_LiveMirrorPairNoFinding(t *testing.T) {
	dir := t.TempDir()
	liveDir := filepath.Join(dir, "live", ".claude", "agents", "moai")
	mirrorDir := filepath.Join(dir, "internal", "template", "templates", ".claude", "agents", "moai")
	if err := os.MkdirAll(liveDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	block := mandateBlockFixture()
	livePath := filepath.Join(liveDir, "sync-auditor.md")
	mirrorPath := filepath.Join(mirrorDir, "sync-auditor.md")
	if err := os.WriteFile(livePath, []byte(block), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mirrorPath, []byte(block), 0o644); err != nil {
		t.Fatal(err)
	}

	vios := checkDuplicateMandateBlocks([]string{livePath, mirrorPath})
	if l := countRule(vios, "LR-07"); l != 0 {
		t.Errorf("live/mirror pair must yield 0 LR-07 findings; got %d: %+v", l, vios)
	}
}

// TestLR07_GenuineDuplicateStillFires proves REQ-LINT-001-003 negative space:
// two DISTINCT live agents (unrelated paths, same name) both carrying the block
// still produce a finding — the dedupe does not mask true duplicates.
func TestLR07_GenuineDuplicateStillFires(t *testing.T) {
	dir := t.TempDir()
	liveDir := filepath.Join(dir, "live", ".claude", "agents", "moai")
	otherDir := filepath.Join(dir, "other", ".claude", "agents", "moai")
	if err := os.MkdirAll(liveDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatal(err)
	}
	block := mandateBlockFixture()
	a := filepath.Join(liveDir, "sync-auditor.md")
	b := filepath.Join(otherDir, "sync-auditor.md")
	if err := os.WriteFile(a, []byte(block), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(block), 0o644); err != nil {
		t.Fatal(err)
	}

	vios := checkDuplicateMandateBlocks([]string{a, b})
	if l := countRule(vios, "LR-07"); l < 1 {
		t.Errorf("genuine same-name duplicate across unrelated paths must still fire LR-07; got %d: %+v", l, vios)
	}
}

// mandateBlockFixture returns an agent file body containing a Skeptical-Evaluator
// Mandate header + >=3 mandate bullets, matching the LR-07 detector patterns.
func mandateBlockFixture() string {
	return "## Skeptical-Evaluator Mandate\n" +
		"- evaluate evidence before accepting claims\n" +
		"- score quality on a fixed rubric\n" +
		"- assess security and robustness explicitly\n"
}

func hasRule(vios []LintViolation, rule string) bool {
	for _, v := range vios {
		if v.Rule == rule {
			return true
		}
	}
	return false
}

func countRule(vios []LintViolation, rule string) int {
	n := 0
	for _, v := range vios {
		if v.Rule == rule {
			n++
		}
	}
	return n
}
