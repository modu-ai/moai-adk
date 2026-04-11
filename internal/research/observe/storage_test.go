package observe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestStorage_Append_Single verifies that a single observation produces one line in the file.
func TestStorage_Append_Single(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	obs := &Observation{
		Type:      ObsCorrection,
		Agent:     "expert-backend",
		Target:    "error-handling",
		Detail:    "error wrapping missing",
		Timestamp: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}

	if err := s.Append(obs); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "observations.jsonl"))
	if err != nil {
		t.Fatalf("file read failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Errorf("line count = %d, want 1", len(lines))
	}
}

// TestStorage_Append_Multiple verifies that multiple observations produce N lines in the file.
func TestStorage_Append_Multiple(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	for i := 0; i < 5; i++ {
		obs := &Observation{
			Type:      ObsSuccess,
			Agent:     "expert-testing",
			Target:    "coverage",
			Detail:    "test passed",
			Timestamp: time.Date(2026, 4, 9, 12, i, 0, 0, time.UTC),
		}
		if err := s.Append(obs); err != nil {
			t.Fatalf("Append[%d] failed: %v", i, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, "observations.jsonl"))
	if err != nil {
		t.Fatalf("file read failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 5 {
		t.Errorf("line count = %d, want 5", len(lines))
	}
}

// TestStorage_LoadAll verifies that LoadAll returns all observations in order.
func TestStorage_LoadAll(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	agents := []string{"agent-a", "agent-b", "agent-c"}
	for i, agent := range agents {
		obs := &Observation{
			Type:      ObsCorrection,
			Agent:     agent,
			Target:    "target",
			Detail:    "detail info",
			Timestamp: time.Date(2026, 4, 9, 12, i, 0, 0, time.UTC),
		}
		if err := s.Append(obs); err != nil {
			t.Fatalf("Append failed: %v", err)
		}
	}

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("len = %d, want 3", len(all))
	}
	for i, agent := range agents {
		if all[i].Agent != agent {
			t.Errorf("all[%d].Agent = %q, want %q", i, all[i].Agent, agent)
		}
	}
}

// TestStorage_LoadSince verifies that LoadSince filters observations after the specified time.
func TestStorage_LoadSince(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	base := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		obs := &Observation{
			Type:      ObsSuccess,
			Agent:     "agent",
			Target:    "target",
			Detail:    "detail",
			Timestamp: base.Add(time.Duration(i) * time.Hour),
		}
		if err := s.Append(obs); err != nil {
			t.Fatalf("Append failed: %v", err)
		}
	}

	// since base+2h → indices 2,3,4 (3 items)
	since := base.Add(2 * time.Hour)
	filtered, err := s.LoadSince(since)
	if err != nil {
		t.Fatalf("LoadSince failed: %v", err)
	}

	if len(filtered) != 3 {
		t.Errorf("len = %d, want 3", len(filtered))
	}
}

// TestStorage_LoadAll_NonExistentFile verifies that LoadAll on a missing file returns empty slice without error.
func TestStorage_LoadAll_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("len = %d, want 0", len(all))
	}
}

// TestStorage_LoadAll_EmptyFile verifies that LoadAll on an empty file returns empty slice without error.
func TestStorage_LoadAll_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	// Create empty file
	if err := os.WriteFile(filepath.Join(dir, "observations.jsonl"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create empty file: %v", err)
	}

	s := NewStorage(dir)
	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("len = %d, want 0", len(all))
	}
}

// TestStorage_LoadAll_CorruptedLine verifies that corrupted lines are skipped and the rest are loaded normally.
func TestStorage_LoadAll_CorruptedLine(t *testing.T) {
	dir := t.TempDir()
	s := NewStorage(dir)

	// Add one valid observation
	obs := &Observation{
		Type:      ObsCorrection,
		Agent:     "agent",
		Target:    "target",
		Detail:    "detail",
		Timestamp: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
	if err := s.Append(obs); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Append a corrupted line
	filePath := filepath.Join(dir, "observations.jsonl")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	if _, err := f.WriteString("this is not valid JSON\n"); err != nil {
		_ = f.Close()
		t.Fatalf("write failed: %v", err)
	}
	_ = f.Close()

	// Add another valid observation
	obs2 := &Observation{
		Type:      ObsSuccess,
		Agent:     "agent-2",
		Target:    "target-2",
		Detail:    "detail 2",
		Timestamp: time.Date(2026, 4, 9, 13, 0, 0, 0, time.UTC),
	}
	if err := s.Append(obs2); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	all, err := s.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll returned error: %v", err)
	}

	// Corrupted line should be skipped; expect 2 valid observations
	if len(all) != 2 {
		t.Errorf("len = %d, want 2", len(all))
	}
}
