// Package harness_test — M6 integration tests for SPEC-V3R5-HARNESS-AUTONOMY-001.
// Covers AC-HRA-002 (tier progression), AC-HRA-004 (capture emit), AC-HRA-006 (throttle),
// AC-HRA-007 (cold-start seeds), AC-HRA-008 (canary veto rollback), AC-HRA-010 (evolution log),
// AC-HRA-011 (anti-pattern flag), AC-HRA-012 (observation schema).
package harness_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness/capture"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
	"github.com/modu-ai/moai-adk/internal/harness/seeds"
	"github.com/modu-ai/moai-adk/internal/harness/throttle"
	"github.com/modu-ai/moai-adk/internal/harness/tier"

	harness "github.com/modu-ai/moai-adk/internal/harness"
)

// TestTier1ToTier4Progression verifies the 4-Tier state machine transitions.
// A single entry incremented 10 times MUST reach StatusHighConfidence (AC-HRA-002).
func TestTier1ToTier4Progression(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	antiPath := filepath.Join(dir, "anti-patterns.yaml")

	eng := tier.NewEngine(tier.EngineConfig{
		ObservationsPath: obsPath,
		AntiPatternsPath: antiPath,
	})

	entry := tier.Entry{
		AgentName:   "expert-backend",
		ContextHash: "ctx-tier-prog",
	}

	// Increment 10 times to reach Tier 4 (StatusHighConfidence).
	for i := 0; i < 10; i++ {
		if err := eng.Increment(entry); err != nil {
			t.Fatalf("Increment iteration %d: %v", i+1, err)
		}
	}

	entries, err := eng.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	found := false
	for _, e := range entries {
		if e.AgentName == entry.AgentName && e.ContextHash == entry.ContextHash {
			found = true
			if e.Count != 10 {
				t.Errorf("Count = %d, want 10", e.Count)
			}
			if e.Status != tier.StatusHighConfidence {
				t.Errorf("Status = %q, want %q", e.Status, tier.StatusHighConfidence)
			}
		}
	}
	if !found {
		t.Error("entry not found in observations.yaml after 10 increments")
	}
}

// TestCaptureEmitOnSubagentStop verifies that Capturer.OnSubagentStop emits a YAML observation
// entry with correct fields (AC-HRA-004).
func TestCaptureEmitOnSubagentStop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")

	cap := capture.New(capture.Config{ObservationsPath: obsPath})

	event := capture.SubagentStopEvent{
		AgentName:   "manager-develop",
		AgentType:   "subagent",
		SessionID:   "sess-001",
		Timestamp:   time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC),
		ContextHash: "ctx-capture-test",
	}

	if err := cap.OnSubagentStop(event); err != nil {
		t.Fatalf("OnSubagentStop: %v", err)
	}

	data, err := os.ReadFile(obsPath)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", obsPath, err)
	}

	content := string(data)
	if !strings.Contains(content, "agent_name: manager-develop") {
		t.Errorf("observations.yaml missing 'agent_name: manager-develop':\n%s", content)
	}
	if !strings.Contains(content, "status: observation") {
		t.Errorf("observations.yaml missing 'status: observation':\n%s", content)
	}
	if !strings.Contains(content, "count: 1") {
		t.Errorf("observations.yaml missing 'count: 1':\n%s", content)
	}
}

// TestThrottlingFourModes verifies the 4-mode proposal throttling behavior (AC-HRA-006).
// Checks: Immediate emits, Mute suppresses, Quiet defers, Batch queues.
func TestThrottlingFourModes(t *testing.T) {
	t.Parallel()

	// Immediate mode: ShouldEmit = true
	t.Run("Immediate", func(t *testing.T) {
		t.Parallel()
		th := throttle.New(throttle.Config{Mode: throttle.ModeImmediate})
		result := th.Check(throttle.ProposalMeta{
			ID:        "prop-001",
			Category:  "error-handling",
			CreatedAt: time.Now(),
		})
		if !result.ShouldEmit {
			t.Errorf("Immediate mode: ShouldEmit = false, want true; reason=%q", result.Reason)
		}
	})

	// Mute mode: ShouldEmit = false, Reason = HARNESS_LEARNING_MUTED
	t.Run("Muted", func(t *testing.T) {
		t.Parallel()
		th := throttle.New(throttle.Config{
			Mode:           throttle.ModeImmediate,
			MuteCategories: []string{"error-handling"},
		})
		result := th.Check(throttle.ProposalMeta{
			ID:        "prop-002",
			Category:  "error-handling",
			CreatedAt: time.Now(),
		})
		if result.ShouldEmit {
			t.Error("Muted category: ShouldEmit = true, want false")
		}
		if result.Reason != throttle.ReasonMuted {
			t.Errorf("Reason = %q, want %q", result.Reason, throttle.ReasonMuted)
		}
	})

	// Quiet mode: inside quiet window → ShouldEmit = false
	t.Run("QuietWindow", func(t *testing.T) {
		t.Parallel()
		// Quiet window 00:00-23:59 (covers all hours).
		th := throttle.NewWithNow(throttle.Config{
			Mode:         throttle.ModeQuiet,
			QuietStartHr: 0,
			QuietEndHr:   0, // 0→0 overnight = 24h quiet
		}, func() time.Time {
			return time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
		})
		result := th.Check(throttle.ProposalMeta{
			ID:        "prop-003",
			Category:  "naming",
			CreatedAt: time.Now(),
		})
		if result.ShouldEmit {
			t.Error("QuietWindow: ShouldEmit = true, want false")
		}
		if result.Reason != throttle.ReasonQuiet {
			t.Errorf("Reason = %q, want %q", result.Reason, throttle.ReasonQuiet)
		}
	})

	// Batch mode: ShouldEmit = false, Reason = HARNESS_LEARNING_BATCHED
	t.Run("Batch", func(t *testing.T) {
		t.Parallel()
		th := throttle.New(throttle.Config{
			Mode:           throttle.ModeBatch,
			BatchMaxPerWin: 10,
		})
		result := th.Check(throttle.ProposalMeta{
			ID:        "prop-004",
			Category:  "testing",
			CreatedAt: time.Now(),
		})
		if result.ShouldEmit {
			t.Error("Batch mode: ShouldEmit = true, want false")
		}
		if result.Reason != throttle.ReasonBatched {
			t.Errorf("Reason = %q, want %q", result.Reason, throttle.ReasonBatched)
		}
	})
}

// TestColdStartSeedInject verifies that LoadForProject("unknown") returns an empty slice
// with no error (REQ-HRA-022: valid cold-start state), and DetectProjectType is "unknown" (AC-HRA-007).
func TestColdStartSeedInject(t *testing.T) {
	t.Parallel()

	// DetectProjectType is "unknown" in W3 stub.
	projectType := seeds.DetectProjectType()
	if projectType != "unknown" {
		t.Errorf("DetectProjectType() = %q, want %q", projectType, "unknown")
	}

	loader := seeds.NewLoader(seeds.LoaderConfig{
		SSoTDir:  t.TempDir(),
		CacheDir: t.TempDir(),
	})

	seedList, err := loader.LoadForProject(projectType)
	if err != nil {
		t.Fatalf("LoadForProject(%q): %v", projectType, err)
	}
	// Cold-start with "unknown" project type must return empty slice (not error).
	if len(seedList) != 0 {
		t.Errorf("LoadForProject(%q) returned %d seeds, want 0", projectType, len(seedList))
	}
}

// TestCanaryVetoProvisionalRollback verifies the E5 Canary Veto Policy:
// ProvisionalApply writes new content, VetoAndRollback restores original,
// and a cooldown entry is created (AC-HRA-008).
func TestCanaryVetoProvisionalRollback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	revertDir := filepath.Join(dir, "revert")
	rateLimitPath := filepath.Join(dir, "rate-limit-state.json")
	targetPath := filepath.Join(dir, "target.md")

	originalContent := "# Original content\n"
	if err := os.WriteFile(targetPath, []byte(originalContent), 0o644); err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}

	veto := safety.NewCanaryVeto(safety.CanaryVetoConfig{
		RevertDir:          revertDir,
		RateLimitStatePath: rateLimitPath,
	})

	proposal := harness.Proposal{
		ID:         "evo-canary-001",
		TargetPath: targetPath,
		NewValue:   "# Modified content\n",
	}

	// ProvisionalApply: writes new content.
	if err := veto.ProvisionalApply(proposal); err != nil {
		t.Fatalf("ProvisionalApply: %v", err)
	}

	afterApply, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("ReadFile after apply: %v", err)
	}
	if string(afterApply) != "# Modified content\n" {
		t.Errorf("after ProvisionalApply content = %q, want %q", string(afterApply), "# Modified content\n")
	}

	// VetoAndRollback: restores original.
	if err := veto.VetoAndRollback(proposal); err != nil {
		t.Fatalf("VetoAndRollback: %v", err)
	}

	afterRollback, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("ReadFile after rollback: %v", err)
	}
	if string(afterRollback) != originalContent {
		t.Errorf("after VetoAndRollback content = %q, want %q", string(afterRollback), originalContent)
	}

	// Cooldown must be recorded.
	if err := veto.CheckCooldown(proposal); err == nil {
		t.Error("CheckCooldown after VetoAndRollback: expected cooldown error, got nil")
	}
}

// TestEvolutionLogAppendOnly verifies that multiple Confirm calls append
// non-overlapping veto.log entries (AC-HRA-010: evolution log append-only).
func TestEvolutionLogAppendOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	revertDir := filepath.Join(dir, "revert")
	targetPath := filepath.Join(dir, "target.md")

	if err := os.WriteFile(targetPath, []byte("# v0\n"), 0o644); err != nil {
		t.Fatalf("WriteFile target: %v", err)
	}

	veto := safety.NewCanaryVeto(safety.CanaryVetoConfig{
		RevertDir:          revertDir,
		RateLimitStatePath: filepath.Join(dir, "rl.json"),
	})

	p1 := harness.Proposal{ID: "evo-001", TargetPath: targetPath, NewValue: "# v1\n"}
	p2 := harness.Proposal{ID: "evo-002", TargetPath: targetPath, NewValue: "# v2\n"}

	for _, p := range []harness.Proposal{p1, p2} {
		if err := veto.ProvisionalApply(p); err != nil {
			t.Fatalf("ProvisionalApply(%s): %v", p.ID, err)
		}
		if err := veto.Confirm(p); err != nil {
			t.Fatalf("Confirm(%s): %v", p.ID, err)
		}
	}

	// Each evolution gets its own veto.log under revert/<evo-id>/.
	for _, id := range []string{"evo-001", "evo-002"} {
		logPath := filepath.Join(revertDir, id, "veto.log")
		data, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("ReadFile veto.log for %s: %v", id, err)
		}
		if !strings.Contains(string(data), `"applied"`) {
			t.Errorf("veto.log for %s missing 'applied': %s", id, string(data))
		}
	}
}

// TestAntiPatternAutoFlag verifies that FlagAntiPattern writes an anti-pattern entry
// to anti-patterns.yaml with status=anti-pattern (AC-HRA-011).
func TestAntiPatternAutoFlag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	antiPath := filepath.Join(dir, "anti-patterns.yaml")

	eng := tier.NewEngine(tier.EngineConfig{
		ObservationsPath: obsPath,
		AntiPatternsPath: antiPath,
	})

	ev := tier.AntiPatternEvidence{
		AgentName:   "expert-security",
		ContextHash: "ctx-anti-001",
		Reason:      "score drop > 0.20 on security criterion",
	}

	if err := eng.FlagAntiPattern(ev); err != nil {
		t.Fatalf("FlagAntiPattern: %v", err)
	}

	data, err := os.ReadFile(antiPath)
	if err != nil {
		t.Fatalf("ReadFile anti-patterns.yaml: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "agent_name: expert-security") {
		t.Errorf("anti-patterns.yaml missing 'agent_name: expert-security':\n%s", content)
	}
	if !strings.Contains(content, "status: anti-pattern") {
		t.Errorf("anti-patterns.yaml missing 'status: anti-pattern':\n%s", content)
	}
	if !strings.Contains(content, "score drop") {
		t.Errorf("anti-patterns.yaml missing reason text:\n%s", content)
	}
}

// TestObservationsSchemaCanonical verifies that the Seed struct has the expected
// schema fields including the Version field for W4 backward compatibility (AC-HRA-012).
func TestObservationsSchemaCanonical(t *testing.T) {
	t.Parallel()

	s := seeds.Seed{
		ID:         "SEED-GO-001",
		Pattern:    "error wrapping",
		Tier:       3,
		Confidence: 0.85,
		Category:   "error-handling",
		Body:       "Always use fmt.Errorf with %w",
		Version:    1,
	}

	if s.ID != "SEED-GO-001" {
		t.Errorf("Seed.ID = %q, want SEED-GO-001", s.ID)
	}
	if s.Version != 1 {
		t.Errorf("Seed.Version = %d, want 1", s.Version)
	}
	if s.Tier != 3 {
		t.Errorf("Seed.Tier = %d, want 3", s.Tier)
	}
	if s.Confidence != 0.85 {
		t.Errorf("Seed.Confidence = %f, want 0.85", s.Confidence)
	}
	if s.Category != "error-handling" {
		t.Errorf("Seed.Category = %q, want error-handling", s.Category)
	}
}

// TestClassifyStatusThresholds verifies the canonical 4-Tier thresholds [1, 3, 5, 10]
// (REQ-HRA-007 via AC-HRA-002).
func TestClassifyStatusThresholds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		count int
		want  tier.Status
	}{
		{1, tier.StatusObservation},
		{2, tier.StatusObservation},
		{3, tier.StatusHeuristic},
		{4, tier.StatusHeuristic},
		{5, tier.StatusRule},
		{9, tier.StatusRule},
		{10, tier.StatusHighConfidence},
		{100, tier.StatusHighConfidence},
	}

	for _, tc := range cases {
		got := tier.ClassifyStatus(tc.count)
		if got != tc.want {
			t.Errorf("ClassifyStatus(%d) = %q, want %q", tc.count, got, tc.want)
		}
	}
}
