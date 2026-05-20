// Package tier — M2 tier engine tests.
package tier_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/tier"
)

// TestTierTransitions is a table-driven test covering 12 state machine transitions
// (REQ-HRA-004).
// M2 RED: fails until tier.go is implemented.
func TestTierTransitions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		currentCount int
		wantStatus   tier.Status
	}{
		{"count=0 → observation", 0, tier.StatusObservation},
		{"count=1 → observation", 1, tier.StatusObservation},
		{"count=2 → observation", 2, tier.StatusObservation},
		{"count=3 → heuristic", 3, tier.StatusHeuristic},
		{"count=4 → heuristic", 4, tier.StatusHeuristic},
		{"count=5 → rule", 5, tier.StatusRule},
		{"count=6 → rule", 6, tier.StatusRule},
		{"count=9 → rule", 9, tier.StatusRule},
		{"count=10 → high-confidence", 10, tier.StatusHighConfidence},
		{"count=11 → high-confidence", 11, tier.StatusHighConfidence},
		{"count=50 → high-confidence", 50, tier.StatusHighConfidence},
		// seed inject starting point (synthetic count=5 → rule, plan.md §2.1)
		{"seed inject count=5 → rule", 5, tier.StatusRule},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tier.ClassifyStatus(tt.currentCount)
			if got != tt.wantStatus {
				t.Errorf("ClassifyStatus(%d) = %q, want %q", tt.currentCount, got, tt.wantStatus)
			}
		})
	}
}

// TestEngine_IncrementAndSave verifies that Engine.Increment persists an entry to
// observations.yaml with correct count and status.
func TestEngine_IncrementAndSave(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")

	eng := tier.NewEngine(tier.EngineConfig{ObservationsPath: obsPath})

	entry := tier.Entry{
		AgentName:   "manager-develop",
		ContextHash: "hash001",
	}

	if err := eng.Increment(entry); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	// Load back and verify
	entries, err := eng.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	got := entries[0]
	if got.Count != 1 {
		t.Errorf("count = %d, want 1", got.Count)
	}
	if got.Status != tier.StatusObservation {
		t.Errorf("status = %q, want %q", got.Status, tier.StatusObservation)
	}
}

// TestEngine_Increment_ExistingEntry verifies that repeated increments on the same
// agent/hash accumulate correctly and trigger status transitions.
func TestEngine_Increment_ExistingEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	eng := tier.NewEngine(tier.EngineConfig{ObservationsPath: obsPath})

	entry := tier.Entry{AgentName: "expert-backend", ContextHash: "repeating"}

	// Increment 10 times → should reach high-confidence
	for i := 0; i < 10; i++ {
		if err := eng.Increment(entry); err != nil {
			t.Fatalf("Increment %d: %v", i, err)
		}
	}

	entries, err := eng.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (deduplicated), got %d", len(entries))
	}
	if entries[0].Count != 10 {
		t.Errorf("count = %d, want 10", entries[0].Count)
	}
	if entries[0].Status != tier.StatusHighConfidence {
		t.Errorf("status = %q, want %q", entries[0].Status, tier.StatusHighConfidence)
	}
}

// TestEngine_AntiPatternFlag verifies that FlagAntiPattern writes to anti-patterns.yaml
// with FROZEN status (REQ-HRA-006).
func TestEngine_AntiPatternFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	eng := tier.NewEngine(tier.EngineConfig{
		ObservationsPath: filepath.Join(dir, "observations.yaml"),
		AntiPatternsPath: filepath.Join(dir, "anti-patterns.yaml"),
	})

	if err := eng.FlagAntiPattern(tier.AntiPatternEvidence{
		AgentName:   "manager-develop",
		ContextHash: "bad-pattern",
		Reason:      "SPEC quality score drop > 0.20",
	}); err != nil {
		t.Fatalf("FlagAntiPattern: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "anti-patterns.yaml"))
	if err != nil {
		t.Fatalf("anti-patterns.yaml not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("anti-patterns.yaml is empty")
	}
}

// TestObservationsSchemaCanonical verifies that the observations.yaml schema uses
// canonical field names (spec.md §1.7 + AC-HRA-012).
func TestObservationsSchemaCanonical(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	eng := tier.NewEngine(tier.EngineConfig{ObservationsPath: obsPath})

	if err := eng.Increment(tier.Entry{AgentName: "expert-backend", ContextHash: "schema-check"}); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	data, _ := os.ReadFile(obsPath)
	content := string(data)

	// Canonical: "timestamp" not "time" / "ts"
	if contains(content, "  time:") || contains(content, "  ts:") {
		t.Errorf("non-canonical time field found; must use 'timestamp':\n%s", content)
	}
	if !contains(content, "timestamp:") {
		t.Errorf("canonical 'timestamp:' field missing:\n%s", content)
	}

	// Canonical: "context_hash" not "contextHash"
	if contains(content, "contextHash:") {
		t.Errorf("camelCase 'contextHash' found; must use 'context_hash':\n%s", content)
	}
}

// TestEngine_ActiveCountCap verifies that when active entries exceed 50,
// oldest observation-status entries are archived (REQ-HRA-017).
func TestEngine_ActiveCountCap(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	obsPath := filepath.Join(dir, "observations.yaml")
	eng := tier.NewEngine(tier.EngineConfig{
		ObservationsPath: obsPath,
		MaxActive:        5, // reduced cap for test speed
	})

	// Insert 7 unique entries → 5 kept, 2 archived
	for i := 0; i < 7; i++ {
		if err := eng.Increment(tier.Entry{
			AgentName:   "agent",
			ContextHash: "hash" + string(rune('a'+i)),
		}); err != nil {
			t.Fatalf("Increment %d: %v", i, err)
		}
	}

	entries, err := eng.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	active := 0
	for _, e := range entries {
		if e.Status != tier.StatusArchived {
			active++
		}
	}
	if active > 5 {
		t.Errorf("active count %d exceeds cap 5", active)
	}
}

// contains checks whether substr appears in s.
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
