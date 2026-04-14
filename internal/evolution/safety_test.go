package evolution_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/evolution"
)

// TestFrozenGuard_BlocksConstitution verifies that moai-constitution.md is
// protected from modification.
func TestFrozenGuard_BlocksConstitution(t *testing.T) {
	target := ".claude/rules/moai/core/moai-constitution.md"
	err := evolution.CheckFrozenGuard(target)
	if err == nil {
		t.Fatalf("expected ErrFrozenPath for %q, got nil", target)
	}
	if err != evolution.ErrFrozenPath {
		t.Fatalf("expected ErrFrozenPath, got %v", err)
	}
}

// TestFrozenGuard_BlocksRootCLAUDEMD verifies that the root CLAUDE.md is
// protected from modification.
func TestFrozenGuard_BlocksRootCLAUDEMD(t *testing.T) {
	target := "CLAUDE.md"
	err := evolution.CheckFrozenGuard(target)
	if err == nil {
		t.Fatalf("expected ErrFrozenPath for %q, got nil", target)
	}
	if err != evolution.ErrFrozenPath {
		t.Fatalf("expected ErrFrozenPath, got %v", err)
	}
}

// TestFrozenGuard_AllowsSkillFile verifies that SKILL.md files in the skills
// directory are permitted targets.
func TestFrozenGuard_AllowsSkillFile(t *testing.T) {
	target := ".claude/skills/moai-lang-go/SKILL.md"
	err := evolution.CheckFrozenGuard(target)
	if err != nil {
		t.Fatalf("expected no error for skill file %q, got %v", target, err)
	}
}

// TestFrozenGuard_BlocksAgentCommonProtocol verifies that agent-common-protocol.md
// is in the frozen zone.
func TestFrozenGuard_BlocksAgentCommonProtocol(t *testing.T) {
	target := ".claude/rules/moai/core/agent-common-protocol.md"
	err := evolution.CheckFrozenGuard(target)
	if err == nil {
		t.Fatalf("expected ErrFrozenPath for %q, got nil", target)
	}
}

// TestFrozenGuard_BlocksYAMLFrontmatter verifies that YAML frontmatter
// modification targets are rejected.
func TestFrozenGuard_BlocksYAMLFrontmatter(t *testing.T) {
	// Frontmatter modification is indicated by a :frontmatter suffix.
	target := ".claude/skills/moai-lang-go/SKILL.md:frontmatter"
	err := evolution.CheckFrozenGuard(target)
	if err == nil {
		t.Fatalf("expected ErrFrozenPath for frontmatter target %q, got nil", target)
	}
}

// TestRateLimiter_AllowsFirst3Proposals verifies that the first three
// proposals in a week are permitted.
func TestRateLimiter_AllowsFirst3Proposals(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	for i := 0; i < evolution.MaxProposalsPerWeek; i++ {
		if err := evolution.CheckRateLimit(projectRoot); err != nil {
			t.Fatalf("proposal %d: expected no rate limit error, got %v", i+1, err)
		}
		if err := evolution.UpdateRateLimit(projectRoot, ""); err != nil {
			t.Fatalf("proposal %d: failed to update rate limit: %v", i+1, err)
		}
	}
}

// TestRateLimiter_Blocks4thProposal verifies that the fourth proposal in a
// week is blocked.
func TestRateLimiter_Blocks4thProposal(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	// Record MaxProposalsPerWeek proposals.
	for i := 0; i < evolution.MaxProposalsPerWeek; i++ {
		if err := evolution.UpdateRateLimit(projectRoot, ""); err != nil {
			t.Fatalf("proposal %d: failed to update rate limit: %v", i+1, err)
		}
	}

	// 4th proposal must be blocked.
	err := evolution.CheckRateLimit(projectRoot)
	if err == nil {
		t.Fatal("expected ErrRateLimit for 4th proposal, got nil")
	}
	if err != evolution.ErrRateLimit {
		t.Fatalf("expected ErrRateLimit, got %v", err)
	}
}

// TestRateLimiter_ResetsAfterWeekBoundary verifies that the counter resets
// when the week changes.
func TestRateLimiter_ResetsAfterWeekBoundary(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	// Fill quota.
	for i := 0; i < evolution.MaxProposalsPerWeek; i++ {
		if err := evolution.UpdateRateLimit(projectRoot, ""); err != nil {
			t.Fatalf("setup: failed to update rate limit: %v", err)
		}
	}

	// Simulate a week reset by writing state with a past week start.
	writeRateStateWithWeekStart(t, projectRoot, lastMonday(time.Now()).AddDate(0, 0, -7))

	// New week — should allow proposals again.
	if err := evolution.CheckRateLimit(projectRoot); err != nil {
		t.Fatalf("expected no error after week reset, got %v", err)
	}
}

// TestRateLimiter_Enforces24hCooldown verifies that two proposals for the
// same file within 24 hours are blocked.
func TestRateLimiter_Enforces24hCooldown(t *testing.T) {
	projectRoot := t.TempDir()
	mustInitMoAI(t, projectRoot)

	targetFile := ".claude/skills/moai-lang-go/SKILL.md"

	// First proposal should be allowed.
	if err := evolution.CheckRateLimit(projectRoot); err != nil {
		t.Fatalf("first proposal: expected no rate limit error, got %v", err)
	}
	if err := evolution.UpdateRateLimit(projectRoot, targetFile); err != nil {
		t.Fatalf("first proposal: failed to update rate limit: %v", err)
	}

	// Check cooldown for same file — should be blocked.
	err := evolution.CheckFileCooldown(projectRoot, targetFile)
	if err == nil {
		t.Fatal("expected ErrRateLimit for file within cooldown, got nil")
	}
	if err != evolution.ErrRateLimit {
		t.Fatalf("expected ErrRateLimit, got %v", err)
	}
}

// mustInitMoAI creates the .moai directory structure needed for rate-limit state.
func mustInitMoAI(t *testing.T, projectRoot string) {
	t.Helper()
	dir := filepath.Join(projectRoot, ".moai", "evolution")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mustInitMoAI: %v", err)
	}
}

// writeRateStateWithWeekStart overwrites the rate-limit state file with the
// given week start date and a full quota (to trigger the week-reset logic).
func writeRateStateWithWeekStart(t *testing.T, projectRoot string, weekStart time.Time) {
	t.Helper()
	state := evolution.RateState{
		WeekStart:         weekStart.UTC().Format("2006-01-02"),
		ProposalsThisWeek: evolution.MaxProposalsPerWeek,
	}
	if err := evolution.WriteRateState(projectRoot, &state); err != nil {
		t.Fatalf("writeRateStateWithWeekStart: %v", err)
	}
}

// lastMonday returns the most recent Monday at midnight UTC on or before t.
func lastMonday(t time.Time) time.Time {
	wd := int(t.UTC().Weekday())
	if wd == 0 {
		wd = 7
	}
	return t.UTC().AddDate(0, 0, -(wd - 1)).Truncate(24 * time.Hour)
}
