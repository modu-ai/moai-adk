// Package safety — Canary Veto tests (M3 RED).
package safety_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	harness "github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
)

// TestCanaryVeto_ProvisionalApply verifies that a Canary veto after provisional apply
// triggers auto-rollback and sets evolution_status=vetoed_by_canary (plan.md §3.3 E5).
// M3 RED: fails until canary_veto.go exists.
func TestCanaryVeto_ProvisionalApply(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	revertDir := filepath.Join(dir, "revert")

	// Create a target file to be "provisionally applied"
	targetPath := filepath.Join(dir, "target_skill.md")
	originalContent := "description: original\n"
	if err := os.WriteFile(targetPath, []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	veto := safety.NewCanaryVeto(safety.CanaryVetoConfig{
		RevertDir:          revertDir,
		RateLimitStatePath: filepath.Join(dir, "rate-limit.json"),
	})

	proposal := harness.Proposal{
		ID:         "evo-001",
		TargetPath: targetPath,
		FieldKey:   "description",
		NewValue:   "description: new value\n",
		CreatedAt:  time.Now().UTC(),
	}

	// Provisional apply: snapshot + write new content
	if err := veto.ProvisionalApply(proposal); err != nil {
		t.Fatalf("ProvisionalApply: %v", err)
	}

	// File should now contain new value
	data, _ := os.ReadFile(targetPath)
	if string(data) != "description: new value\n" {
		t.Errorf("after provisional apply, content = %q, want %q", data, "description: new value\n")
	}

	// Canary veto: auto-rollback + 48h cooldown
	if err := veto.VetoAndRollback(proposal); err != nil {
		t.Fatalf("VetoAndRollback: %v", err)
	}

	// File should be restored to original content
	data, _ = os.ReadFile(targetPath)
	if string(data) != originalContent {
		t.Errorf("after rollback, content = %q, want %q", data, originalContent)
	}

	// evolution_status in veto log should be "vetoed_by_canary"
	vetoLogPath := filepath.Join(dir, "revert", "evo-001", "veto.log")
	logData, err := os.ReadFile(vetoLogPath)
	if err != nil {
		t.Fatalf("veto.log not created: %v", err)
	}
	if !contains(string(logData), "vetoed_by_canary") {
		t.Errorf("veto.log missing 'vetoed_by_canary': %s", logData)
	}
}

// TestCanaryVeto_Cooldown verifies that re-applying within 48h is rejected with
// HARNESS_LEARNING_RATELIMIT_EXCEEDED (AC-HRA-008b).
func TestCanaryVeto_Cooldown(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	statePath := filepath.Join(dir, "rate-limit.json")
	veto := safety.NewCanaryVeto(safety.CanaryVetoConfig{
		RevertDir:          filepath.Join(dir, "revert"),
		RateLimitStatePath: statePath,
	})

	// Record a veto cooldown entry
	proposal := harness.Proposal{ID: "evo-cooldown", TargetPath: "some/path"}
	if err := veto.RecordCooldown(proposal); err != nil {
		t.Fatalf("RecordCooldown: %v", err)
	}

	// Attempt to apply within cooldown
	err := veto.CheckCooldown(proposal)
	if err == nil {
		t.Fatal("expected cooldown error, got nil")
	}
	if !contains(err.Error(), "HARNESS_LEARNING_RATELIMIT_EXCEEDED") {
		t.Errorf("expected HARNESS_LEARNING_RATELIMIT_EXCEEDED in error, got: %v", err)
	}
}

// TestCanaryVeto_SnapshotBeforeApply verifies that the revert directory snapshot
// is created before provisional apply (copy-before-write pattern, plan.md §3.3).
func TestCanaryVeto_SnapshotBeforeApply(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "target.md")
	originalContent := "original content\n"
	if err := os.WriteFile(targetPath, []byte(originalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	veto := safety.NewCanaryVeto(safety.CanaryVetoConfig{
		RevertDir:          filepath.Join(dir, "revert"),
		RateLimitStatePath: filepath.Join(dir, "rate-limit.json"),
	})

	proposal := harness.Proposal{
		ID:         "evo-snapshot",
		TargetPath: targetPath,
		NewValue:   "new content\n",
	}

	if err := veto.ProvisionalApply(proposal); err != nil {
		t.Fatalf("ProvisionalApply: %v", err)
	}

	// Snapshot must exist at revert/evo-snapshot/<filename>
	snapDir := filepath.Join(dir, "revert", "evo-snapshot")
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		t.Fatalf("snapshot dir not created: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("snapshot dir is empty")
	}
}

// BenchmarkL1FrozenGuard verifies that IsFrozen is <10ms p99 (Q1 resolution, plan.md §11.4).
func BenchmarkL1FrozenGuard(b *testing.B) {
	proposal := harness.Proposal{
		ID:         "bench-frozen",
		TargetPath: ".claude/skills/moai-harness/SKILL.md",
		FieldKey:   "description",
		NewValue:   "new val",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = safety.IsFrozen(proposal.TargetPath)
	}
}

// BenchmarkL4RateLimit verifies <100ms p99 (plan.md §11.4).
func BenchmarkL4RateLimit(b *testing.B) {
	dir := b.TempDir()
	rl := safety.NewRateLimiter(filepath.Join(dir, "rate-limit.json"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = rl.CheckLimit()
	}
}

// contains checks whether s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
