package safety

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRateLimiter_CheckSessionLimit verifies the per-session experiment count limit.
func TestRateLimiter_CheckSessionLimit(t *testing.T) {
	tests := []struct {
		name           string
		config         RateLimitConfig
		sessionActions int
		wantErr        bool
	}{
		{
			name:           "below session limit → nil",
			config:         RateLimitConfig{MaxExperimentsPerSession: 10},
			sessionActions: 5,
			wantErr:        false,
		},
		{
			name:           "exactly at session limit → error",
			config:         RateLimitConfig{MaxExperimentsPerSession: 10},
			sessionActions: 10,
			wantErr:        true,
		},
		{
			name:           "above session limit → error",
			config:         RateLimitConfig{MaxExperimentsPerSession: 5},
			sessionActions: 8,
			wantErr:        true,
		},
		{
			name:           "zero session actions → nil",
			config:         RateLimitConfig{MaxExperimentsPerSession: 10},
			sessionActions: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			l := NewRateLimiter(filepath.Join(tmpDir, "actions.jsonl"))
			err := l.CheckSessionLimit(tt.config, tt.sessionActions)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSessionLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrRateLimitExceeded) {
				t.Errorf("CheckSessionLimit() error = %v, want ErrRateLimitExceeded", err)
			}
		})
	}
}

// TestRateLimiter_CheckWeeklyLimit verifies the weekly auto-research limit.
func TestRateLimiter_CheckWeeklyLimit(t *testing.T) {
	t.Run("no records → nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		l := NewRateLimiter(filepath.Join(tmpDir, "actions.jsonl"))
		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}

		err := l.CheckWeeklyLimit(cfg)
		if err != nil {
			t.Errorf("CheckWeeklyLimit() error = %v, want nil", err)
		}
	})

	t.Run("below weekly limit → nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		// Write 2 recent records
		writeActionRecords(t, storePath, []ActionRecord{
			{Type: "auto_research", Timestamp: time.Now().Add(-1 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-2 * time.Hour)},
		})

		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}
		err := l.CheckWeeklyLimit(cfg)
		if err != nil {
			t.Errorf("CheckWeeklyLimit() error = %v, want nil", err)
		}
	})

	t.Run("above weekly limit → ErrRateLimitExceeded", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		// Write 5 recent records (limit: 5)
		records := make([]ActionRecord, 5)
		for i := range records {
			records[i] = ActionRecord{
				Type:      "auto_research",
				Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			}
		}
		writeActionRecords(t, storePath, records)

		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}
		err := l.CheckWeeklyLimit(cfg)
		if err == nil {
			t.Error("CheckWeeklyLimit() error = nil, want ErrRateLimitExceeded")
		}
		if !errors.Is(err, ErrRateLimitExceeded) {
			t.Errorf("CheckWeeklyLimit() error = %v, want ErrRateLimitExceeded", err)
		}
	})

	t.Run("records older than 7 days are excluded from weekly limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		// 3 old records + 2 recent records
		records := []ActionRecord{
			{Type: "auto_research", Timestamp: time.Now().Add(-10 * 24 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-9 * 24 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-8 * 24 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-1 * time.Hour)},
			{Type: "auto_research", Timestamp: time.Now().Add(-2 * time.Hour)},
		}
		writeActionRecords(t, storePath, records)

		cfg := RateLimitConfig{MaxAutoResearchPerWeek: 5}
		err := l.CheckWeeklyLimit(cfg)
		if err != nil {
			t.Errorf("CheckWeeklyLimit() error = %v, want nil (old records should be excluded)", err)
		}
	})
}

// TestRateLimiter_RecordAction verifies that action records are persisted to the JSONL file.
func TestRateLimiter_RecordAction(t *testing.T) {
	t.Run("action record is saved to JSONL file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storePath := filepath.Join(tmpDir, "actions.jsonl")
		l := NewRateLimiter(storePath)

		err := l.RecordAction("experiment")
		if err != nil {
			t.Fatalf("RecordAction() error = %v", err)
		}

		err = l.RecordAction("auto_research")
		if err != nil {
			t.Fatalf("RecordAction() error = %v", err)
		}

		// Read and verify JSONL file
		records := readActionRecords(t, storePath)
		if len(records) != 2 {
			t.Fatalf("record count = %d, want 2", len(records))
		}
		if records[0].Type != "experiment" {
			t.Errorf("records[0].Type = %q, want %q", records[0].Type, "experiment")
		}
		if records[1].Type != "auto_research" {
			t.Errorf("records[1].Type = %q, want %q", records[1].Type, "auto_research")
		}
	})

	t.Run("recording to nonexistent directory returns error", func(t *testing.T) {
		l := NewRateLimiter("/nonexistent/dir/actions.jsonl")
		err := l.RecordAction("experiment")
		if err == nil {
			t.Error("RecordAction() error = nil, want error for nonexistent directory")
		}
	})
}

// writeActionRecords writes action records to a JSONL file for testing.
func writeActionRecords(t *testing.T, path string, records []ActionRecord) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	for _, r := range records {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("failed to encode record: %v", err)
		}
	}
}

// readActionRecords reads action records from a JSONL file for testing.
func readActionRecords(t *testing.T, path string) []ActionRecord {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer func() { _ = f.Close() }()

	var records []ActionRecord
	dec := json.NewDecoder(f)
	for dec.More() {
		var r ActionRecord
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("failed to decode record: %v", err)
		}
		records = append(records, r)
	}
	return records
}
