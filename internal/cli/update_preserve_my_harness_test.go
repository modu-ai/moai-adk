// SPEC-V3R3-HARNESS-001 / T-M3-02
// RED phase: isUserAreaPath does not exist yet — test fails.
// GREEN phase: implement isUserAreaPath + integrate into cleanMoaiManagedPaths → tests pass.

package cli

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// fileSHA256 returns the SHA-256 hex digest of a file.
func fileSHA256(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("fileSHA256: open %s: %v", path, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("fileSHA256: hash %s: %v", path, err)
	}
	return string(h.Sum(nil))
}

// TestIsUserAreaPath verifies the guard function used by the update flow to
// identify project-side paths that must never be overwritten or deleted.
func TestIsUserAreaPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		relPath  string
		wantSkip bool
	}{
		// User customization area — must be preserved
		{".claude/skills/my-harness-test-skill/SKILL.md", ".claude/skills/my-harness-test-skill/SKILL.md", true},
		{".claude/skills/my-harness-foo/bar.md", ".claude/skills/my-harness-foo/bar.md", true},
		{".claude/agents/my-harness/test-agent.md", ".claude/agents/my-harness/test-agent.md", true},
		// MoAI-managed — must NOT be skipped by user-area guard
		{".claude/skills/moai-foundation-cc/SKILL.md", ".claude/skills/moai-foundation-cc/SKILL.md", false},
		{".claude/skills/moai-meta-harness/SKILL.md", ".claude/skills/moai-meta-harness/SKILL.md", false},
		{".claude/agents/moai/spec.md", ".claude/agents/moai/spec.md", false},
		// Other project files — must NOT be skipped
		{"CLAUDE.md", "CLAUDE.md", false},
		{".moai/config/sections/quality.yaml", ".moai/config/sections/quality.yaml", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isUserAreaPath(tt.relPath)
			if got != tt.wantSkip {
				t.Errorf("isUserAreaPath(%q) = %v, want %v", tt.relPath, got, tt.wantSkip)
			}
		})
	}
}

// TestPreserveMyHarnessOnUpdate verifies that moai update does NOT overwrite or
// delete files under .claude/skills/my-harness-* or .claude/agents/my-harness/.
//
// The test:
//  1. Sets up a fixture project in t.TempDir() with a my-harness-test-skill dir.
//  2. Calls cleanMoaiManagedPaths (the stale-file removal step inside update).
//  3. Asserts the user file is still present with identical content.
func TestPreserveMyHarnessOnUpdate(t *testing.T) {
	t.Parallel()

	// --- Setup fixture project ---
	tmpDir := t.TempDir()

	// my-harness skill file
	harnessSkillDir := filepath.Join(tmpDir, ".claude", "skills", "my-harness-test-skill")
	if err := os.MkdirAll(harnessSkillDir, 0o755); err != nil {
		t.Fatalf("mkdir my-harness-test-skill: %v", err)
	}
	harnessSkillFile := filepath.Join(harnessSkillDir, "SKILL.md")
	const skillContent = "test content — user harness skill"
	if err := os.WriteFile(harnessSkillFile, []byte(skillContent), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	// my-harness agent file
	harnessAgentDir := filepath.Join(tmpDir, ".claude", "agents", "my-harness")
	if err := os.MkdirAll(harnessAgentDir, 0o755); err != nil {
		t.Fatalf("mkdir my-harness agents: %v", err)
	}
	harnessAgentFile := filepath.Join(harnessAgentDir, "test-agent.md")
	const agentContent = "test content — user harness agent"
	if err := os.WriteFile(harnessAgentFile, []byte(agentContent), 0o644); err != nil {
		t.Fatalf("write test-agent.md: %v", err)
	}

	// Capture pre-update content hashes
	preSHA256Skill := fileSHA256(t, harnessSkillFile)
	preSHA256Agent := fileSHA256(t, harnessAgentFile)

	// Capture pre-update mtime
	skillInfoBefore, err := os.Stat(harnessSkillFile)
	if err != nil {
		t.Fatalf("stat SKILL.md before: %v", err)
	}
	agentInfoBefore, err := os.Stat(harnessAgentFile)
	if err != nil {
		t.Fatalf("stat test-agent.md before: %v", err)
	}

	// --- Exercise cleanMoaiManagedPaths (the stale-file removal step) ---
	// We discard output — we only care about side effects on the fixture.
	if err := cleanMoaiManagedPaths(tmpDir, io.Discard); err != nil {
		t.Fatalf("cleanMoaiManagedPaths: %v", err)
	}

	// --- Assert: user files are untouched ---
	// SKILL.md must still exist
	if _, err := os.Stat(harnessSkillFile); err != nil {
		t.Errorf("SKILL.md removed by cleanMoaiManagedPaths (must be preserved): %v", err)
	}
	// SKILL.md content must be identical
	postSHA256Skill := fileSHA256(t, harnessSkillFile)
	if preSHA256Skill != postSHA256Skill {
		t.Errorf("SKILL.md content changed: SHA256 before=%x after=%x", preSHA256Skill, postSHA256Skill)
	}
	// SKILL.md mtime must be unchanged
	skillInfoAfter, err := os.Stat(harnessSkillFile)
	if err != nil {
		t.Fatalf("stat SKILL.md after: %v", err)
	}
	if !skillInfoAfter.ModTime().Equal(skillInfoBefore.ModTime()) {
		t.Errorf("SKILL.md mtime changed: before=%v after=%v", skillInfoBefore.ModTime(), skillInfoAfter.ModTime())
	}

	// test-agent.md must still exist
	if _, err := os.Stat(harnessAgentFile); err != nil {
		t.Errorf("test-agent.md removed by cleanMoaiManagedPaths (must be preserved): %v", err)
	}
	// test-agent.md content must be identical
	postSHA256Agent := fileSHA256(t, harnessAgentFile)
	if preSHA256Agent != postSHA256Agent {
		t.Errorf("test-agent.md content changed: SHA256 before=%x after=%x", preSHA256Agent, postSHA256Agent)
	}
	// test-agent.md mtime must be unchanged
	agentInfoAfter, err := os.Stat(harnessAgentFile)
	if err != nil {
		t.Fatalf("stat test-agent.md after: %v", err)
	}
	if !agentInfoAfter.ModTime().Equal(agentInfoBefore.ModTime()) {
		t.Errorf("test-agent.md mtime changed: before=%v after=%v", agentInfoBefore.ModTime(), agentInfoAfter.ModTime())
	}
}
