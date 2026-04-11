package safety

import (
	"testing"
	"time"
)

// TestBaselineFields verifies that Baseline struct fields are set correctly.
func TestBaselineFields(t *testing.T) {
	now := time.Now()
	b := Baseline{
		Target:    "test-target",
		Score:     0.85,
		Timestamp: now,
	}

	if b.Target != "test-target" {
		t.Errorf("Target = %q, want %q", b.Target, "test-target")
	}
	if b.Score != 0.85 {
		t.Errorf("Score = %f, want %f", b.Score, 0.85)
	}
	if !b.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", b.Timestamp, now)
	}
}

// TestRateLimitConfigFields verifies that RateLimitConfig struct fields are set correctly.
func TestRateLimitConfigFields(t *testing.T) {
	cfg := RateLimitConfig{
		MaxExperimentsPerSession: 10,
		MaxAcceptedPerSession:    5,
		MaxAutoResearchPerWeek:   20,
	}

	if cfg.MaxExperimentsPerSession != 10 {
		t.Errorf("MaxExperimentsPerSession = %d, want %d", cfg.MaxExperimentsPerSession, 10)
	}
	if cfg.MaxAcceptedPerSession != 5 {
		t.Errorf("MaxAcceptedPerSession = %d, want %d", cfg.MaxAcceptedPerSession, 5)
	}
	if cfg.MaxAutoResearchPerWeek != 20 {
		t.Errorf("MaxAutoResearchPerWeek = %d, want %d", cfg.MaxAutoResearchPerWeek, 20)
	}
}

// TestActionRecordFields verifies that ActionRecord struct fields are set correctly.
func TestActionRecordFields(t *testing.T) {
	now := time.Now()
	r := ActionRecord{
		Type:      "experiment",
		Timestamp: now,
	}

	if r.Type != "experiment" {
		t.Errorf("Type = %q, want %q", r.Type, "experiment")
	}
	if !r.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", r.Timestamp, now)
	}
}
