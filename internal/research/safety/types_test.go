package safety

import (
	"testing"
	"time"
)

// TestBaselineFields는 Baseline 구조체 필드가 올바르게 설정되는지 검증한다.
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

// TestRateLimitConfigFields는 RateLimitConfig 구조체 필드가 올바르게 설정되는지 검증한다.
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

// TestActionRecordFields는 ActionRecord 구조체 필드가 올바르게 설정되는지 검증한다.
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
